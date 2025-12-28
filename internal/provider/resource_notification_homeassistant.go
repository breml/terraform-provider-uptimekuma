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
	_ resource.Resource                = &NotificationHomeAssistantResource{}
	_ resource.ResourceWithImportState = &NotificationHomeAssistantResource{}
)

// NewNotificationHomeAssistantResource returns a new instance of the Home Assistant notification resource.
func NewNotificationHomeAssistantResource() resource.Resource {
	return &NotificationHomeAssistantResource{}
}

// NotificationHomeAssistantResource defines the resource implementation.
type NotificationHomeAssistantResource struct {
	client *kuma.Client
}

// NotificationHomeAssistantResourceModel describes the resource data model.
type NotificationHomeAssistantResourceModel struct {
	NotificationBaseModel

	HomeAssistantURL     types.String `tfsdk:"home_assistant_url"`
	LongLivedAccessToken types.String `tfsdk:"long_lived_access_token"`
	NotificationService  types.String `tfsdk:"notification_service"`
}

// Metadata returns the metadata for the resource.
func (*NotificationHomeAssistantResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_homeassistant"
}

// Schema returns the schema for the resource.
func (*NotificationHomeAssistantResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Home Assistant notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"home_assistant_url": schema.StringAttribute{
				MarkdownDescription: "Home Assistant URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"long_lived_access_token": schema.StringAttribute{
				MarkdownDescription: "Home Assistant long-lived access token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"notification_service": schema.StringAttribute{
				MarkdownDescription: "Home Assistant notification service name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Home Assistant notification resource with the API client.
func (r *NotificationHomeAssistantResource) Configure(
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

// Create creates a new Home Assistant notification resource.
func (r *NotificationHomeAssistantResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationHomeAssistantResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	homeAssistant := notification.HomeAssistant{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		HomeAssistantDetails: notification.HomeAssistantDetails{
			HomeAssistantURL:     data.HomeAssistantURL.ValueString(),
			LongLivedAccessToken: data.LongLivedAccessToken.ValueString(),
			NotificationService:  data.NotificationService.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, homeAssistant)
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

// Read reads the current state of the Home Assistant notification resource.
func (r *NotificationHomeAssistantResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationHomeAssistantResourceModel

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

	homeAssistant := notification.HomeAssistant{}
	err = base.As(&homeAssistant)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "HomeAssistant"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(homeAssistant.Name)
	data.IsActive = types.BoolValue(homeAssistant.IsActive)
	data.IsDefault = types.BoolValue(homeAssistant.IsDefault)
	data.ApplyExisting = types.BoolValue(homeAssistant.ApplyExisting)

	data.HomeAssistantURL = types.StringValue(homeAssistant.HomeAssistantURL)
	data.LongLivedAccessToken = types.StringValue(homeAssistant.LongLivedAccessToken)
	data.NotificationService = types.StringValue(homeAssistant.NotificationService)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Home Assistant notification resource.
func (r *NotificationHomeAssistantResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationHomeAssistantResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	homeAssistant := notification.HomeAssistant{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		HomeAssistantDetails: notification.HomeAssistantDetails{
			HomeAssistantURL:     data.HomeAssistantURL.ValueString(),
			LongLivedAccessToken: data.LongLivedAccessToken.ValueString(),
			NotificationService:  data.NotificationService.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, homeAssistant)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Home Assistant notification resource.
func (r *NotificationHomeAssistantResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationHomeAssistantResourceModel

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
func (*NotificationHomeAssistantResource) ImportState(
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
