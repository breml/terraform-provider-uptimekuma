data "uptimekuma_monitor_sqlserver" "example_by_name" {
  name = "SQL Server Production Database"
}

data "uptimekuma_monitor_sqlserver" "example_by_id" {
  id = 42
}
