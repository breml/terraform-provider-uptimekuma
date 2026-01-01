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

func TestAccNotificationAliyunsmsDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationAliyunsms")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAliyunsmsDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func TestAccNotificationAliyunsmsDataSourceByID(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationAliyunsms")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAliyunsmsDataSourceByIDConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccNotificationAliyunsmsDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_aliyunsms" "test" {
  name               = %[1]q
  is_active          = true
  access_key_id      = "test-access-key-id"
  secret_access_key  = "test-secret-access-key"
  phone_number       = "+1234567890"
  sign_name          = "TestSign"
  template_code      = "SMS_001"
}

data "uptimekuma_notification_aliyunsms" "test" {
  name = uptimekuma_notification_aliyunsms.test.name
}
`, name)
}

func testAccNotificationAliyunsmsDataSourceByIDConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_aliyunsms" "test" {
  name               = %[1]q
  is_active          = true
  access_key_id      = "test-access-key-id"
  secret_access_key  = "test-secret-access-key"
  phone_number       = "+1234567890"
  sign_name          = "TestSign"
  template_code      = "SMS_001"
}

data "uptimekuma_notification_aliyunsms" "test" {
  id = uptimekuma_notification_aliyunsms.test.id
}
`, name)
}
