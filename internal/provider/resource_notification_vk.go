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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationVKResource{}
	_ resource.ResourceWithImportState = &NotificationVKResource{}
)

// NewNotificationVKResource returns a new instance of the VK notification resource.
func NewNotificationVKResource() resource.Resource {
	return &NotificationVKResource{}
}

// NotificationVKResource defines the resource implementation.
type NotificationVKResource struct {
	client *kuma.Client
}

// NotificationVKResourceModel describes the resource data model.
type NotificationVKResourceModel struct {
	NotificationBaseModel

	AccessToken    types.String `tfsdk:"access_token"`
	PeerID         types.String `tfsdk:"peer_id"`
	APIVersion     types.String `tfsdk:"api_version"`
	DontParseLinks types.Bool   `tfsdk:"dont_parse_links"`
}

// Metadata returns the metadata for the resource.
func (*NotificationVKResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_vk"
}

// Schema returns the schema for the resource.
func (*NotificationVKResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VK (VKontakte) notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				MarkdownDescription: "The VK API user or service access token.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"peer_id": schema.StringAttribute{
				MarkdownDescription: "The recipient. Must be a numeric string: positive for user " +
					"IDs, negative for community IDs, or 2000000000+chat_id for group chats.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"api_version": schema.StringAttribute{
				MarkdownDescription: "The VK API version to use.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("5.199"),
			},
			"dont_parse_links": schema.BoolAttribute{
				MarkdownDescription: "If true, disables link previews in VK notifications.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		}),
	}
}

// Configure configures the VK notification resource with the API client.
func (r *NotificationVKResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new VK notification resource.
func (r *NotificationVKResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationVKResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vk := vkFromModel(&data)

	id, err := r.client.CreateNotification(ctx, vk)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the VK notification resource.
func (r *NotificationVKResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationVKResourceModel

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

	vk := notification.VK{}
	err = base.As(&vk)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "vk"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(vk.Name)
	data.IsActive = types.BoolValue(vk.IsActive)
	data.IsDefault = types.BoolValue(vk.IsDefault)
	data.ApplyExisting = types.BoolValue(vk.ApplyExisting)

	data.AccessToken = types.StringValue(vk.AccessToken)
	data.PeerID = types.StringValue(vk.PeerID)
	data.APIVersion = types.StringValue(vk.APIVersion)
	data.DontParseLinks = types.BoolValue(vk.DontParseLinks)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the VK notification resource.
func (r *NotificationVKResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationVKResourceModel

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

	vk := vkFromModel(&data)
	vk.ID = id

	err := r.client.UpdateNotification(ctx, vk)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the VK notification resource.
func (r *NotificationVKResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationVKResourceModel

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
func (*NotificationVKResource) ImportState(
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

// vkFromModel builds a VK notification from the resource model.
func vkFromModel(data *NotificationVKResourceModel) notification.VK {
	return notification.VK{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		VKDetails: notification.VKDetails{
			AccessToken:    data.AccessToken.ValueString(),
			PeerID:         data.PeerID.ValueString(),
			APIVersion:     data.APIVersion.ValueString(),
			DontParseLinks: data.DontParseLinks.ValueBool(),
		},
	}
}
