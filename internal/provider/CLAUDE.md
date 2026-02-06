# internal/provider Package

This package implements the Terraform provider for Uptime Kuma, including all resources, data sources, and provider configuration.

## Package Overview

- **Package**: `internal/provider`
- **Files**: 339 files total
- **Resources**: 85+ resource types
- **Data Sources**: 81+ data source types (mirroring resources)
- **Size**: One of the largest packages in the codebase

## Provider Structure

### Main Provider ([provider.go:24-30](provider.go))

```go
type UptimeKumaProvider struct {
    version string  // "dev", "test", or release version
}
```

**Type Name**: `uptimekuma` - all resources prefixed with `uptimekuma_`

**Configuration Model**:

```go
type UptimeKumaProviderModel struct {
    Endpoint types.String `tfsdk:"endpoint"`  // Required: Uptime Kuma server URL
    Username types.String `tfsdk:"username"`  // Optional: Login username
    Password types.String `tfsdk:"password"`  // Optional: Login password (sensitive)
}
```

**Environment Variables**:

- `UPTIMEKUMA_ENDPOINT` - Provider endpoint (e.g., `https://uptime-kuma.example.com`)
- `UPTIMEKUMA_USERNAME` - Login username
- `UPTIMEKUMA_PASSWORD` - Login password
- `SOCKETIO_LOG_LEVEL` - Socket.IO client logging level (for debugging)

**Configuration Validation**:

- Endpoint: Always required
- Credentials: Both or neither (if username provided, password required; if password provided, username required)
- Precedence: Terraform config > environment variables

### Provider Configuration ([provider.go:70-143](provider.go))

The `Configure()` method:

1. Reads configuration from Terraform
2. Applies environment variable defaults
3. Validates configuration (endpoint required, credential pairing)
4. Creates client using [../client/client.go](../client/client.go)
5. Passes client to resources via `resp.ResourceData`

**Critical Context Detail**:

```go
kumaCtx := context.Background()  // NOT ctx from Terraform
```

Terraform's context cancels after `Configure()` completes. Socket.IO connection must outlive this method, so
`context.Background()` is used.

## Resource Categories

### Monitor Resources (18 types)

#### HTTP-Based Monitors

Share common HTTP configuration via `MonitorHTTPBaseModel`:

- `uptimekuma_monitor_http` - Basic HTTP/HTTPS monitoring
- `uptimekuma_monitor_http_keyword` - HTTP with keyword presence checking
- `uptimekuma_monitor_http_json_query` - HTTP with JSON response validation
- `uptimekuma_monitor_grpc_keyword` - gRPC with keyword monitoring

#### Network Monitors

- `uptimekuma_monitor_ping` - ICMP ping monitoring
- `uptimekuma_monitor_dns` - DNS resolution monitoring
- `uptimekuma_monitor_tcp_port` - TCP port connectivity
- `uptimekuma_monitor_snmp` - SNMP device monitoring

#### Database Monitors

- `uptimekuma_monitor_postgres` - PostgreSQL database monitoring
- `uptimekuma_monitor_mysql` - MySQL database monitoring
- `uptimekuma_monitor_mongodb` - MongoDB monitoring
- `uptimekuma_monitor_redis` - Redis monitoring
- `uptimekuma_monitor_sqlserver` - SQL Server monitoring

#### IoT/Protocol Monitors

- `uptimekuma_monitor_mqtt` - MQTT protocol monitoring
- `uptimekuma_monitor_push` - Push-based monitoring (for integrations)

#### Specialized Monitors

- `uptimekuma_monitor_real_browser` - Real browser automation (Puppeteer/Playwright)
- `uptimekuma_monitor_docker` - Docker container health monitoring
- `uptimekuma_monitor_steam` - Steam game server monitoring

#### Organization

- `uptimekuma_monitor_group` - Monitor groups for hierarchical organization

### Notification Resources (51 types)

All notification types follow the same pattern with type-specific fields:

#### Webhook

- `uptimekuma_notification_webhook` - Generic webhook notifications

#### Chat Platforms

