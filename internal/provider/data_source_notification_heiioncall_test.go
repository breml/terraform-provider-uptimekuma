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

func TestAccNotificationHeiiOnCallDataSourceByID(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHeiiOnCallDS")
	apiKey := acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)
	triggerID := acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationHeiiOnCallDataSourceByIDConfig(name, apiKey, triggerID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_heiioncall.by_id",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_heiioncall.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationHeiiOnCallDataSourceByIDConfig(
	name string,
	apiKey string,
	triggerID string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_heiioncall" "test" {
  name       = %[1]q
  api_key    = %[2]q
  trigger_id = %[3]q
}

data "uptimekuma_notification_heiioncall" "by_id" {
  id = uptimekuma_notification_heiioncall.test.id
}
`, name, apiKey, triggerID)
}

func TestAccNotificationHeiiOnCallDataSourceByName(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHeiiOnCallDS")
	apiKey := acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)
	triggerID := acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationHeiiOnCallDataSourceByNameConfig(name, apiKey, triggerID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_heiioncall.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_heiioncall.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationHeiiOnCallDataSourceByNameConfig(
	name string,
	apiKey string,
	triggerID string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_heiioncall" "test" {
  name       = %[1]q
  api_key    = %[2]q
  trigger_id = %[3]q
}

data "uptimekuma_notification_heiioncall" "by_name" {
  name = uptimekuma_notification_heiioncall.test.name
}
`, name, apiKey, triggerID)
}
