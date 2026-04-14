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
	_ resource.Resource                = &NotificationSMSPlanetResource{}
	_ resource.ResourceWithImportState = &NotificationSMSPlanetResource{}
)

// NewNotificationSMSPlanetResource returns a new instance of the SMS Planet notification resource.
func NewNotificationSMSPlanetResource() resource.Resource {
	return &NotificationSMSPlanetResource{}
}

// NotificationSMSPlanetResource defines the resource implementation.
type NotificationSMSPlanetResource struct {
	client *kuma.Client
}

// NotificationSMSPlanetResourceModel describes the resource data model.
type NotificationSMSPlanetResourceModel struct {
	NotificationBaseModel

	APIToken     types.String `tfsdk:"api_token"`
	PhoneNumbers types.String `tfsdk:"phone_numbers"`
	SenderName   types.String `tfsdk:"sender_name"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSMSPlanetResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_smsplanet"
}

// Schema returns the schema for the resource.
func (*NotificationSMSPlanetResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SMS Planet notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "SMS Planet API token for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"phone_numbers": schema.StringAttribute{
				MarkdownDescription: "Recipient phone numbers",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"sender_name": schema.StringAttribute{
				MarkdownDescription: "Sender name or identifier",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the SMS Planet notification resource with the API client.
func (r *NotificationSMSPlanetResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new SMS Planet notification resource.
func (r *NotificationSMSPlanetResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSMSPlanetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smsplanet := notification.SMSPlanet{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMSPlanetDetails: notification.SMSPlanetDetails{
			APIToken:     data.APIToken.ValueString(),
			PhoneNumbers: data.PhoneNumbers.ValueString(),
			SenderName:   data.SenderName.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, smsplanet)
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

// Read reads the current state of the SMS Planet notification resource.
func (r *NotificationSMSPlanetResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSMSPlanetResourceModel

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

	smsplanet := notification.SMSPlanet{}
	err = base.As(&smsplanet)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "SMSPlanet"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(smsplanet.Name)
	data.IsActive = types.BoolValue(smsplanet.IsActive)
	data.IsDefault = types.BoolValue(smsplanet.IsDefault)
	data.ApplyExisting = types.BoolValue(smsplanet.ApplyExisting)

	data.APIToken = types.StringValue(smsplanet.APIToken)
	data.PhoneNumbers = types.StringValue(smsplanet.PhoneNumbers)

	if smsplanet.SenderName != "" {
		data.SenderName = types.StringValue(smsplanet.SenderName)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SMS Planet notification resource.
func (r *NotificationSMSPlanetResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSMSPlanetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smsplanet := notification.SMSPlanet{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMSPlanetDetails: notification.SMSPlanetDetails{
			APIToken:     data.APIToken.ValueString(),
			PhoneNumbers: data.PhoneNumbers.ValueString(),
			SenderName:   data.SenderName.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, smsplanet)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SMS Planet notification resource.
func (r *NotificationSMSPlanetResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSMSPlanetResourceModel

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
func (*NotificationSMSPlanetResource) ImportState(
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
