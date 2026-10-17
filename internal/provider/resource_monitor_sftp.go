package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                   = &MonitorSFTPResource{}
	_ resource.ResourceWithImportState    = &MonitorSFTPResource{}
	_ resource.ResourceWithValidateConfig = &MonitorSFTPResource{}
)

// defaultSFTPTimeout is the connection timeout in seconds Uptime Kuma's SFTP
// check falls back to. It is used both as the schema default and as the read
// fallback: the monitor.timeout column is NOT NULL, so this value is stored and
// reads back unchanged, which makes it the only SFTP fallback that can be
// modelled as a Terraform default without causing a perpetual diff.
//
// The value mirrors the unexported defaultSFTPTimeout in the client library
// (monitor/monitor_sftp.go); re-check it when bumping go-uptime-kuma-client.
const defaultSFTPTimeout = 10

// NewMonitorSFTPResource returns a new instance of the SFTP monitor resource.
func NewMonitorSFTPResource() resource.Resource {
	return &MonitorSFTPResource{}
}

// MonitorSFTPResource defines the resource implementation for SFTP monitors.
type MonitorSFTPResource struct {
	client *kuma.Client
}

// MonitorSFTPResourceModel describes the resource data model for SFTP monitors.
type MonitorSFTPResourceModel struct {
	MonitorBaseModel

	Hostname      types.String  `tfsdk:"hostname"`        // SFTP server to connect to.
	Port          types.Int64   `tfsdk:"port"`            // TCP port of the SFTP server.
	Timeout       types.Float64 `tfsdk:"timeout"`         // Connection timeout in seconds.
	SSHUsername   types.String  `tfsdk:"ssh_username"`    // User to authenticate as.
	SSHAuthMethod types.String  `tfsdk:"ssh_auth_method"` // password or privateKey.
	SSHPassword   types.String  `tfsdk:"ssh_password"`    // Password for password authentication.
	SSHPrivateKey types.String  `tfsdk:"ssh_private_key"` // Private key for key authentication.
	SSHPassphrase types.String  `tfsdk:"ssh_passphrase"`  // Passphrase of the private key.
	SFTPPath      types.String  `tfsdk:"sftp_path"`       // Remote path that must exist.
}

// Metadata returns the metadata for the resource.
func (*MonitorSFTPResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_sftp"
}

// Schema returns the schema for the resource.
func (*MonitorSFTPResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SFTP monitor resource. The monitor opens an SSH connection to a host, " +
			"authenticates with a password or a private key and, when `sftp_path` is set, checks " +
			"that the remote path exists. Requires Uptime Kuma 2.5.4 or later.\n\n" +
			"`port`, the credentials and `sftp_path` are stored as SQL NULL while unset, which is " +
			"a different state from setting them to the value the check falls back to. Leave them " +
			"unset to follow the check's own fallbacks.",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "SFTP server IP address or hostname",
				Required:            true,
				Validators: []validator.String{
					nonBlank(),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "TCP port of the SFTP server, between 1 and 65535. While unset, " +
					"the check falls back to 22.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"timeout": monitorTimeoutAttribute(
				"Connection timeout in seconds, at least 0.1, used as both the SSH ready timeout "+
					"and the socket timeout. Fractional values are supported and round-trip unchanged. "+
					"Defaults to 10. Note that the Uptime Kuma web UI assigns 48 to new monitors, so a "+
					"monitor created there and then imported reports 48. "+
					monitorTimeoutFloorDescription+" "+monitorTimeoutClampDescription,
				defaultSFTPTimeout,
				float64validator.AtLeast(0.1),
			),
			"ssh_username": schema.StringAttribute{
				MarkdownDescription: "User to authenticate as",
				Required:            true,
				Validators: []validator.String{
					nonBlank(),
				},
			},
			"ssh_auth_method": schema.StringAttribute{
				MarkdownDescription: "SSH authentication method, either `password` or `privateKey`. " +
					"Defaults to `password`. `privateKey` requires `ssh_private_key`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(monitor.SFTPAuthMethodPassword)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(monitor.SFTPAuthMethodPassword),
						string(monitor.SFTPAuthMethodPrivateKey),
					),
				},
			},
			"ssh_password": schema.StringAttribute{
				MarkdownDescription: "Password used while `ssh_auth_method` is `password`",
				Optional:            true,
				Sensitive:           true,
			},
			"ssh_private_key": schema.StringAttribute{
				MarkdownDescription: "Private key in PEM or OpenSSH format, used while " +
					"`ssh_auth_method` is `privateKey`, where it is required.",
				Optional:  true,
				Sensitive: true,
			},
			"ssh_passphrase": schema.StringAttribute{
				MarkdownDescription: "Passphrase of `ssh_private_key`. Only allowed while " +
					"`ssh_auth_method` is `privateKey`, because password authentication ignores it.",
				Optional:  true,
				Sensitive: true,
			},
			"sftp_path": schema.StringAttribute{
				MarkdownDescription: "Remote path that must exist on the server. While unset, the " +
					"check only verifies that the SSH connection and authentication succeed.",
				Optional: true,
			},
		}),
	}
}

