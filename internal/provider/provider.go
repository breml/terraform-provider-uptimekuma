// Package provider implements the Uptime Kuma Terraform provider.
package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"

	"github.com/breml/terraform-provider-uptimekuma/internal/client"
)

// Ensure UptimeKumaProvider satisfies various provider interfaces.
var (
	_ provider.Provider                   = &UptimeKumaProvider{}
	_ provider.ProviderWithValidateConfig = &UptimeKumaProvider{}
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
	SessionToken      types.String `tfsdk:"session_token"`
	TOTPSecret        types.String `tfsdk:"totp_secret"`
	AuthScheme        types.String `tfsdk:"auth_scheme"`
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
				MarkdownDescription: "Uptime Kuma username. Can be set via `UPTIMEKUMA_USERNAME` environment variable, " +
					"which applies when the attribute is unset or empty.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Uptime Kuma password. Can be set via `UPTIMEKUMA_PASSWORD` environment variable, " +
					"which applies when the attribute is unset or empty.",
				Optional:  true,
				Sensitive: true,
			},
			"session_token": schema.StringAttribute{
				MarkdownDescription: sessionTokenDescription,
				Optional:            true,
				Sensitive:           true,
			},
			"totp_secret": schema.StringAttribute{
				MarkdownDescription: totpSecretDescription,
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					totpSecretValidator{},
				},
			},
			"auth_scheme": schema.StringAttribute{
				MarkdownDescription: authSchemeDescription,
				Optional:            true,
				Validators: []validator.String{
					authSchemeValidator{},
				},
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
	applyEnvironmentDefaults(&data, &resp.Diagnostics)

	// Endpoint is always required to connect to Uptime Kuma.
	if data.Endpoint.IsNull() {
		resp.Diagnostics.AddError("endpoint required", "endpoint is required")
	}

	// A credential that is still unknown would reach the client as an empty
	// string, so it is an error here rather than a login with nothing.
	rejectUnknownCredentials(&data, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	// The credential shape is judged here and not in ValidateConfig, see
	// there: ValidateConfig cannot tell `terraform validate` from a plan, and
	// only here is every credential known. The auth scheme is resolved here
	// because the client configuration needs it.
	validateCredentials(&data, &resp.Diagnostics)
	authScheme := resolveAuthScheme(&data, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	opts := parseClientOptions(&data, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	opts.authScheme = authScheme

	kumaClient := connect(data.Endpoint.ValueString(), buildClientConfig(&data, opts), client.New, &resp.Diagnostics)
	if kumaClient == nil {
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

// newClientFunc creates the client, see client.New. connect takes it as a
// parameter so that a test can stand in for the server.
type newClientFunc func(ctx context.Context, config *client.Config) (*kuma.Client, error)

// connect creates the client for config, and reports a configured session
// token the server refused and the password login took over from. It returns
// nil, with an error in diags, when the connection fails.
//
// The rejection is reported on failure too: the password login may have taken
// over from the refused token before the connection failed for another
// reason, and the connection error alone would not say that the stored token
// has died.
func connect(endpoint string, config *client.Config, newClient newClientFunc, diags *diag.Diagnostics) *kuma.Client {
	// The callback runs on this goroutine before newClient returns, so the
	// rejection is in hand by the time it is read.
	var sessionTokenRejection error

	config.OnSessionTokenRejected = func(err error) { sessionTokenRejection = err }

	// Context is not the one of Configure: Terraform cancels that once
	// Configure returns, and the connection must outlive it.
	kumaClient, err := newClient(context.Background(), config)

	warnRejectedSessionToken(sessionTokenRejection, diags)

	if err != nil {
		diags.AddError(connectionError(endpoint, err))

		return nil
	}

	return kumaClient
}

// connectionError returns the summary and detail of the diagnostic for a
// failed connection to Uptime Kuma. A login failure the client reports with a
// typed error gets a diagnostic of its own that names the next step, since
// transport advice does not help with a rejected password. Every other error,
// including a login rejection the client has no typed error for, gets the
// generic connection advice, and a malformed TOTP secret, which never reached
// the server, a diagnostic of its own.
func connectionError(endpoint string, err error) (summary string, detail string) {
	// The secret is rejected before any connection is attempted, so neither
	// the login nor the transport advice fits it. The schema validator catches
	// a configured one at plan time; this is the one from the environment.
	if errors.Is(err, kuma.ErrInvalidTOTPSecret) {
		return "Invalid TOTP secret", fmt.Sprintf(
			"The TOTP secret in `totp_secret` (or the environment variable `UPTIMEKUMA_TOTP_SECRET`) is "+
				"not a base32 shared secret, so no one-time code can be derived from it. Copy the `secret` "+
				"out of the `otpauth://` URI Uptime Kuma showed while two-factor authentication was set "+
				"up.\n\nUnderlying error: %v",
			err,
		)
	}

	summary, nextStep := loginErrorAdvice(err)
	if summary == "" {
		return "failed to connect to Uptime Kuma", transportErrorDetail(endpoint, err)
	}

	return summary, fmt.Sprintf(
		"Uptime Kuma at %q refused the login.\n\n"+
			"Underlying error: %v\n\n"+
			"%s",
		endpoint,
		err,
		nextStep,
	)
}

// loginErrorAdvice returns the summary and the next step for a login failure
// the client reports with a typed error, or empty strings for any other error.
//
// The order matters twice over. A refused session token followed by a failed
// password login arrives as one error wrapping both sentinels, which
// passwordFallbackAdvice describes ahead of the switch. And kuma.ErrUserInactive
// wraps kuma.ErrInvalidSessionToken, so it is matched first.
func loginErrorAdvice(err error) (summary string, nextStep string) {
	summary, nextStep = passwordFallbackAdvice(err)
	if summary != "" {
		return summary, nextStep
	}

	var rateLimited *kuma.RateLimitedError

	switch {
	case errors.Is(err, kuma.ErrInvalidCredentials):
		return "Uptime Kuma rejected the username or password",
			"Check the provider attributes `username` and `password` (or the environment variables " +
				"`UPTIMEKUMA_USERNAME` and `UPTIMEKUMA_PASSWORD`). The login is not retried, because every " +
				"attempt counts against the server's login rate limit."

	case errors.Is(err, kuma.ErrTwoFactorRequired):
		return "Uptime Kuma requires two-factor authentication",
			"The account has two-factor authentication (2FA) enabled, so the login needs a one-time code. " +
				"Set the TOTP secret of the account in `totp_secret` (or the environment variable " +
				"`UPTIMEKUMA_TOTP_SECRET`) for the provider to derive the code from, or authenticate with " +
				"`session_token`, which needs no code."

	case errors.Is(err, kuma.ErrInvalidTOTPCode):
		return "Uptime Kuma rejected the two-factor authentication code",
			"The one-time code derived from the TOTP secret in `totp_secret` was refused. Uptime Kuma " +
				"2.x refuses a code it has already accepted for the account, and every Terraform command logs " +
				"in on its own: `terraform apply` alone logs in twice, for the plan and for the apply, and " +
				"the second login often falls into the 30 second step of the first. The provider then waits " +
				"for the next step, which fails when `timeout` or `per_attempt_timeout` cannot cover the " +
				"wait: set `timeout` to at least 60s, and leave `per_attempt_timeout` unset or above 35s. " +
				"A code refused in two " +
				"consecutive steps points at the secret itself or the clock of this host. A `session_token` " +
				"next to the password needs no code at all."

	case errors.Is(err, kuma.ErrTwoFactorLocked):
		return "Uptime Kuma locked the two-factor authentication",
			"The server refuses further two-factor codes for this login after too many wrong ones. Check " +
				"the TOTP secret in `totp_secret` and the clock of this host, then run Terraform again, or " +
				"authenticate with `session_token`, which needs no code."

	case errors.Is(err, kuma.ErrUserInactive):
		return "Uptime Kuma account is inactive",
			"The account is deactivated or deleted. Reactivate it in Uptime Kuma or configure the credentials " +
				"of an active account."

	case errors.Is(err, kuma.ErrInvalidSessionToken):
		return "Uptime Kuma rejected the session token",
			"The server rejected the configured `session_token`. On Uptime Kuma 2.x that happens when the " +
				"password of the account was changed meanwhile; on Uptime Kuma 3.x the token also lapses " +
				"after 7 days without use. Replace it with the token of a fresh login, or set the provider " +
				"attributes `username` and `password` next to it (or the environment variables " +
				"`UPTIMEKUMA_USERNAME` and `UPTIMEKUMA_PASSWORD`) for the provider to fall back on."

	case errors.As(err, &rateLimited) && rateLimited.RetryAfter > 0:
		return "Uptime Kuma rate limited the login",
			fmt.Sprintf("Too many logins were attempted from this address. Wait %s before running "+
				"Terraform again.", rateLimited.RetryAfter)

	case errors.Is(err, kuma.ErrRateLimited):
		return "Uptime Kuma rate limited the login",
			"Too many logins were attempted from this address. Wait a minute before running Terraform again."

	case errors.Is(err, kuma.ErrAuthRequired):
		return "Uptime Kuma requires authentication",
			"The server requires a login, but no credentials are configured. Set the provider attributes " +
				"`username` and `password` (or the environment variables `UPTIMEKUMA_USERNAME` and " +
				"`UPTIMEKUMA_PASSWORD`), or `session_token` (or `UPTIMEKUMA_SESSION_TOKEN`)."
	}

	return "", ""
}

// transportErrorDetail returns the diagnostic detail for a connection that
// failed for a reason other than a typed login failure. That is mostly a
// transport problem, but can be a login rejection without a typed error too.
func transportErrorDetail(endpoint string, err error) string {
	return fmt.Sprintf(
		"Could not connect to or log in to Uptime Kuma at %q.\n\n"+
			"Underlying error: %v\n\n"+
			"Common causes:\n"+
			"  - The endpoint URL is incorrect or Uptime Kuma is not reachable from this host.\n"+
			"  - A reverse proxy is not forwarding the Socket.IO/WebSocket connection.\n"+
			"  - TLS certificate issues when using a custom domain (try the direct URL to confirm).\n"+
			"  - The server refused the login for a reason the provider does not recognize, see the "+
			"underlying error.\n\n"+
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

	// authScheme is the resolved `auth_scheme`, see resolveAuthScheme.
	authScheme kuma.AuthScheme
}

// buildClientConfig turns the provider model, with the environment defaults
// applied and validated, and the parsed connection options into the
// configuration of the client.
func buildClientConfig(data *UptimeKumaProviderModel, opts clientOptions) *client.Config {
	return &client.Config{
		Endpoint:             data.Endpoint.ValueString(),
		Username:             data.Username.ValueString(),
		Password:             data.Password.ValueString(),
		SessionToken:         data.SessionToken.ValueString(),
		TOTPSecret:           data.TOTPSecret.ValueString(),
		AuthScheme:           opts.authScheme,
		EnableConnectionPool: true,
		LogLevel:             kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
		ConnectTimeout:       opts.connectTimeout,
		PerAttemptTimeout:    opts.perAttemptTimeout,
		OperationTimeout:     opts.operationTimeout,
		MaxRetries:           opts.maxRetries,
	}
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
func applyEnvironmentDefaults(data *UptimeKumaProviderModel, diags *diag.Diagnostics) {
	envEndpoint := os.Getenv("UPTIMEKUMA_ENDPOINT")
	if data.Endpoint.IsNull() && envEndpoint != "" {
		data.Endpoint = types.StringValue(envEndpoint)
	}

	applyCredentialEnvironmentDefaults(data)

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
			diags.AddWarning(
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
		NewMonitorManualResource,
		NewMonitorSFTPResource,
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

// notificationResources returns the notification resources. The list is split alphabetically
// into two functions to stay within the function length limit.
func notificationResources() []func() resource.Resource {
	return append(notificationResourcesAToM(), notificationResourcesNToZ()...)
}

func notificationResourcesAToM() []func() resource.Resource {
	return []func() resource.Resource{
		NewNotificationResource,
		NewNotification46ElksResource,
		NewNotificationAlertaResource,
		NewNotificationAlertNowResource,
		NewNotificationAliyunsmsResource,
		NewNotificationAmootSMSResource,
		NewNotificationAppriseResource,
		NewNotificationBaleResource,
		NewNotificationBarkResource,
		NewNotificationBearSMSResource,
		NewNotificationBitrix24Resource,
		NewNotificationBrevoResource,
		NewNotificationCallMeBotResource,
		NewNotificationCellsyntResource,
		NewNotificationClickUpResource,
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
		NewNotificationMilkyResource,
	}
}

func notificationResourcesNToZ() []func() resource.Resource {
	return []func() resource.Resource{
		NewNotificationNextcloudTalkResource,
		NewNotificationNotiferyResource,
		NewNotificationNotifyAppResource,
		NewNotificationNostrResource,
		NewNotificationNtfyResource,
		NewNotificationOneBotResource,
		NewNotificationOneChatResource,
		NewNotificationOctopushResource,
		NewNotificationOnesenderResource,
		NewNotificationOoredooResource,
		NewNotificationOpenWaResource,
		NewNotificationOpsgenieResource,
		NewNotificationPagerDutyResource,
		NewNotificationPagerTreeResource,
		NewNotificationPingletResource,
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
		NewNotificationSignalgridResource,
		NewNotificationSlackResource,
		NewNotificationSquadcastResource,
		NewNotificationSMSCResource,
		NewNotificationSMSEagleResource,
		NewNotificationSMSGatewayResource,
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
		NewNotificationTurboSMTPResource,
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
		NewMonitorManualDataSource,
		NewMonitorSFTPDataSource,
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

// notificationDataSources returns the notification data sources. The list is split alphabetically
// into two functions to stay within the function length limit.
func notificationDataSources() []func() datasource.DataSource {
	return append(notificationDataSourcesAToM(), notificationDataSourcesNToZ()...)
}

func notificationDataSourcesAToM() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNotificationDataSource,
		NewNotification46ElksDataSource,
		NewNotificationAlertaDataSource,
		NewNotificationAlertNowDataSource,
		NewNotificationAliyunsmsDataSource,
		NewNotificationAmootSMSDataSource,
		NewNotificationAppriseDataSource,
		NewNotificationBaleDataSource,
		NewNotificationBarkDataSource,
		NewNotificationBearSMSDataSource,
		NewNotificationBitrix24DataSource,
		NewNotificationBrevoDataSource,
		NewNotificationCallMeBotDataSource,
		NewNotificationCellsyntDataSource,
		NewNotificationClickUpDataSource,
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
		NewNotificationMilkyDataSource,
	}
}

func notificationDataSourcesNToZ() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNotificationNextcloudTalkDataSource,
		NewNotificationNotiferyDataSource,
		NewNotificationNotifyAppDataSource,
		NewNotificationNostrDataSource,
		NewNotificationNtfyDataSource,
		NewNotificationOneBotDataSource,
		NewNotificationOneChatDataSource,
		NewNotificationOctopushDataSource,
		NewNotificationOnesenderDataSource,
		NewNotificationOoredooDataSource,
		NewNotificationOpenWaDataSource,
		NewNotificationOpsgenieDataSource,
		NewNotificationPagerDutyDataSource,
		NewNotificationPagerTreeDataSource,
		NewNotificationPingletDataSource,
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
		NewNotificationSignalgridDataSource,
		NewNotificationSlackDataSource,
		NewNotificationSquadcastDataSource,
		NewNotificationSMSCDataSource,
		NewNotificationSMSEagleDataSource,
		NewNotificationSMSGatewayDataSource,
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
		NewNotificationTurboSMTPDataSource,
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
