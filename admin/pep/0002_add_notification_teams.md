# PEP 0001: Add Teams Notification Support

Extend the Terraform provider to support Teams notification resources of Uptime Kuma. This will involve creating a new resource for Teams notifications and updating the provider schema to include this resource.

## Implementation Details

1. **New Resource**: Create a new resource file `resource_notification_teams.go` in the `internal/provider/` directory.
2. **Schema Definition**: Define the schema for the Teams notification resource. Derive the required and optional attributes from the Uptime Kuma API. The source of Uptime Kuma is located in @.scratch/uptime-kuma/. In particular the following files are relevant: @.scratch/uptime-kuma/src/components/notifications/Teams.vue and @.scratch/uptime-kuma/server/notification-providers/teams.js .
3. **CRUD Operations**: Implement the Create, Read, Update, and Delete operations for the Teams notification resource using the `github.com/breml/go-uptime-kuma-client` library.
4. **Testing**: Write unit tests and acceptance tests for the new resource to ensure it works as expected.
