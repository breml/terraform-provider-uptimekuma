# PEP 0001: Add Slack Notification Support

Extend the Terraform provider to support Slack notification resources of Uptime Kuma. This will involve creating a new resource for Slack notifications and updating the provider schema to include this resource.

## Implementation Details

1. **New Resource**: Create a new resource file `resource_notification_slack.go` in the `internal/provider/` directory.
2. **Schema Definition**: Define the schema for the Slack notification resource. Derive the required and optional attributes from the Uptime Kuma API. The source of Uptime Kuma is located in @.scratch/uptime-kuma/. In particular the following files are relevant: @.scratch/uptime-kuma/src/components/notifications/Slack.vue and @.scratch/uptime-kuma/server/notification-providers/slack.js .
3. **CRUD Operations**: Implement the Create, Read, Update, and Delete operations for the Slack notification resource using the `github.com/breml/go-uptime-kuma-client` library.
4. **Testing**: Write unit tests and acceptance tests for the new resource to ensure it works as expected.
