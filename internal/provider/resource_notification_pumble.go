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
	_ resource.Resource                = &NotificationPumbleResource{}
	_ resource.ResourceWithImportState = &NotificationPumbleResource{}
)

// NewNotificationPumbleResource returns a new instance of the Pumble notification resource.
func NewNotificationPumbleResource() resource.Resource {
	return &NotificationPumbleResource{}
}

// NotificationPumbleResource defines the resource implementation for Pumble notifications.
type NotificationPumbleResource struct {
	client *kuma.Client
}

// NotificationPumbleResourceModel describes the resource data model for Pumble notifications.
type NotificationPumbleResourceModel struct {
	NotificationBaseModel

	WebhookURL types.String `tfsdk:"webhook_url"`
}

// Metadata returns the metadata for the resource.
func (*NotificationPumbleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_pumble"
}

// Schema returns the schema for the resource.
func (*NotificationPumbleResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pumble notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "Pumble webhook URL for sending notifications",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Pumble notification resource with the API client.
func (r *NotificationPumbleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Pumble notification resource.
func (r *NotificationPumbleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationPumbleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pumble := notification.Pumble{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PumbleDetails: notification.PumbleDetails{
			WebhookURL: data.WebhookURL.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, pumble)
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

// Read reads the current state of the Pumble notification resource.
func (r *NotificationPumbleResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationPumbleResourceModel

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

	pumble := notification.Pumble{}
	err = base.As(&pumble)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "Pumble"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(pumble.Name)
	data.IsActive = types.BoolValue(pumble.IsActive)
	data.IsDefault = types.BoolValue(pumble.IsDefault)
	data.ApplyExisting = types.BoolValue(pumble.ApplyExisting)

	data.WebhookURL = types.StringValue(pumble.WebhookURL)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Pumble notification resource.
func (r *NotificationPumbleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationPumbleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pumble := notification.Pumble{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PumbleDetails: notification.PumbleDetails{
			WebhookURL: data.WebhookURL.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, pumble)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Pumble notification resource.
func (r *NotificationPumbleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationPumbleResourceModel

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
func (*NotificationPumbleResource) ImportState(
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