- `uptimekuma_notification_slack` - Slack
- `uptimekuma_notification_teams` - Microsoft Teams
- `uptimekuma_notification_discord` - Discord
- `uptimekuma_notification_mattermost` - Mattermost
- `uptimekuma_notification_rocketchat` - Rocket.Chat
- `uptimekuma_notification_googlechat` - Google Chat
- `uptimekuma_notification_matrix` - Matrix
- `uptimekuma_notification_wecom` - WeCom
- `uptimekuma_notification_kook` - KOOK (开黑啦)

#### Push Services

- `uptimekuma_notification_pushover` - Pushover
- `uptimekuma_notification_pushbullet` - Pushbullet
- `uptimekuma_notification_pushplus` - PushPlus
- `uptimekuma_notification_pushy` - Pushy
- `uptimekuma_notification_pushdeer` - PushDeer
- `uptimekuma_notification_bark` - Bark
- `uptimekuma_notification_signal` - Signal
- `uptimekuma_notification_telegram` - Telegram
- `uptimekuma_notification_line` - LINE
- `uptimekuma_notification_linenotify` - LINE Notify
- `uptimekuma_notification_twilio` - Twilio

#### Cloud/Self-Hosted

- `uptimekuma_notification_gotify` - Gotify
- `uptimekuma_notification_apprise` - Apprise (multi-protocol)
- `uptimekuma_notification_ntfy` - ntfy.sh
- `uptimekuma_notification_home_assistant` - Home Assistant
- `uptimekuma_notification_grafana_oncall` - Grafana OnCall

#### Enterprise Alerting

- `uptimekuma_notification_pagerduty` - PagerDuty
- `uptimekuma_notification_opsgenie` - Opsgenie
- `uptimekuma_notification_pagertree` - PagerTree
- `uptimekuma_notification_splunk` - Splunk
- `uptimekuma_notification_flashduty` - FlashDuty
- `uptimekuma_notification_alerta` - Alerta
- `uptimekuma_notification_alertnow` - AlertNow

#### SMS Services

- `uptimekuma_notification_aliyunsms` - Alibaba Cloud SMS
- `uptimekuma_notification_cellsynt` - Cellsynt
- `uptimekuma_notification_clicksendsms` - ClickSend SMS
- `uptimekuma_notification_freemobile` - Free Mobile
- `uptimekuma_notification_callmebot` - CallMeBot
- `uptimekuma_notification_octopush` - Octopush
- `uptimekuma_notification_46elks` - 46elks

#### Regional Platforms

- `uptimekuma_notification_feishu` - Feishu (飞书)
- `uptimekuma_notification_dingding` - DingTalk (钉钉)
- `uptimekuma_notification_serverchan` - ServerChan (Server酱)

#### Specialized

- `uptimekuma_notification_threema` - Threema
- `uptimekuma_notification_nostr` - Nostr protocol
- `uptimekuma_notification_waha` - WAHA (WhatsApp API)
- `uptimekuma_notification_brevo` - Brevo (Sendinblue)
- `uptimekuma_notification_bitrix24` - Bitrix24
- `uptimekuma_notification_evolution` - Evolution API
- `uptimekuma_notification_nextcloud_talk` - Nextcloud Talk
- `uptimekuma_notification_lunasea` - LunaSea
- `uptimekuma_notification_sendgrid` - SendGrid
- `uptimekuma_notification_stackfield` - Stackfield

#### Email

- `uptimekuma_notification_smtp` - SMTP email notifications

#### Generic

- `uptimekuma_notification` - Generic notification (deprecated, use specific types)

### Status Page Resources

- `uptimekuma_status_page` - Public status pages with monitor groups
- `uptimekuma_status_page_incident` - Status page incidents

### Maintenance Resources

- `uptimekuma_maintenance` - Scheduled maintenance windows
- `uptimekuma_maintenance_monitors` - Link monitors to maintenance windows
- `uptimekuma_maintenance_status_pages` - Link status pages to maintenance windows

### Infrastructure Resources

- `uptimekuma_tag` - Tags for monitor organization
- `uptimekuma_proxy` - Proxy configuration for routing
- `uptimekuma_docker_host` - Docker host integration

### Data Sources

Each resource has a corresponding data source (81 total) that allows querying existing resources:

- By ID: `data.uptimekuma_monitor_http.example.id = 123`
- By name: `data.uptimekuma_monitor_http.example.name = "Production API"`

