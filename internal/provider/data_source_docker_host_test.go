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

func TestAccDockerHostDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDockerHost")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDockerHostDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_docker_host.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccDockerHostDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_docker_host.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccDockerHostDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_docker_host" "test" {
  name        = %[1]q
  docker_type = "socket"
  docker_daemon = "/var/run/docker.sock"
}

data "uptimekuma_docker_host" "test" {
  name = uptimekuma_docker_host.test.name
}
`, name)
}

func testAccDockerHostDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_docker_host" "test" {
  name        = %[1]q
  docker_type = "socket"
  docker_daemon = "/var/run/docker.sock"
}

data "uptimekuma_docker_host" "test" {
  id = uptimekuma_docker_host.test.id
}
`, name)
}
