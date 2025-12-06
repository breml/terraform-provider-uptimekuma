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

func TestAccMonitorGrpcKeywordResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGrpcKeywordMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestGrpcKeywordMonitorUpdated")
	grpcURL := "localhost:50051"
	serviceName := "Health"
	method := "Check"
	keyword := "SERVING"
	keywordUpdated := "OK"
	protobuf := `syntax = "proto3";

package grpc.health.v1;

service Health {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
}

message HealthCheckRequest {
  string service = 1;
}

message HealthCheckResponse {
  enum ServingStatus {
    UNKNOWN = 0;
    SERVING = 1;
    NOT_SERVING = 2;
  }
  ServingStatus status = 1;
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorGrpcKeywordResourceConfig(name, grpcURL, serviceName, method, protobuf, keyword, false, 60),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("grpc_url"), knownvalue.StringExact(grpcURL)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("grpc_service_name"), knownvalue.StringExact(serviceName)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("grpc_method"), knownvalue.StringExact(method)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("keyword"), knownvalue.StringExact(keyword)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("invert_keyword"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("grpc_enable_tls"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccMonitorGrpcKeywordResourceConfig(nameUpdated, grpcURL, serviceName, method, protobuf, keywordUpdated, false, 120),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("grpc_url"), knownvalue.StringExact(grpcURL)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("keyword"), knownvalue.StringExact(keywordUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(120)),
				},
			},
		},
	})
}

func testAccMonitorGrpcKeywordResourceConfig(name, grpcURL, serviceName, method, protobuf, keyword string, invertKeyword bool, interval int64) string { //nolint:unparam
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_grpc_keyword" "test" {
  name              = %[1]q
  grpc_url          = %[2]q
  grpc_service_name = %[3]q
  grpc_method       = %[4]q
  grpc_protobuf     = %[5]q
  keyword           = %[6]q
  invert_keyword    = %[7]t
  interval          = %[8]d
  active            = true
}
`, name, grpcURL, serviceName, method, protobuf, keyword, invertKeyword, interval)
}

func TestAccMonitorGrpcKeywordResourceWithInvert(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGrpcKeywordMonitorInvert")
	grpcURL := "localhost:50051"
	serviceName := "Health"
	method := "Check"
	keyword := "NOT_SERVING"
	protobuf := `syntax = "proto3";

package grpc.health.v1;

service Health {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
}

message HealthCheckRequest {
  string service = 1;
}

message HealthCheckResponse {
  enum ServingStatus {
    UNKNOWN = 0;
    SERVING = 1;
    NOT_SERVING = 2;
  }
  ServingStatus status = 1;
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGrpcKeywordResourceConfig(name, grpcURL, serviceName, method, protobuf, keyword, true, 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("keyword"), knownvalue.StringExact(keyword)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("invert_keyword"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestAccMonitorGrpcKeywordResourceWithTLS(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGrpcKeywordMonitorTLS")
	grpcURL := "example.com:443"
	serviceName := "Health"
	method := "Check"
	keyword := "SERVING"
	protobuf := `syntax = "proto3";

package grpc.health.v1;

service Health {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
}

message HealthCheckRequest {
  string service = 1;
}

message HealthCheckResponse {
  enum ServingStatus {
    UNKNOWN = 0;
    SERVING = 1;
    NOT_SERVING = 2;
  }
  ServingStatus status = 1;
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGrpcKeywordResourceConfigWithTLS(name, grpcURL, serviceName, method, protobuf, keyword),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("grpc_url"), knownvalue.StringExact(grpcURL)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("grpc_enable_tls"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func testAccMonitorGrpcKeywordResourceConfigWithTLS(name, grpcURL, serviceName, method, protobuf, keyword string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_grpc_keyword" "test" {
  name              = %[1]q
  grpc_url          = %[2]q
  grpc_service_name = %[3]q
  grpc_method       = %[4]q
  grpc_protobuf     = %[5]q
  keyword           = %[6]q
  grpc_enable_tls   = true
}
`, name, grpcURL, serviceName, method, protobuf, keyword)
}

func TestAccMonitorGrpcKeywordResourceWithBody(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGrpcKeywordMonitorBody")
	grpcURL := "localhost:50051"
	serviceName := "Health"
	method := "Check"
	keyword := "SERVING"
	grpcBody := `{"service":""}`
	protobuf := `syntax = "proto3";

package grpc.health.v1;

service Health {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
}

message HealthCheckRequest {
  string service = 1;
}

message HealthCheckResponse {
  enum ServingStatus {
    UNKNOWN = 0;
    SERVING = 1;
    NOT_SERVING = 2;
  }
  ServingStatus status = 1;
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGrpcKeywordResourceConfigWithBody(name, grpcURL, serviceName, method, protobuf, keyword, grpcBody),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_grpc_keyword.test", tfjsonpath.New("grpc_body"), knownvalue.StringExact(grpcBody)),
				},
			},
		},
	})
}

func testAccMonitorGrpcKeywordResourceConfigWithBody(name, grpcURL, serviceName, method, protobuf, keyword, grpcBody string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_grpc_keyword" "test" {
  name              = %[1]q
  grpc_url          = %[2]q
  grpc_service_name = %[3]q
  grpc_method       = %[4]q
  grpc_protobuf     = %[5]q
  keyword           = %[6]q
  grpc_body         = %[7]q
}
`, name, grpcURL, serviceName, method, protobuf, keyword, grpcBody)
}

func TestAccMonitorGrpcKeywordResourceImport(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGrpcKeywordMonitorImport")
	grpcURL := "localhost:50051"
	serviceName := "Health"
	method := "Check"
	keyword := "SERVING"
	protobuf := `syntax = "proto3";

package grpc.health.v1;

service Health {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
}

message HealthCheckRequest {
  string service = 1;
}

message HealthCheckResponse {
  enum ServingStatus {
    UNKNOWN = 0;
    SERVING = 1;
    NOT_SERVING = 2;
  }
  ServingStatus status = 1;
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGrpcKeywordResourceConfig(name, grpcURL, serviceName, method, protobuf, keyword, false, 60),
			},
			{
				ResourceName:      "uptimekuma_monitor_grpc_keyword.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
