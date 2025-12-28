package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationGotifyResource{}
	_ resource.ResourceWithImportState = &NotificationGotifyResource{}
)

// NewNotificationGotifyResource returns a new instance of the Gotify notification resource.
func NewNotificationGotifyResource() resource.Resource {
	return &NotificationGotifyResource{}
}

// NotificationGotifyResource defines the resource implementation.
type NotificationGotifyResource struct {
	client *kuma.Client
}

// NotificationGotifyResourceModel describes the resource data model.
type NotificationGotifyResourceModel struct {
	NotificationBaseModel

	ServerURL        types.String `tfsdk:"server_url"`
	ApplicationToken types.String `tfsdk:"application_token"`
	Priority         types.Int64  `tfsdk:"priority"`
}

// Metadata returns the metadata for the resource.
func (*NotificationGotifyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_gotify"
}

// Schema returns the schema for the resource.
func (*NotificationGotifyResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gotify notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				MarkdownDescription: "Gotify server URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"application_token": schema.StringAttribute{
				MarkdownDescription: "Gotify application token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Gotify message priority (1-10, default 8)",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(8),
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
		}),
	}
}

// Configure configures the Gotify notification resource with the API client.
func (r *NotificationGotifyResource) Configure(
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

// Create creates a new Gotify notification resource.
func (r *NotificationGotifyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationGotifyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	gotify := notification.Gotify{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GotifyDetails: notification.GotifyDetails{
			ServerURL:        data.ServerURL.ValueString(),
			ApplicationToken: data.ApplicationToken.ValueString(),
			Priority:         int(data.Priority.ValueInt64()),
		},
	}

	id, err := r.client.CreateNotification(ctx, gotify)
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

// Read reads the current state of the Gotify notification resource.
func (r *NotificationGotifyResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationGotifyResourceModel

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

	gotify := notification.Gotify{}
	err = base.As(&gotify)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "gotify"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(gotify.Name)
	data.IsActive = types.BoolValue(gotify.IsActive)
	data.IsDefault = types.BoolValue(gotify.IsDefault)
	data.ApplyExisting = types.BoolValue(gotify.ApplyExisting)

	data.ServerURL = types.StringValue(gotify.ServerURL)
	data.ApplicationToken = types.StringValue(gotify.ApplicationToken)
	data.Priority = types.Int64Value(int64(gotify.Priority))

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Gotify notification resource.
func (r *NotificationGotifyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationGotifyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	gotify := notification.Gotify{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GotifyDetails: notification.GotifyDetails{
			ServerURL:        data.ServerURL.ValueString(),
			ApplicationToken: data.ApplicationToken.ValueString(),
			Priority:         int(data.Priority.ValueInt64()),
		},
	}

	err := r.client.UpdateNotification(ctx, gotify)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Gotify notification resource.
func (r *NotificationGotifyResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationGotifyResourceModel

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
func (*NotificationGotifyResource) ImportState(
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
