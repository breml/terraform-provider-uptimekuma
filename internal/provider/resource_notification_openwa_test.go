package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccNotificationOpenWaResource covers create, update and import, and the shapes of the two
// optional fields: unset, set, and explicitly false and empty, which must not collapse to null.
// Going back to unset checks that the keys are dropped rather than kept by the server, which would
// show up as a non-empty plan. The trailing slash of the API URL has to round-trip verbatim.
func TestAccNotificationOpenWaResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOpenWa")
	nameUpdated := acctest.RandomWithPrefix("NotificationOpenWaUpdated")
	addr := "uptimekuma_notification_openwa.test"
	plain := openWaTestConfig{
		name:             name,
		apiURL:           "http://localhost:2785/",
		apiKey:           "key-1",
		session:          "default",
		chatID:           "00117612345678@c.us",
		useCustomMessage: "null",
		customMessage:    "null",
	}
	custom := openWaTestConfig{
		name:             nameUpdated,
		apiURL:           "http://openwa.example.com:2785",
		apiKey:           "key-2",
		session:          "other session",
		chatID:           "00117612345678@c.us,123456789012345678@g.us,1234567890@lid",
		useCustomMessage: "true",
		customMessage:    `"{{ NAME }}: {{ msg }}"`,
	}
	disabled := custom
	disabled.useCustomMessage = "false"
	disabled.customMessage = `""`
	unset := custom
	unset.useCustomMessage = "null"
	unset.customMessage = "null"

	importStep := resource.TestStep{ResourceName: addr, ImportState: true, ImportStateVerify: true}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: plain.hcl(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("is_default"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("apply_existing"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("api_url"), knownvalue.StringExact(plain.apiURL)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("api_key"), knownvalue.StringExact(plain.apiKey)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("session"), knownvalue.StringExact(plain.session)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("chat_id"), knownvalue.StringExact(plain.chatID)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("use_custom_message"), knownvalue.Null()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("custom_message"), knownvalue.Null()),
				},
			},
			importStep,
			{
				Config: custom.hcl(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("api_url"), knownvalue.StringExact(custom.apiURL)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("api_key"), knownvalue.StringExact(custom.apiKey)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("session"), knownvalue.StringExact(custom.session)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("chat_id"), knownvalue.StringExact(custom.chatID)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("use_custom_message"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(
						addr,
						tfjsonpath.New("custom_message"),
						knownvalue.StringExact("{{ NAME }}: {{ msg }}"),
					),
				},
			},
			importStep,
			{
				Config: disabled.hcl(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("use_custom_message"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("custom_message"), knownvalue.StringExact("")),
				},
			},
			importStep,
			{
				Config: unset.hcl(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("use_custom_message"), knownvalue.Null()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("custom_message"), knownvalue.Null()),
				},
			},
			importStep,
		},
	})
}

// openWaTestConfig holds the attributes of the test resource. useCustomMessage and customMessage are
// HCL expressions, so that null leaves the attribute unset.
type openWaTestConfig struct {
	name             string
	apiURL           string
	apiKey           string
	session          string
	chatID           string
	useCustomMessage string
	customMessage    string
}

func (c openWaTestConfig) hcl() string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_openwa" "test" {
  name               = %[1]q
  api_url            = %[2]q
  api_key            = %[3]q
  session            = %[4]q
  chat_id            = %[5]q
  use_custom_message = %[6]s
  custom_message     = %[7]s
}
`, c.name, c.apiURL, c.apiKey, c.session, c.chatID, c.useCustomMessage, c.customMessage)
}
