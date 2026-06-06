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
	_ resource.Resource                = &NotificationBaleResource{}
	_ resource.ResourceWithImportState = &NotificationBaleResource{}
)

// NewNotificationBaleResource returns a new instance of the Bale notification resource.
func NewNotificationBaleResource() resource.Resource {
	return &NotificationBaleResource{}
}

// NotificationBaleResource defines the resource implementation.
type NotificationBaleResource struct {
	client *kuma.Client
}

// NotificationBaleResourceModel describes the resource data model.
type NotificationBaleResourceModel struct {
	NotificationBaseModel

	BotToken types.String `tfsdk:"bot_token"`
	ChatID   types.String `tfsdk:"chat_id"`
}

// Metadata returns the metadata for the resource.
func (*NotificationBaleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_bale"
}

// Schema returns the schema for the resource.
func (*NotificationBaleResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Bale notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"bot_token": schema.StringAttribute{
				MarkdownDescription: "Bale bot API token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"chat_id": schema.StringAttribute{
				MarkdownDescription: "Target chat ID",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Bale notification resource with the API client.
func (r *NotificationBaleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Bale notification resource.
func (r *NotificationBaleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationBaleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bale := notification.Bale{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		BaleDetails: notification.BaleDetails{
			BotToken: data.BotToken.ValueString(),
			ChatID:   data.ChatID.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, bale)
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

// Read reads the current state of the Bale notification resource.
func (r *NotificationBaleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationBaleResourceModel

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

	bale := notification.Bale{}
	err = base.As(&bale)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "bale"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(bale.Name)
	data.IsActive = types.BoolValue(bale.IsActive)
	data.IsDefault = types.BoolValue(bale.IsDefault)
	data.ApplyExisting = types.BoolValue(bale.ApplyExisting)

	data.BotToken = types.StringValue(bale.BotToken)
	data.ChatID = types.StringValue(bale.ChatID)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Bale notification resource.
func (r *NotificationBaleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationBaleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bale := notification.Bale{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		BaleDetails: notification.BaleDetails{
			BotToken: data.BotToken.ValueString(),
			ChatID:   data.ChatID.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, bale)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Bale notification resource.
func (r *NotificationBaleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationBaleResourceModel

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
func (*NotificationBaleResource) ImportState(
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
