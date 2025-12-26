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

func TestAccDockerHostResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDockerHost")
	nameUpdated := acctest.RandomWithPrefix("TestDockerHostUpdated")
	daemon := "unix:///var/run/docker.sock"
	daemonUpdated := "tcp://localhost:2375"
	dockerType := "socket"
	dockerTypeUpdated := "tcp"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccDockerHostResourceConfig(name, daemon, dockerType),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_docker_host.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_docker_host.test",
						tfjsonpath.New("docker_daemon"),
						knownvalue.StringExact(daemon),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_docker_host.test",
						tfjsonpath.New("docker_type"),
						knownvalue.StringExact(dockerType),
					),
				},
			},
			{
				Config:             testAccDockerHostResourceConfig(nameUpdated, daemonUpdated, dockerTypeUpdated),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_docker_host.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_docker_host.test",
						tfjsonpath.New("docker_daemon"),
						knownvalue.StringExact(daemonUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_docker_host.test",
						tfjsonpath.New("docker_type"),
						knownvalue.StringExact(dockerTypeUpdated),
					),
				},
			},
		},
	})
}

func testAccDockerHostResourceConfig(name, daemon, dockerType string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_docker_host" "test" {
  name           = %[1]q
  docker_daemon  = %[2]q
  docker_type    = %[3]q
}
`, name, daemon, dockerType)
}

func TestAccDockerHostResourceDelete(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDockerHostDelete")
	daemon := "unix:///var/run/docker.sock"
	dockerType := "socket"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDockerHostResourceConfig(name, daemon, dockerType),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_docker_host.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config:             testAccDockerHostResourceConfigEmpty(),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccDockerHostResourceConfigEmpty() string {
	return providerConfig()
}
