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
	_ resource.Resource                = &NotificationWxPusherResource{}
	_ resource.ResourceWithImportState = &NotificationWxPusherResource{}
)

// NewNotificationWxPusherResource returns a new instance of the WxPusher notification resource.
func NewNotificationWxPusherResource() resource.Resource {
	return &NotificationWxPusherResource{}
}

// NotificationWxPusherResource defines the resource implementation.
type NotificationWxPusherResource struct {
	client *kuma.Client
}

// NotificationWxPusherResourceModel describes the resource data model.
type NotificationWxPusherResourceModel struct {
	NotificationBaseModel

	SPT types.String `tfsdk:"spt"`
}

// Metadata returns the metadata for the resource.
func (*NotificationWxPusherResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_wxpusher"
}

// Schema returns the schema for the resource.
func (*NotificationWxPusherResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "WxPusher notification resource. WxPusher delivers the notifications through the " +
			"standalone WxPusher app via the simple push (极简推送) endpoint.",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"spt": schema.StringAttribute{
				MarkdownDescription: "The WxPusher simple push token, for example `SPT_xxxxxxxxxxxx`. Multiple " +
					"tokens are separated by commas: Uptime Kuma trims each one, discards the empty entries and " +
					"delivers to them in batches of at most 10 tokens per request. The value is stored " +
					"verbatim, the provider does not normalize it.",
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the WxPusher notification resource with the API client.
func (r *NotificationWxPusherResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new WxPusher notification resource.
func (r *NotificationWxPusherResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationWxPusherResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	wxPusher := wxPusherFromModel(&data)

	id, err := r.client.CreateNotification(ctx, wxPusher)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the WxPusher notification resource.
func (r *NotificationWxPusherResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationWxPusherResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.WxPusherDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)

		return
	}

	wxPusher := notification.WxPusher{}
	err := base.As(&wxPusher)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.WxPusherDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(wxPusher.Name)
	data.IsActive = types.BoolValue(wxPusher.IsActive)
	data.IsDefault = types.BoolValue(wxPusher.IsDefault)
	data.ApplyExisting = types.BoolValue(wxPusher.ApplyExisting)

	data.SPT = types.StringValue(wxPusher.SPT)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the WxPusher notification resource.
func (r *NotificationWxPusherResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationWxPusherResourceModel

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

	wxPusher := wxPusherFromModel(&data)
	wxPusher.ID = id

	err := r.client.UpdateNotification(ctx, wxPusher)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the WxPusher notification resource.
func (r *NotificationWxPusherResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationWxPusherResourceModel

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
func (*NotificationWxPusherResource) ImportState(
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

// wxPusherFromModel builds a WxPusher notification from the resource model.
func wxPusherFromModel(data *NotificationWxPusherResourceModel) notification.WxPusher {
	return notification.WxPusher{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WxPusherDetails: notification.WxPusherDetails{
			SPT: data.SPT.ValueString(),
		},
	}
}
