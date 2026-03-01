package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/settings"
)

var (
	_ resource.Resource                = &SettingsResource{}
	_ resource.ResourceWithImportState = &SettingsResource{}
)

// NewSettingsResource returns a new instance of the settings resource.
func NewSettingsResource() resource.Resource {
	return &SettingsResource{}
}

// SettingsResource defines the singleton resource for Uptime Kuma server settings.
type SettingsResource struct {
	client   *kuma.Client
	password string
}

// SettingsResourceModel describes the resource data model.
type SettingsResourceModel struct {
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

// Metadata returns the metadata for the resource.
func (*SettingsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_settings"
}

// Schema returns the schema for the resource.
func (*SettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Uptime Kuma server settings. This is a singleton resource " +
			"— only one instance should exist per Uptime Kuma server. " +
			"Deleting this resource only removes it from Terraform state; " +
			"the settings persist on the server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Settings identifier (always `settings`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_timezone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Server timezone (e.g. `Europe/Berlin`).",
			},
			"keep_data_period_days": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Number of days to keep monitoring data.",
			},
			"check_update": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable automatic update checks.",
			},
			"search_engine_index": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Allow search engines to index the status pages.",
			},
			"entry_page": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Entry page shown after login (e.g. `dashboard`).",
			},
			"nscd": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable NSCD (Name Service Cache Daemon) for DNS caching.",
			},
			"tls_expiry_notify_days": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Days before TLS certificate expiry to send notifications (e.g. `[7, 14, 21]`).",
			},
			"trust_proxy": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Trust reverse proxy headers.",
			},
			"primary_base_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Primary base URL for the Uptime Kuma instance.",
			},
			"steam_api_key": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Steam API key for Steam Game Server monitoring.",
			},
			"chrome_executable": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Path to Chrome/Chromium executable for Real Browser monitoring.",
			},
		},
	}
}

