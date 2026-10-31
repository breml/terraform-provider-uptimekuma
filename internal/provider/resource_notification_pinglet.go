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
	_ resource.Resource                = &NotificationPingletResource{}
	_ resource.ResourceWithImportState = &NotificationPingletResource{}
)

// NewNotificationPingletResource returns a new instance of the Pinglet notification resource.
func NewNotificationPingletResource() resource.Resource {
	return &NotificationPingletResource{}
}

// NotificationPingletResource defines the resource implementation.
type NotificationPingletResource struct {
	client *kuma.Client
}

// NotificationPingletResourceModel describes the resource data model.
type NotificationPingletResourceModel struct {
	NotificationBaseModel

	PublishURL      types.String `tfsdk:"publish_url"`
	APIKey          types.String `tfsdk:"api_key"`
	APIKeyWO        types.String `tfsdk:"api_key_wo"`
	APIKeyWOVersion types.String `tfsdk:"api_key_wo_version"`
}

// Metadata returns the metadata for the resource.
func (*NotificationPingletResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_pinglet"
}

// Schema returns the schema for the resource.
func (*NotificationPingletResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pinglet notification resource. The alert is published as a push notification " +
			"to a Pinglet topic. Requires Uptime Kuma 2.5.1 or later.",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"publish_url": schema.StringAttribute{
				MarkdownDescription: "The publish URL of the Pinglet topic, for example " +
					"`https://app.pinglet.co.uk/your-namespace/alerts`. The topic is created on the first " +
					"publish. When sending, Uptime Kuma strips a single trailing slash and always appends " +
					"the `rewrite=uptimekuma` query parameter.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The Pinglet API key, sent as a bearer token.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}

	resp.Schema.Attributes = withWriteOnlySecret(
		resp.Schema.Attributes,
		"api_key",
		"The Pinglet API key, sent as a bearer token",
	)
}

// Configure configures the Pinglet notification resource with the API client.
func (r *NotificationPingletResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Pinglet notification resource.
func (r *NotificationPingletResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationPingletResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pinglet := pingletFromModel(&data)
	pinglet.APIKey = writeOnlyOrPlan(ctx, req.Config, "api_key", data.APIKey, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateNotification(ctx, pinglet)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Pinglet notification resource.
func (r *NotificationPingletResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationPingletResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.PingletDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	pinglet := notification.Pinglet{}
	err := base.As(&pinglet)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.PingletDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(pinglet.Name)
	data.IsActive = types.BoolValue(pinglet.IsActive)
	data.IsDefault = types.BoolValue(pinglet.IsDefault)
	data.ApplyExisting = types.BoolValue(pinglet.ApplyExisting)

	data.PublishURL = types.StringValue(pinglet.PublishURL)

	if data.APIKeyWOVersion.IsNull() {
		data.APIKey = types.StringValue(pinglet.APIKey)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Pinglet notification resource.
func (r *NotificationPingletResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationPingletResourceModel

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

	pinglet := pingletFromModel(&data)
	pinglet.ID = id
	pinglet.APIKey = writeOnlyOrPlan(ctx, req.Config, "api_key", data.APIKey, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNotification(ctx, pinglet)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Pinglet notification resource.
func (r *NotificationPingletResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationPingletResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationPingletResource) ImportState(
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

// pingletFromModel builds a Pinglet notification from the resource model. The caller resolves
// api_key, because its value may come from the write-only api_key_wo, which is only in the config,
// not in the plan.
func pingletFromModel(data *NotificationPingletResourceModel) notification.Pinglet {
	return notification.Pinglet{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PingletDetails: notification.PingletDetails{
			PublishURL: data.PublishURL.ValueString(),
		},
	}
}
