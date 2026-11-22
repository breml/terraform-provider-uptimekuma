package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/ory/dockertest/v3"

	kuma "github.com/breml/go-uptime-kuma-client"

	"github.com/breml/terraform-provider-uptimekuma/internal/client"
)

// apiKeyScript runs with node inside the Uptime Kuma 3.x container, on the
// socket.io-client the image bundles. It connects with the session cookie of
// outOfBandClient, which spends no sign-in of the server's rate limit, and:
//
//   - sets "API and MCP" (apiEnabled) to KUMA_API_ENABLED,
//   - given KUMA_API_KEY_NAME, creates an API key (addAPIKey) and prints it as
//     {"key": "...", "key_id": "..."},
//   - given KUMA_DELETE_KEY_ID, revokes that key (deleteAPIKey).
//
// The go-uptime-kuma-client keeps its addAPIKey and setSettings-with-apiEnabled
// helpers test-only, and has no generic emit, so this is the least invasive way
// to reach those socket events from here. The event names are the ones of
// server/socket-handlers/api-key-socket-handler.js.
const apiKeyScript = `
const fail = (msg) => { console.error(msg); process.exit(1); };
let io;
try {
	({ io } = require("/app/node_modules/socket.io-client"));
} catch (err) {
	fail("this image has no /app/node_modules/socket.io-client, which the test drives the server with: " + err.message);
}
const timer = setTimeout(() => fail("timeout"), 20000);
const socket = io("http://127.0.0.1:3001", {
	transports: ["websocket"],
	reconnection: false,
	extraHeaders: { cookie: "better-auth.session_token=" + process.env.KUMA_SESSION },
});
const emit = (...args) => new Promise((resolve) => socket.emit(...args, resolve));
socket.on("connect_error", (err) => fail("connect_error: " + err.message));
socket.on("loginRequired", () => fail("loginRequired: the session cookie was not accepted"));
socket.on("session", async () => {
	try {
		const settings = await emit("getSettings");
		if (!settings.ok) fail("getSettings: " + settings.msg);
		const enabled = process.env.KUMA_API_ENABLED === "true";
		const saved = await emit("setSettings", { ...settings.data, apiEnabled: enabled }, process.env.KUMA_PASSWORD);
		if (!saved.ok) fail("setSettings: " + saved.msg);
		let result = { key: "", key_id: "" };
		if (process.env.KUMA_API_KEY_NAME) {
			const added = await emit("addAPIKey", { name: process.env.KUMA_API_KEY_NAME });
			if (!added.ok) fail("addAPIKey: " + added.msg);
			result = { key: added.key, key_id: String(added.keyID) };
		}
		if (process.env.KUMA_DELETE_KEY_ID) {
			const deleted = await emit("deleteAPIKey", process.env.KUMA_DELETE_KEY_ID);
			if (!deleted.ok) fail("deleteAPIKey: " + deleted.msg);
		}
		process.stdout.write(JSON.stringify(result));
		clearTimeout(timer);
		socket.close();
		process.exit(0);
	} catch (err) {
		fail(String(err));
	}
});
`

// apiKeyScriptRun is one run of apiKeyScript, see there.
type apiKeyScriptRun struct {
	enabled     bool
	createName  string
	deleteKeyID string
}

// apiKeyScriptResult is what apiKeyScript prints.
type apiKeyScriptResult struct {
	Key   string `json:"key"`
	KeyID string `json:"key_id"`
}

// runAPIKeyScript runs apiKeyScript in the test container.
func runAPIKeyScript(run apiKeyScriptRun) (apiKeyScriptResult, error) {
	var stdout, stderr bytes.Buffer

	exitCode, err := kumaContainer.Exec([]string{"node", "-e", apiKeyScript}, dockertest.ExecOptions{
		Env: []string{
			"KUMA_SESSION=" + outOfBandClient.SessionToken(),
			"KUMA_PASSWORD=" + password,
			fmt.Sprintf("KUMA_API_ENABLED=%t", run.enabled),
			"KUMA_API_KEY_NAME=" + run.createName,
			"KUMA_DELETE_KEY_ID=" + run.deleteKeyID,
		},
		StdOut: &stdout,
		StdErr: &stderr,
	})
	if err != nil {
		return apiKeyScriptResult{}, fmt.Errorf("run api key script: %w", err)
	}

	if exitCode != 0 {
		return apiKeyScriptResult{}, fmt.Errorf("api key script exited with %d: %s", exitCode, stderr.String())
	}

	var result apiKeyScriptResult

	err = json.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		return apiKeyScriptResult{}, fmt.Errorf("api key script printed %q: %w", stdout.String(), err)
	}

	return result, nil
}

// testAccEnableAPIKey turns "API and MCP" on, creates an API key and returns
// it. The cleanup that turns "API and MCP" off again, a setting of the whole
// server, and revokes the key is registered before either happens, so a test
// that fails halfway leaves neither behind.
func testAccEnableAPIKey(t *testing.T) string {
	t.Helper()

	var keyID string

	t.Cleanup(func() {
		_, err := runAPIKeyScript(apiKeyScriptRun{enabled: false, deleteKeyID: keyID})
		if err != nil {
			t.Errorf("turn API and MCP off and revoke the key: %v", err)
		}
	})

	result, err := runAPIKeyScript(apiKeyScriptRun{enabled: true, createName: acctest.RandomWithPrefix("tf-acc")})
	if err != nil {
		t.Fatal(err)
	}

	if result.Key == "" || result.KeyID == "" {
		t.Fatalf("expected the api key script to return a key and its ID, got key %t, ID %q",
			result.Key != "", result.KeyID)
	}

	keyID = result.KeyID

	return result.Key
}

