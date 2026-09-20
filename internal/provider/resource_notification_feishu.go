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
	_ resource.Resource                = &NotificationFeishuResource{}
	_ resource.ResourceWithImportState = &NotificationFeishuResource{}
)

// NewNotificationFeishuResource returns a new instance of the Feishu notification resource.
func NewNotificationFeishuResource() resource.Resource {
	return &NotificationFeishuResource{}
}

// NotificationFeishuResource defines the resource implementation.
type NotificationFeishuResource struct {
	client *kuma.Client
}

// NotificationFeishuResourceModel describes the resource data model.
type NotificationFeishuResourceModel struct {
	NotificationBaseModel

	WebHookURL types.String `tfsdk:"webhook_url"`
}

// Metadata returns the metadata for the resource.
func (*NotificationFeishuResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_feishu"
}

// Schema returns the schema for the resource.
func (*NotificationFeishuResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Feishu notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "Feishu webhook URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Feishu notification resource with the API client.
func (r *NotificationFeishuResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Feishu notification resource.
func (r *NotificationFeishuResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationFeishuResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	feishu := notification.Feishu{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		FeishuDetails: notification.FeishuDetails{
			WebHookURL: data.WebHookURL.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, feishu)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Feishu notification resource.
func (r *NotificationFeishuResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationFeishuResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.FeishuDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	feishu := notification.Feishu{}
	err := base.As(&feishu)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.FeishuDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(feishu.Name)
	data.IsActive = types.BoolValue(feishu.IsActive)
	data.IsDefault = types.BoolValue(feishu.IsDefault)
	data.ApplyExisting = types.BoolValue(feishu.ApplyExisting)

	data.WebHookURL = types.StringValue(feishu.WebHookURL)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Feishu notification resource.
func (r *NotificationFeishuResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationFeishuResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	feishu := notification.Feishu{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		FeishuDetails: notification.FeishuDetails{
			WebHookURL: data.WebHookURL.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, feishu)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Feishu notification resource.
func (r *NotificationFeishuResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationFeishuResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationFeishuResource) ImportState(
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
