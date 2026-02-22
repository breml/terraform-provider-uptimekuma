package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/statuspage"
)

var _ resource.Resource = &StatusPageResource{}

// statusPageIconValidator validates the icon field format.
type statusPageIconValidator struct{}

// Description returns a plain text description of the validator's behavior.
func (statusPageIconValidator) Description(_ context.Context) string {
	return "string must be a data:image/png;base64,... data URI or a URL/path (max 255 characters)"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (statusPageIconValidator) MarkdownDescription(_ context.Context) string {
	return "string must be a `data:image/png;base64,...` data URI or a URL/path (max 255 characters)"
}

// ValidateString checks that the provided string value is a valid status page icon.
func (statusPageIconValidator) ValidateString(
	_ context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if strings.HasPrefix(value, "data:") {
		if !strings.HasPrefix(value, "data:image/png;base64,") {
			resp.Diagnostics.Append(
				diag.NewAttributeErrorDiagnostic(
					req.Path,
					"Invalid icon data URI",
					"Icon data URI must use PNG format: data:image/png;base64,...",
				),
			)
		}

		return
	}

	const maxIconPathLength = 255
	if len(value) > maxIconPathLength {
		resp.Diagnostics.Append(
			diag.NewAttributeErrorDiagnostic(
				req.Path,
				"Icon URL/path too long",
				fmt.Sprintf(
					"Icon URL/path must be at most %d characters, got %d."+
						" Use a data:image/png;base64,... data URI for inline images.",
					maxIconPathLength,
					len(value),
				),
			),
		)
	}
}

func validateStatusPageIcon() validator.String {
	return statusPageIconValidator{}
}

// NewStatusPageResource returns a new instance of the status page resource.
func NewStatusPageResource() resource.Resource {
	return &StatusPageResource{}
}

// StatusPageResource defines the resource implementation.
type StatusPageResource struct {
	client *kuma.Client
}

// StatusPageResourceModel describes the resource data model.
type StatusPageResourceModel struct {
	ID                    types.Int64  `tfsdk:"id"`
	Slug                  types.String `tfsdk:"slug"`
	Title                 types.String `tfsdk:"title"`
	Description           types.String `tfsdk:"description"`
	Icon                  types.String `tfsdk:"icon"`
	Theme                 types.String `tfsdk:"theme"`
	Published             types.Bool   `tfsdk:"published"`
	ShowTags              types.Bool   `tfsdk:"show_tags"`
	DomainNameList        types.List   `tfsdk:"domain_name_list"`
	GoogleAnalyticsID     types.String `tfsdk:"google_analytics_id"`
	CustomCSS             types.String `tfsdk:"custom_css"`
	FooterText            types.String `tfsdk:"footer_text"`
	ShowPoweredBy         types.Bool   `tfsdk:"show_powered_by"`
	ShowCertificateExpiry types.Bool   `tfsdk:"show_certificate_expiry"`
	PublicGroupList       types.List   `tfsdk:"public_group_list"`
}

// PublicGroupModel describes a public group in a status page.
type PublicGroupModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Weight      types.Int64  `tfsdk:"weight"`
	MonitorList types.List   `tfsdk:"monitor_list"`
}

// PublicMonitorModel describes a monitor in a public group.
type PublicMonitorModel struct {
	ID      types.Int64 `tfsdk:"id"`
	SendURL types.Bool  `tfsdk:"send_url"`
}

