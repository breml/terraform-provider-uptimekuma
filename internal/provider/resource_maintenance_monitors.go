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
	_ resource.Resource                = &MaintenanceMonitorsResource{}
	_ resource.ResourceWithImportState = &MaintenanceMonitorsResource{}
)

func NewMaintenanceMonitorsResource() resource.Resource {
	return &MaintenanceMonitorsResource{}
}

type MaintenanceMonitorsResource struct {
	client *kuma.Client
}

type MaintenanceMonitorsResourceModel struct {
	MaintenanceID types.Int64 `tfsdk:"maintenance_id"`
	MonitorIDs    types.List  `tfsdk:"monitor_ids"`
}

func (r *MaintenanceMonitorsResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_monitors"
}

func (r *MaintenanceMonitorsResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Associate monitors with a maintenance window",
		Attributes: map[string]schema.Attribute{
			"maintenance_id": schema.Int64Attribute{
				MarkdownDescription: "Maintenance window ID",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"monitor_ids": schema.ListAttribute{
				MarkdownDescription: "List of monitor IDs to associate",
				Required:            true,
				ElementType:         types.Int64Type,
			},
		},
	}
}

func (r *MaintenanceMonitorsResource) Configure(
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

func (r *MaintenanceMonitorsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MaintenanceMonitorsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var monitorIDs []int64
	resp.Diagnostics.Append(data.MonitorIDs.ElementsAs(ctx, &monitorIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetMonitorMaintenance(ctx, data.MaintenanceID.ValueInt64(), monitorIDs)
	if err != nil {
		resp.Diagnostics.AddError("failed to set monitor maintenance", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaintenanceMonitorsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MaintenanceMonitorsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	monitorIDs, err := r.client.GetMonitorMaintenance(ctx, data.MaintenanceID.ValueInt64())
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read monitor maintenance", err.Error())
		return
	}

	listValue, diags := types.ListValueFrom(ctx, types.Int64Type, monitorIDs)
	resp.Diagnostics.Append(diags...)
	data.MonitorIDs = listValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaintenanceMonitorsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MaintenanceMonitorsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var monitorIDs []int64
	resp.Diagnostics.Append(data.MonitorIDs.ElementsAs(ctx, &monitorIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetMonitorMaintenance(ctx, data.MaintenanceID.ValueInt64(), monitorIDs)
	if err != nil {
		resp.Diagnostics.AddError("failed to update monitor maintenance", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaintenanceMonitorsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MaintenanceMonitorsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetMonitorMaintenance(ctx, data.MaintenanceID.ValueInt64(), []int64{})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete monitor maintenance", err.Error())
		return
	}
}

func (r *MaintenanceMonitorsResource) ImportState(
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
