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

func TestAccMonitorDockerResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDockerMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestDockerMonitorUpdated")
	container := "test-container"
	containerUpdated := "updated-container"
	dockerHost := "unix:///var/run/docker.sock"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorDockerResourceConfig(name, container, dockerHost, 60, 60),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("docker_container"),
						knownvalue.StringExact(container),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorDockerResourceConfig(nameUpdated, containerUpdated, dockerHost, 120, 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("docker_container"),
						knownvalue.StringExact(containerUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_docker.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorDockerResourceConfig(
	name string,
	container string,
	dockerHost string,
	interval int64,
	retryInterval int64,
) string {
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
  interval          = %[5]d
  retry_interval    = %[6]d
  active            = true
}
`, name, container, dockerHost, acctest.RandomWithPrefix("TestDockerHost"), interval, retryInterval)
}

func TestAccMonitorDockerResourceWithDescription(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDockerMonitorDesc")
	container := "test-container"
	description := "Test Docker Monitor"
	dockerHost := "unix:///var/run/docker.sock"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorDockerResourceConfigWithDescription(name, container, dockerHost, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_docker.test",
						tfjsonpath.New("docker_container"),
						knownvalue.StringExact(container),
					),
				},
			},
		},
	})
}

func testAccMonitorDockerResourceConfigWithDescription(
	name string,
	container string,
	dockerHost string,
	description string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_docker_host" "test" {
  name           = %[4]q
  docker_daemon  = %[3]q
  docker_type    = "socket"
}

resource "uptimekuma_monitor_docker" "test" {
  name              = %[1]q
  description       = %[5]q
  docker_host_id    = uptimekuma_docker_host.test.id
  docker_container  = %[2]q
  active            = true
}
`, name, container, dockerHost, acctest.RandomWithPrefix("TestDockerHost"), description)
}
