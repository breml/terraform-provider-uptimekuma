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

func TestAccNotificationGrafanaOncallResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGrafanaOncall")
	nameUpdated := acctest.RandomWithPrefix("NotificationGrafanaOncallUpdated")
	url := "https://grafana-oncall.example.com/integrations/v1/webhook/abc123/"
	urlUpdated := "https://grafana-oncall.example.com/integrations/v1/webhook/def456/"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGrafanaOncallResourceConfig(
					name,
					url,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_grafanaoncall.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_grafanaoncall.test",
						tfjsonpath.New("grafana_oncall_url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_grafanaoncall.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationGrafanaOncallResourceConfig(
					nameUpdated,
					urlUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_grafanaoncall.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_grafanaoncall.test",
						tfjsonpath.New("grafana_oncall_url"),
						knownvalue.StringExact(urlUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_grafanaoncall.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_grafanaoncall.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationGrafanaOncallResourceConfig(
	name string,
	grafanaOncallURL string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_grafanaoncall" "test" {
  name                = %[1]q
  is_active           = true
  grafana_oncall_url  = %[2]q
}
`, name, grafanaOncallURL)
}
