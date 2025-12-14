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

func TestAccNotificationDingDingResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationDingDing")
	nameUpdated := acctest.RandomWithPrefix("NotificationDingDingUpdated")
	secretKey := "test-secret-key-123"
	mentioning := "@user1 @user2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationDingDingResourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("webhook_url"), knownvalue.StringExact("https://oapi.dingtalk.com/robot/send?access_token=abcdefg123456")),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("secret_key"), knownvalue.Null()),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("mentioning"), knownvalue.Null()),
				},
			},
			{
				Config: testAccNotificationDingDingResourceConfigWithOptionalFields(nameUpdated, secretKey, mentioning),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("webhook_url"), knownvalue.StringExact("https://oapi.dingtalk.com/robot/send?access_token=abcdefg123456")),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("secret_key"), knownvalue.StringExact(secretKey)),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("mentioning"), knownvalue.StringExact(mentioning)),
				},
			},
			{
				Config: testAccNotificationDingDingResourceConfig(nameUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("webhook_url"), knownvalue.StringExact("https://oapi.dingtalk.com/robot/send?access_token=abcdefg123456")),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("secret_key"), knownvalue.Null()),
					statecheck.ExpectKnownValue("uptimekuma_notification_dingding.test", tfjsonpath.New("mentioning"), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccNotificationDingDingResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_dingding" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://oapi.dingtalk.com/robot/send?access_token=abcdefg123456"
}
`, name)
}

func testAccNotificationDingDingResourceConfigWithOptionalFields(name, secretKey, mentioning string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_dingding" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://oapi.dingtalk.com/robot/send?access_token=abcdefg123456"
  secret_key  = %[2]q
  mentioning  = %[3]q
}
`, name, secretKey, mentioning)
}
