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
	_ resource.Resource                = &NotificationGorushResource{}
	_ resource.ResourceWithImportState = &NotificationGorushResource{}
)

// NewNotificationGorushResource returns a new instance of the Gorush notification resource.
func NewNotificationGorushResource() resource.Resource {
	return &NotificationGorushResource{}
}

// NotificationGorushResource defines the resource implementation.
type NotificationGorushResource struct {
	client *kuma.Client
}

// NotificationGorushResourceModel describes the resource data model.
type NotificationGorushResourceModel struct {
	NotificationBaseModel

	ServerURL   types.String `tfsdk:"server_url"`
	DeviceToken types.String `tfsdk:"device_token"`
	Platform    types.String `tfsdk:"platform"`
	Title       types.String `tfsdk:"title"`
	Priority    types.String `tfsdk:"priority"`
	Retry       types.Int64  `tfsdk:"retry"`
	Topic       types.String `tfsdk:"topic"`
}

// Metadata returns the metadata for the resource.
func (*NotificationGorushResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_gorush"
}

// Schema returns the schema for the resource.
func (*NotificationGorushResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gorush notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				MarkdownDescription: "Gorush server URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"device_token": schema.StringAttribute{
				MarkdownDescription: "Device token for the push notification",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "Platform (ios, android, huawei, web)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Notification title",
				Optional:            true,
			},
			"priority": schema.StringAttribute{
				MarkdownDescription: "Notification priority",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"retry": schema.Int64Attribute{
				MarkdownDescription: "Number of retries",
				Optional:            true,
				Computed:            true,
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "Notification topic",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Gorush notification resource with the API client.
func (r *NotificationGorushResource) Configure(
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

// Create creates a new Gorush notification resource.
func (r *NotificationGorushResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationGorushResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	gorush := notification.Gorush{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GorushDetails: notification.GorushDetails{
			ServerURL:   data.ServerURL.ValueString(),
			DeviceToken: data.DeviceToken.ValueString(),
			Platform:    data.Platform.ValueString(),
			Title:       data.Title.ValueString(),
			Priority:    data.Priority.ValueString(),
			Retry:       int(data.Retry.ValueInt64()),
			Topic:       data.Topic.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, gorush)
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

// Read reads the current state of the Gorush notification resource.
func (r *NotificationGorushResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationGorushResourceModel

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

	gorush := notification.Gorush{}
	err = base.As(&gorush)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "gorush"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(gorush.Name)
	data.IsActive = types.BoolValue(gorush.IsActive)
	data.IsDefault = types.BoolValue(gorush.IsDefault)
	data.ApplyExisting = types.BoolValue(gorush.ApplyExisting)

	data.ServerURL = types.StringValue(gorush.ServerURL)
	data.DeviceToken = types.StringValue(gorush.DeviceToken)
	data.Platform = types.StringValue(gorush.Platform)
	if gorush.Title != "" {
		data.Title = types.StringValue(gorush.Title)
	}

	if gorush.Priority != "" {
		data.Priority = types.StringValue(gorush.Priority)
	}

	if gorush.Retry > 0 {
		data.Retry = types.Int64Value(int64(gorush.Retry))
	}

	if gorush.Topic != "" {
		data.Topic = types.StringValue(gorush.Topic)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Gorush notification resource.
func (r *NotificationGorushResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationGorushResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	gorush := notification.Gorush{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GorushDetails: notification.GorushDetails{
			ServerURL:   data.ServerURL.ValueString(),
			DeviceToken: data.DeviceToken.ValueString(),
			Platform:    data.Platform.ValueString(),
			Title:       data.Title.ValueString(),
			Priority:    data.Priority.ValueString(),
			Retry:       int(data.Retry.ValueInt64()),
			Topic:       data.Topic.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, gorush)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Gorush notification resource.
func (r *NotificationGorushResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationGorushResourceModel

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
func (*NotificationGorushResource) ImportState(
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
