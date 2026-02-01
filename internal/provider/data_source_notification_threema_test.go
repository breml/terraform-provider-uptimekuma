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

func TestAccNotificationThreemaDataSourceByID(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationThreema")
	senderIdentity := "TESTID123"
	secret := "testsecret123456789"
	recipient := "john@example.com"
	recipientType := "email"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationThreemaDataSourceByIDConfig(
					name,
					senderIdentity,
					secret,
					recipient,
					recipientType,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_threema.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationThreemaDataSourceByIDConfig(
	name string,
	senderIdentity string,
	secret string,
	recipient string,
	recipientType string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_threema" "test" {
  name               = %[1]q
  is_active          = true
  sender_identity    = %[2]q
  secret             = %[3]q
  recipient          = %[4]q
  recipient_type     = %[5]q
}

data "uptimekuma_notification_threema" "test" {
  id = uptimekuma_notification_threema.test.id
}
`, name, senderIdentity, secret, recipient, recipientType)
}
