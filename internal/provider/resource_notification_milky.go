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
	_ resource.Resource                = &NotificationMilkyResource{}
	_ resource.ResourceWithImportState = &NotificationMilkyResource{}
)

// NewNotificationMilkyResource returns a new instance of the Milky notification resource.
func NewNotificationMilkyResource() resource.Resource {
	return &NotificationMilkyResource{}
}

// NotificationMilkyResource defines the resource implementation.
type NotificationMilkyResource struct {
	client *kuma.Client
}

// NotificationMilkyResourceModel describes the resource data model.
type NotificationMilkyResourceModel struct {
	NotificationBaseModel

	HTTPAddr             types.String `tfsdk:"http_addr"`
	AccessToken          types.String `tfsdk:"access_token"`
	AccessTokenWO        types.String `tfsdk:"access_token_wo"`
	AccessTokenWOVersion types.String `tfsdk:"access_token_wo_version"`
	MsgType              types.String `tfsdk:"msg_type"`
	ReceiverID           types.String `tfsdk:"receiver_id"`
}

// Metadata returns the metadata for the resource.
func (*NotificationMilkyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_milky"
}

// Schema returns the schema for the resource.
func (*NotificationMilkyResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Milky (QQ bot protocol) notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"http_addr": schema.StringAttribute{
				MarkdownDescription: "Address of the Milky service (e.g., `http://localhost:3000`). " +
					"Uptime Kuma prepends `http://` if the value does not start with `http`. " +
					"Do not end the address with a slash: Uptime Kuma then skips the `/api/` segment and " +
					"posts the alert to an endpoint that does not exist. The value is stored verbatim.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Access token sent as bearer token to the Milky service.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"msg_type": schema.StringAttribute{
				MarkdownDescription: "Message type: `group` for a QQ group or `private` for a private chat. " +
					"If unset, Uptime Kuma sends a private message.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						notification.MilkyMessageTypeGroup.String(),
						notification.MilkyMessageTypePrivate.String(),
					),
				},
			},
			"receiver_id": schema.StringAttribute{
				MarkdownDescription: "QQ group ID (when `msg_type` is `group`) or user ID (otherwise).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}

	resp.Schema.Attributes = withWriteOnlySecret(
		resp.Schema.Attributes,
		"access_token",
		"Access token sent as bearer token to the Milky service",
	)
}

// Configure configures the Milky notification resource with the API client.
func (r *NotificationMilkyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Milky notification resource.
func (r *NotificationMilkyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationMilkyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	milky := milkyFromModel(&data)
	milky.AccessToken = writeOnlyOrPlan(ctx, req.Config, "access_token", data.AccessToken, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateNotification(ctx, milky)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Milky notification resource.
func (r *NotificationMilkyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationMilkyResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	// OneBot shares the wire field names, so only the type tells the two apart.
	base, found := readNotification(ctx, r.client, id, notification.MilkyDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	milky := notification.Milky{}
	err := base.As(&milky)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.MilkyDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(milky.Name)
	data.IsDefault = types.BoolValue(milky.IsDefault)
	data.ApplyExisting = types.BoolValue(milky.ApplyExisting)

	data.HTTPAddr = types.StringValue(milky.HTTPAddr)

	if data.AccessTokenWOVersion.IsNull() {
		data.AccessToken = types.StringValue(milky.AccessToken)
	}

	// msgType is omitempty on the wire, so an unset message type comes back as the empty string.
	// Mapping it to an empty string instead of null would produce a perpetual diff.
	if milky.MsgType == "" {
		data.MsgType = types.StringNull()
	} else {
		data.MsgType = types.StringValue(milky.MsgType.String())
	}

	data.ReceiverID = types.StringValue(milky.ReceiverID)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Milky notification resource.
func (r *NotificationMilkyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationMilkyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	milky := milkyFromModel(&data)
	milky.ID = data.ID.ValueInt64()
	milky.AccessToken = writeOnlyOrPlan(ctx, req.Config, "access_token", data.AccessToken, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNotification(ctx, milky)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Milky notification resource.
func (r *NotificationMilkyResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationMilkyResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationMilkyResource) ImportState(
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

// milkyFromModel builds a Milky notification from the resource model. The caller resolves
// access_token, which may be write-only and so is not in the model.
//
// msgType is omitempty on the wire, so a null msg_type maps to the zero value, which keeps it out
// of the request entirely.
func milkyFromModel(data *NotificationMilkyResourceModel) notification.Milky {
	return notification.Milky{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      true,
			Name:          data.Name.ValueString(),
		},
		MilkyDetails: notification.MilkyDetails{
			HTTPAddr:   data.HTTPAddr.ValueString(),
			MsgType:    notification.MilkyMessageType(data.MsgType.ValueString()),
			ReceiverID: data.ReceiverID.ValueString(),
		},
	}
}
