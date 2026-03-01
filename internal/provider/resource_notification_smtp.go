package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationSMTPResource{}
	_ resource.ResourceWithImportState = &NotificationSMTPResource{}
)

// NewNotificationSMTPResource returns a new instance of the SMTP notification resource.
func NewNotificationSMTPResource() resource.Resource {
	return &NotificationSMTPResource{}
}

// NotificationSMTPResource defines the resource implementation.
type NotificationSMTPResource struct {
	client *kuma.Client
}

// NotificationSMTPResourceModel describes the resource data model.
type NotificationSMTPResourceModel struct {
	NotificationBaseModel

	Host                 types.String `tfsdk:"host"`
	Port                 types.Int64  `tfsdk:"port"`
	Secure               types.Bool   `tfsdk:"secure"`
	IgnoreTLSError       types.Bool   `tfsdk:"ignore_tls_error"`
	DkimDomain           types.String `tfsdk:"dkim_domain"`
	DkimKeySelector      types.String `tfsdk:"dkim_key_selector"`
	DkimPrivateKey       types.String `tfsdk:"dkim_private_key"`
	DkimHashAlgo         types.String `tfsdk:"dkim_hash_algo"`
	DkimHeaderFieldNames types.String `tfsdk:"dkim_header_field_names"`
	DkimSkipFields       types.String `tfsdk:"dkim_skip_fields"`
	Username             types.String `tfsdk:"username"`
	Password             types.String `tfsdk:"password"`
	From                 types.String `tfsdk:"from"`
	CC                   types.String `tfsdk:"cc"`
	BCC                  types.String `tfsdk:"bcc"`
	To                   types.String `tfsdk:"to"`
	CustomSubject        types.String `tfsdk:"custom_subject"`
	CustomBody           types.String `tfsdk:"custom_body"`
	HTMLBody             types.Bool   `tfsdk:"html_body"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSMTPResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_smtp"
}

// Schema returns the schema for the resource.
func (*NotificationSMTPResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SMTP notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "SMTP server hostname",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "SMTP server port",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(587),
			},
			"secure": schema.BoolAttribute{
				MarkdownDescription: "Enable TLS/SSL for SMTP connection",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ignore_tls_error": schema.BoolAttribute{
				MarkdownDescription: "Ignore TLS certificate errors",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"dkim_domain": schema.StringAttribute{
				MarkdownDescription: "DKIM domain for email signing",
				Optional:            true,
			},
			"dkim_key_selector": schema.StringAttribute{
				MarkdownDescription: "DKIM key selector",
				Optional:            true,
			},
			"dkim_private_key": schema.StringAttribute{
				MarkdownDescription: "DKIM private key for email signing",
				Optional:            true,
				Sensitive:           true,
			},
			"dkim_hash_algo": schema.StringAttribute{
				MarkdownDescription: "DKIM hash algorithm (sha1, sha256)",
				Optional:            true,
			},
			"dkim_header_field_names": schema.StringAttribute{
				MarkdownDescription: "DKIM header field names to sign",
				Optional:            true,
			},
			"dkim_skip_fields": schema.StringAttribute{
				MarkdownDescription: "DKIM fields to skip",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "SMTP username for authentication",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "SMTP password for authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"from": schema.StringAttribute{
				MarkdownDescription: "Sender email address",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cc": schema.StringAttribute{
				MarkdownDescription: "CC email addresses (comma-separated)",
				Optional:            true,
			},
			"bcc": schema.StringAttribute{
				MarkdownDescription: "BCC email addresses (comma-separated)",
				Optional:            true,
			},
			"to": schema.StringAttribute{
				MarkdownDescription: "Recipient email addresses (comma-separated)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"custom_subject": schema.StringAttribute{
				MarkdownDescription: "Custom email subject",
				Optional:            true,
			},
			"custom_body": schema.StringAttribute{
				MarkdownDescription: "Custom email body",
				Optional:            true,
			},
			"html_body": schema.BoolAttribute{
				MarkdownDescription: "Enable HTML formatting in email body",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		}),
	}
}

// Configure configures the SMTP notification resource with the API client.
func (r *NotificationSMTPResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new SMTP notification resource.
func (r *NotificationSMTPResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSMTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smtp := notification.SMTP{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMTPDetails: notification.SMTPDetails{
			Host:                 data.Host.ValueString(),
			Port:                 int(data.Port.ValueInt64()),
			Secure:               data.Secure.ValueBool(),
			IgnoreTLSError:       data.IgnoreTLSError.ValueBool(),
			DkimDomain:           data.DkimDomain.ValueString(),
			DkimKeySelector:      data.DkimKeySelector.ValueString(),
			DkimPrivateKey:       data.DkimPrivateKey.ValueString(),
			DkimHashAlgo:         data.DkimHashAlgo.ValueString(),
			DkimHeaderFieldNames: data.DkimHeaderFieldNames.ValueString(),
			DkimSkipFields:       data.DkimSkipFields.ValueString(),
			Username:             data.Username.ValueString(),
			Password:             data.Password.ValueString(),
			From:                 data.From.ValueString(),
			CC:                   data.CC.ValueString(),
			BCC:                  data.BCC.ValueString(),
			To:                   data.To.ValueString(),
			CustomSubject:        data.CustomSubject.ValueString(),
			CustomBody:           data.CustomBody.ValueString(),
			HTMLBody:             data.HTMLBody.ValueBool(),
		},
	}

	id, err := r.client.CreateNotification(ctx, smtp)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the SMTP notification resource.
func (r *NotificationSMTPResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSMTPResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, err := r.client.GetNotification(ctx, id)
	// Handle error.
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	smtp := notification.SMTP{}
	err = base.As(&smtp)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "smtp"`, err.Error())
		return
	}

	populateSMTPModelFromAPI(&data, id, &smtp)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SMTP notification resource.
