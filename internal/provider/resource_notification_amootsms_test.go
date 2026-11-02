package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

const testAccAmootSMSResourceName = "uptimekuma_notification_amootsms.test"

// testAccAmootSMSChecks asserts the optional attributes. An unset one must read back as null:
// the fields are omitempty on the wire, so any other value would be a perpetual diff.
func testAccAmootSMSChecks(
	usePattern knownvalue.Check,
	code knownvalue.Check,
	ownLine knownvalue.Check,
	line knownvalue.Check,
) []statecheck.StateCheck {
	return []statecheck.StateCheck{
		statecheck.ExpectKnownValue(testAccAmootSMSResourceName, tfjsonpath.New("use_pattern"), usePattern),
		statecheck.ExpectKnownValue(testAccAmootSMSResourceName, tfjsonpath.New("pattern_code_id"), code),
		statecheck.ExpectKnownValue(testAccAmootSMSResourceName, tfjsonpath.New("use_own_line"), ownLine),
		statecheck.ExpectKnownValue(testAccAmootSMSResourceName, tfjsonpath.New("line_number"), line),
	}
}

// TestAccNotificationAmootSMSResource walks through every send mode and back to the plain one, whose
// omitted attributes must clear what the previous step stored.
func TestAccNotificationAmootSMSResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationAmootSMS")
	nameUpdated := acctest.RandomWithPrefix("NotificationAmootSMSUpdated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAmootSMSResourceConfig(name, `
  line_number = "public"`),
				ConfigStateChecks: append(
					testAccAmootSMSChecks(
						knownvalue.Null(), knownvalue.Null(), knownvalue.Null(), knownvalue.StringExact("public"),
					),
					statecheck.ExpectKnownValue(
						testAccAmootSMSResourceName,
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						testAccAmootSMSResourceName,
						tfjsonpath.New("api_token"),
						knownvalue.StringExact("token-1"),
					),
					statecheck.ExpectKnownValue(
						testAccAmootSMSResourceName,
						tfjsonpath.New("mobiles"),
						knownvalue.StringExact("09120000000,09120000001"),
					),
				),
			},
			{
				Config: testAccNotificationAmootSMSResourceConfigSecrets(
					nameUpdated, "token-2", "09120000000, 09120000001", `
  use_pattern     = true
  pattern_code_id = 1234`,
				),
				ConfigStateChecks: append(
					testAccAmootSMSChecks(
						knownvalue.Bool(true), knownvalue.Int64Exact(1234), knownvalue.Null(), knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						testAccAmootSMSResourceName,
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						testAccAmootSMSResourceName,
						tfjsonpath.New("api_token"),
						knownvalue.StringExact("token-2"),
					),
					statecheck.ExpectKnownValue(
						testAccAmootSMSResourceName,
						tfjsonpath.New("mobiles"),
						knownvalue.StringExact("09120000000, 09120000001"),
					),
				),
			},
			{
				ResourceName:      testAccAmootSMSResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationAmootSMSResourceConfig(nameUpdated, `
  use_pattern     = true
  pattern_code_id = 5678
  use_own_line    = true
  line_number     = "50001"`),
				ConfigStateChecks: testAccAmootSMSChecks(
					knownvalue.Bool(true), knownvalue.Int64Exact(5678), knownvalue.Bool(true), knownvalue.StringExact("50001"),
				),
			},
			{
				ResourceName:      testAccAmootSMSResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationAmootSMSResourceConfig(nameUpdated, `
  use_pattern = false
  line_number = "public"`),
				ConfigStateChecks: testAccAmootSMSChecks(
					knownvalue.Bool(false), knownvalue.Null(), knownvalue.Null(), knownvalue.StringExact("public"),
				),
			},
			{
				Config: testAccNotificationAmootSMSResourceConfig(nameUpdated, `
  line_number = "public"`),
				ConfigStateChecks: testAccAmootSMSChecks(
					knownvalue.Null(), knownvalue.Null(), knownvalue.Null(), knownvalue.StringExact("public"),
				),
			},
			{
				ResourceName:      testAccAmootSMSResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationAmootSMSResourceInvalidConfig covers the configurations Uptime Kuma would save
// without complaint and then fail on the first alert.
func TestAccNotificationAmootSMSResourceInvalidConfig(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationAmootSMSInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationAmootSMSResourceConfig(name, ""),
				ExpectError: regexp.MustCompile(`"line_number"\s+must\s+be\s+set`),
			},
			{
				Config: testAccNotificationAmootSMSResourceConfig(name, `
  use_pattern = true`),
				ExpectError: regexp.MustCompile(`"pattern_code_id"\s+must\s+be\s+set`),
			},
			{
				Config: testAccNotificationAmootSMSResourceConfig(name, `
  use_pattern     = true
  pattern_code_id = 1234
  use_own_line    = true`),
				ExpectError: regexp.MustCompile(`"line_number"\s+must\s+be\s+set`),
			},
			{
				Config: testAccNotificationAmootSMSResourceConfig(name, `
  use_pattern     = true
  pattern_code_id = 0`),
				ExpectError: regexp.MustCompile(`pattern_code_id`),
			},
			{
				Config: testAccNotificationAmootSMSResourceConfig(name, `
  line_number = ""`),
				ExpectError: regexp.MustCompile(`must\s+not\s+be\s+blank`),
			},
			{
				Config: testAccNotificationAmootSMSResourceConfig(name, `
  line_number = " "`),
				ExpectError: regexp.MustCompile(`must\s+not\s+be\s+blank`),
			},
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_amootsms" "test" {
  name        = %[1]q
  api_token   = "token-1"
  mobiles     = " , "
  line_number = "public"
}
`, name),
				ExpectError: regexp.MustCompile(`at\s+least\s+one\s+recipient\s+number`),
			},
		},
	})
}

