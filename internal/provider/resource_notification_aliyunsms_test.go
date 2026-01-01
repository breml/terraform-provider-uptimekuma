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

func TestAccNotificationAliyunsmsResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationAliyunsms")
	nameUpdated := acctest.RandomWithPrefix("NotificationAliyunsmsUpdated")
	accessKeyID := "test-access-key-id"
	accessKeyIDUpdated := "test-access-key-id-updated"
	secretAccessKey := "test-secret-access-key"
	secretAccessKeyUpdated := "test-secret-access-key-updated"
	phoneNumber := "+1234567890"
	phoneNumberUpdated := "+0987654321"
	signName := "TestSign"
	signNameUpdated := "TestSignUpdated"
	templateCode := "SMS_001"
	templateCodeUpdated := "SMS_002"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAliyunsmsResourceConfig(
					name,
					accessKeyID,
					secretAccessKey,
					phoneNumber,
					signName,
					templateCode,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("access_key_id"),
						knownvalue.StringExact(accessKeyID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("secret_access_key"),
						knownvalue.StringExact(secretAccessKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact(phoneNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("sign_name"),
						knownvalue.StringExact(signName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("template_code"),
						knownvalue.StringExact(templateCode),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationAliyunsmsResourceConfig(
					nameUpdated,
					accessKeyIDUpdated,
					secretAccessKeyUpdated,
					phoneNumberUpdated,
					signNameUpdated,
					templateCodeUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("access_key_id"),
						knownvalue.StringExact(accessKeyIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("secret_access_key"),
						knownvalue.StringExact(secretAccessKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact(phoneNumberUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("sign_name"),
						knownvalue.StringExact(signNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("template_code"),
						knownvalue.StringExact(templateCodeUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_aliyunsms.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccNotificationAliyunsmsResourceConfig(
	name string,
	accessKeyID string,
	secretAccessKey string,
	phoneNumber string,
	signName string,
	templateCode string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_aliyunsms" "test" {
  name               = %[1]q
  is_active          = true
  access_key_id      = %[2]q
  secret_access_key  = %[3]q
  phone_number       = %[4]q
  sign_name          = %[5]q
  template_code      = %[6]q
}
`, name, accessKeyID, secretAccessKey, phoneNumber, signName, templateCode)
}
