package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationPagerTreeResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPagerTree")
	nameUpdated := acctest.RandomWithPrefix("NotificationPagerTreeUpdated")
	integrationURL := "https://alerts.pagertree.com/api/v2/incidents"
	integrationURLUpdated := "https://alerts.pagertree.com/api/v2/incidents/updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPagerTreeResourceConfig(
					name,
					integrationURL,
					"high",
					"resolve",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagertree.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagertree.test",
						tfjsonpath.New("integration_url"),
						knownvalue.StringExact(integrationURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagertree.test",
						tfjsonpath.New("urgency"),
						knownvalue.StringExact("high"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagertree.test",
						tfjsonpath.New("auto_resolve"),
						knownvalue.StringExact("resolve"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagertree.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationPagerTreeResourceConfig(
					nameUpdated,
					integrationURLUpdated,
					"medium",
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagertree.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagertree.test",
						tfjsonpath.New("integration_url"),
						knownvalue.StringExact(integrationURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagertree.test",
						tfjsonpath.New("urgency"),
						knownvalue.StringExact("medium"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagertree.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_pagertree.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationPagerTreeImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"integration_url"},
			},
		},
	})
}

// testAccNotificationPagerTreeImportStateID extracts the resource ID for import testing.
func testAccNotificationPagerTreeImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_pagertree.test"]
	return rs.Primary.Attributes["id"], nil
}

func testAccNotificationPagerTreeResourceConfig(
	name string, integrationURL string,
	urgency string, autoResolve string,
) string {
	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pagertree" "test" {
  name              = %[1]q
  is_active         = true
  integration_url   = %[2]q
  urgency           = %[3]q
`, name, integrationURL, urgency)

	if autoResolve != "" {
		config += fmt.Sprintf(`  auto_resolve      = %[1]q
`, autoResolve)
	}

	config += `}
`

	return config
}
