package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	kuma "github.com/breml/go-uptime-kuma-client"
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

// TestAccProxyResourceWriteOnly runs non-parallel like every proxy test, see
// TestAccProxyResource.
func TestAccProxyResourceWriteOnly(t *testing.T) {
	host := acctest.RandomWithPrefix("wo") + ".proxy.example.com"

	resource.Test(t, testAccWriteOnlyTestCase(t, writeOnlyTestCase{
		addr:  "uptimekuma_proxy.test",
		attrs: []string{"password"},
		config: func(secret string, version string) string {
			return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_proxy" "test" {
  host     = %[1]q
  port     = 8080
  protocol = "http"
  auth     = true
  username = "monitor"
  active   = true
%[2]s}
`, host, writeOnlyAttrLines(secret, version, "password"))
		},
		serverSecrets: func(ctx context.Context, kumaClient *kuma.Client, id int64) (map[string]string, error) {
			// GetProxy serves from a cache that nothing waits to refresh after a write.
			err := kumaClient.Resync(ctx)
			if err != nil {
				return nil, fmt.Errorf("resync out-of-band client: %w", err)
			}

			p, err := kumaClient.GetProxy(ctx, id)
			if err != nil {
				return nil, fmt.Errorf("get proxy %d: %w", id, err)
			}

			return map[string]string{"password": p.Password}, nil
		},
	}))
}