// Metadata returns the metadata for the resource.
func (*StatusPageResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

// Schema returns the schema for the resource.
//
//nolint:revive // function length is acceptable for Terraform provider schema definitions
func (*StatusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Status page resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Status page ID",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-friendly slug for the status page",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Display title",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Status page description",
				Optional:            true,
			},
			"icon": schema.StringAttribute{
				MarkdownDescription: "Icon for the status page. Accepts a PNG data URI (`data:image/png;base64,...`)" +
					" or a URL/path (max 255 characters). When a data URI is provided," +
					" Uptime Kuma converts it to a file on disk.",
				Optional:   true,
				Validators: []validator.String{validateStatusPageIcon()},
			},
			"theme": schema.StringAttribute{
				MarkdownDescription: "Theme name for styling",
				Optional:            true,
			},
			"published": schema.BoolAttribute{
				MarkdownDescription: "Whether page is publicly visible",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"show_tags": schema.BoolAttribute{
				MarkdownDescription: "Show monitor tags on status page",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"domain_name_list": schema.ListAttribute{
				MarkdownDescription: "Custom domain names",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"google_analytics_id": schema.StringAttribute{
				MarkdownDescription: "Google Analytics tracking ID",
				Optional:            true,
			},
			"custom_css": schema.StringAttribute{
				MarkdownDescription: "Custom CSS styling",
				Optional:            true,
			},
			"footer_text": schema.StringAttribute{
				MarkdownDescription: "Footer content",
				Optional:            true,
			},
			"show_powered_by": schema.BoolAttribute{
				MarkdownDescription: "Display 'Powered by Uptime Kuma'",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"show_certificate_expiry": schema.BoolAttribute{
				MarkdownDescription: "Show certificate expiry dates",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"public_group_list": schema.ListNestedAttribute{
				MarkdownDescription: "Monitor grouping configuration",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Public group ID",
							Computed:            true,
							Optional:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Group display name",
							Required:            true,
						},
						"weight": schema.Int64Attribute{
							MarkdownDescription: "Display order/weight",
							Optional:            true,
						},
						"monitor_list": schema.ListNestedAttribute{
							MarkdownDescription: "Monitors in group",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.Int64Attribute{
										MarkdownDescription: "Monitor ID",
										Required:            true,
									},
									"send_url": schema.BoolAttribute{
										MarkdownDescription: "Include monitor URL in status page",
										Optional:            true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure configures the resource with the API client.
func (r *StatusPageResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	r.client = client
}

// Create creates a new status page resource.
// First creates the base status page via AddStatusPage, then saves configuration with groups/domains.
func (r *StatusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Extract planned configuration from Terraform.
	var data StatusPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// First, create the base status page. Some fields like domain names and groups
	// are managed separately via SaveStatusPage, so we start with title and slug.
	err := r.client.AddStatusPage(ctx, data.Title.ValueString(), data.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create status page", err.Error())
		return
	}

	// Build the status page object with all configuration.
	sp := &statuspage.StatusPage{
		Slug:                  data.Slug.ValueString(),
		Title:                 data.Title.ValueString(),
		Description:           data.Description.ValueString(),
		Icon:                  data.Icon.ValueString(),
		Theme:                 data.Theme.ValueString(),
		Published:             data.Published.ValueBool(),
		ShowTags:              data.ShowTags.ValueBool(),
		GoogleAnalyticsID:     data.GoogleAnalyticsID.ValueString(),
		CustomCSS:             data.CustomCSS.ValueString(),
		FooterText:            data.FooterText.ValueString(),
		ShowPoweredBy:         data.ShowPoweredBy.ValueBool(),
		ShowCertificateExpiry: data.ShowCertificateExpiry.ValueBool(),
		DomainNameList:        []string{},
		PublicGroupList:       []statuspage.PublicGroup{},
	}

	// Populate domain names from configuration.
	r.populateStatusPageDomainNames(ctx, &data, sp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate public groups (monitor groupings) from configuration.
	r.populateStatusPagePublicGroups(ctx, &data, sp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the complete status page configuration and get back group IDs.
	savedGroups, err := r.client.SaveStatusPage(ctx, sp)
	if err != nil {
		resp.Diagnostics.AddError("failed to save status page", err.Error())
		return
	}

	// Retrieve the saved status page to get the assigned ID.
	retrievedSP, err := r.client.GetStatusPage(ctx, data.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read status page after creation", err.Error())
		return
	}

	data.ID = types.Int64Value(retrievedSP.ID)

	updateStatusPageDataAfterSave(ctx, &data, savedGroups, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the resource.
func (r *StatusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StatusPageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sp, err := r.client.GetStatusPage(ctx, data.Slug.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read status page", err.Error())
		return
	}

	data.ID = types.Int64Value(sp.ID)
	data.Title = types.StringValue(sp.Title)
	data.Description = stringOrNullPreserveEmpty(sp.Description, data.Description)

	if !data.Icon.IsNull() && !strings.HasPrefix(data.Icon.ValueString(), "data:") {
		data.Icon = stringOrNullPreserveEmpty(sp.Icon, data.Icon)
	}

	data.Theme = stringOrNullPreserveEmpty(sp.Theme, data.Theme)

	// Note: The Uptime Kuma API's saveStatusPage endpoint does not actually update
	// the published, show_tags, show_powered_by, and show_certificate_expiry fields
	// (see server/socket-handlers/status-page-socket-handler.js line 160-167).
	// Therefore, we don't update these fields from the API response to avoid drift.
	// We keep whatever values are in the Terraform config/state.

	data.GoogleAnalyticsID = stringOrNullPreserveEmpty(sp.GoogleAnalyticsID, data.GoogleAnalyticsID)
	data.CustomCSS = stringOrNullPreserveEmpty(sp.CustomCSS, data.CustomCSS)
	data.FooterText = stringOrNullPreserveEmpty(sp.FooterText, data.FooterText)

	if len(sp.DomainNameList) > 0 {
		domainNames, diags := types.ListValueFrom(ctx, types.StringType, sp.DomainNameList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.DomainNameList = domainNames
	} else {
		data.DomainNameList = types.ListNull(types.StringType)
	}

	// Note: public_group_list is managed entirely by the provider through Create and Update.
	// We do NOT try to read it back from the server because:
	// 1. GetStatusPage doesn't return it (see comment at line 28-29 of statuspage.go)
	// 2. GetStatusPages returns cached data that doesn't include monitor associations
	// Therefore, we preserve whatever is in the current state without modification.
	//
	// The public_group_list in state comes from Create/Update operations which call
	// SaveStatusPage and receive the group IDs in the response.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *StatusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StatusPageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sp := buildStatusPageFromModel(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	savedGroups, err := r.client.SaveStatusPage(ctx, sp)
	if err != nil {
		resp.Diagnostics.AddError("failed to update status page", err.Error())
		return
	}

	updateStatusPageDataAfterSave(ctx, &data, savedGroups, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildStatusPageFromModel(
	ctx context.Context,
	data *StatusPageResourceModel,
	diags *diag.Diagnostics,
) *statuspage.StatusPage {
	sp := &statuspage.StatusPage{
		Slug:                  data.Slug.ValueString(),
		Title:                 data.Title.ValueString(),
		Description:           data.Description.ValueString(),
		Icon:                  data.Icon.ValueString(),
		Theme:                 data.Theme.ValueString(),
		Published:             data.Published.ValueBool(),
		ShowTags:              data.ShowTags.ValueBool(),
		GoogleAnalyticsID:     data.GoogleAnalyticsID.ValueString(),
		CustomCSS:             data.CustomCSS.ValueString(),
		FooterText:            data.FooterText.ValueString(),
		ShowPoweredBy:         data.ShowPoweredBy.ValueBool(),
		ShowCertificateExpiry: data.ShowCertificateExpiry.ValueBool(),
		DomainNameList:        []string{},
		PublicGroupList:       []statuspage.PublicGroup{},
	}

	if !data.DomainNameList.IsNull() {
		var domainNames []string
		diags.Append(data.DomainNameList.ElementsAs(ctx, &domainNames, false)...)
		if diags.HasError() {
			return sp
		}

		sp.DomainNameList = domainNames
	}

	if !data.PublicGroupList.IsNull() {
		var groups []PublicGroupModel
		diags.Append(data.PublicGroupList.ElementsAs(ctx, &groups, false)...)
		if diags.HasError() {
			return sp
		}

		sp.PublicGroupList = convertGroupModelsToAPI(ctx, groups, diags)
	}

	return sp
}

func convertGroupModelsToAPI(
	ctx context.Context,
	groups []PublicGroupModel,
	diags *diag.Diagnostics,
) []statuspage.PublicGroup {
	publicGroups := make([]statuspage.PublicGroup, len(groups))

	for i, group := range groups {
		publicGroup := statuspage.PublicGroup{
			Name:        group.Name.ValueString(),
			Weight:      int(group.Weight.ValueInt64()),
			MonitorList: []statuspage.PublicMonitor{},
		}
		if !group.ID.IsNull() {
			publicGroup.ID = group.ID.ValueInt64()
		}

		if !group.MonitorList.IsNull() {
			var monitors []PublicMonitorModel
			diags.Append(group.MonitorList.ElementsAs(ctx, &monitors, false)...)
			if diags.HasError() {
				return publicGroups
			}

			publicGroup.MonitorList = convertMonitorModelsToAPI(monitors)
		}

		publicGroups[i] = publicGroup
	}

	return publicGroups
}

func convertMonitorModelsToAPI(monitors []PublicMonitorModel) []statuspage.PublicMonitor {
	apiMonitors := make([]statuspage.PublicMonitor, len(monitors))

	for j, monitor := range monitors {
		apiMonitors[j] = statuspage.PublicMonitor{
			ID: monitor.ID.ValueInt64(),
		}
		if !monitor.SendURL.IsNull() {
			sendURL := monitor.SendURL.ValueBool()
			apiMonitors[j].SendURL = &sendURL
		}
	}

	return apiMonitors
}

func updateStatusPageDataAfterSave(
	ctx context.Context,
	data *StatusPageResourceModel,
	savedGroups []statuspage.PublicGroup,
	diags *diag.Diagnostics,
) {
	if data.PublicGroupList.IsNull() {
		return
	}

	data.PublicGroupList = mergeGroupIDsIntoPlan(ctx, data.PublicGroupList, savedGroups, diags)
}

// Delete deletes the resource.
func (r *StatusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StatusPageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStatusPage(ctx, data.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete status page", err.Error())
		return
	}
}

func (*StatusPageResource) populateStatusPageDomainNames(
	ctx context.Context,
	data *StatusPageResourceModel,
	sp *statuspage.StatusPage,
	diags *diag.Diagnostics,
) {
	if !data.DomainNameList.IsNull() {
		var domainNames []string
		diags.Append(data.DomainNameList.ElementsAs(ctx, &domainNames, false)...)
		if !diags.HasError() {
			sp.DomainNameList = domainNames
		}
	}
}

// populateStatusPagePublicGroups populates the public group list from the resource model.
func (r *StatusPageResource) populateStatusPagePublicGroups(
	ctx context.Context,
	data *StatusPageResourceModel,
	sp *statuspage.StatusPage,
	diags *diag.Diagnostics,
) {
	if !data.PublicGroupList.IsNull() {
		var groups []PublicGroupModel
		diags.Append(data.PublicGroupList.ElementsAs(ctx, &groups, false)...)
		if diags.HasError() {
			return
		}

		sp.PublicGroupList = make([]statuspage.PublicGroup, len(groups))
		for i, group := range groups {
			r.populatePublicGroup(ctx, &group, &sp.PublicGroupList[i], diags)
			if diags.HasError() {
				return
			}
		}
	}
}

// populatePublicGroup populates a single public group and its monitors.
func (*StatusPageResource) populatePublicGroup(
	ctx context.Context,
	groupModel *PublicGroupModel,
	publicGroup *statuspage.PublicGroup,
	diags *diag.Diagnostics,
) {
	publicGroup.Name = groupModel.Name.ValueString()
	publicGroup.Weight = int(groupModel.Weight.ValueInt64())
	publicGroup.MonitorList = []statuspage.PublicMonitor{}

	if !groupModel.ID.IsNull() {
		publicGroup.ID = groupModel.ID.ValueInt64()
	}

	if !groupModel.MonitorList.IsNull() {
		var monitors []PublicMonitorModel
		diags.Append(groupModel.MonitorList.ElementsAs(ctx, &monitors, false)...)
		if diags.HasError() {
			return
		}

		publicGroup.MonitorList = make([]statuspage.PublicMonitor, len(monitors))
		for j, monitor := range monitors {
			publicGroup.MonitorList[j] = statuspage.PublicMonitor{
				ID: monitor.ID.ValueInt64(),
			}
			if !monitor.SendURL.IsNull() {
				sendURL := monitor.SendURL.ValueBool()
				publicGroup.MonitorList[j].SendURL = &sendURL
			}
		}
	}
}