func (r *NotificationSMTPResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSMTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smtp := notification.SMTP{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMTPDetails: notification.SMTPDetails{
			Host:                 data.Host.ValueString(),
			Port:                 int(data.Port.ValueInt64()),
			Secure:               data.Secure.ValueBool(),
			IgnoreTLSError:       data.IgnoreTLSError.ValueBool(),
			DkimDomain:           data.DkimDomain.ValueString(),
			DkimKeySelector:      data.DkimKeySelector.ValueString(),
			DkimPrivateKey:       data.DkimPrivateKey.ValueString(),
			DkimHashAlgo:         data.DkimHashAlgo.ValueString(),
			DkimHeaderFieldNames: data.DkimHeaderFieldNames.ValueString(),
			DkimSkipFields:       data.DkimSkipFields.ValueString(),
			Username:             data.Username.ValueString(),
			Password:             data.Password.ValueString(),
			From:                 data.From.ValueString(),
			CC:                   data.CC.ValueString(),
			BCC:                  data.BCC.ValueString(),
			To:                   data.To.ValueString(),
			CustomSubject:        data.CustomSubject.ValueString(),
			CustomBody:           data.CustomBody.ValueString(),
			HTMLBody:             data.HTMLBody.ValueBool(),
		},
	}

	err := r.client.UpdateNotification(ctx, smtp)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SMTP notification resource.
func (r *NotificationSMTPResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSMTPResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*NotificationSMTPResource) ImportState(
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

// populateSMTPModelFromAPI populates the model from API response data.
func populateSMTPModelFromAPI(
	data *NotificationSMTPResourceModel,
	id int64,
	smtp *notification.SMTP,
) {
	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(smtp.Name)
	data.IsActive = types.BoolValue(smtp.IsActive)
	data.IsDefault = types.BoolValue(smtp.IsDefault)
	data.ApplyExisting = types.BoolValue(smtp.ApplyExisting)

	data.Host = types.StringValue(smtp.Host)
	data.Port = types.Int64Value(int64(smtp.Port))
	data.Secure = types.BoolValue(smtp.Secure)
	data.IgnoreTLSError = types.BoolValue(smtp.IgnoreTLSError)

	if smtp.DkimDomain != "" {
		data.DkimDomain = types.StringValue(smtp.DkimDomain)
	}

	if smtp.DkimKeySelector != "" {
		data.DkimKeySelector = types.StringValue(smtp.DkimKeySelector)
	}

	if smtp.DkimPrivateKey != "" {
		data.DkimPrivateKey = types.StringValue(smtp.DkimPrivateKey)
	}

	if smtp.DkimHashAlgo != "" {
		data.DkimHashAlgo = types.StringValue(smtp.DkimHashAlgo)
	}

	if smtp.DkimHeaderFieldNames != "" {
		data.DkimHeaderFieldNames = types.StringValue(smtp.DkimHeaderFieldNames)
	}

	if smtp.DkimSkipFields != "" {
		data.DkimSkipFields = types.StringValue(smtp.DkimSkipFields)
	}

	if smtp.Username != "" {
		data.Username = types.StringValue(smtp.Username)
	}

	if smtp.Password != "" {
		data.Password = types.StringValue(smtp.Password)
	}

	data.From = types.StringValue(smtp.From)
	if smtp.CC != "" {
		data.CC = types.StringValue(smtp.CC)
	}

	if smtp.BCC != "" {
		data.BCC = types.StringValue(smtp.BCC)
	}

	data.To = types.StringValue(smtp.To)

	if smtp.CustomSubject != "" {
		data.CustomSubject = types.StringValue(smtp.CustomSubject)
	}

	if smtp.CustomBody != "" {
		data.CustomBody = types.StringValue(smtp.CustomBody)
	}

	data.HTMLBody = types.BoolValue(smtp.HTMLBody)
}
