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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorTCPPortResource{}
	_ resource.ResourceWithImportState = &MonitorTCPPortResource{}
)

// NewMonitorTCPPortResource returns a new instance of the TCP Port monitor resource.
func NewMonitorTCPPortResource() resource.Resource {
	return &MonitorTCPPortResource{}
}

// MonitorTCPPortResource defines the resource implementation.
type MonitorTCPPortResource struct {
	client *kuma.Client
}

// MonitorTCPPortResourceModel describes the resource data model.
type MonitorTCPPortResourceModel struct {
	MonitorBaseModel

	Hostname                 types.String `tfsdk:"hostname"`
	Port                     types.Int64  `tfsdk:"port"`
	SMTPSecurity             types.String `tfsdk:"smtp_security"`
	ExpiryNotification       types.Bool   `tfsdk:"expiry_notification"`
	DomainExpiryNotification types.Bool   `tfsdk:"domain_expiry_notification"`
}

// Metadata returns the metadata for the resource.
func (*MonitorTCPPortResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_tcp_port"
}

// Schema returns the schema for the resource.
func (*MonitorTCPPortResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
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
			"smtp_security": schema.StringAttribute{
				MarkdownDescription: "TLS mode, the \"SSL/TLS\" option in the Uptime Kuma web UI: `nostarttls` " +
					"(None), `secure` (SSL) or `starttls` (STARTTLS). When unset, the check is plain TCP, and " +
					"a value set on the monitor outside Terraform is cleared on the next apply. Uptime Kuma " +
					"only performs the TLS check for `secure` and `starttls` while `expiry_notification` is " +
					"`true`. The \"expected TLS alert\" option of the web UI takes precedence over all of " +
					"this; the provider does not manage that option, and any update clears it.",
				Optional:   true,
				Validators: smtpSecurityValidators(),
			},
			"expiry_notification": schema.BoolAttribute{
				MarkdownDescription: "Enable the TLS check and certificate expiry notification. Only has an " +
					"effect while `smtp_security` is `secure` or `starttls`, or while the \"expected TLS " +
					"alert\" option set in the web UI is in use.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"domain_expiry_notification": domainExpiryNotificationAttribute(),
		}),
	}
}

// Configure configures the TCP Port monitor resource with the API client.
func (r *MonitorTCPPortResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new TCP Port monitor resource.
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
			Hostname:                 data.Hostname.ValueString(),
			Port:                     int(data.Port.ValueInt64()),
			SMTPSecurity:             strToPtr(data.SMTPSecurity),
			ExpiryNotification:       data.ExpiryNotification.ValueBool(),
			DomainExpiryNotification: data.DomainExpiryNotification.ValueBool(),
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

	id, err := r.client.CreateMonitor(ctx, &tcpPortMonitor)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create TCP Port monitor") {
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// The monitor exists, so record it rather than leaving it unmanaged.
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

		return
	}

	err = handleMonitorActiveStateCreate(ctx, r.client, id, data.Active, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError("failed to apply monitor active state", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the TCP Port monitor resource.
func (r *MonitorTCPPortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorTCPPortResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var tcpPortMonitor monitor.TCPPort
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &tcpPortMonitor)
	if err != nil {
		removeMonitorOnServerMiss(ctx, r.client, err, data.ID.ValueInt64(), "TCP Port monitor", resp)

		return
	}

	if monitorTypeDrifted(
		ctx, data.ID.ValueInt64(), tcpPortMonitor.Type(), tcpPortMonitor.Base.Type(), resp,
	) {
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
	data.SMTPSecurity = ptrToTypes(tcpPortMonitor.SMTPSecurity)
	data.ExpiryNotification = types.BoolValue(tcpPortMonitor.ExpiryNotification)
	data.DomainExpiryNotification = types.BoolValue(tcpPortMonitor.DomainExpiryNotification)

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

	data.Tags = handleMonitorTagsRead(ctx, tcpPortMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the TCP Port monitor resource.
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

	// Get resource from state.
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
			Hostname:                 data.Hostname.ValueString(),
			Port:                     int(data.Port.ValueInt64()),
			SMTPSecurity:             strToPtr(data.SMTPSecurity),
			ExpiryNotification:       data.ExpiryNotification.ValueBool(),
			DomainExpiryNotification: data.DomainExpiryNotification.ValueBool(),
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

	err := r.client.UpdateMonitor(ctx, &tcpPortMonitor)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update TCP Port monitor") {
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	handleMonitorActiveStateUpdate(ctx, r.client, data.ID.ValueInt64(), state.Active, data.Active, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the TCP Port monitor resource.
func (r *MonitorTCPPortResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorTCPPortResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete TCP Port monitor")
}

// ImportState imports an existing resource by ID.
func (*MonitorTCPPortResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
