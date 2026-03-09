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

func TestAccNotificationOneBotResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOneBot")
	nameUpdated := acctest.RandomWithPrefix("NotificationOneBotUpdated")
	httpAddr := "http://localhost:5700"
	httpAddrUpdated := "http://localhost:5701"
	accessToken := "testtoken123"
	accessTokenUpdated := "testtoken456"
	msgType := "group"
	msgTypeUpdated := "private"
	receiverID := "123456789"
	receiverIDUpdated := "987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOneBotResourceConfig(
					name,
					httpAddr,
					accessToken,
					msgType,
					receiverID,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("http_addr"),
						knownvalue.StringExact(httpAddr),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("msg_type"),
						knownvalue.StringExact(msgType),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("receiver_id"),
						knownvalue.StringExact(receiverID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationOneBotResourceConfig(
					nameUpdated,
					httpAddrUpdated,
					accessTokenUpdated,
					msgTypeUpdated,
					receiverIDUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("http_addr"),
						knownvalue.StringExact(httpAddrUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("msg_type"),
						knownvalue.StringExact(msgTypeUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("receiver_id"),
						knownvalue.StringExact(receiverIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onebot.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_onebot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"access_token"},
			},
		},
	})
}

func testAccNotificationOneBotResourceConfig(
	name string,
	httpAddr string,
	accessToken string,
	msgType string,
	receiverID string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_onebot" "test" {
  name         = %[1]q
  is_active    = true
  http_addr    = %[2]q
  access_token = %[3]q
  msg_type     = %[4]q
  receiver_id  = %[5]q
}
`, name, httpAddr, accessToken, msgType, receiverID)
}
