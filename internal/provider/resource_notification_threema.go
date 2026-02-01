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
	_ resource.Resource                = &NotificationThreemaResource{}
	_ resource.ResourceWithImportState = &NotificationThreemaResource{}
)

// NewNotificationThreemaResource returns a new instance of the Threema notification resource.
func NewNotificationThreemaResource() resource.Resource {
	return &NotificationThreemaResource{}
}

// NotificationThreemaResource defines the resource implementation.
type NotificationThreemaResource struct {
	client *kuma.Client
}

// NotificationThreemaResourceModel describes the resource data model.
type NotificationThreemaResourceModel struct {
	NotificationBaseModel

	SenderIdentity types.String `tfsdk:"sender_identity"`
	Secret         types.String `tfsdk:"secret"`
	Recipient      types.String `tfsdk:"recipient"`
	RecipientType  types.String `tfsdk:"recipient_type"`
}

// Metadata returns the metadata for the resource.
func (*NotificationThreemaResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_threema"
}

// Schema returns the schema for the resource.
func (*NotificationThreemaResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Threema notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"sender_identity": schema.StringAttribute{
				MarkdownDescription: "Threema Gateway ID for the sender",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "Threema API secret for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"recipient": schema.StringAttribute{
				MarkdownDescription: "Recipient identifier (ID, phone number, or email address)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"recipient_type": schema.StringAttribute{
				MarkdownDescription: "Type of recipient (identity, phone, or email)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Threema notification resource with the API client.
func (r *NotificationThreemaResource) Configure(
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

// Create creates a new Threema notification resource.
func (r *NotificationThreemaResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationThreemaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	threema := notification.Threema{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		ThreemaDetails: notification.ThreemaDetails{
			SenderIdentity: data.SenderIdentity.ValueString(),
			Secret:         data.Secret.ValueString(),
			Recipient:      data.Recipient.ValueString(),
			RecipientType:  data.RecipientType.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, threema)
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

// Read reads the current state of the Threema notification resource.
func (r *NotificationThreemaResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationThreemaResourceModel

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

	threema := notification.Threema{}
	err = base.As(&threema)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "threema"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(threema.Name)
	data.IsActive = types.BoolValue(threema.IsActive)
	data.IsDefault = types.BoolValue(threema.IsDefault)
	data.ApplyExisting = types.BoolValue(threema.ApplyExisting)

	data.SenderIdentity = types.StringValue(threema.SenderIdentity)
	data.Secret = types.StringValue(threema.Secret)
	data.Recipient = types.StringValue(threema.Recipient)
	data.RecipientType = types.StringValue(threema.RecipientType)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Threema notification resource.
func (r *NotificationThreemaResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationThreemaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	threema := notification.Threema{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		ThreemaDetails: notification.ThreemaDetails{
			SenderIdentity: data.SenderIdentity.ValueString(),
			Secret:         data.Secret.ValueString(),
			Recipient:      data.Recipient.ValueString(),
			RecipientType:  data.RecipientType.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, threema)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Threema notification resource.
func (r *NotificationThreemaResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationThreemaResourceModel

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
func (*NotificationThreemaResource) ImportState(
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
