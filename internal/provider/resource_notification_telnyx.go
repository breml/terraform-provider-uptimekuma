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

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationTelnyxResource{}
	_ resource.ResourceWithImportState = &NotificationTelnyxResource{}
)

// NewNotificationTelnyxResource returns a new instance of the Telnyx notification resource.
func NewNotificationTelnyxResource() resource.Resource {
	return &NotificationTelnyxResource{}
}

// NotificationTelnyxResource defines the resource implementation.
type NotificationTelnyxResource struct {
	client *kuma.Client
}

// NotificationTelnyxResourceModel describes the resource data model.
type NotificationTelnyxResourceModel struct {
	NotificationBaseModel

	APIKey             types.String `tfsdk:"api_key"`
	MessagingProfileID types.String `tfsdk:"messaging_profile_id"`
	PhoneNumber        types.String `tfsdk:"phone_number"`
	ToNumber           types.String `tfsdk:"to_number"`
}

// Metadata returns the metadata for the resource.
func (*NotificationTelnyxResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_telnyx"
}

// Schema returns the schema for the resource.
func (*NotificationTelnyxResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Telnyx notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The Telnyx API key used to authenticate requests.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"messaging_profile_id": schema.StringAttribute{
				MarkdownDescription: "The Telnyx messaging profile ID used to send messages.",
				Optional:            true,
			},
			"phone_number": schema.StringAttribute{
				MarkdownDescription: "The sender phone number in E.164 format.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to_number": schema.StringAttribute{
				MarkdownDescription: "The recipient phone number in E.164 format.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Telnyx notification resource with the API client.
func (r *NotificationTelnyxResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Telnyx notification resource.
func (r *NotificationTelnyxResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationTelnyxResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	telnyx := telnyxFromModel(&data)

	id, err := r.client.CreateNotification(ctx, telnyx)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Telnyx notification resource.
func (r *NotificationTelnyxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationTelnyxResourceModel

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

	telnyx := notification.Telnyx{}
	err = base.As(&telnyx)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "telnyx"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(telnyx.Name)
	data.IsActive = types.BoolValue(telnyx.IsActive)
	data.IsDefault = types.BoolValue(telnyx.IsDefault)
	data.ApplyExisting = types.BoolValue(telnyx.ApplyExisting)

	data.APIKey = types.StringValue(telnyx.APIKey)
	data.MessagingProfileID = ptrToTypes(telnyx.MessagingProfileID)
	data.PhoneNumber = types.StringValue(telnyx.PhoneNumber)
	data.ToNumber = types.StringValue(telnyx.ToNumber)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Telnyx notification resource.
func (r *NotificationTelnyxResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationTelnyxResourceModel

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

	telnyx := telnyxFromModel(&data)
	telnyx.ID = id

	err := r.client.UpdateNotification(ctx, telnyx)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Telnyx notification resource.
func (r *NotificationTelnyxResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationTelnyxResourceModel

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
func (*NotificationTelnyxResource) ImportState(
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

// telnyxFromModel builds a Telnyx notification from the resource model.
func telnyxFromModel(data *NotificationTelnyxResourceModel) notification.Telnyx {
	return notification.Telnyx{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		TelnyxDetails: notification.TelnyxDetails{
			APIKey:             data.APIKey.ValueString(),
			MessagingProfileID: strToPtr(data.MessagingProfileID),
			PhoneNumber:        data.PhoneNumber.ValueString(),
			ToNumber:           data.ToNumber.ValueString(),
		},
	}
}