// testAccIsolatePool closes the pooled connection the provider shares across
// tests, and closes it again once the test ends.
//
// The pool holds one connection, and refuses a configuration that differs
// from the one it was opened with: an api_key provider block must not be
// served the connection the username and password opened, nor the other way
// round. A test with another provider configuration therefore needs the pool
// to itself. It is only safe from a test that is not parallel: Go runs every
// test that is not parallel before it releases a parallel one, so no other
// test uses the pool meanwhile. The next test opens a connection of its own.
func testAccIsolatePool(t *testing.T) {
	t.Helper()

	closePool := func() {
		err := client.GetGlobalPool().Close()
		if err != nil {
			t.Logf("close pooled connection: %v", err)
		}
	}

	closePool()
	t.Cleanup(closePool)
}

func testAccProviderAPIKeyConfig(apiKey string) string {
	return fmt.Sprintf(`
provider "uptimekuma" {
  endpoint   = %[1]q
  api_key    = %[2]q
  user_agent = "terraform-provider-uptimekuma-acc/1.0"
  headers = {
    "X-Acceptance-Test" = "api-key"
  }
}
`, endpoint, apiKey)
}

// TestAccProviderAPIKey runs a provider configured with nothing but an
// endpoint and an API key. On Uptime Kuma 3.x it creates, reads and destroys
// a tag with it; on 2.x, which has no API key login, it asserts the error that
// says so.
//
// Not resource.ParallelTest: it changes "API and MCP", a setting of the whole
// server, and it needs the pooled connection to itself, see
// testAccIsolatePool.
func TestAccProviderAPIKey(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("acceptance test, set %s to run it", resource.EnvTfAcc)
	}

	if outOfBandClient.AuthScheme() != kuma.AuthSchemeHTTP {
		testAccProviderAPIKeyUnsupported(t)

		return
	}

	key := testAccEnableAPIKey(t)
	tagName := acctest.RandomWithPrefix("TestAPIKeyTag")
	notificationName := acctest.RandomWithPrefix("TestAPIKeyNotification")
	resources := testAccProviderAPIKeyResources(tagName, notificationName)

	testAccIsolatePool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// The pool is empty, so this connects, and fails, on its own.
				Config: testAccProviderAPIKeyConfig("uk1_not-a-key-of-this-server") + `
data "uptimekuma_tag" "test" {
  name = "does-not-matter"
}
`,
				ExpectError: regexp.MustCompile(`Uptime Kuma rejected the API key`),
			},
			{
				Config: testAccProviderAPIKeyConfig(key) + resources + `
data "uptimekuma_tag" "test" {
  id = uptimekuma_tag.test.id
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_tag.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(tagName),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_tag.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(tagName),
					),
				},
			},
			{
				// A notification deleted behind Terraform's back is missing
				// from the client's cache, so its Read resyncs before it
				// believes the miss. A client authenticated by an API key holds
				// no session token: the resync connects again with the key, and
				// the plan recreates the notification.
				PreConfig: func() { testAccDeleteNotificationByName(t, notificationName) },
				Config:    testAccProviderAPIKeyConfig(key) + resources,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"uptimekuma_notification_webhook.test", plancheck.ResourceActionCreate,
						),
					},
				},
			},
			{
				// Uptime Kuma asks for the account password whenever settings
				// are written, which a provider with api_key does not have.
				Config: testAccProviderAPIKeyConfig(key) + resources + `
resource "uptimekuma_settings" "test" {
  keep_data_period_days = 180
}
`,
				ExpectError: regexp.MustCompile(`(?s)password required.*authenticates\s+with\s+api_key`),
			},
		},
	})
}

// testAccProviderAPIKeyUnsupported asserts that an Uptime Kuma 2.x server,
// which has no API key login on the socket.io API, is reported as such.
func testAccProviderAPIKeyUnsupported(t *testing.T) {
	t.Helper()

	testAccIsolatePool(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderAPIKeyConfig("uk1_not-a-key-a-2x-server-would-take") + `
data "uptimekuma_tag" "test" {
  name = "does-not-matter"
}
`,
				ExpectError: regexp.MustCompile(`Uptime Kuma does not accept API keys for this connection`),
			},
		},
	})
}

// testAccProviderAPIKeyResources is what TestAccProviderAPIKey manages with
// the key: a tag, read from the server, and a notification, read from the
// client's cache.
func testAccProviderAPIKeyResources(tagName string, notificationName string) string {
	return fmt.Sprintf(`
resource "uptimekuma_tag" "test" {
  name  = %[1]q
  color = "#2563eb"
}

resource "uptimekuma_notification_webhook" "test" {
  name                 = %[2]q
  webhook_url          = "https://example.com/api-key-webhook"
  webhook_content_type = "json"
}
`, tagName, notificationName)
}

// testAccDeleteNotificationByName deletes the notification with name through
// outOfBandClient, behind the provider's back.
func testAccDeleteNotificationByName(t *testing.T, name string) {
	t.Helper()

	ctx := t.Context()

	find := func() int64 {
		for _, notification := range outOfBandClient.GetNotifications(ctx) {
			if notification.Name == name {
				return notification.ID
			}
		}

		return 0
	}

	id := find()
	if id == 0 {
		err := outOfBandClient.Resync(ctx)
		if err != nil {
			t.Fatalf("resync out-of-band client: %v", err)
		}

		id = find()
	}

	if id == 0 {
		t.Fatalf("notification %q not found", name)
	}

	err := outOfBandClient.DeleteNotification(ctx, id)
	if err != nil {
		t.Fatalf("delete notification %q out of band: %v", name, err)
	}
}
