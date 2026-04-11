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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationSMSManagerResource{}
	_ resource.ResourceWithImportState = &NotificationSMSManagerResource{}
)

// NewNotificationSMSManagerResource returns a new instance of the SMS Manager notification resource.
func NewNotificationSMSManagerResource() resource.Resource {
	return &NotificationSMSManagerResource{}
}

// NotificationSMSManagerResource defines the resource implementation.
type NotificationSMSManagerResource struct {
	client *kuma.Client
}

// NotificationSMSManagerResourceModel describes the resource data model.
type NotificationSMSManagerResourceModel struct {
	NotificationBaseModel

	APIKey      types.String `tfsdk:"api_key"`
	Numbers     types.String `tfsdk:"numbers"`
	MessageType types.String `tfsdk:"message_type"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSMSManagerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_smsmanager"
}

// Schema returns the schema for the resource.
func (*NotificationSMSManagerResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SMS Manager notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "SMS Manager API key for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"numbers": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"message_type": schema.StringAttribute{
				MarkdownDescription: "Message gateway type",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("sms"),
			},
		}),
	}
}

// Configure configures the SMS Manager notification resource with the API client.
func (r *NotificationSMSManagerResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new SMS Manager notification resource.
func (r *NotificationSMSManagerResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSMSManagerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smsmanager := notification.SMSManager{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMSManagerDetails: notification.SMSManagerDetails{
			APIKey:      data.APIKey.ValueString(),
			Numbers:     data.Numbers.ValueString(),
			MessageType: data.MessageType.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, smsmanager)
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

// Read reads the current state of the SMS Manager notification resource.
func (r *NotificationSMSManagerResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSMSManagerResourceModel

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

	smsmanager := notification.SMSManager{}
	err = base.As(&smsmanager)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "SMSManager"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(smsmanager.Name)
	data.IsActive = types.BoolValue(smsmanager.IsActive)
	data.IsDefault = types.BoolValue(smsmanager.IsDefault)
	data.ApplyExisting = types.BoolValue(smsmanager.ApplyExisting)

	data.APIKey = types.StringValue(smsmanager.APIKey)
	data.Numbers = types.StringValue(smsmanager.Numbers)
	data.MessageType = types.StringValue(smsmanager.MessageType)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SMS Manager notification resource.
func (r *NotificationSMSManagerResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSMSManagerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smsmanager := notification.SMSManager{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMSManagerDetails: notification.SMSManagerDetails{
			APIKey:      data.APIKey.ValueString(),
			Numbers:     data.Numbers.ValueString(),
			MessageType: data.MessageType.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, smsmanager)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SMS Manager notification resource.
func (r *NotificationSMSManagerResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSMSManagerResourceModel

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
func (*NotificationSMSManagerResource) ImportState(
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
