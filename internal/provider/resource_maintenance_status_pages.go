package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var (
	_ resource.Resource                = &MaintenanceStatusPagesResource{}
	_ resource.ResourceWithImportState = &MaintenanceStatusPagesResource{}
)

func NewMaintenanceStatusPagesResource() resource.Resource {
	return &MaintenanceStatusPagesResource{}
}

type MaintenanceStatusPagesResource struct {
	client *kuma.Client
}

type MaintenanceStatusPagesResourceModel struct {
	MaintenanceID types.Int64 `tfsdk:"maintenance_id"`
	StatusPageIDs types.List  `tfsdk:"status_page_ids"`
}

func (r *MaintenanceStatusPagesResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_status_pages"
}

func (r *MaintenanceStatusPagesResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Associate status pages with a maintenance window",
		Attributes: map[string]schema.Attribute{
			"maintenance_id": schema.Int64Attribute{
				MarkdownDescription: "Maintenance window ID",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status_page_ids": schema.ListAttribute{
				MarkdownDescription: "List of status page IDs to associate",
				Required:            true,
				ElementType:         types.Int64Type,
			},
		},
	}
}

func (r *MaintenanceStatusPagesResource) Configure(
	ctx context.Context,
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

func (r *MaintenanceStatusPagesResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MaintenanceStatusPagesResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var statusPageIDs []int64
	resp.Diagnostics.Append(data.StatusPageIDs.ElementsAs(ctx, &statusPageIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetMaintenanceStatusPage(ctx, data.MaintenanceID.ValueInt64(), statusPageIDs)
	if err != nil {
		resp.Diagnostics.AddError("failed to set maintenance status pages", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaintenanceStatusPagesResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MaintenanceStatusPagesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	statusPageIDs, err := r.client.GetMaintenanceStatusPage(ctx, data.MaintenanceID.ValueInt64())
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read maintenance status pages", err.Error())
		return
	}

	listValue, diags := types.ListValueFrom(ctx, types.Int64Type, statusPageIDs)
	resp.Diagnostics.Append(diags...)
	data.StatusPageIDs = listValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaintenanceStatusPagesResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MaintenanceStatusPagesResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var statusPageIDs []int64
	resp.Diagnostics.Append(data.StatusPageIDs.ElementsAs(ctx, &statusPageIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetMaintenanceStatusPage(ctx, data.MaintenanceID.ValueInt64(), statusPageIDs)
	if err != nil {
		resp.Diagnostics.AddError("failed to update maintenance status pages", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaintenanceStatusPagesResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MaintenanceStatusPagesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetMaintenanceStatusPage(ctx, data.MaintenanceID.ValueInt64(), []int64{})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete maintenance status pages", err.Error())
		return
	}
}

func (r *MaintenanceStatusPagesResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid maintenance ID (integer), got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("maintenance_id"), id)...)
}
