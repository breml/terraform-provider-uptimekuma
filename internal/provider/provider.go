// Package provider implements the Uptime Kuma Terraform provider.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"

	"github.com/breml/terraform-provider-uptimekuma/internal/client"
)

// Ensure UptimeKumaProvider satisfies various provider interfaces.
var (
	_ provider.Provider = &UptimeKumaProvider{}
)

// UptimeKumaProvider defines the provider implementation.
type UptimeKumaProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// UptimeKumaProviderModel describes the provider data model.
type UptimeKumaProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// Metadata returns the metadata for the provider.
func (p *UptimeKumaProvider) Metadata(
	_ context.Context,
	_ provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "uptimekuma"
	resp.Version = p.version
}

// Schema returns the schema for the provider.
func (*UptimeKumaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Uptime Kuma endpoint",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Uptime Kuma username",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Uptime Kuma password",
				Optional:            true,
			},
		},
	}
}

// Configure configures the provider with the API client.
func (*UptimeKumaProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var data UptimeKumaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate configuration
	// Endpoint is always required to connect to Uptime Kuma
	// Username and password are optional (client will skip login if both are empty)
	// However, if either username or password is provided, both must be present
	hasUsername := !data.Username.IsNull()
	hasPassword := !data.Password.IsNull()

	if data.Endpoint.IsNull() {
		resp.Diagnostics.AddError("endpoint required", "endpoint is required")
	}

	// If credentials are partially provided, require both
	if hasUsername && !hasPassword {
		resp.Diagnostics.AddError("password required", "password is required when username is provided")
	}

	if hasPassword && !hasUsername {
		resp.Diagnostics.AddError("username required", "username is required when password is provided")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Uptime Kuma client configuration for data sources and resources
	// We can not use the context from Terraform, since it gets cancelled too early.
	// ValueString() returns "" for null values, which client library handles gracefully
	kumaClient, err := client.New(context.Background(), &client.Config{
		Endpoint:             data.Endpoint.ValueString(),
		Username:             data.Username.ValueString(),
		Password:             data.Password.ValueString(),
		EnableConnectionPool: true,
		LogLevel:             kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create client", err.Error())
		return
	}

	// Context is cancelled on shutdown - you can use defer or goroutine
	go func() {
		<-ctx.Done()
		client.GetGlobalPool().Release()
	}()

	resp.DataSourceData = kumaClient
	resp.ResourceData = kumaClient
}

// Resources returns the list of resources for the provider.
func (*UptimeKumaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNotificationResource,
		NewNotificationAppriseResource,
		NewNotificationDingDingResource,
		NewNotificationDiscordResource,
		NewNotificationFeishuResource,
		NewNotificationGoogleChatResource,
		NewNotificationGotifyResource,
		NewNotificationGrafanaOncallResource,
		NewNotificationHomeAssistantResource,
		NewNotificationMatrixResource,
		NewNotificationMattermostResource,
		NewNotificationNtfyResource,
		NewNotificationOpsgenieResource,
		NewNotificationPagerDutyResource,
		NewNotificationSlackResource,
		NewNotificationTeamsResource,
		NewNotificationWebhookResource,
		NewMonitorHTTPResource,
		NewMonitorHTTPKeywordResource,
		NewMonitorGrpcKeywordResource,
		NewMonitorHTTPJSONQueryResource,
		NewMonitorGroupResource,
		NewMonitorPingResource,
		NewMonitorDNSResource,
		NewMonitorPushResource,
		NewMonitorRealBrowserResource,
		NewMonitorPostgresResource,
		NewMonitorRedisResource,
		NewMonitorTCPPortResource,
		NewProxyResource,
		NewTagResource,
		NewDockerHostResource,
		NewMaintenanceResource,
		NewMaintenanceMonitorsResource,
		NewMaintenanceStatusPagesResource,
		NewStatusPageResource,
		NewStatusPageIncidentResource,
	}
}

// DataSources returns the list of data sources for the provider.
func (*UptimeKumaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMaintenancesDataSource,
		NewTagDataSource,
		NewNotificationDataSource,
		NewNotificationAppriseDataSource,
		NewNotificationDingDingDataSource,
		NewNotificationDiscordDataSource,
		NewNotificationFeishuDataSource,
		NewNotificationGoogleChatDataSource,
		NewNotificationGotifyDataSource,
		NewNotificationGrafanaOncallDataSource,
		NewNotificationHomeAssistantDataSource,
		NewNotificationMatrixDataSource,
		NewNotificationMattermostDataSource,
		NewNotificationNtfyDataSource,
		NewNotificationOpsgenieDataSource,
		NewNotificationPagerDutyDataSource,
		NewNotificationSlackDataSource,
		NewNotificationTeamsDataSource,
		NewNotificationWebhookDataSource,
		NewMonitorHTTPDataSource,
		NewMonitorHTTPKeywordDataSource,
		NewMonitorGrpcKeywordDataSource,
		NewMonitorHTTPJSONQueryDataSource,
		NewMonitorGroupDataSource,
		NewMonitorPingDataSource,
		NewMonitorDNSDataSource,
		NewMonitorPushDataSource,
		NewMonitorRealBrowserDataSource,
		NewMonitorPostgresDataSource,
		NewMonitorRedisDataSource,
		NewMonitorTCPPortDataSource,
		NewProxyDataSource,
		NewDockerHostDataSource,
		NewMaintenanceDataSource,
		NewMaintenanceMonitorsDataSource,
		NewMaintenanceStatusPagesDataSource,
		NewStatusPageDataSource,
	}
}

// New returns a new instance of the provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UptimeKumaProvider{
			version: version,
		}
	}
}
