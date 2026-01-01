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

func TestAccNotificationAlertaResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationAlerta")
	nameUpdated := acctest.RandomWithPrefix("NotificationAlertaUpdated")
	apiEndpoint := "https://alerta.example.com"
	apiEndpointUpdated := "https://alerta2.example.com"
	apiKey := "test-api-key-12345"
	apiKeyUpdated := "test-api-key-67890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAlertaResourceConfig(
					name,
					apiEndpoint,
					apiKey,
					"production",
					"alert",
					"ok",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("api_endpoint"),
						knownvalue.StringExact(apiEndpoint),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("environment"),
						knownvalue.StringExact("production"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("alert_state"),
						knownvalue.StringExact("alert"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("recover_state"),
						knownvalue.StringExact("ok"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationAlertaResourceConfig(
					nameUpdated,
					apiEndpointUpdated,
					apiKeyUpdated,
					"staging",
					"critical",
					"resolved",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("api_endpoint"),
						knownvalue.StringExact(apiEndpointUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("environment"),
						knownvalue.StringExact("staging"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("alert_state"),
						knownvalue.StringExact("critical"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("recover_state"),
						knownvalue.StringExact("resolved"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_alerta.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccNotificationAlertaResourceConfig(
	name string,
	apiEndpoint string,
	apiKey string,
	environment string,
	alertState string,
	recoverState string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_alerta" "test" {
  name             = %[1]q
  is_active        = true
  api_endpoint     = %[2]q
  api_key          = %[3]q
  environment      = %[4]q
  alert_state      = %[5]q
  recover_state    = %[6]q
}
`, name, apiEndpoint, apiKey, environment, alertState, recoverState)
}
