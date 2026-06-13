package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationWhatsapp360messengerResource{}
	_ resource.ResourceWithImportState = &NotificationWhatsapp360messengerResource{}
)

// NewNotificationWhatsapp360messengerResource returns a new instance of the WhatsApp 360messenger
// notification resource.
func NewNotificationWhatsapp360messengerResource() resource.Resource {
	return &NotificationWhatsapp360messengerResource{}
}

// NotificationWhatsapp360messengerResource defines the resource implementation.
type NotificationWhatsapp360messengerResource struct {
	client *kuma.Client
}

// NotificationWhatsapp360messengerResourceModel describes the resource data model.
type NotificationWhatsapp360messengerResourceModel struct {
	NotificationBaseModel

	AuthToken   types.String `tfsdk:"auth_token"`
	Recipient   types.String `tfsdk:"recipient"`
	GroupIDs    types.List   `tfsdk:"group_ids"`
	GroupID     types.String `tfsdk:"group_id"`
	UseTemplate types.Bool   `tfsdk:"use_template"`
	Template    types.String `tfsdk:"template"`
}

// Metadata returns the metadata for the resource.
func (*NotificationWhatsapp360messengerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_whatsapp360messenger"
}

// Schema returns the schema for the resource.
func (*NotificationWhatsapp360messengerResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "WhatsApp 360messenger notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "The Bearer authentication token for the 360messenger API.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"recipient": schema.StringAttribute{
				MarkdownDescription: "A comma- or semicolon-separated list of phone numbers to send WhatsApp " +
					"notifications to.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"group_ids": schema.ListAttribute{
				MarkdownDescription: "A list of WhatsApp group IDs to send notifications to. " +
					"Conflicts with `group_id`.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRoot("group_id")),
				},
			},
			"group_id": schema.StringAttribute{
				MarkdownDescription: "Legacy single WhatsApp group ID, kept for backwards compatibility. " +
					"Conflicts with `group_ids`.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("group_ids")),
				},
			},
			"use_template": schema.BoolAttribute{
				MarkdownDescription: "If true, use a custom message template for WhatsApp notifications.",
				Optional:            true,
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "The custom message template for WhatsApp notifications.",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the WhatsApp 360messenger notification resource with the API client.
func (r *NotificationWhatsapp360messengerResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new WhatsApp 360messenger notification resource.
func (r *NotificationWhatsapp360messengerResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationWhatsapp360messengerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	whatsapp360messenger := whatsapp360messengerFromModel(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateNotification(ctx, whatsapp360messenger)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the WhatsApp 360messenger notification resource.
func (r *NotificationWhatsapp360messengerResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationWhatsapp360messengerResourceModel

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

	whatsapp360messenger := notification.Whatsapp360messenger{}

	err = base.As(&whatsapp360messenger)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "Whatsapp360messenger"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(whatsapp360messenger.Name)
	data.IsActive = types.BoolValue(whatsapp360messenger.IsActive)
	data.IsDefault = types.BoolValue(whatsapp360messenger.IsDefault)
	data.ApplyExisting = types.BoolValue(whatsapp360messenger.ApplyExisting)

	data.AuthToken = types.StringValue(whatsapp360messenger.AuthToken)
	data.Recipient = types.StringValue(whatsapp360messenger.Recipient)
	data.GroupID = ptrToTypes(whatsapp360messenger.GroupID)
	data.UseTemplate = boolPtrToTypes(whatsapp360messenger.UseTemplate)
	data.Template = ptrToTypes(whatsapp360messenger.Template)

	if len(whatsapp360messenger.GroupIDs) > 0 {
		groupIDs, diags := types.ListValueFrom(ctx, types.StringType, whatsapp360messenger.GroupIDs)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		data.GroupIDs = groupIDs
	} else {
		data.GroupIDs = types.ListNull(types.StringType)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the WhatsApp 360messenger notification resource.
func (r *NotificationWhatsapp360messengerResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationWhatsapp360messengerResourceModel

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

	whatsapp360messenger := whatsapp360messengerFromModel(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	whatsapp360messenger.ID = id

	err := r.client.UpdateNotification(ctx, whatsapp360messenger)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the WhatsApp 360messenger notification resource.
func (r *NotificationWhatsapp360messengerResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationWhatsapp360messengerResourceModel

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
func (*NotificationWhatsapp360messengerResource) ImportState(
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

// whatsapp360messengerFromModel builds a Whatsapp360messenger notification from the resource model.
func whatsapp360messengerFromModel(
	ctx context.Context,
	data *NotificationWhatsapp360messengerResourceModel,
	diags *diag.Diagnostics,
) notification.Whatsapp360messenger {
	whatsapp360messenger := notification.Whatsapp360messenger{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		Whatsapp360messengerDetails: notification.Whatsapp360messengerDetails{
			AuthToken:   data.AuthToken.ValueString(),
			Recipient:   data.Recipient.ValueString(),
			GroupID:     strToPtr(data.GroupID),
			UseTemplate: boolToPtr(data.UseTemplate),
			Template:    strToPtr(data.Template),
		},
	}

	if !data.GroupIDs.IsNull() && !data.GroupIDs.IsUnknown() {
		var groupIDs []string

		diags.Append(data.GroupIDs.ElementsAs(ctx, &groupIDs, false)...)

		if !diags.HasError() {
			whatsapp360messenger.GroupIDs = groupIDs
		}
	}

	return whatsapp360messenger
}
