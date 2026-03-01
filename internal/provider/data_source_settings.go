package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &SettingsDataSource{}

// NewSettingsDataSource returns a new instance of the settings data source.
func NewSettingsDataSource() datasource.DataSource {
	return &SettingsDataSource{}
}

// SettingsDataSource reads the current Uptime Kuma server settings.
type SettingsDataSource struct {
	client *kuma.Client
}

// SettingsDataSourceModel describes the data model for the settings data source.
type SettingsDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	ServerTimezone      types.String `tfsdk:"server_timezone"`
	KeepDataPeriodDays  types.Int64  `tfsdk:"keep_data_period_days"`
	CheckUpdate         types.Bool   `tfsdk:"check_update"`
	SearchEngineIndex   types.Bool   `tfsdk:"search_engine_index"`
	EntryPage           types.String `tfsdk:"entry_page"`
	NSCD                types.Bool   `tfsdk:"nscd"`
	TLSExpiryNotifyDays types.List   `tfsdk:"tls_expiry_notify_days"`
	TrustProxy          types.Bool   `tfsdk:"trust_proxy"`
	PrimaryBaseURL      types.String `tfsdk:"primary_base_url"`
	SteamAPIKey         types.String `tfsdk:"steam_api_key"`
	ChromeExecutable    types.String `tfsdk:"chrome_executable"`
}

// Metadata returns the metadata for the data source.
func (*SettingsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_settings"
}

// Schema returns the schema for the data source.
func (*SettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read the current Uptime Kuma server settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Settings identifier (always `settings`).",
			},
			"server_timezone": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server timezone (e.g. `Europe/Berlin`).",
			},
			"keep_data_period_days": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of days to keep monitoring data.",
			},
			"check_update": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether automatic update checks are enabled.",
			},
			"search_engine_index": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether search engines are allowed to index the status pages.",
			},
			"entry_page": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Entry page shown after login (e.g. `dashboard`).",
			},
			"nscd": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether NSCD (Name Service Cache Daemon) is enabled for DNS caching.",
			},
			"tls_expiry_notify_days": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Computed:            true,
				MarkdownDescription: "Days before TLS certificate expiry to send notifications.",
			},
			"trust_proxy": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether reverse proxy headers are trusted.",
			},
			"primary_base_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Primary base URL for the Uptime Kuma instance.",
			},
			"steam_api_key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Steam API key for Steam Game Server monitoring.",
			},
			"chrome_executable": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Path to Chrome/Chromium executable for Real Browser monitoring.",
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *SettingsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current server settings.
func (d *SettingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	s, err := d.client.GetSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to read settings", err.Error())
		return
	}

	var data SettingsDataSourceModel

	data.ID = types.StringValue("settings")
	data.ServerTimezone = types.StringValue(s.ServerTimezone)
	data.KeepDataPeriodDays = types.Int64Value(int64(s.KeepDataPeriodDays))
	data.CheckUpdate = types.BoolValue(s.CheckUpdate)
	data.SearchEngineIndex = types.BoolValue(s.SearchEngineIndex)
	data.EntryPage = types.StringValue(s.EntryPage)
	data.NSCD = types.BoolValue(s.NSCD)
	data.TrustProxy = types.BoolValue(s.TrustProxy)
	data.PrimaryBaseURL = types.StringValue(s.PrimaryBaseURL)
	data.SteamAPIKey = types.StringValue(s.SteamAPIKey)
	data.ChromeExecutable = types.StringValue(s.ChromeExecutable)

	tlsDays, tlsDiags := types.ListValueFrom(ctx, types.Int64Type, s.TLSExpiryNotifyDays)
	resp.Diagnostics.Append(tlsDiags...)
	data.TLSExpiryNotifyDays = tlsDays

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
