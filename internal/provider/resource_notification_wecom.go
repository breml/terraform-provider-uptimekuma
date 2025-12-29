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
	_ resource.Resource                = &NotificationWeComResource{}
	_ resource.ResourceWithImportState = &NotificationWeComResource{}
)

// NewNotificationWeComResource returns a new instance of the WeCom notification resource.
func NewNotificationWeComResource() resource.Resource {
	return &NotificationWeComResource{}
}

// NotificationWeComResource defines the resource implementation.
type NotificationWeComResource struct {
	client *kuma.Client
}

// NotificationWeComResourceModel describes the resource data model.
type NotificationWeComResourceModel struct {
	NotificationBaseModel

	BotKey types.String `tfsdk:"bot_key"`
}

// Metadata returns the metadata for the resource.
func (*NotificationWeComResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_wecom"
}

// Schema returns the schema for the resource.
func (*NotificationWeComResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "WeCom notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"bot_key": schema.StringAttribute{
				MarkdownDescription: "WeCom bot key",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the WeCom notification resource with the API client.
func (r *NotificationWeComResource) Configure(
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

// Create creates a new WeCom notification resource.
func (r *NotificationWeComResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationWeComResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	wecom := notification.WeCom{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WeComDetails: notification.WeComDetails{
			BotKey: data.BotKey.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, wecom)
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

// Read reads the current state of the WeCom notification resource.
func (r *NotificationWeComResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationWeComResourceModel

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

	wecom := notification.WeCom{}
	err = base.As(&wecom)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "WeCom"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(wecom.Name)
	data.IsActive = types.BoolValue(wecom.IsActive)
	data.IsDefault = types.BoolValue(wecom.IsDefault)
	data.ApplyExisting = types.BoolValue(wecom.ApplyExisting)

	data.BotKey = types.StringValue(wecom.BotKey)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the WeCom notification resource.
func (r *NotificationWeComResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationWeComResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	wecom := notification.WeCom{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WeComDetails: notification.WeComDetails{
			BotKey: data.BotKey.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, wecom)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the WeCom notification resource.
func (r *NotificationWeComResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationWeComResourceModel

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
func (*NotificationWeComResource) ImportState(
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