## Base Models and Patterns

### MonitorBaseModel ([resource_monitor_base.go:29-44](resource_monitor_base.go))

Common fields for ALL monitor types:

```go
type MonitorBaseModel struct {
    ID              types.Int64  // Computed
    Name            types.String // Required
    Description     types.String // Optional
    Parent          types.Int64  // Parent monitor group ID (for hierarchical organization)
    Interval        types.Int64  // Check interval in seconds (default: 60)
    RetryInterval   types.Int64  // Retry interval when failing (default: 60)
    ResendInterval  types.Int64  // Notification resend interval (default: 0)
    MaxRetries      types.Int64  // Max retries before marking down (default: 3)
    UpsideDown      types.Bool   // Invert status logic (down=up, up=down) (default: false)
    Active          types.Bool   // Monitor actively checking (default: true)
    NotificationIDs types.List   // List of notification channel IDs to alert
    Tags            types.List   // List of tags (MonitorTagModel objects)
}
```

**Helper Function**: `withMonitorBaseAttributes(attrs map[string]schema.Attribute)` ([resource_monitor_base.go:48](resource_monitor_base.go))

Adds all base attributes to any monitor schema. Usage:

```go
func (r *MonitorHTTPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
            "url": schema.StringAttribute{...},
            // ... type-specific fields
        }),
    }
}
```

**Tag Model**:

```go
type MonitorTagModel struct {
    TagID types.Int64  `tfsdk:"tag_id"`  // Tag identifier
    Value types.String `tfsdk:"value"`   // Optional tag value
}
```

### MonitorHTTPBaseModel ([resource_monitor_http_base.go](resource_monitor_http_base.go))

Shared HTTP configuration for HTTP/keyword/JSON query/gRPC monitors:

```go
type MonitorHTTPBaseModel struct {
    URL                 types.String  // Required: Target URL
    Timeout             types.Int64   // Request timeout (default: 48s)
    Method              types.String  // HTTP method (default: GET)
    ExpiryNotification  types.Bool    // Alert on SSL cert expiry
    IgnoreTLS           types.Bool    // Skip TLS verification
    MaxRedirects        types.Int64   // Max redirects to follow (default: 10)
    AcceptedStatusCodes types.List    // Accepted HTTP status codes (default: ["200-299"])
    ProxyID             types.Int64   // Proxy to use for requests
    HTTPBodyEncoding    types.String  // Body encoding (default: "json")
    Body                types.String  // Request body
    Headers             types.String  // Request headers (JSON format)
    AuthMethod          types.String  // "", "basic", "ntlm", "mtls", "oauth2-cc"

    // Basic Auth
    BasicAuthUser       types.String
    BasicAuthPass       types.String

    // NTLM Auth
    AuthDomain          types.String
    AuthWorkstation     types.String

    // mTLS
    TLSCert             types.String  // Client certificate
    TLSKey              types.String  // Client private key
    TLSCa               types.String  // CA certificate

    // OAuth2 Client Credentials
    OAuthAuthMethod     types.String  // "client_secret_basic" or "client_secret_post"
    OAuthTokenURL       types.String
    OAuthClientID       types.String
    OAuthClientSecret   types.String
    OAuthScopes         types.String

    CacheBust           types.Bool    // Add cache-busting query param
}
```

**Helper Function**: `withHTTPMonitorBaseAttributes(attrs map[string]schema.Attribute)`

Each HTTP attribute has a dedicated getter:

- `httpURLAttribute()`, `httpTimeoutAttribute()`, `httpMethodAttribute()`, etc.
- Allows fine-grained control over individual attribute configuration

### NotificationBaseModel ([resource_notification_base.go:12-18](resource_notification_base.go))

Common fields for ALL notification types:

```go
type NotificationBaseModel struct {
    ID            types.Int64  // Computed
    Name          types.String // Required: Notification name
    IsActive      types.Bool   // Active status (default: true)
    IsDefault     types.Bool   // Default notification for new monitors (default: false)
    ApplyExisting types.Bool   // Apply to existing monitors retroactively (default: false)
}
```

**Helper Function**: `withNotificationBaseAttributes(attrs map[string]schema.Attribute)` ([resource_notification_base.go:22](resource_notification_base.go))

**Usage Pattern**:

