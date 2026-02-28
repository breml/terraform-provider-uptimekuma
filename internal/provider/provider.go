// Package provider implements the Uptime Kuma Terraform provider.
package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	Endpoint   types.String `tfsdk:"endpoint"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	Timeout    types.String `tfsdk:"timeout"`
	MaxRetries types.Int64  `tfsdk:"max_retries"`
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
			"timeout": schema.StringAttribute{
				MarkdownDescription: "Connection timeout as a Go duration string (e.g. `30s`, `2m`). " +
					"Can be set via `UPTIMEKUMA_TIMEOUT` environment variable.",
				Optional: true,
			},
			"max_retries": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of connection retry attempts (default: `5`). " +
					"Can be set via `UPTIMEKUMA_MAX_RETRIES` environment variable.",
				Optional: true,
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
	applyEnvironmentDefaults(&data, resp)

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

	var connectTimeout time.Duration

	timeoutStr := strings.TrimSpace(data.Timeout.ValueString())
	if !data.Timeout.IsNull() && timeoutStr != "" {
		var parseErr error

		connectTimeout, parseErr = time.ParseDuration(timeoutStr)
		if parseErr != nil {
			resp.Diagnostics.AddError(
				"invalid timeout",
				fmt.Sprintf("failed to parse timeout %q: %s", data.Timeout.ValueString(), parseErr.Error()),
			)

			return
		}
	}

	maxRetries := 5

	if !data.MaxRetries.IsNull() {
		maxRetries = int(data.MaxRetries.ValueInt64())
	}

	if maxRetries < 0 {
		resp.Diagnostics.AddError(
			"invalid max_retries",
			fmt.Sprintf("max_retries must be non-negative, got %d", maxRetries),
		)

		return
	}

	kumaClient, err := client.New(context.Background(), &client.Config{
		Endpoint:             data.Endpoint.ValueString(),
		Username:             data.Username.ValueString(),
		Password:             data.Password.ValueString(),
		EnableConnectionPool: true,
		LogLevel:             kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
		ConnectTimeout:       connectTimeout,
		MaxRetries:           maxRetries,
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

	pd := &providerData{
		client:   kumaClient,
		password: data.Password.ValueString(),
	}

	resp.DataSourceData = pd
	resp.ResourceData = pd
}

// applyEnvironmentDefaults applies environment variable defaults to the provider model.
// Terraform config values take precedence over environment variables.
func applyEnvironmentDefaults(data *UptimeKumaProviderModel, resp *provider.ConfigureResponse) {
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

	envTimeout := os.Getenv("UPTIMEKUMA_TIMEOUT")
	if data.Timeout.IsNull() && envTimeout != "" {
		data.Timeout = types.StringValue(envTimeout)
	}

	envMaxRetries := os.Getenv("UPTIMEKUMA_MAX_RETRIES")
	if data.MaxRetries.IsNull() && envMaxRetries != "" {
		val, err := strconv.ParseInt(envMaxRetries, 10, 64)
		if err == nil {
			data.MaxRetries = types.Int64Value(val)
		} else {
			resp.Diagnostics.AddWarning(
				"invalid UPTIMEKUMA_MAX_RETRIES",
				fmt.Sprintf("invalid UPTIMEKUMA_MAX_RETRIES value %q; ignore value from environment variable", envMaxRetries),
			)
		}
	}
}

// Resources returns the list of resources for the provider.
func (*UptimeKumaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNotificationResource,
		NewNotification46ElksResource,
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
		NewNotificationGoAlertResource,
		NewNotificationGoogleChatResource,
		NewNotificationGotifyResource,
		NewNotificationGorushResource,
		NewNotificationGrafanaOncallResource,
		NewNotificationGTXMessagingResource,
		NewNotificationHeiiOnCallResource,
		NewNotificationHomeAssistantResource,
		NewNotificationKeepResource,
		NewNotificationKookResource,
		NewNotificationLineResource,
		NewNotificationLunaseaResource,
		NewNotificationLinenotifyResource,
		NewNotificationMatrixResource,
		NewNotificationMattermostResource,
		NewNotificationNextcloudTalkResource,
		NewNotificationNostrResource,
		NewNotificationNtfyResource,
		NewNotificationOctopushResource,
		NewNotificationOpsgenieResource,
		NewNotificationPagerDutyResource,
		NewNotificationPagerTreeResource,
		NewNotificationPushbulletResource,
		NewNotificationPushDeerResource,
		NewNotificationPushoverResource,
		NewNotificationPushPlusResource,
		NewNotificationPushyResource,
		NewNotificationRocketChatResource,
		NewNotificationSendgridResource,
		NewNotificationServerChanResource,
		NewNotificationSignalResource,
		NewNotificationSlackResource,
		NewNotificationStackfieldResource,
		NewNotificationThreemaResource,
		NewNotificationSplunkResource,
		NewNotificationSMTPResource,
		NewNotificationTeamsResource,
		NewNotificationTelegramResource,
		NewNotificationTwilioResource,
		NewNotificationWAHAResource,
		NewNotificationWebhookResource,
		NewNotificationWeComResource,
		NewMonitorHTTPResource,
		NewMonitorHTTPKeywordResource,
		NewMonitorGrpcKeywordResource,
		NewMonitorHTTPJSONQueryResource,
		NewMonitorGroupResource,
		NewMonitorPingResource,
		NewMonitorDNSResource,
		NewMonitorSNMPResource,
		NewMonitorPushResource,
		NewMonitorRealBrowserResource,
		NewMonitorPostgresResource,
		NewMonitorMySQLResource,
		NewMonitorMongoDBResource,
		NewMonitorRedisResource,
		NewMonitorSQLServerResource,
		NewMonitorTCPPortResource,
		NewMonitorDockerResource,
		NewMonitorMQTTResource,
		NewProxyResource,
		NewTagResource,
		NewDockerHostResource,
		NewMaintenanceResource,
		NewMaintenanceMonitorsResource,
		NewMaintenanceStatusPagesResource,
		NewSettingsResource,
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
		NewNotification46ElksDataSource,
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
		NewNotificationGoAlertDataSource,
		NewNotificationGoogleChatDataSource,
		NewNotificationGotifyDataSource,
		NewNotificationGorushDataSource,
		NewNotificationGrafanaOncallDataSource,
		NewNotificationGTXMessagingDataSource,
		NewNotificationHeiiOnCallDataSource,
		NewNotificationHomeAssistantDataSource,
		NewNotificationKeepDataSource,
		NewNotificationKookDataSource,
		NewNotificationLineDataSource,
		NewNotificationLunaseaDataSource,
		NewNotificationLinenotifyDataSource,
		NewNotificationMatrixDataSource,
		NewNotificationMattermostDataSource,
		NewNotificationNextcloudTalkDataSource,
		NewNotificationNostrDataSource,
		NewNotificationNtfyDataSource,
		NewNotificationOctopushDataSource,
		NewNotificationOpsgenieDataSource,
		NewNotificationPagerDutyDataSource,
		NewNotificationPagerTreeDataSource,
		NewNotificationPushbulletDataSource,
		NewNotificationPushDeerDataSource,
		NewNotificationPushoverDataSource,
		NewNotificationPushPlusDataSource,
		NewNotificationPushyDataSource,
		NewNotificationRocketChatDataSource,
		NewNotificationSendgridDataSource,
		NewNotificationServerChanDataSource,
		NewNotificationSignalDataSource,
		NewNotificationSlackDataSource,
		NewNotificationStackfieldDataSource,
		NewNotificationThreemaDataSource,
		NewNotificationSplunkDataSource,
		NewNotificationSMTPDataSource,
		NewNotificationTeamsDataSource,
		NewNotificationTelegramDataSource,
		NewNotificationTwilioDataSource,
		NewNotificationWAHADataSource,
		NewNotificationWebhookDataSource,
		NewNotificationWeComDataSource,
		NewMonitorHTTPDataSource,
		NewMonitorHTTPKeywordDataSource,
		NewMonitorGrpcKeywordDataSource,
		NewMonitorHTTPJSONQueryDataSource,
		NewMonitorGroupDataSource,
		NewMonitorPingDataSource,
		NewMonitorDNSDataSource,
		NewMonitorSNMPDataSource,
		NewMonitorPushDataSource,
		NewMonitorRealBrowserDataSource,
		NewMonitorPostgresDataSource,
		NewMonitorMySQLDataSource,
		NewMonitorMongoDBDataSource,
		NewMonitorRedisDataSource,
		NewMonitorSQLServerDataSource,
		NewMonitorTCPPortDataSource,
		NewMonitorDockerDataSource,
		NewMonitorMQTTDataSource,
		NewProxyDataSource,
		NewDockerHostDataSource,
		NewMaintenanceDataSource,
		NewMaintenanceMonitorsDataSource,
		NewMaintenanceStatusPagesDataSource,
		NewSettingsDataSource,
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
