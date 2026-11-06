package provider

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

const (
	applyExistingMonitorAddr      = "uptimekuma_monitor_push.test"
	applyExistingNotificationAddr = "uptimekuma_notification_webhook.test"
)

// TestAccNotificationResourceApplyExisting verifies that apply_existing does not
// cause a perpetual diff and acts as a one-time action: the provider sends
// applyExisting only on create and on a false to true change, so Uptime Kuma
// attaches the notification to every monitor then, but not on other updates.
//
// It uses resource.Test, not resource.ParallelTest, because apply_existing
// attaches the notification to every monitor of the instance.
func TestAccNotificationResourceApplyExisting(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationApplyExisting")
	nameUpdated := acctest.RandomWithPrefix("NotificationApplyExistingUpdated")
	kumaClient := testAccOutOfBandClient(t)

	applyExisting := func(want bool) statecheck.StateCheck {
		return statecheck.ExpectKnownValue(
			applyExistingNotificationAddr,
			tfjsonpath.New("apply_existing"),
			knownvalue.Bool(want),
		)
	}

	// Every change after create must be an in-place update: a replacement would
	// attach the new notification on create and mask a wrong re-attach.
	expectUpdate := resource.ConfigPlanChecks{
		PreApply: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(applyExistingNotificationAddr, plancheck.ResourceActionUpdate),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create attaches the notification to the existing monitor. The
				// implicit post-apply plan check proves that Read keeps true.
				Config:            testAccNotificationApplyExistingConfig(name, true),
				ConfigStateChecks: []statecheck.StateCheck{applyExisting(true)},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationAttached(t, kumaClient),
					testAccDetachMonitorNotifications(t, kumaClient),
				),
			},
			{
				// An unrelated update must not attach the notification again.
				Config:            testAccNotificationApplyExistingConfig(nameUpdated, true),
				ConfigPlanChecks:  expectUpdate,
				ConfigStateChecks: []statecheck.StateCheck{applyExisting(true)},
				Check:             testAccCheckNotificationDetached(t, kumaClient),
			},
			{
				Config:            testAccNotificationApplyExistingConfig(nameUpdated, false),
				ConfigPlanChecks:  expectUpdate,
				ConfigStateChecks: []statecheck.StateCheck{applyExisting(false)},
				Check:             testAccCheckNotificationDetached(t, kumaClient),
			},
			{
				// A false to true change attaches it again.
				Config:            testAccNotificationApplyExistingConfig(nameUpdated, true),
				ConfigPlanChecks:  expectUpdate,
				ConfigStateChecks: []statecheck.StateCheck{applyExisting(true)},
				Check:             testAccCheckNotificationAttached(t, kumaClient),
			},
			{
				// Import leaves apply_existing null and Read falls back to false,
				// while the pre-import state holds true, so the attribute is
				// excluded from the verify and checked explicitly instead.
				ResourceName:            applyExistingNotificationAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_existing"},
				ImportStateCheck:        testAccCheckImportedApplyExistingFalse,
			},
		},
	})
}

// testAccNotificationApplyExistingConfig returns a push monitor and a webhook
// notification created after it. The monitor ignores notification_ids, since
// apply_existing and the test attach and detach the notification out of band.
// It is paused, so it never goes down and fires the example webhook.
func testAccNotificationApplyExistingConfig(name string, applyExisting bool) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_push" "test" {
  name     = %[1]q
  interval = 60
  active   = false

  lifecycle {
    ignore_changes = [notification_ids]
  }
}

resource "uptimekuma_notification_webhook" "test" {
  name                 = %[1]q
  webhook_url          = "https://example.com/webhook"
  webhook_content_type = "json"
  apply_existing       = %[2]t

  depends_on = [uptimekuma_monitor_push.test]
}
`, name, applyExisting)
}

// testAccCheckImportedApplyExistingFalse asserts that an imported notification
// starts with apply_existing false, as the attribute description promises.
func testAccCheckImportedApplyExistingFalse(states []*terraform.InstanceState) error {
	if len(states) != 1 {
		return fmt.Errorf("import produced %d resources, want 1", len(states))
	}

	got := states[0].Attributes["apply_existing"]
	if got != "false" {
		return fmt.Errorf("imported apply_existing is %q, want %q", got, "false")
	}

	return nil
}

// testAccCheckNotificationAttached asserts that Uptime Kuma has the test
// notification attached to the test monitor.
func testAccCheckNotificationAttached(t *testing.T, kumaClient *kuma.Client) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		return testAccNotificationAttachment(t, kumaClient, s, true)
	}
}

// testAccCheckNotificationDetached asserts that Uptime Kuma does not have the
// test notification attached to the test monitor.
func testAccCheckNotificationDetached(t *testing.T, kumaClient *kuma.Client) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		return testAccNotificationAttachment(t, kumaClient, s, false)
	}
}

//revive:disable-next-line:flag-parameter // want is the expected outcome, not a control flag.
func testAccNotificationAttachment(t *testing.T, kumaClient *kuma.Client, s *terraform.State, want bool) error {
	t.Helper()

	monitorID, err := testAccStateID(s, applyExistingMonitorAddr)
	if err != nil {
		return err
	}

	notificationID, err := testAccStateID(s, applyExistingNotificationAddr)
	if err != nil {
		return err
	}

	var push monitor.Push

	err = kumaClient.GetMonitorAs(t.Context(), monitorID, &push)
	if err != nil {
		return fmt.Errorf("get monitor %d: %w", monitorID, err)
	}

	got := slices.Contains(push.NotificationIDs, notificationID)
	if got != want {
		return fmt.Errorf(
			"notification %d attached to monitor %d is %t, want %t (attached: %v)",
			notificationID, monitorID, got, want, push.NotificationIDs,
		)
	}

	return nil
}

// testAccDetachMonitorNotifications detaches all notifications from the test
// monitor out of band, as a user would to silence one monitor.
func testAccDetachMonitorNotifications(t *testing.T, kumaClient *kuma.Client) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		monitorID, err := testAccStateID(s, applyExistingMonitorAddr)
		if err != nil {
			return err
		}

		var push monitor.Push

		err = kumaClient.GetMonitorAs(t.Context(), monitorID, &push)
		if err != nil {
			return fmt.Errorf("get monitor %d: %w", monitorID, err)
		}

		push.NotificationIDs = nil

		err = kumaClient.UpdateMonitor(t.Context(), &push)
		if err != nil {
			return fmt.Errorf("detach notifications from monitor %d: %w", monitorID, err)
		}

		return nil
	}
}