```go
type NotificationSlackResourceModel struct {
    NotificationBaseModel  // Embedded
    WebhookURL    types.String
    Username      types.String
    IconEmoji     types.String
    Channel       types.String
    // ... type-specific fields
}

func (r *NotificationSlackResource) Schema(...) {
    resp.Schema = schema.Schema{
        Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
            "webhook_url": schema.StringAttribute{Required: true},
            // ... type-specific fields
        }),
    }
}
```

### Status Page Models ([resource_status_page.go](resource_status_page.go))

Status pages support nested groups and monitors:

```go
type StatusPageResourceModel struct {
    ID                    types.Int64
    Slug                  types.String  // Immutable (RequiresReplace)
    Title                 types.String
    Description           types.String
    Theme                 types.String
    PublishedAt           types.String
    ShowTags              types.Bool
    DomainNames           types.List
    CustomCSS             types.String
    FooterText            types.String
    ShowPoweredBy         types.Bool
    GoogleAnalyticsID     types.String
    PublicGroupList       types.List    // List of PublicGroupModel
}

type PublicGroupModel struct {
    ID          types.Int64
    Name        types.String
    Weight      types.Int64   // Display order
    MonitorList types.List    // List of PublicMonitorModel
}

type PublicMonitorModel struct {
    ID      types.Int64
    SendURL types.Bool        // Show URL in status page
}
```

