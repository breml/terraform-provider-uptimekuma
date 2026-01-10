package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationLunaseaResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationLunasea")
	nameUpdated := acctest.RandomWithPrefix("NotificationLunaseaUpdated")
	userID := "user123"
	userIDUpdated := "user456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationLunaseaResourceConfigUser(name, userID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("target"),
						knownvalue.StringExact("user"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("lunasea_user_id"),
						knownvalue.StringExact(userID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationLunaseaResourceConfigUser(nameUpdated, userIDUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("lunasea_user_id"),
						knownvalue.StringExact(userIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func TestAccNotificationLunaseaResourceDevice(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationLunaseaDevice")
	nameUpdated := acctest.RandomWithPrefix("NotificationLunaseaDeviceUpdated")
	deviceID := "device123"
	deviceIDUpdated := "device456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationLunaseaResourceConfigDevice(name, deviceID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("target"),
						knownvalue.StringExact("device"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("device"),
						knownvalue.StringExact(deviceID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationLunaseaResourceConfigDevice(nameUpdated, deviceIDUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("device"),
						knownvalue.StringExact(deviceIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_lunasea.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func TestAccNotificationLunaseaResourceImportState(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationLunasea")
	userID := "user789"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationLunaseaResourceConfigUser(name, userID),
			},
			{
				ResourceName:      "uptimekuma_notification_lunasea.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationLunaseaResourceConfigUser(name string, userID string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_lunasea" "test" {
  name             = %[1]q
  is_active        = true
  target           = "user"
  lunasea_user_id  = %[2]q
}
`, name, userID)
}

func testAccNotificationLunaseaResourceConfigDevice(name string, deviceID string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_lunasea" "test" {
  name      = %[1]q
  is_active = true
  target    = "device"
  device    = %[2]q
}
`, name, deviceID)
}

func TestAccNotificationLunaseaResourceValidation_MissingUserID(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationLunaseaValidation")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationLunaseaResourceConfigMissingUserID(name),
				ExpectError: regexp.MustCompile("lunasea_user_id is required when target is 'user'"),
			},
		},
	})
}

func TestAccNotificationLunaseaResourceValidation_MissingDevice(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationLunaseaValidation")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationLunaseaResourceConfigMissingDevice(name),
				ExpectError: regexp.MustCompile("device is required when target is 'device'"),
			},
		},
	})
}

func testAccNotificationLunaseaResourceConfigMissingUserID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_lunasea" "test" {
  name      = %[1]q
  is_active = true
  target    = "user"
}
`, name)
}

func testAccNotificationLunaseaResourceConfigMissingDevice(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_lunasea" "test" {
  name      = %[1]q
  is_active = true
  target    = "device"
}
`, name)
}