// Configure configures the SFTP monitor resource with the API client.
func (r *MonitorSFTPResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// ValidateConfig validates the configuration for the SFTP monitor resource.
//
// Which credentials the check uses depends on the value of ssh_auth_method rather than on it being
// set at all, which stringvalidator.AlsoRequires and ConflictsWith cannot express.
func (*MonitorSFTPResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config MonitorSFTPResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	validateSFTPCredentials(&config, resp)
}

// validateSFTPCredentials reports credentials that do not match the configured auth method.
func validateSFTPCredentials(config *MonitorSFTPResourceModel, resp *resource.ValidateConfigResponse) {
	if config.SSHAuthMethod.IsUnknown() {
		// The value is only known after apply, so there is nothing to validate against.
		return
	}

	// A null ssh_auth_method means the schema default, which is password authentication.
	authMethod := string(monitor.SFTPAuthMethodPassword)
	if !config.SSHAuthMethod.IsNull() {
		authMethod = config.SSHAuthMethod.ValueString()
	}

	switch authMethod {
	case string(monitor.SFTPAuthMethodPrivateKey):
		if config.SSHPrivateKey.IsNull() {
			resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				path.Root("ssh_private_key"),
				"Missing Attribute Configuration",
				"When \"ssh_auth_method\" is set to \"privateKey\", the attribute \"ssh_private_key\" "+
					"must be set. Without it the check fails on every heartbeat.",
			))
		}

	case string(monitor.SFTPAuthMethodPassword):
		if !config.SSHPassphrase.IsNull() && !config.SSHPassphrase.IsUnknown() {
			resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				path.Root("ssh_passphrase"),
				"Invalid Attribute Combination",
				"The attribute \"ssh_passphrase\" is only used when \"ssh_auth_method\" is set to "+
					"\"privateKey\" and must not be set otherwise.",
			))
		}

	default:
		// An unexpected auth method is left to the schema validator.
	}
}

