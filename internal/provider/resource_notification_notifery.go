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
	_ resource.Resource                = &NotificationNotiferyResource{}
	_ resource.ResourceWithImportState = &NotificationNotiferyResource{}
)

// NewNotificationNotiferyResource returns a new instance of the Notifery notification resource.
func NewNotificationNotiferyResource() resource.Resource {
	return &NotificationNotiferyResource{}
}

// NotificationNotiferyResource defines the resource implementation.
type NotificationNotiferyResource struct {
	client *kuma.Client
}

// NotificationNotiferyResourceModel describes the resource data model.
type NotificationNotiferyResourceModel struct {
	NotificationBaseModel

	APIKey types.String `tfsdk:"api_key"`
	Title  types.String `tfsdk:"title"`
	Group  types.String `tfsdk:"group"`
}

// Metadata returns the metadata for the resource.
func (*NotificationNotiferyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_notifery"
}

// Schema returns the schema for the resource.
func (*NotificationNotiferyResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Notifery notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for authentication with Notifery service",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the notification",
				Optional:            true,
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "Notification group for organizing notifications",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Notifery notification resource with the API client.
func (r *NotificationNotiferyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Notifery notification resource.
func (r *NotificationNotiferyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationNotiferyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	notifery := notification.Notifery{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		NotiferyDetails: notification.NotiferyDetails{
			APIKey: data.APIKey.ValueString(),
			Title:  data.Title.ValueString(),
			Group:  data.Group.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, notifery)
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

// Read reads the current state of the Notifery notification resource.
func (r *NotificationNotiferyResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationNotiferyResourceModel

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

	notifery := notification.Notifery{}
	err = base.As(&notifery)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "notifery"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(notifery.Name)
	data.IsActive = types.BoolValue(notifery.IsActive)
	data.IsDefault = types.BoolValue(notifery.IsDefault)
	data.ApplyExisting = types.BoolValue(notifery.ApplyExisting)

	data.APIKey = types.StringValue(notifery.APIKey)

	if notifery.Title != "" {
		data.Title = types.StringValue(notifery.Title)
	}

	if notifery.Group != "" {
		data.Group = types.StringValue(notifery.Group)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Notifery notification resource.
func (r *NotificationNotiferyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationNotiferyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	notifery := notification.Notifery{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		NotiferyDetails: notification.NotiferyDetails{
			APIKey: data.APIKey.ValueString(),
			Title:  data.Title.ValueString(),
			Group:  data.Group.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, notifery)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Notifery notification resource.
func (r *NotificationNotiferyResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationNotiferyResourceModel

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
func (*NotificationNotiferyResource) ImportState(
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
