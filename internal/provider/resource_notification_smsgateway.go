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
	_ resource.Resource                = &NotificationSMSGatewayResource{}
	_ resource.ResourceWithImportState = &NotificationSMSGatewayResource{}
)

// NewNotificationSMSGatewayResource returns a new instance of the SMS Gateway notification resource.
func NewNotificationSMSGatewayResource() resource.Resource {
	return &NotificationSMSGatewayResource{}
}

// NotificationSMSGatewayResource defines the resource implementation.
type NotificationSMSGatewayResource struct {
	client *kuma.Client
}

// NotificationSMSGatewayResourceModel describes the resource data model.
type NotificationSMSGatewayResourceModel struct {
	NotificationBaseModel

	URL        types.String `tfsdk:"url"`
	APIKey     types.String `tfsdk:"api_key"`
	Recipients types.String `tfsdk:"recipients"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSMSGatewayResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_smsgateway"
}

// Schema returns the schema for the resource.
func (*NotificationSMSGatewayResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SMS Gateway notification resource. SMS Gateway " +
			"([mattboston/sms-gateway](https://github.com/mattboston/sms-gateway)) is a self-hosted server that " +
			"sends text messages through an attached GSM modem and is reached over HTTP.",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The base URL of the SMS Gateway server without an API path, for example " +
					"`http://localhost:8080`. Uptime Kuma strips trailing slashes and appends `/api/v1/sms/send` " +
					"itself. The value is stored verbatim, the provider does not normalize it.",
				Required: true,
				Validators: []validator.String{
					validateURL(),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key of the SMS Gateway server, sent as the `X-API-Key` header.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"recipients": schema.StringAttribute{
				MarkdownDescription: "One or more recipient phone numbers in international format, separated by " +
					"commas, for example `+15551234567, +15559876543`. Uptime Kuma trims each number and discards " +
					"the empty entries; a value that leaves no number would send nothing while still reporting " +
					"success, so such a value, blank or only commas, is rejected. The value is stored verbatim, " +
					"the provider does not normalize it.",
				Required: true,
				Validators: []validator.String{
					nonEmptyCommaList(),
				},
			},
		}),
	}
}

// Configure configures the SMS Gateway notification resource with the API client.
func (r *NotificationSMSGatewayResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new SMS Gateway notification resource.
func (r *NotificationSMSGatewayResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSMSGatewayResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smsGateway := smsGatewayFromModel(&data)

	id, err := r.client.CreateNotification(ctx, smsGateway)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the SMS Gateway notification resource.
func (r *NotificationSMSGatewayResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSMSGatewayResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.SMSGatewayDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	smsGateway := notification.SMSGateway{}
	err := base.As(&smsGateway)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.SMSGatewayDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(smsGateway.Name)
	data.IsActive = types.BoolValue(smsGateway.IsActive)
	data.IsDefault = types.BoolValue(smsGateway.IsDefault)
	data.ApplyExisting = types.BoolValue(smsGateway.ApplyExisting)

	data.URL = types.StringValue(smsGateway.URL)
	data.APIKey = types.StringValue(smsGateway.APIKey)
	data.Recipients = types.StringValue(smsGateway.Recipients)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SMS Gateway notification resource.
func (r *NotificationSMSGatewayResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSMSGatewayResourceModel

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

	smsGateway := smsGatewayFromModel(&data)
	smsGateway.ID = id

	err := r.client.UpdateNotification(ctx, smsGateway)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SMS Gateway notification resource.
func (r *NotificationSMSGatewayResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSMSGatewayResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationSMSGatewayResource) ImportState(
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

// smsGatewayFromModel builds an SMS Gateway notification from the resource model.
func smsGatewayFromModel(data *NotificationSMSGatewayResourceModel) notification.SMSGateway {
	return notification.SMSGateway{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMSGatewayDetails: notification.SMSGatewayDetails{
			URL:        data.URL.ValueString(),
			APIKey:     data.APIKey.ValueString(),
			Recipients: data.Recipients.ValueString(),
		},
	}
}
