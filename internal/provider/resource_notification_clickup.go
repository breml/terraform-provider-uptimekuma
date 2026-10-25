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
	_ resource.Resource                = &NotificationClickUpResource{}
	_ resource.ResourceWithImportState = &NotificationClickUpResource{}
)

// NewNotificationClickUpResource returns a new instance of the ClickUp notification resource.
func NewNotificationClickUpResource() resource.Resource {
	return &NotificationClickUpResource{}
}

// NotificationClickUpResource defines the resource implementation.
type NotificationClickUpResource struct {
	client *kuma.Client
}

// NotificationClickUpResourceModel describes the resource data model.
type NotificationClickUpResourceModel struct {
	NotificationBaseModel

	Token          types.String `tfsdk:"token"`
	TokenWO        types.String `tfsdk:"token_wo"`
	TokenWOVersion types.String `tfsdk:"token_wo_version"`
	WorkspaceID    types.String `tfsdk:"workspace_id"`
	ChannelID      types.String `tfsdk:"channel_id"`
	DisableURL     types.Bool   `tfsdk:"disable_url"`
}

// Metadata returns the metadata for the resource.
func (*NotificationClickUpResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_clickup"
}

// Schema returns the schema for the resource.
func (*NotificationClickUpResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ClickUp notification resource. ClickUp is a project management platform; the " +
			"notification is posted as a Markdown message into one of its chat channels.",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "The ClickUp personal access token (`pk_…`). Uptime Kuma sends it verbatim as the " +
					"`Authorization` header, so it must not carry a `Bearer` prefix.",
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "The id of the workspace holding the chat channel, the segment of the " +
					"workspace URL right after `app.clickup.com/`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"channel_id": schema.StringAttribute{
				MarkdownDescription: "The id of the chat channel the message is posted to, the segment of the " +
					"channel URL right after `/chat/r/`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"disable_url": schema.BoolAttribute{
				MarkdownDescription: "If true, the monitor address is left out of the message body. Leaving it " +
					"unset behaves like `false`.",
				Optional: true,
			},
		}),
	}

	resp.Schema.Attributes = withWriteOnlySecret(
		resp.Schema.Attributes,
		"token",
		"The ClickUp personal access token (`pk_…`), without a `Bearer` prefix",
	)
}

// Configure configures the ClickUp notification resource with the API client.
func (r *NotificationClickUpResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new ClickUp notification resource.
func (r *NotificationClickUpResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationClickUpResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clickUp := clickUpFromModel(&data)
	clickUp.Token = writeOnlyOrPlan(ctx, req.Config, "token", data.Token, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateNotification(ctx, clickUp)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the ClickUp notification resource.
func (r *NotificationClickUpResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationClickUpResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.ClickUpDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	clickUp := notification.ClickUp{}
	err := base.As(&clickUp)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.ClickUpDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(clickUp.Name)
	data.IsActive = types.BoolValue(clickUp.IsActive)
	data.IsDefault = types.BoolValue(clickUp.IsDefault)
	data.ApplyExisting = types.BoolValue(clickUp.ApplyExisting)

	if data.TokenWOVersion.IsNull() {
		data.Token = types.StringValue(clickUp.Token)
	}

	data.WorkspaceID = types.StringValue(clickUp.WorkspaceID)
	data.ChannelID = types.StringValue(clickUp.ChannelID)
	data.DisableURL = boolPtrToTypes(clickUp.DisableURL)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the ClickUp notification resource.
func (r *NotificationClickUpResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationClickUpResourceModel

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

	clickUp := clickUpFromModel(&data)
	clickUp.Token = writeOnlyOrPlan(ctx, req.Config, "token", data.Token, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	clickUp.ID = id

	err := r.client.UpdateNotification(ctx, clickUp)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the ClickUp notification resource.
func (r *NotificationClickUpResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationClickUpResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationClickUpResource) ImportState(
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

// clickUpFromModel builds a ClickUp notification from the resource model.
func clickUpFromModel(data *NotificationClickUpResourceModel) notification.ClickUp {
	return notification.ClickUp{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		ClickUpDetails: notification.ClickUpDetails{
			Token:       data.Token.ValueString(),
			WorkspaceID: data.WorkspaceID.ValueString(),
			ChannelID:   data.ChannelID.ValueString(),
			DisableURL:  boolToPtr(data.DisableURL),
		},
	}
}
