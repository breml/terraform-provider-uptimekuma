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
	_ resource.Resource                = &NotificationGTXMessagingResource{}
	_ resource.ResourceWithImportState = &NotificationGTXMessagingResource{}
)

// NewNotificationGTXMessagingResource returns a new instance of the GTX Messaging notification resource.
func NewNotificationGTXMessagingResource() resource.Resource {
	return &NotificationGTXMessagingResource{}
}

// NotificationGTXMessagingResource defines the resource implementation.
type NotificationGTXMessagingResource struct {
	client *kuma.Client
}

// NotificationGTXMessagingResourceModel describes the resource data model.
type NotificationGTXMessagingResourceModel struct {
	NotificationBaseModel

	APIKey types.String `tfsdk:"api_key"`
	From   types.String `tfsdk:"from"`
	To     types.String `tfsdk:"to"`
}

// Metadata returns the metadata for the resource.
func (*NotificationGTXMessagingResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_gtxmessaging"
}

// Schema returns the schema for the resource.
func (*NotificationGTXMessagingResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "GTX Messaging notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "GTX Messaging API key",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"from": schema.StringAttribute{
				MarkdownDescription: "Sender ID",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the GTX Messaging notification resource with the API client.
func (r *NotificationGTXMessagingResource) Configure(
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

// Create creates a new GTX Messaging notification resource.
func (r *NotificationGTXMessagingResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationGTXMessagingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	gtxmessaging := notification.GTXMessaging{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GTXMessagingDetails: notification.GTXMessagingDetails{
			APIKey: data.APIKey.ValueString(),
			From:   data.From.ValueString(),
			To:     data.To.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, gtxmessaging)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the GTX Messaging notification resource.
func (r *NotificationGTXMessagingResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationGTXMessagingResourceModel

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

	gtxmessaging := notification.GTXMessaging{}
	err = base.As(&gtxmessaging)
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "gtxmessaging"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(gtxmessaging.Name)
	data.IsActive = types.BoolValue(gtxmessaging.IsActive)
	data.IsDefault = types.BoolValue(gtxmessaging.IsDefault)
	data.ApplyExisting = types.BoolValue(gtxmessaging.ApplyExisting)

	data.APIKey = types.StringValue(gtxmessaging.APIKey)
	data.From = types.StringValue(gtxmessaging.From)
	data.To = types.StringValue(gtxmessaging.To)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the GTX Messaging notification resource.
func (r *NotificationGTXMessagingResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationGTXMessagingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	gtxmessaging := notification.GTXMessaging{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GTXMessagingDetails: notification.GTXMessagingDetails{
			APIKey: data.APIKey.ValueString(),
			From:   data.From.ValueString(),
			To:     data.To.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, gtxmessaging)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the GTX Messaging notification resource.
func (r *NotificationGTXMessagingResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationGTXMessagingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*NotificationGTXMessagingResource) ImportState(
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
