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

func TestAccNotificationWhapiDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWhapi")
	apiURL := "https://gate.whapi.cloud"
	authToken := "test-auth-token-123"
	recipient := "1234567890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWhapiDataSourceByNameConfig(
					name,
					apiURL,
					authToken,
					recipient,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_whapi.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationWhapiDataSourceByIDConfig(
					name,
					apiURL,
					authToken,
					recipient,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_whapi.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationWhapiDataSourceByNameConfig(
	name string,
	apiURL string,
	authToken string,
	recipient string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_whapi" "test" {
  name       = %[1]q
  is_active  = true
  api_url    = %[2]q
  auth_token = %[3]q
  recipient  = %[4]q
}

data "uptimekuma_notification_whapi" "test" {
  name = uptimekuma_notification_whapi.test.name
}
`, name, apiURL, authToken, recipient)
}

func testAccNotificationWhapiDataSourceByIDConfig(
	name string,
	apiURL string,
	authToken string,
	recipient string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_whapi" "test" {
  name       = %[1]q
  is_active  = true
  api_url    = %[2]q
  auth_token = %[3]q
  recipient  = %[4]q
}

data "uptimekuma_notification_whapi" "test" {
  id = uptimekuma_notification_whapi.test.id
}
`, name, apiURL, authToken, recipient)
}