**Important**: Status pages were subject to perpetual diff bug (fixed in v0.1.6).
See [Status Page Helpers](#status-page-helpers) section below.

## Helper Functions

### Monitor Helpers

[resource_monitor_helpers.go](resource_monitor_helpers.go)

```go
// Convert Terraform types.String to Go *string (nil if null/unknown)
func strToPtr(s types.String) *string

// Convert Go *string to Terraform types.String (null if nil)
func ptrToTypes(s *string) types.String
```

**Usage**: Converting between Terraform's type system and Go's pointer-based optionals.

### Tag Handling

Tags are added via separate API calls after monitor creation (not part of monitor creation request).

**Create Pattern** ([resource_monitor_base.go](resource_monitor_base.go)):

```go
func handleMonitorTagsCreate(ctx context.Context, client *kuma.Client, monitorID int64, tags []MonitorTagModel, diags *diag.Diagnostics)
```

Iterates through tags and calls `client.AddMonitorTag(ctx, tagID, monitorID, value)` for each.

**Read Pattern**:

```go
func handleMonitorTagsRead(ctx context.Context, monitorTags []tag.MonitorTag, diags *diag.Diagnostics) types.List
```

Converts API tag list to Terraform `types.List` of `MonitorTagModel`.

### Data Source Helpers

[datasource_monitor_helpers.go](datasource_monitor_helpers.go)

```go
// Find monitor by name and type
func findMonitorByName(ctx context.Context, client *kuma.Client, name string, monitorType string, diags *diag.Diagnostics) (*kuma.Monitor, error)

// Validate either ID or name is provided (not both, not neither)
func validateMonitorDataSourceInput(data *DataSourceModel, diags *diag.Diagnostics) bool
```

Similar helpers exist for notifications:

- `findNotificationByName()`
- `validateNotificationDataSourceInput()`

### Status Page Helpers

[status_page_helpers.go](status_page_helpers.go)

**Critical fix for v0.1.6**: Status pages had a perpetual diff issue where unknown group/monitor IDs caused Terraform to
think state differed from config on every plan.

**Solution**: `convertUnknownIDsToNull()` function family

```go
// Main entry point: converts all unknown IDs to null in group list
func convertUnknownIDsToNull(ctx context.Context, groupList types.List, diags *diag.Diagnostics) types.List

// Internal helpers
func deserializeGroupsForConversion(ctx context.Context, groupList types.List, diags *diag.Diagnostics) []PublicGroupModel
func convertGroupsUnknownToNull(groups []PublicGroupModel) []PublicGroupModel
func convertMonitorListUnknownToNull(monitors []PublicMonitorModel) []PublicMonitorModel
func buildGroupListFromModels(ctx context.Context, groups []PublicGroupModel, diags *diag.Diagnostics) types.List

// Null list constructors (with proper types)
func nullGroupList() types.List
func nullMonitorList() types.List

// Type definitions for nested objects
func groupListAttrType() attr.Type
func monitorListAttrType() attr.Type
```

**Usage** (in Create/Update handlers):

```go
// After creating/updating status page
data.PublicGroupList = convertUnknownIDsToNull(ctx, data.PublicGroupList, &resp.Diagnostics)
resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
```

This ensures Terraform sees explicit nulls instead of unknowns, preventing false diff detection.

## Resource Lifecycle Patterns

### Generic Resource Structure

```go
type {Type}Resource struct {
    client *kuma.Client  // Set during Configure()
}

type {Type}ResourceModel struct {
    // Embedded base model (if applicable)
    MonitorBaseModel       // For monitors
    NotificationBaseModel  // For notifications

    // Type-specific fields
    SpecificField1 types.String
    SpecificField2 types.Int64
    // ...
}
```

### Metadata

```go
func (r *{Type}Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_{resource_name}"
}
```

### Schema

```go
func (r *{Type}Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "...",
        Attributes: withBaseAttributes(map[string]schema.Attribute{
            "field1": schema.StringAttribute{Required: true},
            "field2": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(10)},
        }),
    }
}
```

**Plan Modifiers**:

- `int64planmodifier.UseStateForUnknown()` - For computed IDs
- `stringplanmodifier.RequiresReplace()` - For immutable fields (e.g., status page slug)

**Defaults**:

- `int64default.StaticInt64(value)`
- `booldefault.StaticBool(value)`
- `stringdefault.StaticString(value)`

### Configure

```go
func (r *{Type}Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*kuma.Client)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *kuma.Client, got: %T", req.ProviderData),
        )
        return
    }

    r.client = client
}
```

### Create

```go
func (r *{Type}Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data {Type}ResourceModel

    // 1. Read plan
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // 2. Build API object from Terraform model
    apiObject := buildAPIObject(ctx, &data, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    // 3. Call API
    id, err := r.client.CreateResource(ctx, apiObject)
    if err != nil {
        resp.Diagnostics.AddError("Create failed", fmt.Sprintf("Error: %s", err.Error()))
        return
    }

    // 4. Update state with computed values
    data.ID = types.Int64Value(id)

    // 5. Handle tags if monitor (separate API calls)
    if monitor && hasTags {
        handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
    }

    // 6. Handle status page groups if applicable
    if statusPage {
        data.PublicGroupList = convertUnknownIDsToNull(ctx, data.PublicGroupList, &resp.Diagnostics)
    }

    // 7. Save state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Read

```go
func (r *{Type}Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data {Type}ResourceModel

    // 1. Read current state
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // 2. Call API to get current resource
    id := data.ID.ValueInt64()

    // For monitors (polymorphic)
    var apiObject MonitorType
    err := r.client.GetMonitorAs(ctx, id, &apiObject)

    // For notifications (type erasure)
    baseNotification, err := r.client.GetNotification(ctx, id)

    // 3. Handle 404 (resource deleted externally)
    if errors.Is(err, kuma.ErrNotFound) {
        resp.State.RemoveResource(ctx)
        return
    }
    if err != nil {
        resp.Diagnostics.AddError("Read failed", err.Error())
        return
    }

    // 4. Map API object to Terraform model
    // Keep non-sensitive computed values from state if not returned by API
    mapAPIToModel(ctx, apiObject, &data, &resp.Diagnostics)

    // 5. Save state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Update

```go
func (r *{Type}Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var data {Type}ResourceModel

    // 1. Read plan
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // 2. Build API object
    apiObject := buildAPIObject(ctx, &data, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    // 3. Call API
    err := r.client.UpdateResource(ctx, apiObject)
    if err != nil {
        resp.Diagnostics.AddError("Update failed", err.Error())
        return
    }

    // 4. Handle post-update operations (tags, status page groups, etc.)
    // ...

    // 5. Save state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Delete

```go
func (r *{Type}Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var data {Type}ResourceModel

    // 1. Read current state
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // 2. Call API
    id := data.ID.ValueInt64()
    err := r.client.DeleteResource(ctx, id)
    if err != nil {
        resp.Diagnostics.AddError("Delete failed", err.Error())
        return
    }

    // 3. State automatically removed by Terraform
}
```

### ImportState (Optional)

```go
func (r *{Type}Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Parse import ID (usually just the resource ID)
    id, err := strconv.ParseInt(req.ID, 10, 64)
    if err != nil {
        resp.Diagnostics.AddError("Import failed", fmt.Sprintf("Invalid ID format: %s", err.Error()))
        return
    }

    // Set ID in state
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)

    // Terraform will automatically call Read() to populate remaining fields
}
```

## State Management Patterns

### resp.Diagnostics Pattern

**Never use direct returns** when errors occur. Always add to diagnostics:

```go
// Bad
if err != nil {
    return err
}

// Good
if err != nil {
    resp.Diagnostics.AddError("Operation failed", err.Error())
    return
}
```

**Early return after diagnostics**:

```go
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
    return  // Don't proceed if there are errors
}
```

### Error Wrapping

All errors from external packages must be wrapped:

```go
if err != nil {
    return fmt.Errorf("create monitor: %w", err)
}
```

### Sentinel Error Handling

```go
import "errors"

