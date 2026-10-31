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
	_ resource.Resource                = &NotificationBearSMSResource{}
	_ resource.ResourceWithImportState = &NotificationBearSMSResource{}
)

// NewNotificationBearSMSResource returns a new instance of the BearSMS notification resource.
func NewNotificationBearSMSResource() resource.Resource {
	return &NotificationBearSMSResource{}
}

// NotificationBearSMSResource defines the resource implementation.
type NotificationBearSMSResource struct {
	client *kuma.Client
}

// NotificationBearSMSResourceModel describes the resource data model.
type NotificationBearSMSResourceModel struct {
	NotificationBaseModel

	Username         types.String `tfsdk:"username"`
	HashKey          types.String `tfsdk:"hash_key"`
	HashKeyWO        types.String `tfsdk:"hash_key_wo"`
	HashKeyWOVersion types.String `tfsdk:"hash_key_wo_version"`
	SenderID         types.String `tfsdk:"sender_id"`
	PhoneNumber      types.String `tfsdk:"phone_number"`
}

// Metadata returns the metadata for the resource.
func (*NotificationBearSMSResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_bearsms"
}

// Schema returns the schema for the resource.
func (*NotificationBearSMSResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "BearSMS notification resource. The alert is sent as an SMS through the " +
			"BearSMS gateway. Requires Uptime Kuma 2.5.1 or later.",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "The BearSMS account username.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"hash_key": schema.StringAttribute{
				MarkdownDescription: "The BearSMS hash key belonging to the account.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"sender_id": schema.StringAttribute{
				MarkdownDescription: "The sender name or number shown to the recipient, at most 11 characters. " +
					"The sender has to be approved in the BearSMS account. The length limit mirrors the " +
					"Uptime Kuma web UI; Uptime Kuma's server enforces neither. If unset, Uptime Kuma sends " +
					"no sender with the request and BearSMS uses the account's default sender.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 11),
				},
			},
			"phone_number": schema.StringAttribute{
				MarkdownDescription: "The recipient phone number with country code and without a `+` or `00` " +
					"prefix, for example `9725XXXXXXXX`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}

	resp.Schema.Attributes = withWriteOnlySecret(
		resp.Schema.Attributes,
		"hash_key",
		"The BearSMS hash key belonging to the account",
	)
}

// Configure configures the BearSMS notification resource with the API client.
func (r *NotificationBearSMSResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new BearSMS notification resource.
func (r *NotificationBearSMSResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationBearSMSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bearSMS := bearSMSFromModel(&data)
	bearSMS.HashKey = writeOnlyOrPlan(ctx, req.Config, "hash_key", data.HashKey, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateNotification(ctx, bearSMS)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the BearSMS notification resource.
func (r *NotificationBearSMSResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationBearSMSResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.BearSMSDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	bearSMS := notification.BearSMS{}
	err := base.As(&bearSMS)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.BearSMSDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(bearSMS.Name)
	data.IsDefault = types.BoolValue(bearSMS.IsDefault)
	data.ApplyExisting = types.BoolValue(bearSMS.ApplyExisting)

	data.Username = types.StringValue(bearSMS.Username)

	if data.HashKeyWOVersion.IsNull() {
		data.HashKey = types.StringValue(bearSMS.HashKey)
	}

	// A sender cleared in the Uptime Kuma UI is stored as "" and decodes to a non-nil pointer to "".
	// stringOrNullPtr maps that to null, so a null configuration does not show a perpetual diff.
	data.SenderID = stringOrNullPtr(bearSMS.SenderID)
	data.PhoneNumber = types.StringValue(bearSMS.PhoneNumber)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the BearSMS notification resource.
func (r *NotificationBearSMSResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationBearSMSResourceModel

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

	bearSMS := bearSMSFromModel(&data)
	bearSMS.ID = id
	bearSMS.HashKey = writeOnlyOrPlan(ctx, req.Config, "hash_key", data.HashKey, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNotification(ctx, bearSMS)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the BearSMS notification resource.
func (r *NotificationBearSMSResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationBearSMSResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationBearSMSResource) ImportState(
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

// bearSMSFromModel builds a BearSMS notification from the resource model. The caller resolves
// hash_key, because its value may come from the write-only hash_key_wo, which is only in the
// config, not in the plan.
//
// bearsmsSenderId is omitempty on the wire, so a null sender maps to nil, which keeps it out of the
// request entirely.
func bearSMSFromModel(data *NotificationBearSMSResourceModel) notification.BearSMS {
	return notification.BearSMS{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      true,
			Name:          data.Name.ValueString(),
		},
		BearSMSDetails: notification.BearSMSDetails{
			Username:    data.Username.ValueString(),
			SenderID:    strToPtr(data.SenderID),
			PhoneNumber: data.PhoneNumber.ValueString(),
		},
	}
}
