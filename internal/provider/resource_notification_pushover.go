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
	_ resource.Resource                = &NotificationPushoverResource{}
	_ resource.ResourceWithImportState = &NotificationPushoverResource{}
)

// NewNotificationPushoverResource returns a new instance of the Pushover notification resource.
func NewNotificationPushoverResource() resource.Resource {
	return &NotificationPushoverResource{}
}

// NotificationPushoverResource defines the resource implementation.
type NotificationPushoverResource struct {
	client *kuma.Client
}

// NotificationPushoverResourceModel describes the resource data model.
type NotificationPushoverResourceModel struct {
	NotificationBaseModel

	UserKey  types.String `tfsdk:"user_key"`
	AppToken types.String `tfsdk:"app_token"`
	Sounds   types.String `tfsdk:"sounds"`
	SoundsUp types.String `tfsdk:"sounds_up"`
	Priority types.String `tfsdk:"priority"`
	Title    types.String `tfsdk:"title"`
	Device   types.String `tfsdk:"device"`
	TTL      types.String `tfsdk:"ttl"`
}

// Metadata returns the metadata for the resource.
func (*NotificationPushoverResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_pushover"
}

// Schema returns the schema for the resource.
func (*NotificationPushoverResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pushover notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"user_key": schema.StringAttribute{
				MarkdownDescription: "Pushover user key",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"app_token": schema.StringAttribute{
				MarkdownDescription: "Pushover application token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"sounds": schema.StringAttribute{
				MarkdownDescription: "Notification sound",
				Optional:            true,
			},
			"sounds_up": schema.StringAttribute{
				MarkdownDescription: "Notification sound when monitor is up",
				Optional:            true,
			},
			"priority": schema.StringAttribute{
				MarkdownDescription: "Priority level for notifications",
				Optional:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Notification title",
				Optional:            true,
			},
			"device": schema.StringAttribute{
				MarkdownDescription: "Device name to receive notifications",
				Optional:            true,
			},
			"ttl": schema.StringAttribute{
				MarkdownDescription: "Time to live for notifications",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Pushover notification resource with the API client.
func (r *NotificationPushoverResource) Configure(
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

// Create creates a new Pushover notification resource.
func (r *NotificationPushoverResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationPushoverResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pushover := notification.Pushover{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PushoverDetails: notification.PushoverDetails{
			UserKey:  data.UserKey.ValueString(),
			AppToken: data.AppToken.ValueString(),
			Sounds:   data.Sounds.ValueString(),
			SoundsUp: data.SoundsUp.ValueString(),
			Priority: data.Priority.ValueString(),
			Title:    data.Title.ValueString(),
			Device:   data.Device.ValueString(),
			TTL:      data.TTL.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, pushover)
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

// Read reads the current state of the Pushover notification resource.
func (r *NotificationPushoverResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationPushoverResourceModel

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

	pushover := notification.Pushover{}
	err = base.As(&pushover)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "pushover"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(pushover.Name)
	data.IsActive = types.BoolValue(pushover.IsActive)
	data.IsDefault = types.BoolValue(pushover.IsDefault)
	data.ApplyExisting = types.BoolValue(pushover.ApplyExisting)

	data.UserKey = types.StringValue(pushover.UserKey)
	data.AppToken = types.StringValue(pushover.AppToken)

	if pushover.Sounds != "" {
		data.Sounds = types.StringValue(pushover.Sounds)
	}

	if pushover.SoundsUp != "" {
		data.SoundsUp = types.StringValue(pushover.SoundsUp)
	}

	if pushover.Priority != "" {
		data.Priority = types.StringValue(pushover.Priority)
	}

	if pushover.Title != "" {
		data.Title = types.StringValue(pushover.Title)
	}

	if pushover.Device != "" {
		data.Device = types.StringValue(pushover.Device)
	}

	if pushover.TTL != "" {
		data.TTL = types.StringValue(pushover.TTL)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Pushover notification resource.
func (r *NotificationPushoverResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationPushoverResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pushover := notification.Pushover{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PushoverDetails: notification.PushoverDetails{
			UserKey:  data.UserKey.ValueString(),
			AppToken: data.AppToken.ValueString(),
			Sounds:   data.Sounds.ValueString(),
			SoundsUp: data.SoundsUp.ValueString(),
			Priority: data.Priority.ValueString(),
			Title:    data.Title.ValueString(),
			Device:   data.Device.ValueString(),
			TTL:      data.TTL.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, pushover)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Pushover notification resource.
func (r *NotificationPushoverResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationPushoverResourceModel

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
func (*NotificationPushoverResource) ImportState(
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
