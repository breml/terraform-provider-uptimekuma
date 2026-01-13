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
				MarkdownDescription: "Uptime Kuma endpoint. Can be set via `UPTIMEKUMA_ENDPOINT` environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Uptime Kuma username. Can be set via `UPTIMEKUMA_USERNAME` environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Uptime Kuma password. Can be set via `UPTIMEKUMA_PASSWORD` environment variable.",
				Optional:            true,
				Sensitive:           true,
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

	// Apply environment variable defaults where Terraform config is not provided.
	// Precedence: Terraform config > environment variables > nothing
	applyEnvironmentDefaults(&data)

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

// applyEnvironmentDefaults applies environment variable defaults to the provider model.
// Terraform config values take precedence over environment variables.
func applyEnvironmentDefaults(data *UptimeKumaProviderModel) {
	envEndpoint := os.Getenv("UPTIMEKUMA_ENDPOINT")
	if data.Endpoint.IsNull() && envEndpoint != "" {
		data.Endpoint = types.StringValue(envEndpoint)
	}

	envUsername := os.Getenv("UPTIMEKUMA_USERNAME")
	if data.Username.IsNull() && envUsername != "" {
		data.Username = types.StringValue(envUsername)
	}

	envPassword := os.Getenv("UPTIMEKUMA_PASSWORD")
	if data.Password.IsNull() && envPassword != "" {
		data.Password = types.StringValue(envPassword)
	}
}

// Resources returns the list of resources for the provider.
func (*UptimeKumaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNotificationResource,
		NewNotificationAlertaResource,
		NewNotificationAlertNowResource,
		NewNotificationAliyunsmsResource,
		NewNotificationAppriseResource,
		NewNotificationBarkResource,
		NewNotificationBitrix24Resource,
		NewNotificationBrevoResource,
		NewNotificationCallMeBotResource,
		NewNotificationCellsyntResource,
		NewNotificationClicksendSmsResource,
		NewNotificationDingDingResource,
		NewNotificationDiscordResource,
		NewNotificationEvolutionResource,
		NewNotificationFeishuResource,
		NewNotificationFlashDutyResource,
		NewNotificationFreemobileResource,
		NewNotificationGoogleChatResource,
		NewNotificationGotifyResource,
		NewNotificationGrafanaOncallResource,
		NewNotificationHomeAssistantResource,
		NewNotificationKookResource,
		NewNotificationLineResource,
		NewNotificationLunaseaResource,
		NewNotificationLinenotifyResource,
		NewNotificationMatrixResource,
		NewNotificationMattermostResource,
		NewNotificationNostrResource,
		NewNotificationNtfyResource,
		NewNotificationOctopushResource,
		NewNotificationOpsgenieResource,
		NewNotificationPagerDutyResource,
		NewNotificationPagerTreeResource,
		NewNotificationPushbulletResource,
		NewNotificationPushoverResource,
		NewNotificationRocketChatResource,
		NewNotificationSignalResource,
		NewNotificationSlackResource,
		NewNotificationSplunkResource,
		NewNotificationSMTPResource,
		NewNotificationTeamsResource,
		NewNotificationTelegramResource,
		NewNotificationTwilioResource,
		NewNotificationWebhookResource,
		NewNotificationWeComResource,
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
		NewMonitorDockerResource,
		NewMonitorMQTTResource,
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
		NewNotificationAlertaDataSource,
		NewNotificationAlertNowDataSource,
		NewNotificationAliyunsmsDataSource,
		NewNotificationAppriseDataSource,
		NewNotificationBarkDataSource,
		NewNotificationBitrix24DataSource,
		NewNotificationBrevoDataSource,
		NewNotificationCallMeBotDataSource,
		NewNotificationCellsyntDataSource,
		NewNotificationClicksendSmsDataSource,
		NewNotificationDingDingDataSource,
		NewNotificationDiscordDataSource,
		NewNotificationEvolutionDataSource,
		NewNotificationFeishuDataSource,
		NewNotificationFlashDutyDataSource,
		NewNotificationFreemobileDataSource,
		NewNotificationGoogleChatDataSource,
		NewNotificationGotifyDataSource,
		NewNotificationGrafanaOncallDataSource,
		NewNotificationHomeAssistantDataSource,
		NewNotificationKookDataSource,
		NewNotificationLineDataSource,
		NewNotificationLunaseaDataSource,
		NewNotificationLinenotifyDataSource,
		NewNotificationMatrixDataSource,
		NewNotificationMattermostDataSource,
		NewNotificationNostrDataSource,
		NewNotificationNtfyDataSource,
		NewNotificationOctopushDataSource,
		NewNotificationOpsgenieDataSource,
		NewNotificationPagerDutyDataSource,
		NewNotificationPagerTreeDataSource,
		NewNotificationPushbulletDataSource,
		NewNotificationPushoverDataSource,
		NewNotificationRocketChatDataSource,
		NewNotificationSignalDataSource,
		NewNotificationSlackDataSource,
		NewNotificationSplunkDataSource,
		NewNotificationSMTPDataSource,
		NewNotificationTeamsDataSource,
		NewNotificationTelegramDataSource,
		NewNotificationTwilioDataSource,
		NewNotificationWebhookDataSource,
		NewNotificationWeComDataSource,
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
		NewMonitorDockerDataSource,
		NewMonitorMQTTDataSource,
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
