package provider

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// serverSecretsFunc returns what Uptime Kuma stores for each secret attribute of the resource
// with the given ID, keyed by attribute name.
type serverSecretsFunc func(ctx context.Context, kumaClient *kuma.Client, id int64) (map[string]string, error)

// writeOnlyTestCase describes the write-only secrets of one resource for
// testAccWriteOnlyTestCase.
type writeOnlyTestCase struct {
	// addr is the resource address, e.g. "uptimekuma_notification_smtp.test".
	addr string

	// attrs are the secret attributes whose <attr>_wo variant config sets.
	attrs []string

	// config returns a configuration that sets <attr>_wo for every attribute in attrs and
	// <attr>_wo_version to version. The value it sets for an attribute is
	// writeOnlyTestSecret(attr, secret).
	config func(secret string, version string) string

	// serverSecrets reads the stored secrets back from Uptime Kuma.
	serverSecrets serverSecretsFunc

	// singleton marks a resource without a numeric ID, such as uptimekuma_settings;
	// serverSecrets is then called with ID 0.
	singleton bool

	// importOmits are the attributes of attrs whose Read never refreshes the secret, so an
	// import leaves them unset rather than storing the server's value.
	importOmits []string

	// noImport marks a resource without import support, such as uptimekuma_proxy, which skips
	// the import step.
	noImport bool
}

// writeOnlyTestSecret derives a distinct secret per attribute, so that a resource sending one
// attribute's value for another fails the server check.
func writeOnlyTestSecret(attr string, secret string) string {
	return attr + "-" + secret
}

// testAccWriteOnlyTestCase returns a test case that creates the resource with write-only
// secrets, checks they reach the server but not state, that a changed write-only value alone
// plans nothing, and that bumping the version sends it. The full set of transitions, the
// validators are covered once, by TestAccNotificationSMTPResourcePasswordWO. The last step
// imports the resource, which has no <attr>_wo_version to go by and so stores the server's
// value of every secret whose Read refreshes it.
func testAccWriteOnlyTestCase(t *testing.T, tc writeOnlyTestCase) resource.TestCase {
	t.Helper()

	kumaClient := testAccOutOfBandClient(t)

	absentFromState := func(version string) []statecheck.StateCheck {
		checks := make([]statecheck.StateCheck, 0, 3*len(tc.attrs))

		for _, attr := range tc.attrs {
			checks = append(checks,
				statecheck.ExpectKnownValue(tc.addr, tfjsonpath.New(attr+writeOnlySuffix), knownvalue.Null()),
				statecheck.ExpectKnownValue(tc.addr, tfjsonpath.New(attr+writeOnlyVersionSuffix),
					knownvalue.StringExact(version)),
			)
		}

		return checks
	}

	testCase := resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		Steps: []resource.TestStep{
			{
				Config:            tc.config("first", "1"),
				ConfigStateChecks: append(absentFromState("1"), storedSecretsAbsent(tc)...),
				Check:             testAccCheckServerSecrets(t, kumaClient, tc, "first"),
			},
			// A changed write-only value alone is invisible to the plan: nothing is sent.
			{
				Config: tc.config("second", "1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: storedSecretsAbsent(tc),
				Check:             testAccCheckServerSecrets(t, kumaClient, tc, "first"),
			},
			// Bumping the version sends the current write-only value.
			{
				Config: tc.config("second", "2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(tc.addr, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: append(absentFromState("2"), storedSecretsAbsent(tc)...),
				Check:             testAccCheckServerSecrets(t, kumaClient, tc, "second"),
			},
		},
	}

	if !tc.noImport {
		testCase.Steps = append(testCase.Steps, resource.TestStep{
			ResourceName:     tc.addr,
			ImportState:      true,
			ImportStateCheck: testAccCheckImportedSecrets(tc, "second"),
		})
	}

	return testCase
}

// testAccCheckImportedSecrets asserts the state an import produces: every secret attribute
// holds the server's value unless tc.importOmits names it, and no <attr>_wo or
// <attr>_wo_version is set.
func testAccCheckImportedSecrets(tc writeOnlyTestCase, secret string) resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) != 1 {
			return fmt.Errorf("import of %s produced %d resources, want 1", tc.addr, len(states))
		}

		attrs := states[0].Attributes

		for _, attr := range tc.attrs {
			want := writeOnlyTestSecret(attr, secret)
			if slices.Contains(tc.importOmits, attr) {
				want = ""
			}

			if attrs[attr] != want {
				return fmt.Errorf("imported %s of %s is %q, want %q", attr, tc.addr, attrs[attr], want)
			}

			for _, unset := range []string{attr + writeOnlySuffix, attr + writeOnlyVersionSuffix} {
				if attrs[unset] != "" {
					return fmt.Errorf("imported %s of %s is %q, want unset", unset, tc.addr, attrs[unset])
				}
			}
		}

		return nil
	}
}

