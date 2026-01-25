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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationSendgridResource{}
	_ resource.ResourceWithImportState = &NotificationSendgridResource{}
)

// NewNotificationSendgridResource returns a new instance of the SendGrid notification resource.
func NewNotificationSendgridResource() resource.Resource {
	return &NotificationSendgridResource{}
}

// NotificationSendgridResource defines the resource implementation.
type NotificationSendgridResource struct {
	client *kuma.Client
}

// NotificationSendgridResourceModel describes the resource data model.
type NotificationSendgridResourceModel struct {
	NotificationBaseModel

	APIKey    types.String `tfsdk:"api_key"`
	ToEmail   types.String `tfsdk:"to_email"`
	FromEmail types.String `tfsdk:"from_email"`
	Subject   types.String `tfsdk:"subject"`
	CcEmail   types.String `tfsdk:"cc_email"`
	BccEmail  types.String `tfsdk:"bcc_email"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSendgridResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_sendgrid"
}

// Schema returns the schema for the resource.
func (*NotificationSendgridResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SendGrid notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "SendGrid API key",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to_email": schema.StringAttribute{
				MarkdownDescription: "Recipient email address",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"from_email": schema.StringAttribute{
				MarkdownDescription: "Sender email address",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "Email subject",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cc_email": schema.StringAttribute{
				MarkdownDescription: "CC email addresses",
				Optional:            true,
			},
			"bcc_email": schema.StringAttribute{
				MarkdownDescription: "BCC email addresses",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the SendGrid notification resource with the API client.
func (r *NotificationSendgridResource) Configure(
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

// Create creates a new SendGrid notification resource.
func (r *NotificationSendgridResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSendgridResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sendgrid := notification.SendGrid{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SendGridDetails: notification.SendGridDetails{
			APIKey:    data.APIKey.ValueString(),
			ToEmail:   data.ToEmail.ValueString(),
			FromEmail: data.FromEmail.ValueString(),
			Subject:   data.Subject.ValueString(),
			CcEmail:   data.CcEmail.ValueString(),
			BccEmail:  data.BccEmail.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, sendgrid)
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

// Read reads the current state of the SendGrid notification resource.
func (r *NotificationSendgridResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSendgridResourceModel

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

	sendgrid := notification.SendGrid{}
	err = base.As(&sendgrid)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "sendgrid"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(sendgrid.Name)
	data.IsActive = types.BoolValue(sendgrid.IsActive)
	data.IsDefault = types.BoolValue(sendgrid.IsDefault)
	data.ApplyExisting = types.BoolValue(sendgrid.ApplyExisting)

	data.APIKey = types.StringValue(sendgrid.APIKey)
	data.ToEmail = types.StringValue(sendgrid.ToEmail)
	data.FromEmail = types.StringValue(sendgrid.FromEmail)
	data.Subject = types.StringValue(sendgrid.Subject)

	if sendgrid.CcEmail != "" {
		data.CcEmail = types.StringValue(sendgrid.CcEmail)
	}

	if sendgrid.BccEmail != "" {
		data.BccEmail = types.StringValue(sendgrid.BccEmail)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SendGrid notification resource.
func (r *NotificationSendgridResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSendgridResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sendgrid := notification.SendGrid{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SendGridDetails: notification.SendGridDetails{
			APIKey:    data.APIKey.ValueString(),
			ToEmail:   data.ToEmail.ValueString(),
			FromEmail: data.FromEmail.ValueString(),
			Subject:   data.Subject.ValueString(),
			CcEmail:   data.CcEmail.ValueString(),
			BccEmail:  data.BccEmail.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, sendgrid)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SendGrid notification resource.
func (r *NotificationSendgridResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSendgridResourceModel

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
func (*NotificationSendgridResource) ImportState(
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