err := r.client.GetMonitor(ctx, id)
if errors.Is(err, kuma.ErrNotFound) {
    resp.State.RemoveResource(ctx)  // Resource deleted externally
    return
}
```

### List/Nested Object Handling

**Creating Lists**:

```go
listValue, diags := types.ListValueFrom(ctx, types.StringType, []string{"val1", "val2"})
resp.Diagnostics.Append(diags...)
data.StringList = listValue
```

**Reading Lists**:

```go
var stringSlice []string
diags := data.StringList.ElementsAs(ctx, &stringSlice, false)
resp.Diagnostics.Append(diags...)
```

**Nested Objects**:

```go
// Define type
tagsAttrType := types.ObjectType{
    AttrTypes: map[string]attr.Type{
        "tag_id": types.Int64Type,
        "value":  types.StringType,
    },
}

// Create list of objects
listValue, diags := types.ListValueFrom(ctx, tagsAttrType, tagModels)
```

### Unknown vs Null

- **Unknown**: Value not yet known (during plan phase)
- **Null**: Explicitly no value

**Perpetual Diff Issue**: Unknowns in nested objects cause Terraform to think state differs from config. Solution:
convert unknowns to nulls after Create/Update (see [Status Page Helpers](#status-page-helpers)).

## Client Integration Patterns

### Monitor Operations

```go
// Create monitor (returns ID)
id, err := r.client.CreateMonitor(ctx, &monitor)

// Read monitor (polymorphic - use .As() for specific type)
var httpMonitor kuma.MonitorHTTP
err := r.client.GetMonitorAs(ctx, id, &httpMonitor)

// Update monitor
err := r.client.UpdateMonitor(ctx, monitor)

// Delete monitor
err := r.client.DeleteMonitor(ctx, id)

// Get all monitors
monitors, err := r.client.GetMonitors(ctx)
```

### Notification Operations

```go
// Create notification (returns ID)
id, err := r.client.CreateNotification(ctx, notification)

// Read notification (returns base notification)
baseNotification, err := r.client.GetNotification(ctx, id)

// Update notification
err := r.client.UpdateNotification(ctx, notification)

// Delete notification
err := r.client.DeleteNotification(ctx, id)

// Get all notifications
notifications, err := r.client.GetNotifications(ctx)
```

### Tag Operations

```go
// Add tag to monitor (returns tag assignment, ignore return value usually)
_, err := r.client.AddMonitorTag(ctx, tagID, monitorID, value)

