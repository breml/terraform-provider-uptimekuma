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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationWebhookResource{}
	_ resource.ResourceWithImportState = &NotificationWebhookResource{}
)

// NewNotificationWebhookResource returns a new instance of the Webhook notification resource.
func NewNotificationWebhookResource() resource.Resource {
	return &NotificationWebhookResource{}
}

// NotificationWebhookResource defines the resource implementation.
type NotificationWebhookResource struct {
	client *kuma.Client
}

// NotificationWebhookResourceModel describes the resource data model.
type NotificationWebhookResourceModel struct {
	NotificationBaseModel

	WebhookURL               types.String `tfsdk:"webhook_url"`
	WebhookContentType       types.String `tfsdk:"webhook_content_type"`
	WebhookCustomBody        types.String `tfsdk:"webhook_custom_body"`
	WebhookAdditionalHeaders types.Map    `tfsdk:"webhook_additional_headers"`
}

// Metadata returns the metadata for the resource.
func (*NotificationWebhookResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_webhook"
}

// Schema returns the schema for the resource.
func (*NotificationWebhookResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Webhook notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "Webhook endpoint URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"webhook_content_type": schema.StringAttribute{
				MarkdownDescription: "Content type for the webhook payload. Supported values: `json`, `form-data`, `custom`. Defaults to `json`",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("json"),
				Validators: []validator.String{
					stringvalidator.OneOf("json", "form-data", "custom"),
				},
			},
			"webhook_custom_body": schema.StringAttribute{
				MarkdownDescription: "Custom JSON body template (only used when webhook_content_type is `custom`). Supports template variables like `{{ msg }}` and `{{ monitorJSON['name'] }}`",
				Optional:            true,
			},
			"webhook_additional_headers": schema.MapAttribute{
				MarkdownDescription: "Additional HTTP headers to send with the webhook request (e.g., Authorization headers)",
				ElementType:         types.StringType,
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Webhook notification resource with the API client.
func (r *NotificationWebhookResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	r.client = client
}

// Create creates a new Webhook notification resource.
func (r *NotificationWebhookResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationWebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	headers := make(notification.WebhookAdditionalHeaders)
	if !data.WebhookAdditionalHeaders.IsNull() && !data.WebhookAdditionalHeaders.IsUnknown() {
		diags := data.WebhookAdditionalHeaders.ElementsAs(ctx, &headers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	webhook := notification.Webhook{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WebhookDetails: notification.WebhookDetails{
			WebhookURL:               data.WebhookURL.ValueString(),
			WebhookContentType:       data.WebhookContentType.ValueString(),
			WebhookCustomBody:        data.WebhookCustomBody.ValueString(),
			WebhookAdditionalHeaders: headers,
		},
	}

	id, err := r.client.CreateNotification(ctx, webhook)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Created webhook notification", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Webhook notification resource.
func (r *NotificationWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationWebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, err := r.client.GetNotification(ctx, id)
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	webhook := notification.Webhook{}
	err = base.As(&webhook)
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "webhook"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(webhook.Name)
	data.IsActive = types.BoolValue(webhook.IsActive)
	data.IsDefault = types.BoolValue(webhook.IsDefault)
	data.ApplyExisting = types.BoolValue(webhook.ApplyExisting)

	data.WebhookURL = types.StringValue(webhook.WebhookURL)
	data.WebhookContentType = types.StringValue(webhook.WebhookContentType)

	if webhook.WebhookCustomBody != "" {
		data.WebhookCustomBody = types.StringValue(webhook.WebhookCustomBody)
	} else {
		data.WebhookCustomBody = types.StringNull()
	}

	if len(webhook.WebhookAdditionalHeaders) > 0 {
		headersMap, diags := types.MapValueFrom(ctx, types.StringType, webhook.WebhookAdditionalHeaders)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.WebhookAdditionalHeaders = headersMap
	} else {
		data.WebhookAdditionalHeaders = types.MapNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Webhook notification resource.
func (r *NotificationWebhookResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationWebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	headers := make(notification.WebhookAdditionalHeaders)
	if !data.WebhookAdditionalHeaders.IsNull() && !data.WebhookAdditionalHeaders.IsUnknown() {
		diags := data.WebhookAdditionalHeaders.ElementsAs(ctx, &headers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	webhook := notification.Webhook{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WebhookDetails: notification.WebhookDetails{
			WebhookURL:               data.WebhookURL.ValueString(),
			WebhookContentType:       data.WebhookContentType.ValueString(),
			WebhookCustomBody:        data.WebhookCustomBody.ValueString(),
			WebhookAdditionalHeaders: headers,
		},
	}

	err := r.client.UpdateNotification(ctx, webhook)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	tflog.Info(ctx, "Updated webhook notification", map[string]any{"id": data.ID.ValueInt64()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Webhook notification resource.
func (r *NotificationWebhookResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationWebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}

	tflog.Info(ctx, "Deleted webhook notification", map[string]any{"id": data.ID.ValueInt64()})
}

// ImportState imports an existing resource by ID.
func (*NotificationWebhookResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
