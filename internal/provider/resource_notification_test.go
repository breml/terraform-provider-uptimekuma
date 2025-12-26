package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationResource(t *testing.T) {
	name := acctest.RandomWithPrefix("Notification")
	nameUpdated := acctest.RandomWithPrefix("NotificationUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNotificationResourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("ntfy"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification.test",
						tfjsonpath.New("config"),
						knownvalue.StringFunc(func(v string) error {
							configMap := make(map[string]any, 5)
							err := json.Unmarshal([]byte(v), &configMap)
							if err != nil {
								return err
							}

							if fmt.Sprintf("%v", configMap["ntfyAuthenticationMethod"]) != "none" {
								return fmt.Errorf(
									`"ntfyAuthenticationMethod" != "none", got: %v`,
									configMap["ntfyAuthenticationMethod"],
								)
							}

							if fmt.Sprintf("%v", configMap["ntfyserverurl"]) != "https://ntfy.sh/" {
								return fmt.Errorf(
									`"ntfyserverurl" != "https://ntfy.sh/", got: %v`,
									configMap["ntfyserverurl"],
								)
							}

							if fmt.Sprintf("%v", configMap["ntfyPriority"]) != "5" {
								return fmt.Errorf(`"ntfyPriority" != "5", got: %v`, configMap["ntfyPriority"])
							}

							if fmt.Sprintf("%v", configMap["ntfytopic"]) != name {
								return fmt.Errorf(`"ntfytopic" != %q, got: %v`, name, configMap["ntfytopic"])
							}

							return nil
						}),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccNotificationResourceConfig(nameUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("ntfy"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification.test",
						tfjsonpath.New("config"),
						knownvalue.StringFunc(func(v string) error {
							configMap := make(map[string]any, 5)
							err := json.Unmarshal([]byte(v), &configMap)
							if err != nil {
								return err
							}

							if fmt.Sprintf("%v", configMap["ntfyAuthenticationMethod"]) != "none" {
								return fmt.Errorf(
									`"ntfyAuthenticationMethod" != "none", got: %v`,
									configMap["ntfyAuthenticationMethod"],
								)
							}

							if fmt.Sprintf("%v", configMap["ntfyserverurl"]) != "https://ntfy.sh/" {
								return fmt.Errorf(
									`"ntfyserverurl" != "https://ntfy.sh/", got: %v`,
									configMap["ntfyserverurl"],
								)
							}

							if fmt.Sprintf("%v", configMap["ntfyPriority"]) != "5" {
								return fmt.Errorf(`"ntfyPriority" != "5", got: %v`, configMap["ntfyPriority"])
							}

							if fmt.Sprintf("%v", configMap["ntfytopic"]) != nameUpdated {
								return fmt.Errorf(`"ntfytopic" != %q, got: %v`, nameUpdated, configMap["ntfytopic"])
							}

							return nil
						}),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccNotificationResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification" "test" {
  name      = %[1]q
  is_active = true
  type      = "ntfy"
  config = jsonencode({
    ntfyAuthenticationMethod = "none"
    ntfyserverurl            = "https://ntfy.sh/"
    ntfyPriority             = 5
    ntfytopic                = %[1]q
  })
}
`, name)
}