// Configure configures the resource with the API client and password.
func (r *SettingsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	pd, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf(
				"Expected *providerData, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	r.client = pd.client
	r.password = pd.password
}

// Create applies the planned settings to the server.
func (r *SettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SettingsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If no attributes are explicitly set, adopt existing server settings
	// without performing a write (import-like behavior).
	if hasExplicitlySetFields(&data) {
		r.applySettings(ctx, &data, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		r.readSettings(ctx, &data, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	data.ID = types.StringValue("settings")

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current server settings.
func (r *SettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SettingsResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readSettings(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies the planned settings to the server.
func (r *SettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SettingsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applySettings(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete is a no-op for singleton settings — it only removes from Terraform state.
func (*SettingsResource) Delete(
	_ context.Context,
	_ resource.DeleteRequest,
	_ *resource.DeleteResponse,
) {
	// Settings are a server-wide singleton and cannot be deleted.
	// Removing from state is sufficient.
}

// ImportState imports the settings resource into Terraform state.
func (*SettingsResource) ImportState(
	ctx context.Context,
	_ resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Set the synthetic ID; Read() will populate the rest.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "settings")...)
}

// hasExplicitlySetFields reports whether any user-configurable attribute
// in the plan is explicitly set (non-null and non-unknown).
func hasExplicitlySetFields(data *SettingsResourceModel) bool {
	return (!data.ServerTimezone.IsNull() && !data.ServerTimezone.IsUnknown()) ||
		(!data.KeepDataPeriodDays.IsNull() && !data.KeepDataPeriodDays.IsUnknown()) ||
		(!data.CheckUpdate.IsNull() && !data.CheckUpdate.IsUnknown()) ||
		(!data.SearchEngineIndex.IsNull() && !data.SearchEngineIndex.IsUnknown()) ||
		(!data.EntryPage.IsNull() && !data.EntryPage.IsUnknown()) ||
		(!data.NSCD.IsNull() && !data.NSCD.IsUnknown()) ||
		(!data.TLSExpiryNotifyDays.IsNull() && !data.TLSExpiryNotifyDays.IsUnknown()) ||
		(!data.TrustProxy.IsNull() && !data.TrustProxy.IsUnknown()) ||
		(!data.PrimaryBaseURL.IsNull() && !data.PrimaryBaseURL.IsUnknown()) ||
		(!data.SteamAPIKey.IsNull() && !data.SteamAPIKey.IsUnknown()) ||
		(!data.ChromeExecutable.IsNull() && !data.ChromeExecutable.IsUnknown())
}

// applySettings reads current settings from the server, merges
// user-specified values from the plan, and writes them back.
func (r *SettingsResource) applySettings(
	ctx context.Context,
	data *SettingsResourceModel,
	diags *diag.Diagnostics,
) {
	if r.password == "" {
		diags.AddError(
			"password required",
			"The provider password must be configured to update settings. "+
				"Set the password in the provider configuration or via the UPTIMEKUMA_PASSWORD environment variable.",
		)

		return
	}

	// Read current settings to preserve values the user did not specify.
	current, err := r.client.GetSettings(ctx)
	if err != nil {
		diags.AddError("failed to read settings", fmt.Sprintf("could not read current settings: %s", err.Error()))
		return
	}

	merged := mergeSettingsFromPlan(ctx, data, current, diags)
	if diags.HasError() {
		return
	}

	err = r.client.SetSettings(ctx, merged, r.password)
	if err != nil {
		diags.AddError("failed to update settings", fmt.Sprintf("failed to update settings: %s", err.Error()))
		return
	}

	// Re-read to capture any server-side normalisation.
	r.readSettings(ctx, data, diags)
}

// readSettings fetches settings from the server and populates the model.
func (r *SettingsResource) readSettings(
	ctx context.Context,
	data *SettingsResourceModel,
	diags *diag.Diagnostics,
) {
	s, err := r.client.GetSettings(ctx)
	if err != nil {
		diags.AddError("failed to read settings", err.Error())
		return
	}

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
	diags.Append(tlsDiags...)
	data.TLSExpiryNotifyDays = tlsDays
}

// mergeSettingsFromPlan overlays user-specified plan values onto
// the current server settings, preserving unspecified fields.
func mergeSettingsFromPlan(
	ctx context.Context,
	data *SettingsResourceModel,
	current *settings.Settings,
	diags *diag.Diagnostics,
) settings.Settings {
	merged := *current

	mergeStringFields(data, &merged)
	mergeBoolFields(data, &merged)
	mergeNumericAndListFields(ctx, data, &merged, diags)

	return merged
}

// mergeStringFields overlays string plan values onto the merged settings.
func mergeStringFields(data *SettingsResourceModel, merged *settings.Settings) {
	if !data.ServerTimezone.IsNull() && !data.ServerTimezone.IsUnknown() {
		merged.ServerTimezone = data.ServerTimezone.ValueString()
	}

	if !data.EntryPage.IsNull() && !data.EntryPage.IsUnknown() {
		merged.EntryPage = data.EntryPage.ValueString()
	}

	if !data.PrimaryBaseURL.IsNull() && !data.PrimaryBaseURL.IsUnknown() {
		merged.PrimaryBaseURL = data.PrimaryBaseURL.ValueString()
	}

	if !data.SteamAPIKey.IsNull() && !data.SteamAPIKey.IsUnknown() {
		merged.SteamAPIKey = data.SteamAPIKey.ValueString()
	}

	if !data.ChromeExecutable.IsNull() && !data.ChromeExecutable.IsUnknown() {
		merged.ChromeExecutable = data.ChromeExecutable.ValueString()
	}
}

// mergeBoolFields overlays boolean plan values onto the merged settings.
func mergeBoolFields(data *SettingsResourceModel, merged *settings.Settings) {
	if !data.CheckUpdate.IsNull() && !data.CheckUpdate.IsUnknown() {
		merged.CheckUpdate = data.CheckUpdate.ValueBool()
	}

	if !data.SearchEngineIndex.IsNull() && !data.SearchEngineIndex.IsUnknown() {
		merged.SearchEngineIndex = data.SearchEngineIndex.ValueBool()
	}

	if !data.NSCD.IsNull() && !data.NSCD.IsUnknown() {
		merged.NSCD = data.NSCD.ValueBool()
	}

	if !data.TrustProxy.IsNull() && !data.TrustProxy.IsUnknown() {
		merged.TrustProxy = data.TrustProxy.ValueBool()
	}
}

// mergeNumericAndListFields overlays numeric and list plan values onto the merged settings.
func mergeNumericAndListFields(
	ctx context.Context,
	data *SettingsResourceModel,
	merged *settings.Settings,
	diags *diag.Diagnostics,
) {
	if !data.KeepDataPeriodDays.IsNull() && !data.KeepDataPeriodDays.IsUnknown() {
		merged.KeepDataPeriodDays = int(data.KeepDataPeriodDays.ValueInt64())
	}

	if !data.TLSExpiryNotifyDays.IsNull() && !data.TLSExpiryNotifyDays.IsUnknown() {
		var days64 []int64
		diags.Append(data.TLSExpiryNotifyDays.ElementsAs(ctx, &days64, false)...)
		if diags.HasError() {
			return
		}

		days := make([]int, len(days64))
		for i, d := range days64 {
			days[i] = int(d)
		}

		merged.TLSExpiryNotifyDays = days
	}
}