// Tags are included in monitor response during Read
```

## Testing Infrastructure

### Test Setup ([main_test.go:15-100](main_test.go))

**TestMain Lifecycle**:

1. Check `TF_ACC` environment variable (only run acceptance tests if set)
2. Create Docker pool and ping daemon
3. Run `louislam/uptime-kuma:2` container on port 3001
4. Set 480-second expiration for auto-cleanup
5. Wait for Kuma to be ready (exponential backoff, max 2 minutes)
6. Create initial client and perform autosetup
7. Close initial connection
8. **Enable connection pooling** for tests (`enableConnectionPool = true`)
9. Run tests (all share pooled connection)
10. Cleanup: Close pool, purge container

**Global Variables** (used by all tests):

```go
var (
    endpoint             string  // e.g., "http://localhost:32768"
    username             string  // "admin"
    password             string  // "password123"
    enableConnectionPool bool    // true during tests
)
```

### Provider Configuration Helper

```go
func providerConfig() string {
    return fmt.Sprintf(`
provider "uptimekuma" {
  endpoint = %[1]q
  username = %[2]q
  password = %[3]q
}
`, endpoint, username, password)
}
```

### Acceptance Test Pattern

```go
func TestAcc{Type}Resource(t *testing.T) {
    name := acctest.RandomWithPrefix("Test{Type}")
    nameUpdated := acctest.RandomWithPrefix("Test{Type}Updated")

    resource.Test(t, resource.TestCase{
        PreCheck: func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create
            {
                Config: testAcc{Type}ResourceConfig(name, field1, field2, ...),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "uptimekuma_{type}.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact(name),
                    ),
                    statecheck.ExpectKnownValue(
                        "uptimekuma_{type}.test",
                        tfjsonpath.New("field1"),
                        knownvalue.StringExact(field1),
                    ),
                    // ... more checks
                },
            },
            // Step 2: Update
            {
                Config: testAcc{Type}ResourceConfig(nameUpdated, field1Updated, ...),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "uptimekuma_{type}.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact(nameUpdated),
                    ),
                    // ... verify updated values
                },
            },
            // Step 3: Import (optional)
            {
                ResourceName:      "uptimekuma_{type}.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}

func testAcc{Type}ResourceConfig(name string, field1 string, ...) string {
    return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_{type}" "test" {
  name   = %[1]q
  field1 = %[2]q
  // ...
}
`, name, field1, ...)
}
```

**State Check Patterns**:

- `knownvalue.StringExact(value)` - Exact string match
- `knownvalue.Int64Exact(value)` - Exact int64 match
- `knownvalue.Bool(value)` - Boolean match
- `knownvalue.NotNull()` - Value is not null
- `knownvalue.ListSizeExact(n)` - List has exactly n elements

### Data Source Testing

Similar pattern but typically:

1. Create resource
2. Query via data source (by ID or name)
3. Verify returned attributes match resource

```go
func TestAcc{Type}DataSource(t *testing.T) {
    name := acctest.RandomWithPrefix("Test{Type}")

    resource.Test(t, resource.TestCase{
        PreCheck: func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAcc{Type}DataSourceConfig(name),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "data.uptimekuma_{type}.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact(name),
                    ),
                },
            },
        },
    })
}

func testAcc{Type}DataSourceConfig(name string) string {
    return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_{type}" "test" {
  name = %[1]q
}

data "uptimekuma_{type}" "test" {
  name = uptimekuma_{type}.test.name
}
`, name)
}
```

## Recent Changes and Patterns

### v0.1.6: Status Page Perpetual Diff Fix

**Problem**: Unknown IDs in `public_group_list` caused perpetual diffs. Terraform saw unknowns in state and thought
config changed.

**Solution**: `convertUnknownIDsToNull()` function ([status_page_helpers.go:15](status_page_helpers.go))

- Recursively converts unknown IDs to explicit nulls
- Runs after Create/Update before saving state
- Requires careful type management with nested objects

**PR**: #227

### Recent Feature Additions

Last 20 commits show active development:

**New Notifications**:

- 46elks (#224)
- WAHA (WhatsApp API) (#222)
- Threema (#220)
- Stackfield (#219)
- ServerChan (#211)
- SendGrid (#210)
- Pushy (#209)
- PushPlus (#208)
- PushDeer (#207)
- Nextcloud Talk (#203)

**New Monitors**:

- MongoDB (#218)
- MySQL (#215)
- SNMP (#213)
- SQL Server (#212)

**Pattern**: Each new resource follows established base model pattern with minimal code duplication.

## Design Patterns and Principles

### Composition over Inheritance

- Base models embedded in specific models
- Helper functions add common attributes to schemas
- Avoids deep inheritance hierarchies
- Go's struct embedding provides clean composition

### Polymorphic Resource Handling

**Monitors**: Use `.As()` type assertion pattern from client library

```go
var httpMonitor kuma.MonitorHTTP
err := client.GetMonitorAs(ctx, id, &httpMonitor)
```

**Notifications**: Use type-specific fields in embedded base model

```go
type NotificationSlackResourceModel struct {
    NotificationBaseModel  // Common fields
    WebhookURL types.String  // Type-specific
}
```

### Early Return Pattern

Reduce nesting by returning early on errors:

```go
if err != nil {
    resp.Diagnostics.AddError("...", err.Error())
    return
}
// Continue happy path
```

### State Management Pattern

```go
// 1. Deserialize from plan/state
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
    return
}

