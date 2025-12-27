package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccProxyResource(t *testing.T) {
	host := "proxy.example.com"
	hostUpdated := "proxy-updated.example.com"
	port := "8080"
	portUpdated := "3128"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccProxyResourceConfig(host, port, "http"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("host"),
						knownvalue.StringExact(host),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(8080),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("protocol"),
						knownvalue.StringExact("http"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("auth"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				Config:             testAccProxyResourceConfig(hostUpdated, portUpdated, "https"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("host"),
						knownvalue.StringExact(hostUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(3128),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("protocol"),
						knownvalue.StringExact("https"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func TestAccProxyResourceWithAuth(t *testing.T) {
	host := "proxy.example.com"
	port := "8080"
	username := "testuser"
	password := "testpass"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccProxyResourceConfigWithAuth(host, port, "socks5", username, password),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("host"),
						knownvalue.StringExact(host),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(8080),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("protocol"),
						knownvalue.StringExact("socks5"),
					),
					statecheck.ExpectKnownValue("uptimekuma_proxy.test", tfjsonpath.New("auth"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
				},
			},
		},
	})
}

func TestAccProxyResourceAuthToggle(t *testing.T) {
	host := "proxy.example.com"
	port := "8080"
	username := "testuser"
	password := "testpass"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccProxyResourceConfigWithAuth(host, port, "http", username, password),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_proxy.test", tfjsonpath.New("auth"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
				},
			},
			{
				Config:             testAccProxyResourceConfig(host, port, "http"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("auth"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue("uptimekuma_proxy.test", tfjsonpath.New("username"), knownvalue.Null()),
					statecheck.ExpectKnownValue("uptimekuma_proxy.test", tfjsonpath.New("password"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccProxyResourceWithDefault(t *testing.T) {
	host := "proxy.example.com"
	port := "8080"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccProxyResourceConfigWithDefault(host, port, "http"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("host"),
						knownvalue.StringExact(host),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("default"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func TestAccProxyResourceDelete(t *testing.T) {
	host := "proxy.example.com"
	port := "8080"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyResourceConfig(host, port, "http"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_proxy.test",
						tfjsonpath.New("host"),
						knownvalue.StringExact(host),
					),
				},
			},
			{
				Config:             testAccProxyResourceConfigEmpty(),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccProxyResourceConfig(host string, port string, protocol string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_proxy" "test" {
  host     = %[1]q
  port     = %[2]s
  protocol = %[3]q
  active   = true
}
`, host, port, protocol)
}

func testAccProxyResourceConfigWithAuth(
	host string,
	port string,
	protocol string,
	username string,
	password string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_proxy" "test" {
  host     = %[1]q
  port     = %[2]s
  protocol = %[3]q
  auth     = true
  username = %[4]q
  password = %[5]q
  active   = true
}
`, host, port, protocol, username, password)
}

func testAccProxyResourceConfigWithDefault(host string, port string, protocol string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_proxy" "test" {
  host     = %[1]q
  port     = %[2]s
  protocol = %[3]q
  active   = true
  default  = true
}
`, host, port, protocol)
}

func testAccProxyResourceConfigEmpty() string {
	return providerConfig()
}
