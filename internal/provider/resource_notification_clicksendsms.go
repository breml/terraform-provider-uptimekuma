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
	_ resource.Resource                = &NotificationClicksendSmsResource{}
	_ resource.ResourceWithImportState = &NotificationClicksendSmsResource{}
)

// NewNotificationClicksendSmsResource returns a new instance of the ClickSend SMS notification resource.
func NewNotificationClicksendSmsResource() resource.Resource {
	return &NotificationClicksendSmsResource{}
}

// NotificationClicksendSmsResource defines the resource implementation.
type NotificationClicksendSmsResource struct {
	client *kuma.Client
}

// NotificationClicksendSmsResourceModel describes the resource data model.
type NotificationClicksendSmsResourceModel struct {
	NotificationBaseModel

	Login      types.String `tfsdk:"login"`
	Password   types.String `tfsdk:"password"`
	ToNumber   types.String `tfsdk:"to_number"`
	SenderName types.String `tfsdk:"sender_name"`
}

// Metadata returns the metadata for the resource.
func (*NotificationClicksendSmsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_clicksendsms"
}

// Schema returns the schema for the resource.
func (*NotificationClicksendSmsResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ClickSend SMS notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"login": schema.StringAttribute{
				MarkdownDescription: "ClickSend account username",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "ClickSend API key",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to_number": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"sender_name": schema.StringAttribute{
				MarkdownDescription: "Sender name or phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the ClickSend SMS notification resource with the API client.
func (r *NotificationClicksendSmsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new ClickSend SMS notification resource.
func (r *NotificationClicksendSmsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationClicksendSmsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clicksendSms := notification.ClickSendSMS{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		ClickSendSMSDetails: notification.ClickSendSMSDetails{
			Login:      data.Login.ValueString(),
			Password:   data.Password.ValueString(),
			ToNumber:   data.ToNumber.ValueString(),
			SenderName: data.SenderName.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, clicksendSms)
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

// Read reads the current state of the ClickSend SMS notification resource.
func (r *NotificationClicksendSmsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationClicksendSmsResourceModel

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

	clicksendSms := notification.ClickSendSMS{}
	err = base.As(&clicksendSms)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "clicksendsms"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(clicksendSms.Name)
	data.IsActive = types.BoolValue(clicksendSms.IsActive)
	data.IsDefault = types.BoolValue(clicksendSms.IsDefault)
	data.ApplyExisting = types.BoolValue(clicksendSms.ApplyExisting)

	data.Login = types.StringValue(clicksendSms.Login)
	data.Password = types.StringValue(clicksendSms.Password)
	data.ToNumber = types.StringValue(clicksendSms.ToNumber)
	data.SenderName = types.StringValue(clicksendSms.SenderName)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the ClickSend SMS notification resource.
func (r *NotificationClicksendSmsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationClicksendSmsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clicksendSms := notification.ClickSendSMS{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		ClickSendSMSDetails: notification.ClickSendSMSDetails{
			Login:      data.Login.ValueString(),
			Password:   data.Password.ValueString(),
			ToNumber:   data.ToNumber.ValueString(),
			SenderName: data.SenderName.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, clicksendSms)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the ClickSend SMS notification resource.
func (r *NotificationClicksendSmsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationClicksendSmsResourceModel

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
func (*NotificationClicksendSmsResource) ImportState(
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
