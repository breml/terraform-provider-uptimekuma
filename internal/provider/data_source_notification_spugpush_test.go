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

func TestAccNotificationSpugPushDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSpugPushDS")
	templateKey := "test-template-key-xxxxx"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSpugPushDataSourceConfig(name, templateKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_spugpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_spugpush.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccNotificationSpugPushDataSourceConfig(name string, templateKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_spugpush" "test" {
  name         = %[1]q
  is_active    = true
  template_key = %[2]q
}

data "uptimekuma_notification_spugpush" "test" {
  name = uptimekuma_notification_spugpush.test.name
}
`, name, templateKey)
}