// Create creates a new SFTP monitor resource.
func (r *MonitorSFTPResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorSFTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sftpMonitor := buildSFTPMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &sftpMonitor)
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create SFTP monitor") {
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the SFTP monitor resource.
func (r *MonitorSFTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorSFTPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var sftpMonitor monitor.SFTP
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &sftpMonitor)
	if err != nil {
		removeMonitorOnServerMiss(ctx, r.client, err, data.ID.ValueInt64(), "SFTP monitor", resp)

		return
	}

	if monitorTypeDrifted(
		ctx, data.ID.ValueInt64(), sftpMonitor.Type(), sftpMonitor.Base.Type(), resp,
	) {
		return
	}

	populateSFTPModel(ctx, &sftpMonitor, &data, &resp.Diagnostics)
	populateSFTPOptionalFields(ctx, &sftpMonitor, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SFTP monitor resource.
func (r *MonitorSFTPResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorSFTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorSFTPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sftpMonitor := buildSFTPMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	sftpMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &sftpMonitor)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update SFTP monitor") {
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SFTP monitor resource.
func (r *MonitorSFTPResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorSFTPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete SFTP monitor")
}

// ImportState imports an existing resource by ID.
func (*MonitorSFTPResource) ImportState(
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

// buildSFTPMonitor constructs an SFTP monitor API object from the Terraform resource model.
func buildSFTPMonitor(
	ctx context.Context,
	data *MonitorSFTPResourceModel,
	diags *diag.Diagnostics,
) monitor.SFTP {
	sftpMonitor := monitor.SFTP{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		SFTPDetails: monitor.SFTPDetails{
			Hostname:      data.Hostname.ValueString(),
			Port:          int64ToPtr(data.Port),
			Timeout:       float64ToPtr(data.Timeout),
			SSHUsername:   data.SSHUsername.ValueString(),
			SSHAuthMethod: monitor.SFTPAuthMethod(data.SSHAuthMethod.ValueString()),
			SSHPassword:   strToPtr(data.SSHPassword),
			SSHPrivateKey: strToPtr(data.SSHPrivateKey),
			SSHPassphrase: strToPtr(data.SSHPassphrase),
			SFTPPath:      strToPtr(data.SFTPPath),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		sftpMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		sftpMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if diags.HasError() {
			return sftpMonitor
		}

		sftpMonitor.NotificationIDs = notificationIDs
	}

	return sftpMonitor
}

// populateSFTPModel populates the base fields of the Terraform model from the API response.
func populateSFTPModel(
	ctx context.Context,
	sftpMonitor *monitor.SFTP,
	data *MonitorSFTPResourceModel,
	diags *diag.Diagnostics,
) {
	data.Name = types.StringValue(sftpMonitor.Name)
	if sftpMonitor.Description != nil {
		data.Description = types.StringValue(*sftpMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(sftpMonitor.Interval)
	data.RetryInterval = types.Int64Value(sftpMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(sftpMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(sftpMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(sftpMonitor.UpsideDown)
	data.Active = types.BoolValue(sftpMonitor.IsActive)
	data.Hostname = types.StringValue(sftpMonitor.Hostname)
	data.Port = int64PtrToTypes(sftpMonitor.Port)
	data.Timeout = timeoutValueOrDefault(
		ctx, sftpMonitor.ID, sftpMonitor.Timeout, defaultSFTPTimeout, sftpMonitor.Type(), diags,
	)
	data.SSHUsername = types.StringValue(sftpMonitor.SSHUsername)

	// The server reads a NULL column back as password, so an empty value only
	// shows up for a monitor that was never saved through the server.
	authMethod := sftpMonitor.SSHAuthMethod
	if authMethod == "" {
		authMethod = monitor.SFTPAuthMethodPassword
	}

	data.SSHAuthMethod = types.StringValue(string(authMethod))

	// The server returns the credentials to the owning user's socket, so they
	// round-trip and an import picks them up as well.
	data.SSHPassword = ptrToTypes(sftpMonitor.SSHPassword)
	data.SSHPrivateKey = ptrToTypes(sftpMonitor.SSHPrivateKey)
	data.SSHPassphrase = ptrToTypes(sftpMonitor.SSHPassphrase)
	data.SFTPPath = ptrToTypes(sftpMonitor.SFTPPath)
}

// populateSFTPOptionalFields populates optional and computed fields from the API response.
func populateSFTPOptionalFields(
	ctx context.Context,
	sftpMonitor *monitor.SFTP,
	data *MonitorSFTPResourceModel,
	diags *diag.Diagnostics,
) {
	if sftpMonitor.Parent != nil {
		data.Parent = types.Int64Value(*sftpMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(sftpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, sftpMonitor.NotificationIDs)
		diags.Append(d...)

		if diags.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, sftpMonitor.Tags, data.Tags, diags)
}
