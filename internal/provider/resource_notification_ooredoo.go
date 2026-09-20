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
	_ resource.Resource                = &NotificationOoredooResource{}
	_ resource.ResourceWithImportState = &NotificationOoredooResource{}
)

// NewNotificationOoredooResource returns a new instance of the Ooredoo notification resource.
func NewNotificationOoredooResource() resource.Resource {
	return &NotificationOoredooResource{}
}

// NotificationOoredooResource defines the resource implementation.
type NotificationOoredooResource struct {
	client *kuma.Client
}

// NotificationOoredooResourceModel describes the resource data model.
type NotificationOoredooResourceModel struct {
	NotificationBaseModel

	Username    types.String `tfsdk:"username"`
	AccessKey   types.String `tfsdk:"access_key"`
	BearerToken types.String `tfsdk:"bearer_token"`
	ToNumber    types.String `tfsdk:"to_number"`
	ServerURL   types.String `tfsdk:"server_url"`
}

// Metadata returns the metadata for the resource.
func (*NotificationOoredooResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_ooredoo"
}

// Schema returns the schema for the resource.
func (*NotificationOoredooResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Ooredoo Maldives bulk SMS notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "The Ooredoo bulk SMS account username.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"access_key": schema.StringAttribute{
				MarkdownDescription: "The Ooredoo account access key. Uptime Kuma Base64 encodes it before " +
					"it is passed on to the gateway.",
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"bearer_token": schema.StringAttribute{
				MarkdownDescription: "The token authenticating the request against the gateway. It is sent " +
					"as an `Authorization: Bearer` header.",
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to_number": schema.StringAttribute{
				MarkdownDescription: "One or more recipient phone numbers, separated by comma, semicolon or " +
					"whitespace. Since whitespace separates recipients, an individual number must not contain " +
					"spaces. Uptime Kuma strips every `+` character and prefixes bare 7 digit numbers with the " +
					"Maldives country code `960`. The numbers are not validated when the notification is " +
					"saved: Uptime Kuma only rejects it at send time if no recipient remains at all.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"server_url": schema.StringAttribute{
				MarkdownDescription: "The Ooredoo API endpoint. If unset, Uptime Kuma falls back to " +
					"`https://o-papi1-lb01.ooredoo.mv/bulk_sms/v2`.",
				Optional: true,
				Validators: []validator.String{
					validateURL(),
				},
			},
		}),
	}
}

// Configure configures the Ooredoo notification resource with the API client.
func (r *NotificationOoredooResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Ooredoo notification resource.
func (r *NotificationOoredooResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationOoredooResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ooredoo := ooredooFromModel(&data)

	id, err := r.client.CreateNotification(ctx, ooredoo)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Ooredoo notification resource.
func (r *NotificationOoredooResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationOoredooResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.OoredooDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	ooredoo := notification.Ooredoo{}
	err := base.As(&ooredoo)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.OoredooDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(ooredoo.Name)
	data.IsActive = types.BoolValue(ooredoo.IsActive)
	data.IsDefault = types.BoolValue(ooredoo.IsDefault)
	data.ApplyExisting = types.BoolValue(ooredoo.ApplyExisting)

	data.Username = types.StringValue(ooredoo.Username)
	data.AccessKey = types.StringValue(ooredoo.AccessKey)
	data.BearerToken = types.StringValue(ooredoo.BearerToken)
	data.ToNumber = types.StringValue(ooredoo.ToNumber)

	// ooredooServerUrl is a *string, so omitempty drops it only when it is nil, not when it points
	// at the empty string. A notification whose server URL was cleared in the Uptime Kuma UI comes
	// back as "", and keeping that in state against a null configuration would be a perpetual diff.
	if ooredoo.ServerURL == nil || *ooredoo.ServerURL == "" {
		data.ServerURL = types.StringNull()
	} else {
		data.ServerURL = types.StringValue(*ooredoo.ServerURL)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Ooredoo notification resource.
func (r *NotificationOoredooResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationOoredooResourceModel

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

	ooredoo := ooredooFromModel(&data)
	ooredoo.ID = id

	err := r.client.UpdateNotification(ctx, ooredoo)
	// Handle error.
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Ooredoo notification resource.
func (r *NotificationOoredooResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationOoredooResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationOoredooResource) ImportState(
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

// ooredooFromModel builds an Ooredoo notification from the resource model.
//
// ooredooServerUrl is omitempty on the wire, so a null server_url maps to nil, which keeps it out
// of the request entirely and lets Uptime Kuma apply its own default endpoint.
func ooredooFromModel(data *NotificationOoredooResourceModel) notification.Ooredoo {
	return notification.Ooredoo{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OoredooDetails: notification.OoredooDetails{
			Username:    data.Username.ValueString(),
			AccessKey:   data.AccessKey.ValueString(),
			BearerToken: data.BearerToken.ValueString(),
			ToNumber:    data.ToNumber.ValueString(),
			ServerURL:   strToPtr(data.ServerURL),
		},
	}
}
