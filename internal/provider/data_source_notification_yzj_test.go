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

func TestAccNotificationYZJDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationYZJ")
	webhookURL := "https://yzj.example.com/webhook/test"
	token := "test-token-placeholder"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationYZJDataSourceConfig(name, webhookURL, token),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_yzj.by_id",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_yzj.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationYZJDataSourceConfig(name, webhookURL, token),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_yzj.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_yzj.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationYZJDataSourceConfig(name string, webhookURL string, token string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_yzj" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
  token       = %[3]q
}

data "uptimekuma_notification_yzj" "by_id" {
  id = uptimekuma_notification_yzj.test.id
}

data "uptimekuma_notification_yzj" "by_name" {
  name = uptimekuma_notification_yzj.test.name
}
`, name, webhookURL, token)
}
