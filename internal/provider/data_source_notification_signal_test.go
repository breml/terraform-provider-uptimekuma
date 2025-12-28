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

func TestAccNotificationSignalDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationSignal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSignalDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_signal.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationSignalDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_signal.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSignalDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_signal" "test" {
  name       = %[1]q
  is_active  = true
  url        = "http://signal.example.com:8080"
  number     = "+1234567890"
  recipients = "+9876543210"
}

data "uptimekuma_notification_signal" "test" {
  name = uptimekuma_notification_signal.test.name
}
`, name)
}

func testAccNotificationSignalDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_signal" "test" {
  name       = %[1]q
  is_active  = true
  url        = "http://signal.example.com:8080"
  number     = "+1234567890"
  recipients = "+9876543210"
}

data "uptimekuma_notification_signal" "test" {
  id = uptimekuma_notification_signal.test.id
}
`, name)
}
