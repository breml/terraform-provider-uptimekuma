# PEP 0009: Refactor Monitor Base Pattern

Eliminate code duplication across monitor resources by introducing a base pattern similar to the existing notification base pattern. Currently, all monitor resources (HTTP, Ping, DNS, Group) repeat the same ~11 base fields in their models and ~60-70 lines of identical schema definitions.

## Problem Statement

### Current Repetition

Each monitor resource currently duplicates:

1. **Model fields** - Base fields are repeated in every `MonitorXXXResourceModel`:
   ```go
   ID              types.Int64  `tfsdk:"id"`
   Name            types.String `tfsdk:"name"`
   Description     types.String `tfsdk:"description"`
   Parent          types.Int64  `tfsdk:"parent"`
   Interval        types.Int64  `tfsdk:"interval"`
   RetryInterval   types.Int64  `tfsdk:"retry_interval"`
   ResendInterval  types.Int64  `tfsdk:"resend_interval"`
   MaxRetries      types.Int64  `tfsdk:"max_retries"`
   UpsideDown      types.Bool   `tfsdk:"upside_down"`
   Active          types.Bool   `tfsdk:"active"`
   NotificationIDs types.List   `tfsdk:"notification_ids"`
   ```

2. **Schema attributes** - Each `Schema()` method repeats identical definitions with validators, defaults, and plan modifiers for all base fields

3. **Mapping logic** - CRUD operations repeat the same conversion logic between Terraform types and client types for base fields

### Why This Matters

- **Maintenance burden**: Any change to base fields requires updates across 4+ files
- **Inconsistency risk**: Easy to have subtle differences between resources
- **Scalability**: Each new monitor type adds more duplication
- **Contrast with notifications**: Notification resources already follow DRY principles with `NotificationBaseModel` and `withNotificationBaseAttributes()`

## Implementation Details

### 1. Create Monitor Base Model

Create a new file `internal/provider/resource_monitor_base.go` with:

```go
package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MonitorBaseModel contains common fields shared across all monitor types.
// This corresponds to the monitor.Base struct in the go-uptime-kuma-client.
type MonitorBaseModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Parent          types.Int64  `tfsdk:"parent"`
	Interval        types.Int64  `tfsdk:"interval"`
	RetryInterval   types.Int64  `tfsdk:"retry_interval"`
	ResendInterval  types.Int64  `tfsdk:"resend_interval"`
	MaxRetries      types.Int64  `tfsdk:"max_retries"`
	UpsideDown      types.Bool   `tfsdk:"upside_down"`
	Active          types.Bool   `tfsdk:"active"`
	NotificationIDs types.List   `tfsdk:"notification_ids"`
}

// withMonitorBaseAttributes adds common monitor attributes to the provided attribute map.
// This is intended to be used in the Schema() method of monitor resources.
func withMonitorBaseAttributes(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	attrs["id"] = schema.Int64Attribute{
		Computed:            true,
		MarkdownDescription: "Monitor identifier",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	}

	attrs["name"] = schema.StringAttribute{
		MarkdownDescription: "Friendly name",
		Required:            true,
	}

	attrs["description"] = schema.StringAttribute{
		MarkdownDescription: "Description",
		Optional:            true,
	}

	attrs["parent"] = schema.Int64Attribute{
		MarkdownDescription: "Parent monitor ID for hierarchical organization",
		Optional:            true,
	}

	attrs["interval"] = schema.Int64Attribute{
		MarkdownDescription: "Heartbeat interval in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(60),
		Validators: []validator.Int64{
			int64validator.Between(20, 2073600),
		},
	}

	attrs["retry_interval"] = schema.Int64Attribute{
		MarkdownDescription: "Retry interval in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(60),
		Validators: []validator.Int64{
			int64validator.Between(20, 2073600),
		},
	}

	attrs["resend_interval"] = schema.Int64Attribute{
		MarkdownDescription: "Resend interval in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(0),
	}

	attrs["max_retries"] = schema.Int64Attribute{
		MarkdownDescription: "Maximum number of retries",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(3),
		Validators: []validator.Int64{
			int64validator.Between(0, 10),
		},
	}

	attrs["upside_down"] = schema.BoolAttribute{
		MarkdownDescription: "Invert monitor status (treat DOWN as UP and vice versa)",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	}

	attrs["active"] = schema.BoolAttribute{
		MarkdownDescription: "Monitor is active",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(true),
	}

	attrs["notification_ids"] = schema.ListAttribute{
		MarkdownDescription: "List of notification IDs",
		ElementType:         types.Int64Type,
		Optional:            true,
	}

	return attrs
}
```

