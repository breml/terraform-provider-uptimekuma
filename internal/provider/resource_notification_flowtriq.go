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
	_ resource.Resource                = &NotificationFlowtriqResource{}
	_ resource.ResourceWithImportState = &NotificationFlowtriqResource{}
)

// NewNotificationFlowtriqResource returns a new instance of the Flowtriq notification resource.
func NewNotificationFlowtriqResource() resource.Resource {
	return &NotificationFlowtriqResource{}
}

// NotificationFlowtriqResource defines the resource implementation.
type NotificationFlowtriqResource struct {
	client *kuma.Client
}

// NotificationFlowtriqResourceModel describes the resource data model.
type NotificationFlowtriqResourceModel struct {
	NotificationBaseModel

	WebhookURL types.String `tfsdk:"webhook_url"`
	APIKey     types.String `tfsdk:"api_key"`
}

// Metadata returns the metadata for the resource.
func (*NotificationFlowtriqResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_flowtriq"
}

// Schema returns the schema for the resource.
func (*NotificationFlowtriqResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Flowtriq notification resource. Flowtriq is a DDoS detection and incident " +
			"platform, which receives the notifications via a webhook.",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "The Flowtriq webhook endpoint the notification is posted to, for " +
					"example `https://app.flowtriq.com/api/webhooks/0123456789abcdef`.",
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					validateURL(),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key authenticating the request against Flowtriq. Uptime Kuma " +
					"sends it as the `X-API-Key` header. If unset, the header is omitted.",
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Flowtriq notification resource with the API client.
func (r *NotificationFlowtriqResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Flowtriq notification resource.
func (r *NotificationFlowtriqResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationFlowtriqResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	flowtriq := flowtriqFromModel(&data)

	id, err := r.client.CreateNotification(ctx, flowtriq)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Flowtriq notification resource.
func (r *NotificationFlowtriqResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationFlowtriqResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.FlowtriqDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)

		return
	}

	flowtriq := notification.Flowtriq{}
	err := base.As(&flowtriq)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.FlowtriqDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(flowtriq.Name)
	data.IsActive = types.BoolValue(flowtriq.IsActive)
	data.IsDefault = types.BoolValue(flowtriq.IsDefault)
	data.ApplyExisting = types.BoolValue(flowtriq.ApplyExisting)

	data.WebhookURL = types.StringValue(flowtriq.WebhookURL)

	// flowtriqApiKey is a *string, so omitempty drops it only when it is nil, not when it points at
	// the empty string. A notification whose API key was cleared in the Uptime Kuma UI comes back as
	// "", and keeping that in state against a null configuration would be a perpetual diff.
	if flowtriq.APIKey == nil || *flowtriq.APIKey == "" {
		data.APIKey = types.StringNull()
	} else {
		data.APIKey = types.StringValue(*flowtriq.APIKey)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Flowtriq notification resource.
func (r *NotificationFlowtriqResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationFlowtriqResourceModel

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

	flowtriq := flowtriqFromModel(&data)
	flowtriq.ID = id

	err := r.client.UpdateNotification(ctx, flowtriq)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Flowtriq notification resource.
func (r *NotificationFlowtriqResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationFlowtriqResourceModel

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
func (*NotificationFlowtriqResource) ImportState(
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

// flowtriqFromModel builds a Flowtriq notification from the resource model.
//
// flowtriqApiKey is omitempty on the wire, so a null api_key maps to nil, which keeps it out of the
// request entirely and lets Uptime Kuma omit the X-API-Key header.
func flowtriqFromModel(data *NotificationFlowtriqResourceModel) notification.Flowtriq {
	return notification.Flowtriq{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		FlowtriqDetails: notification.FlowtriqDetails{
			WebhookURL: data.WebhookURL.ValueString(),
			APIKey:     strToPtr(data.APIKey),
		},
	}
}
