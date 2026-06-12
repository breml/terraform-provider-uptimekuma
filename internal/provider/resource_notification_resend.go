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
	_ resource.Resource                = &NotificationResendResource{}
	_ resource.ResourceWithImportState = &NotificationResendResource{}
)

// NewNotificationResendResource returns a new instance of the Resend notification resource.
func NewNotificationResendResource() resource.Resource {
	return &NotificationResendResource{}
}

// NotificationResendResource defines the resource implementation.
type NotificationResendResource struct {
	client *kuma.Client
}

// NotificationResendResourceModel describes the resource data model.
type NotificationResendResourceModel struct {
	NotificationBaseModel

	APIKey    types.String `tfsdk:"api_key"`
	FromEmail types.String `tfsdk:"from_email"`
	FromName  types.String `tfsdk:"from_name"`
	ToEmail   types.String `tfsdk:"to_email"`
	Subject   types.String `tfsdk:"subject"`
}

// Metadata returns the metadata for the resource.
func (*NotificationResendResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_resend"
}

// Schema returns the schema for the resource.
func (*NotificationResendResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resend notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The Resend API key used to send notifications.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"from_email": schema.StringAttribute{
				MarkdownDescription: "The sender email address.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"from_name": schema.StringAttribute{
				MarkdownDescription: "The sender display name.",
				Optional:            true,
			},
			"to_email": schema.StringAttribute{
				MarkdownDescription: "The recipient email address(es). Multiple recipients can be specified " +
					"as a comma-separated list.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "The subject of the email notification.",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Resend notification resource with the API client.
func (r *NotificationResendResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Resend notification resource.
func (r *NotificationResendResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationResendResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resend := resendFromModel(&data)

	id, err := r.client.CreateNotification(ctx, resend)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Resend notification resource.
func (r *NotificationResendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationResendResourceModel

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

	resend := notification.Resend{}
	err = base.As(&resend)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "Resend"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(resend.Name)
	data.IsActive = types.BoolValue(resend.IsActive)
	data.IsDefault = types.BoolValue(resend.IsDefault)
	data.ApplyExisting = types.BoolValue(resend.ApplyExisting)

	data.APIKey = types.StringValue(resend.APIKey)
	data.FromEmail = types.StringValue(resend.FromEmail)
	data.ToEmail = types.StringValue(resend.ToEmail)

	if resend.FromName != "" {
		data.FromName = types.StringValue(resend.FromName)
	} else {
		data.FromName = types.StringNull()
	}

	if resend.Subject != "" {
		data.Subject = types.StringValue(resend.Subject)
	} else {
		data.Subject = types.StringNull()
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Resend notification resource.
func (r *NotificationResendResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationResendResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if id == 0 {
		resp.Diagnostics.AddError(
			"Invalid resource state",
			"Cannot update notification: resource ID is missing from state. This is a provider bug.",
		)

		return
	}

	resend := resendFromModel(&data)
	resend.ID = id

	err := r.client.UpdateNotification(ctx, resend)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Resend notification resource.
func (r *NotificationResendResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationResendResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			return
		}

		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*NotificationResendResource) ImportState(
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

// resendFromModel builds a Resend notification from the resource model.
func resendFromModel(data *NotificationResendResourceModel) notification.Resend {
	return notification.Resend{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		ResendDetails: notification.ResendDetails{
			APIKey:    data.APIKey.ValueString(),
			FromEmail: data.FromEmail.ValueString(),
			FromName:  data.FromName.ValueString(),
			ToEmail:   data.ToEmail.ValueString(),
			Subject:   data.Subject.ValueString(),
		},
	}
}
