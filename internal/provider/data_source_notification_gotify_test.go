package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationGotifyDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGotify")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGotifyDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.uptimekuma_notification_gotify.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_notification_gotify.test",
						"name",
						name,
					),
				),
			},
		},
	})
}

func testAccNotificationGotifyDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_gotify" "test" {
  name              = %[1]q
  is_active         = true
  server_url        = "https://gotify.example.com"
  application_token = "AGe0Ks4WV5fEJkX"
  priority          = 8
}

data "uptimekuma_notification_gotify" "test" {
  name = uptimekuma_notification_gotify.test.name
}
`, name)
}
