package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/breml/go-uptime-kuma-client/monitor"
)

func TestIPFamilyToAPI(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value types.String
		want  monitor.IPFamily
	}{
		"auto":  {value: types.StringValue("auto"), want: monitor.IPFamilyAuto},
		"empty": {value: types.StringValue(""), want: monitor.IPFamilyAuto},
		"null":  {value: types.StringNull(), want: monitor.IPFamilyAuto},
		"ipv4":  {value: types.StringValue("ipv4"), want: monitor.IPFamilyIPv4},
		"ipv6":  {value: types.StringValue("ipv6"), want: monitor.IPFamilyIPv6},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := ipFamilyToAPI(tc.value)
			if got != tc.want {
				t.Errorf("ipFamilyToAPI(%s) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestIPFamilyFromAPI(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		apiValue monitor.IPFamily
		prior    types.String
		fallback string
		want     types.String
	}{
		"auto without prior value": {
			apiValue: monitor.IPFamilyAuto,
			prior:    types.StringNull(),
			fallback: "auto",
			want:     types.StringValue("auto"),
		},
		"auto without prior value and empty fallback": {
			apiValue: monitor.IPFamilyAuto,
			prior:    types.StringNull(),
			fallback: "",
			want:     types.StringValue(""),
		},
		"auto with unknown prior value": {
			apiValue: monitor.IPFamilyAuto,
			prior:    types.StringUnknown(),
			fallback: "auto",
			want:     types.StringValue("auto"),
		},
		"auto keeps prior auto": {
			apiValue: monitor.IPFamilyAuto,
			prior:    types.StringValue("auto"),
			fallback: "",
			want:     types.StringValue("auto"),
		},
		"auto keeps prior empty string": {
			apiValue: monitor.IPFamilyAuto,
			prior:    types.StringValue(""),
			fallback: "auto",
			want:     types.StringValue(""),
		},
		"auto replaces prior ipv4 changed outside Terraform": {
			apiValue: monitor.IPFamilyAuto,
			prior:    types.StringValue("ipv4"),
			fallback: "auto",
			want:     types.StringValue("auto"),
		},
		"auto replaces prior ipv6 with empty fallback": {
			apiValue: monitor.IPFamilyAuto,
			prior:    types.StringValue("ipv6"),
			fallback: "",
			want:     types.StringValue(""),
		},
		"ipv6 replaces prior empty string": {
			apiValue: monitor.IPFamilyIPv6,
			prior:    types.StringValue(""),
			fallback: "",
			want:     types.StringValue("ipv6"),
		},
		"ipv4 without prior value": {
			apiValue: monitor.IPFamilyIPv4,
			prior:    types.StringNull(),
			fallback: "auto",
			want:     types.StringValue("ipv4"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := ipFamilyFromAPI(tc.apiValue, tc.prior, tc.fallback)
			if !got.Equal(tc.want) {
				t.Errorf(
					"ipFamilyFromAPI(%q, %s, %q) = %s, want %s",
					tc.apiValue, tc.prior, tc.fallback, got, tc.want,
				)
			}
		})
	}
}

func TestAccMonitorHTTPResourceIPFamily(t *testing.T) {
	testAccMonitorIPFamily(t, "uptimekuma_monitor_http", `
  url = "https://httpbin.org/status/200"
`)
}

func TestAccMonitorHTTPKeywordResourceIPFamily(t *testing.T) {
	testAccMonitorIPFamily(t, "uptimekuma_monitor_http_keyword", `
  url     = "https://httpbin.org/html"
  keyword = "Herman"
`)
}

func TestAccMonitorHTTPJSONQueryResourceIPFamily(t *testing.T) {
	testAccMonitorIPFamily(t, "uptimekuma_monitor_http_json_query", `
  url            = "https://httpbin.org/json"
  json_path      = "slideshow.author"
  expected_value = "Yours Truly"
`)
}

// testAccMonitorIPFamily walks ip_family of the given monitor resource type
// through ipv4, ipv6 and the omitted attribute (`auto`), importing it on ipv6
// and on auto, and checks that the empty string is rejected. The data source
// is checked both by id and by name.
func testAccMonitorIPFamily(t *testing.T, resourceType string, body string) {
	t.Helper()

	name := acctest.RandomWithPrefix("TestIPFamily")
	address := resourceType + ".test"

	check := func(value string) []statecheck.StateCheck {
		return []statecheck.StateCheck{
			statecheck.ExpectKnownValue(address, tfjsonpath.New("ip_family"), knownvalue.StringExact(value)),
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// The data sources are read after the monitor is created, because
				// they reference its attributes.
				Config: testAccMonitorIPFamilyConfig(resourceType, name, body, `ip_family = "ipv4"`) +
					fmt.Sprintf(`
data %[1]q "by_id" {
  id = %[2]s.id
}

data %[1]q "by_name" {
  name = %[2]s.name
}
`, resourceType, address),
				ConfigStateChecks: append(
					check("ipv4"),
					statecheck.ExpectKnownValue(
						"data."+resourceType+".by_id", tfjsonpath.New("ip_family"), knownvalue.StringExact("ipv4"),
					),
					statecheck.ExpectKnownValue(
						"data."+resourceType+".by_name", tfjsonpath.New("ip_family"), knownvalue.StringExact("ipv4"),
					),
				),
			},
			{
				Config:            testAccMonitorIPFamilyConfig(resourceType, name, body, `ip_family = "ipv6"`),
				ConfigStateChecks: check("ipv6"),
			},
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccMonitorIPFamilyConfig(resourceType, name, body, `ip_family = ""`),
				ExpectError: regexp.MustCompile(`Attribute ip_family value must be one of`),
			},
			{
				Config:            testAccMonitorIPFamilyConfig(resourceType, name, body, ""),
				ConfigStateChecks: check("auto"),
			},
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorIPFamilyConfig(resourceType string, name string, body string, ipFamily string) string {
	return providerConfig() + fmt.Sprintf(`
resource %[1]q "test" {
  name = %[2]q
%[3]s
  %[4]s
}
`, resourceType, name, body, ipFamily)
}
