package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	// Ensure MonitorSMTPResource satisfies various resource interfaces.
	_ resource.Resource                = &MonitorSMTPResource{}
	_ resource.ResourceWithImportState = &MonitorSMTPResource{}
)

// NewMonitorSMTPResource returns a new instance of the SMTP monitor resource.
func NewMonitorSMTPResource() resource.Resource {
	return &MonitorSMTPResource{}
}

// MonitorSMTPResource defines the resource implementation.
type MonitorSMTPResource struct {
	client *kuma.Client
}

// MonitorSMTPResourceModel describes the resource data model for SMTP monitors.
type MonitorSMTPResourceModel struct {
	MonitorBaseModel

	Hostname     types.String `tfsdk:"hostname"`      // SMTP server hostname or IP.
	Port         types.Int64  `tfsdk:"port"`          // SMTP server port.
	SMTPSecurity types.String `tfsdk:"smtp_security"` // SMTP security mode.
}

// Metadata returns the metadata for the resource.
func (*MonitorSMTPResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_smtp"
}

// Schema returns the schema for the resource.
func (*MonitorSMTPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SMTP monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "SMTP server hostname or IP address",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "SMTP server port",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(587),
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"smtp_security": schema.StringAttribute{
				MarkdownDescription: "SMTP security mode (None, STARTTLS, or TLS)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("None"),
				Validators: []validator.String{
					stringvalidator.OneOf("None", "STARTTLS", "TLS"),
				},
			},
		}),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorSMTPResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new resource.
func (r *MonitorSMTPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorSMTPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smtpMonitor := buildSMTPMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &smtpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create SMTP monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// buildSMTPMonitor constructs an SMTP monitor from the Terraform resource model.
func buildSMTPMonitor(ctx context.Context, data *MonitorSMTPResourceModel, diags *diag.Diagnostics) monitor.SMTP {
	smtpMonitor := monitor.SMTP{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		SMTPDetails: monitor.SMTPDetails{
			Hostname: data.Hostname.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		smtpMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		smtpMonitor.Parent = &parent
	}

	if !data.Port.IsNull() {
		port := data.Port.ValueInt64()
		smtpMonitor.Port = &port
	}

	if !data.SMTPSecurity.IsNull() {
		security := data.SMTPSecurity.ValueString()
		smtpMonitor.SMTPSecurity = &security
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			smtpMonitor.NotificationIDs = notificationIDs
		}
	}

	return smtpMonitor
}

// populateSMTPMonitorFields populates the SMTP monitor fields from the API response.
// Extracts SMTP-specific fields from the API response into the model.
// Handles base monitor fields and all SMTP configuration options.
func populateSMTPMonitorFields(smtpMonitor *monitor.SMTP, m *MonitorSMTPResourceModel) {
	m.Name = types.StringValue(smtpMonitor.Name)
	if smtpMonitor.Description != nil {
		m.Description = types.StringValue(*smtpMonitor.Description)
	} else {
		m.Description = types.StringNull()
	}

	m.Interval = types.Int64Value(smtpMonitor.Interval)
	m.RetryInterval = types.Int64Value(smtpMonitor.RetryInterval)
	m.ResendInterval = types.Int64Value(smtpMonitor.ResendInterval)
	m.MaxRetries = types.Int64Value(smtpMonitor.MaxRetries)
	m.UpsideDown = types.BoolValue(smtpMonitor.UpsideDown)
	m.Active = types.BoolValue(smtpMonitor.IsActive)
	m.Hostname = types.StringValue(smtpMonitor.Hostname)
	m.SMTPSecurity = ptrToTypes(smtpMonitor.SMTPSecurity)
}

// populateSMTPOptionalFields populates optional fields for SMTP monitor.
// Handles parent group, port, and notification IDs.
// Converts null API values to Terraform null types appropriately.
func populateSMTPOptionalFields(
	ctx context.Context,
	smtpMonitor *monitor.SMTP,
	m *MonitorSMTPResourceModel,
	diags *diag.Diagnostics,
) {
	// Set parent monitor group if present.
	if smtpMonitor.Parent != nil {
		m.Parent = types.Int64Value(*smtpMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	// Set port if configured.
	if smtpMonitor.Port != nil {
		m.Port = types.Int64Value(*smtpMonitor.Port)
	} else {
		m.Port = types.Int64Null()
	}

	// Convert notification IDs list if present.
	if len(smtpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, smtpMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// Read reads the current state of the resource.
func (r *MonitorSMTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorSMTPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var smtpMonitor monitor.SMTP
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &smtpMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read SMTP monitor", err.Error())
		return
	}

	populateSMTPMonitorFields(&smtpMonitor, &data)
	populateSMTPOptionalFields(ctx, &smtpMonitor, &data, &resp.Diagnostics)

	data.Tags = handleMonitorTagsRead(ctx, smtpMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MonitorSMTPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorSMTPResourceModel
	var state MonitorSMTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smtpMonitor := buildSMTPMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	smtpMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &smtpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update SMTP monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *MonitorSMTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorSMTPResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API.
	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete SMTP monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorSMTPResource) ImportState(
	// Import monitor by ID.
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