### 2. Update Existing Monitor Resources

Each monitor resource model will be refactored to embed `MonitorBaseModel`:

**Before** (e.g., `resource_monitor_ping.go`):
```go
type MonitorPingResourceModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Parent          types.Int64  `tfsdk:"parent"`
	Interval        types.Int64  `tfsdk:"interval"`
	RetryInterval   types.Int64  `tfsdk:"retry_interval"`
	ResendInterval  types.Int64  `tfsdk:"resend_interval"`
	MaxRetries      types.Int64  `tfsdk:"max_retries"`
	UpsideDown      types.Bool   `tfsdk:"upside_down"`
	Active          types.Bool   `tfsdk:"active"`
	Hostname        types.String `tfsdk:"hostname"`
	PacketSize      types.Int64  `tfsdk:"packet_size"`
	NotificationIDs types.List   `tfsdk:"notification_ids"`
}
```

**After**:
```go
type MonitorPingResourceModel struct {
	MonitorBaseModel
	Hostname   types.String `tfsdk:"hostname"`
	PacketSize types.Int64  `tfsdk:"packet_size"`
}
```

### 3. Update Schema Definitions

Each resource's `Schema()` method will use `withMonitorBaseAttributes()`:

**Before**:
```go
func (r *MonitorPingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Ping monitor resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{...},
			"name": schema.StringAttribute{...},
			// ... 60+ lines of base attributes ...
			"hostname": schema.StringAttribute{...},
			"packet_size": schema.Int64Attribute{...},
		},
	}
}
```

**After**:
```go
func (r *MonitorPingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Ping monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address to ping",
				Required:            true,
			},
			"packet_size": schema.Int64Attribute{
				MarkdownDescription: "Ping packet size in bytes",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(56),
				Validators: []validator.Int64{
					int64validator.Between(1, 65500),
				},
			},
		}),
	}
}
```

### 4. CRUD Operations Remain Explicit

**Decision**: Keep CRUD operation logic explicit in each resource for clarity and maintainability. While the model and schema are simplified, the Create/Read/Update/Delete methods will continue to explicitly handle all field mappings.

This approach:
- Maintains code readability
- Makes it clear what each monitor type is doing
- Reduces abstraction complexity
- Allows monitor-specific handling without fighting a framework

### 5. Resources to Update

Apply this pattern to all existing monitor resources:
- `resource_monitor_http.go`
- `resource_monitor_ping.go`
- `resource_monitor_dns.go`
- `resource_monitor_group.go`

### 6. Testing

- Ensure all existing tests continue to pass without modification
- Tests should be unaware of the internal refactoring
- Run full test suite: `make test` and `make testacc`

### 7. Documentation

No documentation changes required - this is an internal refactoring with no user-facing changes.

## Benefits

1. **Reduced code duplication**: Eliminates ~250+ lines of repetitive code across 4 resources
2. **Easier maintenance**: Base field changes only need updates in one place
3. **Consistency**: All monitors guaranteed to have identical base field definitions
4. **Scalability**: Future monitor types will require less boilerplate
5. **Alignment**: Follows the pattern already established for notifications
6. **Maintainability**: Struct embedding mirrors the go-uptime-kuma-client design

## Implementation Order

1. Create `resource_monitor_base.go` with `MonitorBaseModel` and `withMonitorBaseAttributes()`
2. Update `resource_monitor_group.go` (simplest, good starting point)
3. Update `resource_monitor_ping.go`
4. Update `resource_monitor_dns.go`
5. Update `resource_monitor_http.go` (most complex)
6. Run `make fmt`, `make lint`, `make test`, `make testacc` after each change
7. Commit with message describing the refactoring

## Future Considerations

- When adding new monitor types, they should use this base pattern from the start
- If additional base fields are identified, add them to `MonitorBaseModel` and `withMonitorBaseAttributes()`
- Consider creating similar helper functions for repeated conversion logic if a clear pattern emerges during future development
