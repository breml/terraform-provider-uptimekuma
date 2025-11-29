resource "uptimekuma_monitor_grpc_keyword" "example" {
  name              = "gRPC Service Monitoring"
  hostname          = "grpc.example.com"
  port              = 50051
  grpc_service_name = "myapp.v1.Health"
  keyword           = "SERVING"
  interval          = 60
  timeout           = 30
  max_retries       = 2
  upside_down       = false
  active            = true
}
