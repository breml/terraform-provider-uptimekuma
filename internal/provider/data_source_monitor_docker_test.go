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

func TestAccMonitorDockerDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDockerMonitor")
	container := "test-container"
	dockerHost := "unix:///var/run/docker.sock"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorDockerDataSourceConfig(name, container, dockerHost),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_docker.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_docker.test",
						tfjsonpath.New("docker_container"),
						knownvalue.StringExact(container),
					),
				},
			},
		},
	})
}

func testAccMonitorDockerDataSourceConfig(name string, container string, dockerHost string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_docker_host" "test" {
  name           = %[4]q
  docker_daemon  = %[3]q
  docker_type    = "socket"
}

resource "uptimekuma_monitor_docker" "test" {
  name              = %[1]q
  docker_host_id    = uptimekuma_docker_host.test.id
  docker_container  = %[2]q
  active            = true
}

data "uptimekuma_monitor_docker" "test" {
  name = uptimekuma_monitor_docker.test.name
}
`, name, container, dockerHost, acctest.RandomWithPrefix("TestDockerHost"))
}
