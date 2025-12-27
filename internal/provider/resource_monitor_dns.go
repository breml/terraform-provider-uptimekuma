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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorDNSResource{}
	_ resource.ResourceWithImportState = &MonitorDNSResource{}
)

// NewMonitorDNSResource returns a new instance of the DNS monitor resource.
func NewMonitorDNSResource() resource.Resource {
	return &MonitorDNSResource{}
}

// MonitorDNSResource defines the resource implementation.
type MonitorDNSResource struct {
	client *kuma.Client
}

// MonitorDNSResourceModel describes the resource data model.
type MonitorDNSResourceModel struct {
	MonitorBaseModel

	Hostname         types.String `tfsdk:"hostname"`
	DNSResolveServer types.String `tfsdk:"dns_resolve_server"`
	DNSResolveType   types.String `tfsdk:"dns_resolve_type"`
	Port             types.Int64  `tfsdk:"port"`
}

// Metadata returns the metadata for the resource.
func (*MonitorDNSResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_dns"
}

// Schema returns the schema for the resource.
func (*MonitorDNSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "DNS monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Domain name to resolve",
				Required:            true,
			},
			"dns_resolve_server": schema.StringAttribute{
				MarkdownDescription: "DNS resolver server IP address",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("1.1.1.1"),
			},
			"dns_resolve_type": schema.StringAttribute{
				MarkdownDescription: "DNS record type to query",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("A"),
				Validators: []validator.String{
					stringvalidator.OneOf("A", "AAAA", "CAA", "CNAME", "MX", "NS", "PTR", "SOA", "SRV", "TXT"),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "DNS resolver port",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(53),
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
		}),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorDNSResource) Configure(
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

// Create creates a new resource.
func (r *MonitorDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorDNSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dnsMonitor := monitor.DNS{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		DNSDetails: monitor.DNSDetails{
			Hostname:       data.Hostname.ValueString(),
			ResolverServer: data.DNSResolveServer.ValueString(),
			ResolveType:    monitor.DNSResolveType(data.DNSResolveType.ValueString()),
			Port:           int(data.Port.ValueInt64()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		dnsMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		dnsMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		dnsMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, dnsMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create DNS monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the resource.
func (r *MonitorDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorDNSResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var dnsMonitor monitor.DNS
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &dnsMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read DNS monitor", err.Error())
		return
	}

	data.Name = types.StringValue(dnsMonitor.Name)
	if dnsMonitor.Description != nil {
		data.Description = types.StringValue(*dnsMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(dnsMonitor.Interval)
	data.RetryInterval = types.Int64Value(dnsMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(dnsMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(dnsMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(dnsMonitor.UpsideDown)
	data.Active = types.BoolValue(dnsMonitor.IsActive)
	data.Hostname = types.StringValue(dnsMonitor.Hostname)
	data.DNSResolveServer = types.StringValue(dnsMonitor.ResolverServer)
	data.DNSResolveType = types.StringValue(string(dnsMonitor.ResolveType))
	data.Port = types.Int64Value(int64(dnsMonitor.Port))

	if dnsMonitor.Parent != nil {
		data.Parent = types.Int64Value(*dnsMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(dnsMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, dnsMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, dnsMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MonitorDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorDNSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorDNSResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dnsMonitor := monitor.DNS{
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
		DNSDetails: monitor.DNSDetails{
			Hostname:       data.Hostname.ValueString(),
			ResolverServer: data.DNSResolveServer.ValueString(),
			ResolveType:    monitor.DNSResolveType(data.DNSResolveType.ValueString()),
			Port:           int(data.Port.ValueInt64()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		dnsMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		dnsMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		dnsMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, dnsMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update DNS monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *MonitorDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorDNSResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete DNS monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorDNSResource) ImportState(
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