// TestAccNotificationAmootSMSResourceEmptyLineNumber covers a line number stored as "" outside of
// Terraform: it must read back as null, so a configuration without one plans no change.
func TestAccNotificationAmootSMSResourceEmptyLineNumber(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationAmootSMSEmptyLine")
	kumaClient := testAccOutOfBandClient(t)
	config := testAccNotificationAmootSMSResourceConfig(name, `
  use_pattern     = true
  pattern_code_id = 1234`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  testAccStoreAmootSMSEmptyLineNumber(t, kumaClient),
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: testAccAmootSMSChecks(
					knownvalue.Bool(true), knownvalue.Int64Exact(1234), knownvalue.Null(), knownvalue.Null(),
				),
			},
		},
	})
}

// testAccStoreAmootSMSEmptyLineNumber stores an empty line number through the out-of-band client.
func testAccStoreAmootSMSEmptyLineNumber(t *testing.T, kumaClient *kuma.Client) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		id, err := testAccStateID(s, testAccAmootSMSResourceName)
		if err != nil {
			return err
		}

		err = kumaClient.Resync(t.Context())
		if err != nil {
			return fmt.Errorf("resync out-of-band client: %w", err)
		}

		var amootSMS notification.AmootSMS

		err = kumaClient.GetNotificationAs(t.Context(), id, &amootSMS)
		if err != nil {
			return fmt.Errorf("get notification %d: %w", id, err)
		}

		amootSMS.LineNumber = new("")

		err = kumaClient.UpdateNotification(t.Context(), amootSMS)
		if err != nil {
			return fmt.Errorf("update notification %d: %w", id, err)
		}

		return nil
	}
}

func testAccNotificationAmootSMSResourceConfig(name string, extra string) string {
	return testAccNotificationAmootSMSResourceConfigSecrets(name, "token-1", "09120000000,09120000001", extra)
}

func testAccNotificationAmootSMSResourceConfigSecrets(name string, token string, mobiles string, extra string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_amootsms" "test" {
  name      = %[1]q
  api_token = %[2]q
  mobiles   = %[3]q%[4]s
}
`, name, token, mobiles, extra)
}
