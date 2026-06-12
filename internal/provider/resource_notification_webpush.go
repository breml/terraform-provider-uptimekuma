package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationWebpushResource{}
	_ resource.ResourceWithImportState = &NotificationWebpushResource{}
)

// NewNotificationWebpushResource returns a new instance of the Web Push notification resource.
func NewNotificationWebpushResource() resource.Resource {
	return &NotificationWebpushResource{}
}

// NotificationWebpushResource defines the resource implementation.
type NotificationWebpushResource struct {
	client *kuma.Client
}

// NotificationWebpushResourceModel describes the resource data model.
type NotificationWebpushResourceModel struct {
	NotificationBaseModel

	Subscription types.Object `tfsdk:"subscription"`
}

// WebpushSubscriptionModel describes the W3C PushSubscription nested object data model.
type WebpushSubscriptionModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Keys     types.Object `tfsdk:"keys"`
}

// WebpushSubscriptionKeysModel describes the Web Push subscription encryption keys data model.
type WebpushSubscriptionKeysModel struct {
	P256dh types.String `tfsdk:"p256dh"`
	Auth   types.String `tfsdk:"auth"`
}

// Metadata returns the metadata for the resource.
func (*NotificationWebpushResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_webpush"
}

// Schema returns the schema for the resource.
func (*NotificationWebpushResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Web Push (W3C Push API) notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"subscription": schema.SingleNestedAttribute{
				MarkdownDescription: "W3C PushSubscription identifying the push endpoint",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						MarkdownDescription: "Push service endpoint URL",
						Required:            true,
						Sensitive:           true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"keys": schema.SingleNestedAttribute{
						MarkdownDescription: "Encryption keys for the push subscription",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"p256dh": schema.StringAttribute{
								MarkdownDescription: "P-256 Diffie-Hellman public key",
								Required:            true,
								Sensitive:           true,
								Validators: []validator.String{
									stringvalidator.LengthAtLeast(1),
								},
							},
							"auth": schema.StringAttribute{
								MarkdownDescription: "Authentication secret",
								Required:            true,
								Sensitive:           true,
								Validators: []validator.String{
									stringvalidator.LengthAtLeast(1),
								},
							},
						},
					},
				},
			},
		}),
	}
}

// Configure configures the Web Push notification resource with the API client.
func (r *NotificationWebpushResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Web Push notification resource.
func (r *NotificationWebpushResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationWebpushResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	subscription := expandWebpushSubscription(ctx, data.Subscription, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	webpush := notification.Webpush{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WebpushDetails: notification.WebpushDetails{
			Subscription: subscription,
		},
	}

	id, err := r.client.CreateNotification(ctx, webpush)
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

// Read reads the current state of the Web Push notification resource.
func (r *NotificationWebpushResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationWebpushResourceModel

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

	webpush := notification.Webpush{}
	err = base.As(&webpush)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "webpush"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(webpush.Name)
	data.IsActive = types.BoolValue(webpush.IsActive)
	data.IsDefault = types.BoolValue(webpush.IsDefault)
	data.ApplyExisting = types.BoolValue(webpush.ApplyExisting)

	data.Subscription = flattenWebpushSubscription(webpush.Subscription, &resp.Diagnostics)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Web Push notification resource.
func (r *NotificationWebpushResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationWebpushResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	subscription := expandWebpushSubscription(ctx, data.Subscription, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	webpush := notification.Webpush{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WebpushDetails: notification.WebpushDetails{
			Subscription: subscription,
		},
	}

	err := r.client.UpdateNotification(ctx, webpush)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Web Push notification resource.
func (r *NotificationWebpushResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationWebpushResourceModel

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
func (*NotificationWebpushResource) ImportState(
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

// webpushSubscriptionKeysAttrTypes returns the attribute types for the subscription keys nested object.
func webpushSubscriptionKeysAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"p256dh": types.StringType,
		"auth":   types.StringType,
	}
}

// webpushSubscriptionAttrTypes returns the attribute types for the subscription nested object.
func webpushSubscriptionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"endpoint": types.StringType,
		"keys":     types.ObjectType{AttrTypes: webpushSubscriptionKeysAttrTypes()},
	}
}

// expandWebpushSubscription converts the Terraform subscription object into the client library type.
func expandWebpushSubscription(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) notification.WebpushSubscription {
	var subscription WebpushSubscriptionModel

	diags.Append(obj.As(ctx, &subscription, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return notification.WebpushSubscription{}
	}

	var keys WebpushSubscriptionKeysModel

	diags.Append(subscription.Keys.As(ctx, &keys, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return notification.WebpushSubscription{}
	}

	return notification.WebpushSubscription{
		Endpoint: subscription.Endpoint.ValueString(),
		Keys: notification.WebpushSubscriptionKeys{
			P256dh: keys.P256dh.ValueString(),
			Auth:   keys.Auth.ValueString(),
		},
	}
}

// flattenWebpushSubscription converts the client library subscription type into a Terraform object value.
func flattenWebpushSubscription(
	subscription notification.WebpushSubscription,
	diags *diag.Diagnostics,
) types.Object {
	keys, d := types.ObjectValue(webpushSubscriptionKeysAttrTypes(), map[string]attr.Value{
		"p256dh": types.StringValue(subscription.Keys.P256dh),
		"auth":   types.StringValue(subscription.Keys.Auth),
	})
	diags.Append(d...)

	subscriptionObj, d := types.ObjectValue(webpushSubscriptionAttrTypes(), map[string]attr.Value{
		"endpoint": types.StringValue(subscription.Endpoint),
		"keys":     keys,
	})
	diags.Append(d...)

	return subscriptionObj
}
