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
	_ resource.Resource                = &NotificationGoAlertResource{}
	_ resource.ResourceWithImportState = &NotificationGoAlertResource{}
)

// NewNotificationGoAlertResource returns a new instance of the GoAlert notification resource.
func NewNotificationGoAlertResource() resource.Resource {
	return &NotificationGoAlertResource{}
}

// NotificationGoAlertResource defines the resource implementation.
type NotificationGoAlertResource struct {
	client *kuma.Client
}

// NotificationGoAlertResourceModel describes the resource data model.
type NotificationGoAlertResourceModel struct {
	NotificationBaseModel

	BaseURL types.String `tfsdk:"base_url"`
	Token   types.String `tfsdk:"token"`
}

// Metadata returns the metadata for the resource.
func (*NotificationGoAlertResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_goalert"
}

// Schema returns the schema for the resource.
func (*NotificationGoAlertResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "GoAlert notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "GoAlert base URL",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "GoAlert integration token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the GoAlert notification resource with the API client.
func (r *NotificationGoAlertResource) Configure(
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

// Create creates a new GoAlert notification resource.
func (r *NotificationGoAlertResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationGoAlertResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	goalert := notification.GoAlert{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GoAlertDetails: notification.GoAlertDetails{
			BaseURL: data.BaseURL.ValueString(),
			Token:   data.Token.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, goalert)
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

// Read reads the current state of the GoAlert notification resource.
func (r *NotificationGoAlertResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationGoAlertResourceModel

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

	goalert := notification.GoAlert{}
	err = base.As(&goalert)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "goalert"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(goalert.Name)
	data.IsActive = types.BoolValue(goalert.IsActive)
	data.IsDefault = types.BoolValue(goalert.IsDefault)
	data.ApplyExisting = types.BoolValue(goalert.ApplyExisting)

	data.BaseURL = types.StringValue(goalert.BaseURL)
	data.Token = types.StringValue(goalert.Token)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the GoAlert notification resource.
func (r *NotificationGoAlertResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationGoAlertResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	goalert := notification.GoAlert{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GoAlertDetails: notification.GoAlertDetails{
			BaseURL: data.BaseURL.ValueString(),
			Token:   data.Token.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, goalert)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the GoAlert notification resource.
func (r *NotificationGoAlertResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationGoAlertResourceModel

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
func (*NotificationGoAlertResource) ImportState(
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
