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

func TestAccNotificationSMSEagleDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSEagle")
	url := "https://192.168.1.100"
	token := "test-api-token-123"
	recipientTo := "+1234567890"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSEagleDataSourceConfig(
					name,
					url,
					token,
					recipientTo,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smseagle.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smseagle.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSMSEagleDataSourceConfig(
	name string,
	url string,
	token string,
	recipientTo string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smseagle" "test" {
  name           = %[1]q
  is_active      = true
  url            = %[2]q
  token          = %[3]q
  recipient_type = "smseagle-to"
  recipient_to   = %[4]q
  api_type       = "smseagle-apiv2"
}

data "uptimekuma_notification_smseagle" "by_name" {
  name = uptimekuma_notification_smseagle.test.name
}

data "uptimekuma_notification_smseagle" "by_id" {
  id = uptimekuma_notification_smseagle.test.id
}
`, name, url, token, recipientTo)
}
