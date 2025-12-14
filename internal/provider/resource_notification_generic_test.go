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

func TestAccNotificationGenericResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGeneric")
	nameUpdated := acctest.RandomWithPrefix("NotificationGenericUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGenericResourceConfig(name, "custom-type", map[string]interface{}{"url": "http://example.com/webhook"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_generic.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_notification_generic.test", tfjsonpath.New("type"), knownvalue.StringExact("custom-type")),
					statecheck.ExpectKnownValue("uptimekuma_notification_generic.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_generic.test", tfjsonpath.New("config"), knownvalue.StringFunc(func(v string) error {
						configMap := make(map[string]interface{})
						if err := json.Unmarshal([]byte(v), &configMap); err != nil {
							return err
						}
						if fmt.Sprintf("%v", configMap["url"]) != "http://example.com/webhook" {
							return fmt.Errorf(`"url" != "http://example.com/webhook", got: %v`, configMap["url"])
						}
						return nil
					})),
				},
			},
			{
				Config: testAccNotificationGenericResourceConfig(nameUpdated, "another-type", map[string]interface{}{"url": "http://example.com/webhook2", "token": "secret"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_generic.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_generic.test", tfjsonpath.New("type"), knownvalue.StringExact("another-type")),
					statecheck.ExpectKnownValue("uptimekuma_notification_generic.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_generic.test", tfjsonpath.New("config"), knownvalue.StringFunc(func(v string) error {
						configMap := make(map[string]interface{})
						if err := json.Unmarshal([]byte(v), &configMap); err != nil {
							return err
						}
						if fmt.Sprintf("%v", configMap["url"]) != "http://example.com/webhook2" {
							return fmt.Errorf(`"url" != "http://example.com/webhook2", got: %v`, configMap["url"])
						}
						if fmt.Sprintf("%v", configMap["token"]) != "secret" {
							return fmt.Errorf(`"token" != "secret", got: %v`, configMap["token"])
						}
						return nil
					})),
				},
			},
		},
	})
}

func testAccNotificationGenericResourceConfig(name, notificationType string, config map[string]interface{}) string {
	configJSON, _ := json.Marshal(config)
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_generic" "test" {
  name      = %[1]q
  is_active = true
  type      = %[2]q
  config    = %[3]q
}
`, name, notificationType, string(configJSON))
}
