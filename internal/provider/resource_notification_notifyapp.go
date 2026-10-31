package provider

import (
	"context"
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
	_ resource.Resource                = &NotificationNotifyAppResource{}
	_ resource.ResourceWithImportState = &NotificationNotifyAppResource{}
)

// NewNotificationNotifyAppResource returns a new instance of the Notify! notification resource.
func NewNotificationNotifyAppResource() resource.Resource {
	return &NotificationNotifyAppResource{}
}

// NotificationNotifyAppResource defines the resource implementation.
type NotificationNotifyAppResource struct {
	client *kuma.Client
}

// NotificationNotifyAppResourceModel describes the resource data model.
type NotificationNotifyAppResourceModel struct {
	NotificationBaseModel

	DeviceID       types.String `tfsdk:"device_id"`
	Token          types.String `tfsdk:"token"`
	TokenWO        types.String `tfsdk:"token_wo"`
	TokenWOVersion types.String `tfsdk:"token_wo_version"`
	IconURL        types.String `tfsdk:"icon_url"`
}

// Metadata returns the metadata for the resource.
func (*NotificationNotifyAppResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_notifyapp"
}

// Schema returns the schema for the resource.
func (*NotificationNotifyAppResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Notify! notification resource. Notify! is a push notification app. " +
			"Requires Uptime Kuma 2.5.4 or later.",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				MarkdownDescription: "The Notify! device or group ID the push is sent to.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The token that authenticates the push.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"icon_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the icon shown with the push.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}

	resp.Schema.Attributes = withWriteOnlySecret(resp.Schema.Attributes, "token", "Notify! token")
}

// Configure configures the Notify! notification resource with the API client.
func (r *NotificationNotifyAppResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Notify! notification resource.
func (r *NotificationNotifyAppResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationNotifyAppResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	notifyApp := notifyAppFromModel(&data)
	notifyApp.Token = writeOnlyOrPlan(ctx, req.Config, "token", data.Token, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateNotification(ctx, notifyApp)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Notify! notification resource.
func (r *NotificationNotifyAppResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationNotifyAppResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.NotifyAppDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	notifyApp := notification.NotifyApp{}
	err := base.As(&notifyApp)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.NotifyAppDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(notifyApp.Name)
	data.IsDefault = types.BoolValue(notifyApp.IsDefault)
	data.ApplyExisting = types.BoolValue(notifyApp.ApplyExisting)
	data.DeviceID = types.StringValue(notifyApp.DeviceID)
	data.IconURL = ptrToTypes(notifyApp.IconURL)

	if data.TokenWOVersion.IsNull() {
		data.Token = types.StringValue(notifyApp.Token)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Notify! notification resource.
func (r *NotificationNotifyAppResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationNotifyAppResourceModel

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

	notifyApp := notifyAppFromModel(&data)
	notifyApp.Token = writeOnlyOrPlan(ctx, req.Config, "token", data.Token, &resp.Diagnostics)
	notifyApp.ID = id

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNotification(ctx, notifyApp)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Notify! notification resource.
func (r *NotificationNotifyAppResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationNotifyAppResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationNotifyAppResource) ImportState(
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

// notifyAppFromModel builds a NotifyApp notification from the resource model.
func notifyAppFromModel(data *NotificationNotifyAppResourceModel) notification.NotifyApp {
	return notification.NotifyApp{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      true,
			Name:          data.Name.ValueString(),
		},
		NotifyAppDetails: notification.NotifyAppDetails{
			DeviceID: data.DeviceID.ValueString(),
			Token:    data.Token.ValueString(),
			IconURL:  strToPtr(data.IconURL),
		},
	}
}
