# PEP 0008: Add Ping Monitor Resource

Extend the Terraform provider to support Ping monitor resources of Uptime Kuma. This will involve creating a new resource for Ping monitors and updating the provider schema to include this resource.

## Implementation Details

1. **New Resource**: Create a new resource file `resource_monitor_ping.go` in the `internal/provider/` directory.
2. **Schema Definition**: Define the schema for the Ping monitor resource. Derive the required and optional attributes from the Uptime Kuma API. The source of Uptime Kuma is located in @.scratch/uptime-kuma/. In particular the following files are relevant: @.scratch/uptime-kuma/src/pages/EditMonitor.vue and @.scratch/uptime-kuma/server/model/monitor.js and @.scratch/uptime-kuma/server/server.js .
3. **CRUD Operations**: Implement the Create, Read, Update, and Delete operations for the Ping monitor resource using the `github.com/breml/go-uptime-kuma-client` library.
4. **Testing**: Write unit tests and acceptance tests for the new resource to ensure it works as expected.
5. **Documentation**: Update the provider documentation to include information about the new Ping monitor resource. This will involve adding examples and usage instructions.
