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
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/ory/dockertest/v3"

	kuma "github.com/breml/go-uptime-kuma-client"

	"github.com/breml/terraform-provider-uptimekuma/internal/client"
)

// apiKeyScript runs with node inside the Uptime Kuma 3.x container, on the
// socket.io-client the image bundles. It connects with the session cookie of
// outOfBandClient, which spends no sign-in of the server's rate limit, sets
// "API and MCP" (apiEnabled) to KUMA_API_ENABLED and, given KUMA_API_KEY_NAME,
// creates an API key and prints it as {"key": "..."}.
//
// The go-uptime-kuma-client keeps its addAPIKey and setSettings-with-apiEnabled
// helpers test-only, and has no generic emit, so this is the least invasive way
// to reach those socket events from here.
const apiKeyScript = `
const { io } = require("/app/node_modules/socket.io-client");
const fail = (msg) => { console.error(msg); process.exit(1); };
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
		let key = "";
		if (process.env.KUMA_API_KEY_NAME) {
			const added = await emit("addAPIKey", { name: process.env.KUMA_API_KEY_NAME });
			if (!added.ok) fail("addAPIKey: " + added.msg);
			key = added.key;
		}
		process.stdout.write(JSON.stringify({ key }));
		clearTimeout(timer);
		socket.close();
		process.exit(0);
	} catch (err) {
		fail(String(err));
	}
});
`

// runAPIKeyScript runs apiKeyScript in the test container and returns the key
// it created, if it was asked to create one.
func runAPIKeyScript(t *testing.T, enabled bool, keyName string) string {
	t.Helper()

	var stdout, stderr bytes.Buffer

	exitCode, err := kumaContainer.Exec([]string{"node", "-e", apiKeyScript}, dockertest.ExecOptions{
		Env: []string{
			"KUMA_SESSION=" + outOfBandClient.SessionToken(),
			"KUMA_PASSWORD=" + password,
			fmt.Sprintf("KUMA_API_ENABLED=%t", enabled),
			"KUMA_API_KEY_NAME=" + keyName,
		},
		StdOut: &stdout,
		StdErr: &stderr,
	})
	if err != nil || exitCode != 0 {
		t.Fatalf("api key script failed with exit code %d: %v: %s", exitCode, err, stderr.String())
	}

	var result struct {
		Key string `json:"key"`
	}

	err = json.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		t.Fatalf("api key script printed %q: %v", stdout.String(), err)
	}

	return result.Key
}

// testAccEnableAPIKey turns "API and MCP" on, creates an API key and returns
// it. "API and MCP" is turned off again when the test ends, because it is a
// setting of the whole server.
func testAccEnableAPIKey(t *testing.T) string {
	t.Helper()

	key := runAPIKeyScript(t, true, acctest.RandomWithPrefix("tf-acc"))
	if key == "" {
		t.Fatal("expected the api key script to return a key")
	}

	t.Cleanup(func() { runAPIKeyScript(t, false, "") })

	return key
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
				Config: testAccProviderAPIKeyConfig(key) + fmt.Sprintf(`
resource "uptimekuma_tag" "test" {
  name  = %[1]q
  color = "#2563eb"
}

data "uptimekuma_tag" "test" {
  id = uptimekuma_tag.test.id
}
`, tagName),
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
				// Uptime Kuma asks for the account password whenever settings
				// are written, which a provider with api_key does not have.
				Config: testAccProviderAPIKeyConfig(key) + fmt.Sprintf(`
resource "uptimekuma_tag" "test" {
  name  = %[1]q
  color = "#2563eb"
}

resource "uptimekuma_settings" "test" {
  keep_data_period_days = 180
}
`, tagName),
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
