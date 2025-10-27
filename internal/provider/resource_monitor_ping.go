package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ resource.Resource = &MonitorPingResource{}

func NewMonitorPingResource() resource.Resource {
	return &MonitorPingResource{}
}

type MonitorPingResource struct {
	client *kuma.Client
}

type MonitorPingResourceModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Parent          types.Int64  `tfsdk:"parent"`
	Interval        types.Int64  `tfsdk:"interval"`
	RetryInterval   types.Int64  `tfsdk:"retry_interval"`
	ResendInterval  types.Int64  `tfsdk:"resend_interval"`
	MaxRetries      types.Int64  `tfsdk:"max_retries"`
	UpsideDown      types.Bool   `tfsdk:"upside_down"`
	Active          types.Bool   `tfsdk:"active"`
	Hostname        types.String `tfsdk:"hostname"`
	PacketSize      types.Int64  `tfsdk:"packet_size"`
	NotificationIDs types.List   `tfsdk:"notification_ids"`
}

func (r *MonitorPingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_ping"
}

func (r *MonitorPingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Ping monitor resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Monitor identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Friendly name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description",
				Optional:            true,
			},
			"parent": schema.Int64Attribute{
				MarkdownDescription: "Parent monitor ID for hierarchical organization",
				Optional:            true,
			},
			"interval": schema.Int64Attribute{
				MarkdownDescription: "Heartbeat interval in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				Validators: []validator.Int64{
					int64validator.Between(20, 2073600),
				},
			},
			"retry_interval": schema.Int64Attribute{
				MarkdownDescription: "Retry interval in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				Validators: []validator.Int64{
					int64validator.Between(20, 2073600),
				},
			},
			"resend_interval": schema.Int64Attribute{
				MarkdownDescription: "Resend interval in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"max_retries": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of retries",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3),
				Validators: []validator.Int64{
					int64validator.Between(0, 10),
				},
			},
			"upside_down": schema.BoolAttribute{
				MarkdownDescription: "Invert monitor status (treat DOWN as UP and vice versa)",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Monitor is active",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address to ping",
				Required:            true,
			},
			"packet_size": schema.Int64Attribute{
				MarkdownDescription: "Ping packet size in bytes",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(56),
				Validators: []validator.Int64{
					int64validator.Between(1, 65500),
				},
			},
			"notification_ids": schema.ListAttribute{
				MarkdownDescription: "List of notification IDs",
				ElementType:         types.Int64Type,
				Optional:            true,
			},
		},
	}
}

func (r *MonitorPingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *kuma.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *MonitorPingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorPingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pingMonitor := monitor.Ping{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		PingDetails: monitor.PingDetails{
			Hostname:   data.Hostname.ValueString(),
			PacketSize: int(data.PacketSize.ValueInt64()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		pingMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		pingMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		pingMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, pingMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Ping monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorPingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorPingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var pingMonitor monitor.Ping
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &pingMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Ping monitor", err.Error())
		return
	}

	data.Name = types.StringValue(pingMonitor.Name)
	if pingMonitor.Description != nil {
		data.Description = types.StringValue(*pingMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(pingMonitor.Interval)
	data.RetryInterval = types.Int64Value(pingMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(pingMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(pingMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(pingMonitor.UpsideDown)
	data.Active = types.BoolValue(pingMonitor.IsActive)
	data.Hostname = types.StringValue(pingMonitor.Hostname)
	data.PacketSize = types.Int64Value(int64(pingMonitor.PacketSize))

	if pingMonitor.Parent != nil {
		data.Parent = types.Int64Value(*pingMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(pingMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, pingMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorPingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorPingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pingMonitor := monitor.Ping{
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
		PingDetails: monitor.PingDetails{
			Hostname:   data.Hostname.ValueString(),
			PacketSize: int(data.PacketSize.ValueInt64()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		pingMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		pingMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		pingMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, pingMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update Ping monitor", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorPingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorPingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Ping monitor", err.Error())
		return
	}
}
