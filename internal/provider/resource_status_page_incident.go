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

var _ resource.Resource = &StatusPageIncidentResource{}

func NewStatusPageIncidentResource() resource.Resource {
	return &StatusPageIncidentResource{}
}

// StatusPageIncidentResource manages incidents on status pages.
//
// NOTE: The Uptime Kuma API and client have limited incident management capabilities:
//   - No GetIncident method: The Read operation cannot fetch incident state from the API,
//     so drift detection is limited. The resource maintains state but cannot verify it.
//   - No DeleteIncident method: The Delete operation can only unpin incidents. Incidents
//     persist on the status page after Terraform destroys the resource.
//   - PostIncident handles both create and update operations using the incident ID.
//
// These limitations mean this resource provides best-effort management of incidents
// but cannot guarantee full CRUD semantics or accurate drift detection.
type StatusPageIncidentResource struct {
	client *kuma.Client
}

type StatusPageIncidentResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	StatusPageSlug types.String `tfsdk:"status_page_slug"`
	Title          types.String `tfsdk:"title"`
	Content        types.String `tfsdk:"content"`
	Style          types.String `tfsdk:"style"`
	Pin            types.Bool   `tfsdk:"pin"`
}

func (r *StatusPageIncidentResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_status_page_incident"
}

func (r *StatusPageIncidentResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Status page incident resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Incident ID",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status_page_slug": schema.StringAttribute{
				MarkdownDescription: "Reference to status page slug",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Incident title",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Incident description",
				Required:            true,
			},
			"style": schema.StringAttribute{
				MarkdownDescription: "Incident style/severity (e.g., info, warning, danger, primary, light, dark)",
				Optional:            true,
			},
			"pin": schema.BoolAttribute{
				MarkdownDescription: "Pin incident to top of status page",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *StatusPageIncidentResource) Configure(
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

func (r *StatusPageIncidentResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data StatusPageIncidentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	incident := &statuspage.Incident{
		Title:   data.Title.ValueString(),
		Content: data.Content.ValueString(),
		Style:   data.Style.ValueString(),
		Pin:     data.Pin.ValueBool(),
	}

	err := r.client.PostIncident(ctx, data.StatusPageSlug.ValueString(), incident)
	if err != nil {
		resp.Diagnostics.AddError("failed to create incident", err.Error())
		return
	}

	data.ID = types.Int64Value(incident.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusPageIncidentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StatusPageIncidentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusPageIncidentResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data StatusPageIncidentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	incident := &statuspage.Incident{
		ID:      data.ID.ValueInt64(),
		Title:   data.Title.ValueString(),
		Content: data.Content.ValueString(),
		Style:   data.Style.ValueString(),
		Pin:     data.Pin.ValueBool(),
	}

	err := r.client.PostIncident(ctx, data.StatusPageSlug.ValueString(), incident)
	if err != nil {
		resp.Diagnostics.AddError("failed to update incident", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusPageIncidentResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data StatusPageIncidentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Pin.ValueBool() {
		err := r.client.UnpinIncident(ctx, data.StatusPageSlug.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to unpin incident", err.Error())
			return
		}
	}
}
