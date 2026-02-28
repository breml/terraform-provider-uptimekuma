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
	_ resource.Resource                = &NotificationBitrix24Resource{}
	_ resource.ResourceWithImportState = &NotificationBitrix24Resource{}
)

// NewNotificationBitrix24Resource returns a new instance of the Bitrix24 notification resource.
func NewNotificationBitrix24Resource() resource.Resource {
	return &NotificationBitrix24Resource{}
}

// NotificationBitrix24Resource defines the resource implementation.
type NotificationBitrix24Resource struct {
	client *kuma.Client
}

// NotificationBitrix24ResourceModel describes the resource data model.
type NotificationBitrix24ResourceModel struct {
	NotificationBaseModel

	WebhookURL         types.String `tfsdk:"webhook_url"`
	NotificationUserID types.String `tfsdk:"notification_user_id"`
}

// Metadata returns the metadata for the resource.
func (*NotificationBitrix24Resource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_bitrix24"
}

// Schema returns the schema for the resource.
func (*NotificationBitrix24Resource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Bitrix24 notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "Bitrix24 webhook URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"notification_user_id": schema.StringAttribute{
				MarkdownDescription: "Bitrix24 user ID to receive the notification",
				Required:            true,
			},
		}),
	}
}

// Configure configures the Bitrix24 notification resource with the API client.
func (r *NotificationBitrix24Resource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Bitrix24 notification resource.
func (r *NotificationBitrix24Resource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationBitrix24ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bitrix24 := notification.Bitrix24{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		Bitrix24Details: notification.Bitrix24Details{
			WebhookURL:         data.WebhookURL.ValueString(),
			NotificationUserID: data.NotificationUserID.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, bitrix24)
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

// Read reads the current state of the Bitrix24 notification resource.
func (r *NotificationBitrix24Resource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationBitrix24ResourceModel

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

	bitrix24 := notification.Bitrix24{}
	err = base.As(&bitrix24)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "bitrix24"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(bitrix24.Name)
	data.IsActive = types.BoolValue(bitrix24.IsActive)
	data.IsDefault = types.BoolValue(bitrix24.IsDefault)
	data.ApplyExisting = types.BoolValue(bitrix24.ApplyExisting)

	data.WebhookURL = types.StringValue(bitrix24.WebhookURL)
	data.NotificationUserID = types.StringValue(bitrix24.NotificationUserID)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Bitrix24 notification resource.
func (r *NotificationBitrix24Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationBitrix24ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bitrix24 := notification.Bitrix24{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		Bitrix24Details: notification.Bitrix24Details{
			WebhookURL:         data.WebhookURL.ValueString(),
			NotificationUserID: data.NotificationUserID.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, bitrix24)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Bitrix24 notification resource.
func (r *NotificationBitrix24Resource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationBitrix24ResourceModel

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
func (*NotificationBitrix24Resource) ImportState(
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