// storedSecretsAbsent checks that no stored secret attribute carries the server's value. A
// computed attribute with a default holds the empty string rather than null.
func storedSecretsAbsent(tc writeOnlyTestCase) []statecheck.StateCheck {
	checks := make([]statecheck.StateCheck, 0, len(tc.attrs))

	for _, attr := range tc.attrs {
		checks = append(checks, statecheck.ExpectKnownValue(tc.addr, tfjsonpath.New(attr), nullOrEmptyString{}))
	}

	return checks
}

// nullOrEmptyString is a knownvalue.Check for a string attribute that is null or empty.
type nullOrEmptyString struct{}

// CheckValue implements knownvalue.Check.
func (nullOrEmptyString) CheckValue(value any) error {
	if value == nil || value == "" {
		return nil
	}

	return fmt.Errorf("expected null or empty string, got %v", value)
}

// String implements knownvalue.Check.
func (nullOrEmptyString) String() string {
	return "null or empty string"
}

// testAccCheckServerSecrets asserts the secrets Uptime Kuma stores for the resource, which is
// the only place a write-only value can be observed.
func testAccCheckServerSecrets(
	t *testing.T,
	kumaClient *kuma.Client,
	tc writeOnlyTestCase,
	secret string,
) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		var id int64

		if !tc.singleton {
			var err error

			id, err = testAccStateID(s, tc.addr)
			if err != nil {
				return err
			}
		}

		got, err := tc.serverSecrets(t.Context(), kumaClient, id)
		if err != nil {
			return fmt.Errorf("read secrets of %s: %w", tc.addr, err)
		}

		for _, attr := range tc.attrs {
			want := writeOnlyTestSecret(attr, secret)
			if got[attr] != want {
				return fmt.Errorf("server value of %s for %s is %q, want %q", attr, tc.addr, got[attr], want)
			}
		}

		return nil
	}
}

// writeOnlyAttrLines returns HCL lines that set <attr>_wo to writeOnlyTestSecret(attr, secret)
// and <attr>_wo_version to version for every attribute.
func writeOnlyAttrLines(secret string, version string, attrs ...string) string {
	var b strings.Builder

	for _, attr := range attrs {
		fmt.Fprintf(&b, "  %s_wo = %q\n  %s_wo_version = %q\n", attr, writeOnlyTestSecret(attr, secret), attr, version)
	}

	return b.String()
}

// testAccStateID returns the numeric ID of the resource at addr.
func testAccStateID(s *terraform.State, addr string) (int64, error) {
	rs, ok := s.RootModule().Resources[addr]
	if !ok {
		return 0, fmt.Errorf("resource %s not found in state", addr)
	}

	id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse id %q: %w", rs.Primary.ID, err)
	}

	return id, nil
}

// notificationSecrets returns a serverSecretsFunc for a notification of type T. The
// out-of-band client serves notifications from a cache that the server's notificationList
// broadcast refreshes, and nothing waits for that broadcast after a write, so the cache is
// resynced before it is read.
func notificationSecrets[T any](secrets func(n *T) map[string]string) serverSecretsFunc {
	return func(ctx context.Context, kumaClient *kuma.Client, id int64) (map[string]string, error) {
		err := kumaClient.Resync(ctx)
		if err != nil {
			return nil, fmt.Errorf("resync out-of-band client: %w", err)
		}

		var n T

		err = kumaClient.GetNotificationAs(ctx, id, &n)
		if err != nil {
			return nil, fmt.Errorf("get notification %d: %w", id, err)
		}

		return secrets(&n), nil
	}
}

// monitorSecrets returns a serverSecretsFunc for a monitor of type T. GetMonitorAs asks the
// server, so no resync is needed.
func monitorSecrets[T any](secrets func(m *T) map[string]string) serverSecretsFunc {
	return func(ctx context.Context, kumaClient *kuma.Client, id int64) (map[string]string, error) {
		var m T

		err := kumaClient.GetMonitorAs(ctx, id, &m)
		if err != nil {
			return nil, fmt.Errorf("get monitor %d: %w", id, err)
		}

		return secrets(&m), nil
	}
}
