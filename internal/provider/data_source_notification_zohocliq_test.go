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

func TestAccNotificationZohoCliqDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationZohoCliq")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationZohoCliqDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_zohocliq.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_zohocliq.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationZohoCliqDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_zohocliq" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://cliq.zoho.com/company/api/v2/channelsbyname/general/message?zapikey=test-key"
}

data "uptimekuma_notification_zohocliq" "by_name" {
  name = uptimekuma_notification_zohocliq.test.name
}

data "uptimekuma_notification_zohocliq" "by_id" {
  id = uptimekuma_notification_zohocliq.test.id
}
`, name)
}
