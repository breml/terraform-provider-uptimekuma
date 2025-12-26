package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorTCPPortResource{}
	_ resource.ResourceWithImportState = &MonitorTCPPortResource{}
)

func NewMonitorTCPPortResource() resource.Resource {
	return &MonitorTCPPortResource{}
}

type MonitorTCPPortResource struct {
	client *kuma.Client
}

type MonitorTCPPortResourceModel struct {
	MonitorBaseModel

	Hostname types.String `tfsdk:"hostname"`
	Port     types.Int64  `tfsdk:"port"`
}

func (r *MonitorTCPPortResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_tcp_port"
}

func (r *MonitorTCPPortResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "TCP Port monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address to monitor",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "TCP port number to monitor",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
		}),
	}
}

func (r *MonitorTCPPortResource) Configure(
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

func (r *MonitorTCPPortResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorTCPPortResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tcpPortMonitor := monitor.TCPPort{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		TCPPortDetails: monitor.TCPPortDetails{
			Hostname: data.Hostname.ValueString(),
			Port:     int(data.Port.ValueInt64()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		tcpPortMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		tcpPortMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		tcpPortMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, tcpPortMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create TCP Port monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorTCPPortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorTCPPortResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var tcpPortMonitor monitor.TCPPort
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &tcpPortMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read TCP Port monitor", err.Error())
		return
	}

	data.Name = types.StringValue(tcpPortMonitor.Name)
	if tcpPortMonitor.Description != nil {
		data.Description = types.StringValue(*tcpPortMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(tcpPortMonitor.Interval)
	data.RetryInterval = types.Int64Value(tcpPortMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(tcpPortMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(tcpPortMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(tcpPortMonitor.UpsideDown)
	data.Active = types.BoolValue(tcpPortMonitor.IsActive)
	data.Hostname = types.StringValue(tcpPortMonitor.Hostname)
	data.Port = types.Int64Value(int64(tcpPortMonitor.Port))

	if tcpPortMonitor.Parent != nil {
		data.Parent = types.Int64Value(*tcpPortMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(tcpPortMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, tcpPortMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, tcpPortMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorTCPPortResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorTCPPortResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorTCPPortResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tcpPortMonitor := monitor.TCPPort{
		Base: monitor.Base{
			ID:             data.ID.ValueInt64(),
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		TCPPortDetails: monitor.TCPPortDetails{
			Hostname: data.Hostname.ValueString(),
			Port:     int(data.Port.ValueInt64()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		tcpPortMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		tcpPortMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		tcpPortMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, tcpPortMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update TCP Port monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorTCPPortResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorTCPPortResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete TCP Port monitor", err.Error())
		return
	}
}

func (r *MonitorTCPPortResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
