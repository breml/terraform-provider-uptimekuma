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
	_ resource.Resource                = &NotificationOpenWaResource{}
	_ resource.ResourceWithImportState = &NotificationOpenWaResource{}
)

// NewNotificationOpenWaResource returns a new instance of the OpenWA notification resource.
func NewNotificationOpenWaResource() resource.Resource {
	return &NotificationOpenWaResource{}
}

// NotificationOpenWaResource defines the resource implementation.
type NotificationOpenWaResource struct {
	client *kuma.Client
}

// NotificationOpenWaResourceModel describes the resource data model.
type NotificationOpenWaResourceModel struct {
	NotificationBaseModel

	APIURL           types.String `tfsdk:"api_url"`
	APIKey           types.String `tfsdk:"api_key"`
	APIKeyWO         types.String `tfsdk:"api_key_wo"`
	APIKeyWOVersion  types.String `tfsdk:"api_key_wo_version"`
	Session          types.String `tfsdk:"session"`
	ChatID           types.String `tfsdk:"chat_id"`
	UseCustomMessage types.Bool   `tfsdk:"use_custom_message"`
	CustomMessage    types.String `tfsdk:"custom_message"`
}

// Metadata returns the metadata for the resource.
func (*NotificationOpenWaResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_openwa"
}

// Schema returns the schema for the resource.
func (*NotificationOpenWaResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OpenWA (self-hosted WhatsApp gateway) notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Base URL of the OpenWA gateway (e.g., `http://localhost:2785/`). " +
					"Uptime Kuma strips trailing slashes when sending. The value is stored verbatim.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key sent as `X-Api-Key` header to the OpenWA gateway.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"session": schema.StringAttribute{
				MarkdownDescription: "Name of the OpenWA session. Uptime Kuma URL-encodes it into the request path.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"chat_id": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of WhatsApp chat IDs " +
					"(e.g., `00117612345678@c.us,123456789012345678@g.us,1234567890@lid`). " +
					"Uptime Kuma trims each entry and drops empty ones. Sending fails if no chat ID remains.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"use_custom_message": schema.BoolAttribute{
				MarkdownDescription: "If `true`, render `custom_message` instead of sending the default message " +
					"(see `custom_message`). If `false` or unset, Uptime Kuma sends the default message.",
				Optional: true,
			},
			"custom_message": schema.StringAttribute{
				MarkdownDescription: "Custom message template. Uptime Kuma only renders it if " +
					"`use_custom_message` is `true` and the trimmed template is not empty, " +
					"otherwise it sends the default message.",
				Optional: true,
			},
		}),
	}

	resp.Schema.Attributes = withWriteOnlySecret(
		resp.Schema.Attributes,
		"api_key",
		"API key sent as `X-Api-Key` header to the OpenWA gateway",
	)
}

// Configure configures the OpenWA notification resource with the API client.
func (r *NotificationOpenWaResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new OpenWA notification resource.
func (r *NotificationOpenWaResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationOpenWaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	openWa := openWaFromModel(&data)
	openWa.APIKey = writeOnlyOrPlan(ctx, req.Config, "api_key", data.APIKey, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateNotification(ctx, openWa)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the OpenWA notification resource.
func (r *NotificationOpenWaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationOpenWaResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.OpenWaDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	openWa := notification.OpenWa{}
	err := base.As(&openWa)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.OpenWaDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(openWa.Name)
	data.IsActive = types.BoolValue(openWa.IsActive)
	data.IsDefault = types.BoolValue(openWa.IsDefault)
	data.ApplyExisting = types.BoolValue(openWa.ApplyExisting)

	data.APIURL = types.StringValue(openWa.APIURL)

	if data.APIKeyWOVersion.IsNull() {
		data.APIKey = types.StringValue(openWa.APIKey)
	}

	data.Session = types.StringValue(openWa.Session)
	data.ChatID = types.StringValue(openWa.ChatID)
	data.UseCustomMessage = types.BoolPointerValue(openWa.UseCustomMessage)
	data.CustomMessage = types.StringPointerValue(openWa.CustomMessage)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the OpenWA notification resource.
func (r *NotificationOpenWaResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationOpenWaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	openWa := openWaFromModel(&data)
	openWa.ID = data.ID.ValueInt64()
	openWa.APIKey = writeOnlyOrPlan(ctx, req.Config, "api_key", data.APIKey, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNotification(ctx, openWa)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the OpenWA notification resource.
func (r *NotificationOpenWaResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationOpenWaResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationOpenWaResource) ImportState(
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

// openWaFromModel builds an OpenWA notification from the resource model. The caller resolves
// api_key: a write-only value is only available in the config, not in the plan this model is
// read from.
//
// Both optional fields are omitempty pointers on the wire: null maps to nil, which keeps them out of
// the request entirely, whereas an explicit false or empty string is sent as is.
func openWaFromModel(data *NotificationOpenWaResourceModel) notification.OpenWa {
	return notification.OpenWa{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OpenWaDetails: notification.OpenWaDetails{
			APIURL:           data.APIURL.ValueString(),
			Session:          data.Session.ValueString(),
			ChatID:           data.ChatID.ValueString(),
			UseCustomMessage: data.UseCustomMessage.ValueBoolPointer(),
			CustomMessage:    data.CustomMessage.ValueStringPointer(),
		},
	}
}
