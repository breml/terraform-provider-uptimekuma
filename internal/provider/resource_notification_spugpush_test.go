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

func TestAccNotificationSpugPushResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSpugPush")
	nameUpdated := acctest.RandomWithPrefix("NotificationSpugPushUpdated")
	templateKey := "test-template-key-xxxxx"
	templateKeyUpdated := "test-template-key-yyyyy"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSpugPushResourceConfig(name, templateKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_spugpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_spugpush.test",
						tfjsonpath.New("template_key"),
						knownvalue.StringExact(templateKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_spugpush.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationSpugPushResourceConfig(nameUpdated, templateKeyUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_spugpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_spugpush.test",
						tfjsonpath.New("template_key"),
						knownvalue.StringExact(templateKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_spugpush.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_spugpush.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_key"},
			},
		},
	})
}

func testAccNotificationSpugPushResourceConfig(name string, templateKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_spugpush" "test" {
  name         = %[1]q
  is_active    = true
  template_key = %[2]q
}
`, name, templateKey)
}
