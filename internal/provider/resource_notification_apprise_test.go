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

func TestAccNotificationAppriseResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationApprise")
	nameUpdated := acctest.RandomWithPrefix("NotificationAppriseUpdated")
	appriseURL := "json://localhost:8000/path"
	appriseURLUpdated := "json://localhost:8001/path"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAppriseResourceConfig(name, appriseURL, "Test Apprise"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("apprise_url"), knownvalue.StringExact(appriseURL)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("title"), knownvalue.StringExact("Test Apprise")),
				},
			},
			{
				Config: testAccNotificationAppriseResourceConfig(nameUpdated, appriseURLUpdated, "Updated Apprise"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("apprise_url"), knownvalue.StringExact(appriseURLUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("title"), knownvalue.StringExact("Updated Apprise")),
				},
			},
		},
	})
}

func testAccNotificationAppriseResourceConfig(name, appriseURL, title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_apprise" "test" {
  name        = %[1]q
  is_active   = true
  apprise_url = %[2]q
  title       = %[3]q
}
`, name, appriseURL, title)
}
