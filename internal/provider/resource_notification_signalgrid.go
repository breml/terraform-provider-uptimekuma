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
	_ resource.Resource                = &NotificationSignalgridResource{}
	_ resource.ResourceWithImportState = &NotificationSignalgridResource{}
)

// NewNotificationSignalgridResource returns a new instance of the Signalgrid notification resource.
func NewNotificationSignalgridResource() resource.Resource {
	return &NotificationSignalgridResource{}
}

// NotificationSignalgridResource defines the resource implementation.
type NotificationSignalgridResource struct {
	client *kuma.Client
}

// NotificationSignalgridResourceModel describes the resource data model.
type NotificationSignalgridResourceModel struct {
	NotificationBaseModel

	ClientKey          types.String `tfsdk:"client_key"`
	ClientKeyWO        types.String `tfsdk:"client_key_wo"`
	ClientKeyWOVersion types.String `tfsdk:"client_key_wo_version"`
	Channel            types.String `tfsdk:"channel"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSignalgridResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_signalgrid"
}

// Schema returns the schema for the resource.
func (*NotificationSignalgridResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Signalgrid notification resource. Signalgrid is a push notification service. " +
			"Requires Uptime Kuma 2.5.4 or later.",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"client_key": schema.StringAttribute{
				MarkdownDescription: "The Signalgrid client key.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"channel": schema.StringAttribute{
				MarkdownDescription: "The Signalgrid channel the push is sent to.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}

	resp.Schema.Attributes = withWriteOnlySecret(resp.Schema.Attributes, "client_key", "Signalgrid client key")
}

// Configure configures the Signalgrid notification resource with the API client.
func (r *NotificationSignalgridResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Signalgrid notification resource.
func (r *NotificationSignalgridResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSignalgridResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	signalgrid := signalgridFromModel(&data)
	signalgrid.ClientKey = writeOnlyOrPlan(ctx, req.Config, "client_key", data.ClientKey, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateNotification(ctx, signalgrid)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Signalgrid notification resource.
func (r *NotificationSignalgridResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSignalgridResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.SignalgridDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	signalgrid := notification.Signalgrid{}
	err := base.As(&signalgrid)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.SignalgridDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(signalgrid.Name)
	data.IsDefault = types.BoolValue(signalgrid.IsDefault)
	data.ApplyExisting = types.BoolValue(signalgrid.ApplyExisting)
	data.Channel = types.StringValue(signalgrid.Channel)

	if data.ClientKeyWOVersion.IsNull() {
		data.ClientKey = types.StringValue(signalgrid.ClientKey)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Signalgrid notification resource.
func (r *NotificationSignalgridResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSignalgridResourceModel

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

	signalgrid := signalgridFromModel(&data)
	signalgrid.ClientKey = writeOnlyOrPlan(ctx, req.Config, "client_key", data.ClientKey, &resp.Diagnostics)
	signalgrid.ID = id

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNotification(ctx, signalgrid)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Signalgrid notification resource.
func (r *NotificationSignalgridResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSignalgridResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationSignalgridResource) ImportState(
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

// signalgridFromModel builds a Signalgrid notification from the resource model.
func signalgridFromModel(data *NotificationSignalgridResourceModel) notification.Signalgrid {
	return notification.Signalgrid{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      true,
			Name:          data.Name.ValueString(),
		},
		SignalgridDetails: notification.SignalgridDetails{
			ClientKey: data.ClientKey.ValueString(),
			Channel:   data.Channel.ValueString(),
		},
	}
}