// 2. Build API object
apiObject := buildObject(ctx, &data, &resp.Diagnostics)

// 3. Call API
result, err := r.client.Operation(ctx, apiObject)
if err != nil {
    resp.Diagnostics.AddError("...", err.Error())
    return
}

// 4. Update state
data.ID = types.Int64Value(result.ID)

// 5. Save state
resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
```

## Integration Points with Other Modules

### internal/client

Provider uses [../client](../client/CLAUDE.md) package for:

- Client creation with retry logic
- Connection pooling during tests
- Context management (`context.Background()` vs Terraform context)

### go-uptime-kuma-client

External dependency providing:

- `kuma.Client` type
- Monitor and notification types
- Socket.IO-based API communication
- `.As()` pattern for polymorphic monitor handling

### terraform-plugin-framework

Terraform SDK providing:

- Provider, Resource, DataSource interfaces
- Schema definition with attributes, validators, defaults
- Plan modifiers for computed values
- types package (types.String, types.Int64, types.Bool, types.List)

### terraform-plugin-testing

Testing framework providing:

- `resource.Test()` and `resource.TestCase`
- State checks with `statecheck` and `knownvalue`
- Provider factory registration
- Import state testing

## File Organization

### Naming Conventions

- `provider.go` - Provider definition and configuration
- `resource_{type}.go` - Resource implementation
- `resource_{type}_base.go` - Base model for resource category
- `resource_{type}_helpers.go` - Helper functions for resource category
- `datasource_{type}.go` - Data source implementation
- `datasource_{type}_helpers.go` - Data source helper functions
- `{type}_helpers.go` - General helpers (e.g., status_page_helpers.go)
- `main_test.go` - Test setup and Docker integration
- `provider_test.go` - Provider-level tests
- `resource_{type}_test.go` - Resource acceptance tests
- `datasource_{type}_test.go` - Data source acceptance tests

### File Size

Most files are 200-400 lines. Large files:

- `provider.go` (338 lines) - 85+ resource + 81+ data source registrations
- `resource_monitor_http.go` (300+ lines) - Heavily documented
- `resource_monitor_http_base.go` (300+ lines) - Many HTTP configuration options
- `resource_status_page.go` (350+ lines) - Complex nested structures

## Code Quality Considerations

See [../../CODE_STYLE.md](../../CODE_STYLE.md) for:

- Function complexity limits (max 50 statements/100 lines)
- Parameter limits (max 6 parameters)
- Error wrapping requirements
- Naming conventions
- Linting configuration

**Provider-Specific Considerations**:

- CRUD methods often approach complexity limits - extract helpers when needed
- Use early returns to reduce nesting
- Break large functions into focused helpers
- Group related parameters into structs if exceeding 6-parameter limit

## Common Patterns Summary

1. **Base models** embedded in specific types with helper functions for schema building
2. **Tag handling** via separate API calls after monitor creation
3. **Error handling** with `errors.Is()` for sentinel errors and early returns
4. **State management** via `resp.Diagnostics`, never direct returns
5. **Nested objects** require careful type management and unknown-to-null conversion
6. **Connection pooling** for acceptance tests to prevent rate limiting
7. **Context management**: `context.Background()` in provider, Terraform context in resources
8. **Data sources** support both ID and name lookups for flexibility
9. **Imports** use simple ID parsing and automatic Read() call
10. **Testing** uses Docker containers with connection pooling and state checks

## Future Enhancements

Potential improvements:

1. **Code generation**: 50+ similar notification types could be code-generated
2. **Shared test utilities**: Extract common test patterns into helpers
3. **Better error messages**: Add more context to diagnostic errors
4. **Validation**: Add more schema validators for common patterns (URLs, ranges, etc.)
5. **Documentation**: Auto-generate more resource documentation from schema
