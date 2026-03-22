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
	_ resource.Resource                = &NotificationOnesenderResource{}
	_ resource.ResourceWithImportState = &NotificationOnesenderResource{}
)

// NewNotificationOnesenderResource returns a new instance of the OneSender notification resource.
func NewNotificationOnesenderResource() resource.Resource {
	return &NotificationOnesenderResource{}
}

// NotificationOnesenderResource defines the resource implementation.
type NotificationOnesenderResource struct {
	client *kuma.Client
}

// NotificationOnesenderResourceModel describes the resource data model.
type NotificationOnesenderResourceModel struct {
	NotificationBaseModel

	URL          types.String `tfsdk:"url"`
	Token        types.String `tfsdk:"token"`
	Receiver     types.String `tfsdk:"receiver"`
	TypeReceiver types.String `tfsdk:"type_receiver"`
}

// Metadata returns the metadata for the resource.
func (*NotificationOnesenderResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_onesender"
}

// Schema returns the schema for the resource.
func (*NotificationOnesenderResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OneSender notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "OneSender API endpoint URL for sending messages",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "API token for authentication with OneSender service",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"receiver": schema.StringAttribute{
				MarkdownDescription: "Recipient identifier (phone number or group ID)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"type_receiver": schema.StringAttribute{
				MarkdownDescription: "Type of receiver: `private` for individual or `group` for group",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("private", "group"),
				},
			},
		}),
	}
}

// Configure configures the OneSender notification resource with the API client.
func (r *NotificationOnesenderResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new OneSender notification resource.
func (r *NotificationOnesenderResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationOnesenderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	onesender := notification.OneSender{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OneSenderDetails: notification.OneSenderDetails{
			URL:          data.URL.ValueString(),
			Token:        data.Token.ValueString(),
			Receiver:     data.Receiver.ValueString(),
			TypeReceiver: data.TypeReceiver.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, onesender)
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

// Read reads the current state of the OneSender notification resource.
func (r *NotificationOnesenderResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationOnesenderResourceModel

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

	onesender := notification.OneSender{}
	err = base.As(&onesender)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "onesender"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(onesender.Name)
	data.IsActive = types.BoolValue(onesender.IsActive)
	data.IsDefault = types.BoolValue(onesender.IsDefault)
	data.ApplyExisting = types.BoolValue(onesender.ApplyExisting)

	data.URL = types.StringValue(onesender.URL)
	data.Token = types.StringValue(onesender.Token)
	data.Receiver = types.StringValue(onesender.Receiver)
	data.TypeReceiver = types.StringValue(onesender.TypeReceiver)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the OneSender notification resource.
func (r *NotificationOnesenderResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationOnesenderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	onesender := notification.OneSender{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OneSenderDetails: notification.OneSenderDetails{
			URL:          data.URL.ValueString(),
			Token:        data.Token.ValueString(),
			Receiver:     data.Receiver.ValueString(),
			TypeReceiver: data.TypeReceiver.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, onesender)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the OneSender notification resource.
func (r *NotificationOnesenderResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationOnesenderResourceModel

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
func (*NotificationOnesenderResource) ImportState(
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
