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

func TestAccMonitorGRPCKeywordDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGRPCKeywordMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGRPCKeywordDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
			{
				Config: testAccMonitorGRPCKeywordDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
		},
	})
}

func testAccMonitorGRPCKeywordDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_grpc_keyword" "test" {
  name              = %[1]q
  grpc_url          = "grpc.example.com:50051"
  keyword           = "success"
  grpc_service_name = "example.Service"
  grpc_method       = "Check"
  grpc_protobuf     = "syntax = \"proto3\";"
}

data "uptimekuma_monitor_grpc_keyword" "test" {
  name = uptimekuma_monitor_grpc_keyword.test.name
}
`, name)
}

func testAccMonitorGRPCKeywordDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_grpc_keyword" "test" {
  name              = %[1]q
  grpc_url          = "grpc.example.com:50051"
  keyword           = "success"
  grpc_service_name = "example.Service"
  grpc_method       = "Check"
  grpc_protobuf     = "syntax = \"proto3\";"
}

data "uptimekuma_monitor_grpc_keyword" "test" {
  id = uptimekuma_monitor_grpc_keyword.test.id
}
`, name)
}
