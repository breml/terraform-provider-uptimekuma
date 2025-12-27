package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/statuspage"
)

var _ resource.Resource = &StatusPageResource{}

func NewStatusPageResource() resource.Resource {
	return &StatusPageResource{}
}

type StatusPageResource struct {
	client *kuma.Client
}

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

type PublicGroupModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Weight      types.Int64  `tfsdk:"weight"`
	MonitorList types.List   `tfsdk:"monitor_list"`
}

type PublicMonitorModel struct {
	ID      types.Int64 `tfsdk:"id"`
	SendURL types.Bool  `tfsdk:"send_url"`
}

func (r *StatusPageResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (r *StatusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				MarkdownDescription: "Base64-encoded icon image",
				Optional:            true,
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

func (r *StatusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StatusPageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddStatusPage(ctx, data.Title.ValueString(), data.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to create status page", err.Error())
		return
	}

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
		resp.Diagnostics.Append(data.DomainNameList.ElementsAs(ctx, &domainNames, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		sp.DomainNameList = domainNames
	}

	if !data.PublicGroupList.IsNull() {
		var groups []PublicGroupModel
		resp.Diagnostics.Append(data.PublicGroupList.ElementsAs(ctx, &groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		sp.PublicGroupList = make([]statuspage.PublicGroup, len(groups))
		for i, group := range groups {
			publicGroup := statuspage.PublicGroup{
				Name:        group.Name.ValueString(),
				Weight:      int(group.Weight.ValueInt64()),
				MonitorList: []statuspage.PublicMonitor{},
			}
			if !group.ID.IsNull() {
				publicGroup.ID = group.ID.ValueInt64()
			}

			sp.PublicGroupList[i] = publicGroup

			if !group.MonitorList.IsNull() {
				var monitors []PublicMonitorModel
				resp.Diagnostics.Append(group.MonitorList.ElementsAs(ctx, &monitors, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				sp.PublicGroupList[i].MonitorList = make([]statuspage.PublicMonitor, len(monitors))
				for j, monitor := range monitors {
					sp.PublicGroupList[i].MonitorList[j] = statuspage.PublicMonitor{
						ID: monitor.ID.ValueInt64(),
					}
					if !monitor.SendURL.IsNull() {
						sendURL := monitor.SendURL.ValueBool()
						sp.PublicGroupList[i].MonitorList[j].SendURL = &sendURL
					}
				}
			}
		}
	}

	savedGroups, err := r.client.SaveStatusPage(ctx, sp)
	if err != nil {
		resp.Diagnostics.AddError("failed to save status page", err.Error())
		return
	}

	// Read back the status page to get the generated ID
	retrievedSP, err := r.client.GetStatusPage(ctx, data.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read status page after creation", err.Error())
		return
	}

	data.ID = types.Int64Value(retrievedSP.ID)

	planPublic := data.PublicGroupList

	// Build public_group_list from the savedGroups response when available
	data.PublicGroupList = buildPublicGroupListFromSaved(ctx, savedGroups, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If server didn't return groups, preserve config but convert unknown IDs to null so values are known after create
	if len(savedGroups) == 0 && !planPublic.IsNull() {
		data.PublicGroupList = convertUnknownIDsToNull(ctx, planPublic, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StatusPageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sp, err := r.client.GetStatusPage(ctx, data.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read status page", err.Error())
		return
	}

	data.ID = types.Int64Value(sp.ID)
	data.Title = types.StringValue(sp.Title)
	data.Description = stringOrNull(sp.Description)

	if !data.Icon.IsNull() {
		data.Icon = stringOrNull(sp.Icon)
	}

	data.Theme = stringOrNull(sp.Theme)

	// Note: The Uptime Kuma API's saveStatusPage endpoint does not actually update
	// the published, show_tags, show_powered_by, and show_certificate_expiry fields
	// (see server/socket-handlers/status-page-socket-handler.js line 160-167).
	// Therefore, we don't update these fields from the API response to avoid drift.
	// We keep whatever values are in the Terraform config/state.

	data.GoogleAnalyticsID = stringOrNull(sp.GoogleAnalyticsID)
	data.CustomCSS = stringOrNull(sp.CustomCSS)
	data.FooterText = stringOrNull(sp.FooterText)

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

func (r *StatusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StatusPageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
		resp.Diagnostics.Append(data.DomainNameList.ElementsAs(ctx, &domainNames, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		sp.DomainNameList = domainNames
	}

	if !data.PublicGroupList.IsNull() {
		var groups []PublicGroupModel
		resp.Diagnostics.Append(data.PublicGroupList.ElementsAs(ctx, &groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		sp.PublicGroupList = make([]statuspage.PublicGroup, len(groups))
		for i, group := range groups {
			publicGroup := statuspage.PublicGroup{
				Name:        group.Name.ValueString(),
				Weight:      int(group.Weight.ValueInt64()),
				MonitorList: []statuspage.PublicMonitor{},
			}
			if !group.ID.IsNull() {
				publicGroup.ID = group.ID.ValueInt64()
			}

			sp.PublicGroupList[i] = publicGroup

			if !group.MonitorList.IsNull() {
				var monitors []PublicMonitorModel
				resp.Diagnostics.Append(group.MonitorList.ElementsAs(ctx, &monitors, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				sp.PublicGroupList[i].MonitorList = make([]statuspage.PublicMonitor, len(monitors))
				for j, monitor := range monitors {
					sp.PublicGroupList[i].MonitorList[j] = statuspage.PublicMonitor{
						ID: monitor.ID.ValueInt64(),
					}
					if !monitor.SendURL.IsNull() {
						sendURL := monitor.SendURL.ValueBool()
						sp.PublicGroupList[i].MonitorList[j].SendURL = &sendURL
					}
				}
			}
		}
	}

	savedGroups, err := r.client.SaveStatusPage(ctx, sp)
	if err != nil {
		resp.Diagnostics.AddError("failed to update status page", err.Error())
		return
	}

	// If the server returned group IDs, construct a known public_group_list from the response
	if len(savedGroups) > 0 {
		data.PublicGroupList = buildPublicGroupListFromSaved(ctx, savedGroups, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	} else if !data.PublicGroupList.IsNull() {
		// If server didn't return groups, preserve config but ensure unknown IDs are set to null
		data.PublicGroupList = convertUnknownIDsToNull(ctx, data.PublicGroupList, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
