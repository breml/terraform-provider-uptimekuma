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
	_ resource.Resource                = &NotificationPushDeerResource{}
	_ resource.ResourceWithImportState = &NotificationPushDeerResource{}
)

// NewNotificationPushDeerResource returns a new instance of the PushDeer notification resource.
func NewNotificationPushDeerResource() resource.Resource {
	return &NotificationPushDeerResource{}
}

// NotificationPushDeerResource defines the resource implementation.
type NotificationPushDeerResource struct {
	client *kuma.Client
}

// NotificationPushDeerResourceModel describes the resource data model.
type NotificationPushDeerResourceModel struct {
	NotificationBaseModel

	Key    types.String `tfsdk:"key"`
	Server types.String `tfsdk:"server"`
}

// Metadata returns the metadata for the resource.
func (*NotificationPushDeerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_pushdeer"
}

// Schema returns the schema for the resource.
func (*NotificationPushDeerResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PushDeer notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "PushDeer push key",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"server": schema.StringAttribute{
				MarkdownDescription: "Custom PushDeer server URL (optional, defaults to https://api2.pushdeer.com)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("https://api2.pushdeer.com"),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validateURL(),
				},
			},
		}),
	}
}

// Configure configures the PushDeer notification resource with the API client.
func (r *NotificationPushDeerResource) Configure(
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

// Create creates a new PushDeer notification resource.
func (r *NotificationPushDeerResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationPushDeerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pushDeer := notification.PushDeer{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PushDeerDetails: notification.PushDeerDetails{
			Key:    data.Key.ValueString(),
			Server: data.Server.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, pushDeer)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the PushDeer notification resource.
func (r *NotificationPushDeerResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationPushDeerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, err := r.client.GetNotification(ctx, id)
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	pushDeer := notification.PushDeer{}
	err = base.As(&pushDeer)
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "pushdeer"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(pushDeer.Name)
	data.IsActive = types.BoolValue(pushDeer.IsActive)
	data.IsDefault = types.BoolValue(pushDeer.IsDefault)
	data.ApplyExisting = types.BoolValue(pushDeer.ApplyExisting)

	data.Key = types.StringValue(pushDeer.Key)
	data.Server = types.StringValue(pushDeer.Server)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the PushDeer notification resource.
func (r *NotificationPushDeerResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationPushDeerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pushDeer := notification.PushDeer{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PushDeerDetails: notification.PushDeerDetails{
			Key:    data.Key.ValueString(),
			Server: data.Server.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, pushDeer)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the PushDeer notification resource.
func (r *NotificationPushDeerResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationPushDeerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*NotificationPushDeerResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
