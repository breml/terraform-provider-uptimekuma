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
	Endpoint          types.String `tfsdk:"endpoint"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	Timeout           types.String `tfsdk:"timeout"`
	PerAttemptTimeout types.String `tfsdk:"per_attempt_timeout"`
	OperationTimeout  types.String `tfsdk:"operation_timeout"`
	MaxRetries        types.Int64  `tfsdk:"max_retries"`
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
				MarkdownDescription: "Overall connection timeout as a Go duration string (e.g. `30s`, `2m`). " +
					"Bounds the total time spent attempting to connect to Uptime Kuma, including all retry " +
					"attempts and backoff. Defaults to `30s` if not specified. " +
					"Can be set via `UPTIMEKUMA_TIMEOUT` environment variable.",
				Optional: true,
			},
			"per_attempt_timeout": schema.StringAttribute{
				MarkdownDescription: "Optional per-attempt connection timeout as a Go duration string " +
					"(e.g. `5s`, `10s`). Caps the time spent on each individual connection attempt. The " +
					"effective per-attempt timeout is the smaller of this value and the remaining `timeout` " +
					"budget. When unset, each attempt may use the full remaining `timeout` budget. " +
					"Can be set via `UPTIMEKUMA_PER_ATTEMPT_TIMEOUT` environment variable.",
				Optional: true,
			},
			"operation_timeout": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(
					"Timeout for a single Uptime Kuma command as a Go duration string "+
						"(e.g. `30s`, `2m`) (default: `%s`). Bounds how long the provider waits for the "+
						"server to acknowledge one create, read, update or delete and to confirm it, so "+
						"an instance that accepts a command and never answers it fails the operation "+
						"instead of blocking the apply indefinitely. It also bounds the login and setup "+
						"performed while connecting, so during connect the effective bound is whichever "+
						"of `timeout` and `operation_timeout` expires first. The bound is on the round "+
						"trip and not on the whole call: a write that has to queue behind other writes "+
						"to the same list waits its turn outside this budget. A value of `0s` means "+
						"\"not configured\" and falls back to the default; the bound cannot be disabled, "+
						"so set a large value instead. "+
						"Can be set via `UPTIMEKUMA_OPERATION_TIMEOUT` environment variable.",
					// Rendered in seconds rather than via Duration.String, which
					// would print a minute as "1m0s" next to the other timeouts.
					fmt.Sprintf("%ds", int64(client.DefaultOperationTimeout/time.Second)),
				),
				Optional: true,
			},
			"max_retries": schema.Int64Attribute{
				MarkdownDescription: fmt.Sprintf(
					"Maximum number of connection retry attempts (default: `%d`). "+
						"All retry attempts must complete within the overall `timeout` budget. "+
						"Can be set via `UPTIMEKUMA_MAX_RETRIES` environment variable.",
					defaultMaxRetries,
				),
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

	opts := parseClientOptions(&data, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	kumaClient, err := client.New(context.Background(), &client.Config{
		Endpoint:             data.Endpoint.ValueString(),
		Username:             data.Username.ValueString(),
		Password:             data.Password.ValueString(),
		EnableConnectionPool: true,
		LogLevel:             kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
		ConnectTimeout:       opts.connectTimeout,
		PerAttemptTimeout:    opts.perAttemptTimeout,
		OperationTimeout:     opts.operationTimeout,
		MaxRetries:           opts.maxRetries,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"failed to connect to Uptime Kuma",
			connectionErrorDetail(data.Endpoint.ValueString(), err),
		)

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

func connectionErrorDetail(endpoint string, err error) string {
	return fmt.Sprintf(
		"Could not establish a connection to Uptime Kuma at %q.\n\n"+
			"Underlying error: %v\n\n"+
			"Common causes:\n"+
			"  - The endpoint URL is incorrect or Uptime Kuma is not reachable from this host.\n"+
			"  - A reverse proxy is not forwarding the Socket.IO/WebSocket connection.\n"+
			"  - TLS certificate issues when using a custom domain (try the direct URL to confirm).\n"+
			"  - Two-factor authentication (2FA) is enabled on the account (currently not supported).\n\n"+
			"To collect diagnostics, re-run with the following environment variables set and share the "+
			"output when reporting the issue:\n"+
			"  TF_LOG=DEBUG SOCKETIO_LOG_LEVEL=DEBUG terraform plan",
		endpoint,
		err,
	)
}

// clientOptions holds parsed and validated provider connection options.
type clientOptions struct {
	connectTimeout    time.Duration
	perAttemptTimeout time.Duration
	operationTimeout  time.Duration
	maxRetries        int
}

// parseClientOptions extracts and validates timeout, per_attempt_timeout,
// operation_timeout and max_retries from the provider model.
func parseClientOptions(
	data *UptimeKumaProviderModel,
	resp *provider.ConfigureResponse,
) clientOptions {
	var opts clientOptions

	opts.connectTimeout = parseDurationAttribute(data.Timeout, "timeout", resp)
	if resp.Diagnostics.HasError() {
		return opts
	}

	opts.perAttemptTimeout = parseDurationAttribute(data.PerAttemptTimeout, "per_attempt_timeout", resp)
	if resp.Diagnostics.HasError() {
		return opts
	}

	opts.operationTimeout = parseDurationAttribute(data.OperationTimeout, "operation_timeout", resp)
	if resp.Diagnostics.HasError() {
		return opts
	}

	warnUnusedOperationTimeout(data, opts.operationTimeout, resp)

	if opts.perAttemptTimeout > 0 && opts.connectTimeout > 0 && opts.perAttemptTimeout >= opts.connectTimeout {
		resp.Diagnostics.AddWarning(
			"per_attempt_timeout has no effect",
			fmt.Sprintf(
				"per_attempt_timeout (%s) is greater than or equal to timeout (%s); "+
					"each attempt is already bounded by the remaining overall budget, "+
					"so per_attempt_timeout will never be applied",
				opts.perAttemptTimeout, opts.connectTimeout,
			),
		)
	}

	opts.maxRetries = defaultMaxRetries

	if !data.MaxRetries.IsNull() {
		opts.maxRetries = int(data.MaxRetries.ValueInt64())
	}

	if opts.maxRetries < 0 {
		resp.Diagnostics.AddError(
			"invalid max_retries",
			fmt.Sprintf("max_retries must be non-negative, got %d", opts.maxRetries),
		)

		return clientOptions{}
	}

	return opts
}

// defaultMaxRetries mirrors the client package default. It is used both as
// the fallback in parseClientOptions when the user does not provide an
// explicit value, and to render the value in the `max_retries` schema
// description so the documented default cannot drift from the runtime one.
const defaultMaxRetries = 3

// warnUnusedOperationTimeout warns when operation_timeout is set to an explicit
// zero. kuma.WithOperationTimeout documents zero as the opt-out that leaves a
// command bounded only by the caller's context, so a user who writes `0s` most
// likely means "no bound" - but the provider resolves zero to
// client.DefaultOperationTimeout instead, because Terraform gives a provider no
// deadline and an unbounded command would block the apply for good. Saying so
// beats silently applying a duration that appears nowhere in the configuration.
func warnUnusedOperationTimeout(
	data *UptimeKumaProviderModel,
	parsed time.Duration,
	resp *provider.ConfigureResponse,
) {
	if parsed != 0 || data.OperationTimeout.IsNull() {
		return
	}

	if strings.TrimSpace(data.OperationTimeout.ValueString()) == "" {
		return
	}

	resp.Diagnostics.AddWarning(
		"operation_timeout does not disable the bound",
		fmt.Sprintf(
			"operation_timeout is set to %q, which the provider treats as \"not configured\" and "+
				"resolves to the default of %s rather than leaving operations unbounded. "+
				"Omit the attribute to use the default, or set a large explicit value such as \"30m\"",
			data.OperationTimeout.ValueString(), client.DefaultOperationTimeout,
		),
	)
}

// parseDurationAttribute parses a Go duration string from a Terraform string
// attribute. Empty/null values yield a zero duration (meaning "use the default").
// Negative values produce a diagnostic error.
func parseDurationAttribute(
	attr types.String,
	name string,
	resp *provider.ConfigureResponse,
) time.Duration {
	value := strings.TrimSpace(attr.ValueString())
	if attr.IsNull() || value == "" {
		return 0
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("invalid %s", name),
			fmt.Sprintf("failed to parse %s %q: %s", name, attr.ValueString(), err.Error()),
		)

		return 0
	}

	if parsed < 0 {
		resp.Diagnostics.AddError(
			fmt.Sprintf("invalid %s", name),
			fmt.Sprintf("%s must be non-negative, got %s", name, parsed),
		)

		return 0
	}

	return parsed
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

	envPerAttemptTimeout := os.Getenv("UPTIMEKUMA_PER_ATTEMPT_TIMEOUT")
	if data.PerAttemptTimeout.IsNull() && envPerAttemptTimeout != "" {
		data.PerAttemptTimeout = types.StringValue(envPerAttemptTimeout)
	}

	envOperationTimeout := os.Getenv("UPTIMEKUMA_OPERATION_TIMEOUT")
	if data.OperationTimeout.IsNull() && envOperationTimeout != "" {
		data.OperationTimeout = types.StringValue(envOperationTimeout)
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
	resources := notificationResources()

	resources = append(
		resources,
		NewMonitorHTTPResource,
		NewMonitorHTTPKeywordResource,
		NewMonitorGrpcKeywordResource,
		NewMonitorHTTPJSONQueryResource,
		NewMonitorWebsocketUpgradeResource,
		NewMonitorGroupResource,
		NewMonitorPingResource,
		NewMonitorDNSResource,
		NewMonitorSNMPResource,
		NewMonitorRadiusResource,
		NewMonitorKafkaProducerResource,
		NewMonitorPushResource,
		NewMonitorRealBrowserResource,
		NewMonitorPostgresResource,
		NewMonitorMySQLResource,
		NewMonitorOracleDBResource,
		NewMonitorMongoDBResource,
		NewMonitorRabbitMQResource,
		NewMonitorRedisResource,
		NewMonitorSQLServerResource,
		NewMonitorGameDigResource,
		NewMonitorGlobalpingResource,
		NewMonitorNTPResource,
		NewMonitorSteamResource,
		NewMonitorTailscalePingResource,
		NewMonitorTCPPortResource,
		NewMonitorSIPOptionsResource,
		NewMonitorSystemServiceResource,
		NewMonitorPM2Resource,
		NewMonitorDockerResource,
		NewMonitorMQTTResource,
		NewMonitorSMTPResource,
		NewProxyResource,
		NewTagResource,
		NewDockerHostResource,
		NewMaintenanceResource,
		NewMaintenanceMonitorsResource,
		NewMaintenanceStatusPagesResource,
		NewSettingsResource,
		NewStatusPageResource,
		NewStatusPageIncidentResource,
	)

	return resources
}

func notificationResources() []func() resource.Resource {
	return []func() resource.Resource{
		NewNotificationResource,
		NewNotification46ElksResource,
		NewNotificationAlertaResource,
		NewNotificationAlertNowResource,
		NewNotificationAliyunsmsResource,
		NewNotificationAppriseResource,
		NewNotificationBaleResource,
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
		NewNotificationFlowtriqResource,
		NewNotificationFluxerResource,
		NewNotificationFreemobileResource,
		NewNotificationGoAlertResource,
		NewNotificationGoogleChatResource,
		NewNotificationGoogleSheetsResource,
		NewNotificationGotifyResource,
		NewNotificationGorushResource,
		NewNotificationGrafanaOncallResource,
		NewNotificationGTXMessagingResource,
		NewNotificationHaloPSAResource,
		NewNotificationHeiiOnCallResource,
		NewNotificationHomeAssistantResource,
		NewNotificationJiraServiceManagementResource,
		NewNotificationKeepResource,
		NewNotificationKookResource,
		NewNotificationLineResource,
		NewNotificationLunaseaResource,
		NewNotificationMatrixResource,
		NewNotificationMattermostResource,
		NewNotificationMaxResource,
		NewNotificationNextcloudTalkResource,
		NewNotificationNotiferyResource,
		NewNotificationNostrResource,
		NewNotificationNtfyResource,
		NewNotificationOneBotResource,
		NewNotificationOneChatResource,
		NewNotificationOctopushResource,
		NewNotificationOnesenderResource,
		NewNotificationOoredooResource,
		NewNotificationOpsgenieResource,
		NewNotificationPagerDutyResource,
		NewNotificationPagerTreeResource,
		NewNotificationPlivoResource,
		NewNotificationPumbleResource,
		NewNotificationPushbulletResource,
		NewNotificationPromoSMSResource,
		NewNotificationPushDeerResource,
		NewNotificationPushoverResource,
		NewNotificationPushPlusResource,
		NewNotificationPushyResource,
		NewNotificationResendResource,
		NewNotificationRocketChatResource,
		NewNotificationSendgridResource,
		NewNotificationServerChanResource,
		NewNotificationSerwersmsResource,
		NewNotificationSevenioResource,
		NewNotificationSIGNL4Resource,
		NewNotificationSignalResource,
		NewNotificationSlackResource,
		NewNotificationSquadcastResource,
		NewNotificationSMSCResource,
		NewNotificationSMSEagleResource,
		NewNotificationSMSIRResource,
		NewNotificationSMSManagerResource,
		NewNotificationSMSPartnerResource,
		NewNotificationSMSPlanetResource,
		NewNotificationStackfieldResource,
		NewNotificationTechulusPushResource,
		NewNotificationThreemaResource,
		NewNotificationSplunkResource,
		NewNotificationSpugPushResource,
		NewNotificationSMTPResource,
		NewNotificationTeamsResource,
		NewNotificationTelegramResource,
		NewNotificationTelnyxResource,
		NewNotificationTeltonikaResource,
		NewNotificationTwilioResource,
		NewNotificationVKResource,
		NewNotificationWAHAResource,
		NewNotificationWhapiResource,
		NewNotificationWhatsapp360messengerResource,
		NewNotificationWPushResource,
		NewNotificationWebhookResource,
		NewNotificationWebpushResource,
		NewNotificationWeComResource,
		NewNotificationWxPusherResource,
		NewNotificationYZJResource,
		NewNotificationZohoCliqResource,
	}
}

// DataSources returns the list of data sources for the provider.
func (*UptimeKumaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	dataSources := notificationDataSources()

	dataSources = append(
		dataSources,
		NewMonitorHTTPDataSource,
		NewMonitorHTTPKeywordDataSource,
		NewMonitorGrpcKeywordDataSource,
		NewMonitorHTTPJSONQueryDataSource,
		NewMonitorWebsocketUpgradeDataSource,
		NewMonitorGroupDataSource,
		NewMonitorPingDataSource,
		NewMonitorDNSDataSource,
		NewMonitorSNMPDataSource,
		NewMonitorRadiusDataSource,
		NewMonitorKafkaProducerDataSource,
		NewMonitorPushDataSource,
		NewMonitorRealBrowserDataSource,
		NewMonitorPostgresDataSource,
		NewMonitorMySQLDataSource,
		NewMonitorOracleDBDataSource,
		NewMonitorMongoDBDataSource,
		NewMonitorRabbitMQDataSource,
		NewMonitorRedisDataSource,
		NewMonitorSQLServerDataSource,
		NewMonitorGameDigDataSource,
		NewMonitorGlobalpingDataSource,
		NewMonitorNTPDataSource,
		NewMonitorSteamDataSource,
		NewMonitorTailscalePingDataSource,
		NewMonitorTCPPortDataSource,
		NewMonitorSIPOptionsDataSource,
		NewMonitorSystemServiceDataSource,
		NewMonitorPM2DataSource,
		NewMonitorDockerDataSource,
		NewMonitorMQTTDataSource,
		NewMonitorSMTPDataSource,
		NewProxyDataSource,
		NewDockerHostDataSource,
		NewMaintenanceDataSource,
		NewMaintenanceMonitorsDataSource,
		NewMaintenanceStatusPagesDataSource,
		NewMaintenancesDataSource,
		NewSettingsDataSource,
		NewStatusPageDataSource,
		NewTagDataSource,
	)

	return dataSources
}

func notificationDataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNotificationDataSource,
		NewNotification46ElksDataSource,
		NewNotificationAlertaDataSource,
		NewNotificationAlertNowDataSource,
		NewNotificationAliyunsmsDataSource,
		NewNotificationAppriseDataSource,
		NewNotificationBaleDataSource,
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
		NewNotificationFlowtriqDataSource,
		NewNotificationFluxerDataSource,
		NewNotificationFreemobileDataSource,
		NewNotificationGoAlertDataSource,
		NewNotificationGoogleChatDataSource,
		NewNotificationGoogleSheetsDataSource,
		NewNotificationGotifyDataSource,
		NewNotificationGorushDataSource,
		NewNotificationGrafanaOncallDataSource,
		NewNotificationGTXMessagingDataSource,
		NewNotificationHaloPSADataSource,
		NewNotificationHeiiOnCallDataSource,
		NewNotificationHomeAssistantDataSource,
		NewNotificationJiraServiceManagementDataSource,
		NewNotificationKeepDataSource,
		NewNotificationKookDataSource,
		NewNotificationLineDataSource,
		NewNotificationLunaseaDataSource,
		NewNotificationMatrixDataSource,
		NewNotificationMattermostDataSource,
		NewNotificationMaxDataSource,
		NewNotificationNextcloudTalkDataSource,
		NewNotificationNotiferyDataSource,
		NewNotificationNostrDataSource,
		NewNotificationNtfyDataSource,
		NewNotificationOneBotDataSource,
		NewNotificationOneChatDataSource,
		NewNotificationOctopushDataSource,
		NewNotificationOnesenderDataSource,
		NewNotificationOoredooDataSource,
		NewNotificationOpsgenieDataSource,
		NewNotificationPagerDutyDataSource,
		NewNotificationPagerTreeDataSource,
		NewNotificationPlivoDataSource,
		NewNotificationPumbleDataSource,
		NewNotificationPushbulletDataSource,
		NewNotificationPromoSMSDataSource,
		NewNotificationPushDeerDataSource,
		NewNotificationPushoverDataSource,
		NewNotificationPushPlusDataSource,
		NewNotificationPushyDataSource,
		NewNotificationResendDataSource,
		NewNotificationRocketChatDataSource,
		NewNotificationSendgridDataSource,
		NewNotificationServerChanDataSource,
		NewNotificationSerwersmsDataSource,
		NewNotificationSevenioDataSource,
		NewNotificationSIGNL4DataSource,
		NewNotificationSignalDataSource,
		NewNotificationSlackDataSource,
		NewNotificationSquadcastDataSource,
		NewNotificationSMSCDataSource,
		NewNotificationSMSEagleDataSource,
		NewNotificationSMSIRDataSource,
		NewNotificationSMSManagerDataSource,
		NewNotificationSMSPartnerDataSource,
		NewNotificationSMSPlanetDataSource,
		NewNotificationStackfieldDataSource,
		NewNotificationTechulusPushDataSource,
		NewNotificationThreemaDataSource,
		NewNotificationSplunkDataSource,
		NewNotificationSpugPushDataSource,
		NewNotificationSMTPDataSource,
		NewNotificationTeamsDataSource,
		NewNotificationTelegramDataSource,
		NewNotificationTelnyxDataSource,
		NewNotificationTeltonikaDataSource,
		NewNotificationTwilioDataSource,
		NewNotificationVKDataSource,
		NewNotificationWAHADataSource,
		NewNotificationWhapiDataSource,
		NewNotificationWhatsapp360messengerDataSource,
		NewNotificationWPushDataSource,
		NewNotificationWebhookDataSource,
		NewNotificationWebpushDataSource,
		NewNotificationWeComDataSource,
		NewNotificationWxPusherDataSource,
		NewNotificationYZJDataSource,
		NewNotificationZohoCliqDataSource,
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
