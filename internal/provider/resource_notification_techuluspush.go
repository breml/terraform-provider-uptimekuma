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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationTechulusPushResource{}
	_ resource.ResourceWithImportState = &NotificationTechulusPushResource{}
)

// NewNotificationTechulusPushResource returns a new instance of the TechulusPush notification resource.
func NewNotificationTechulusPushResource() resource.Resource {
	return &NotificationTechulusPushResource{}
}

// NotificationTechulusPushResource defines the resource implementation.
type NotificationTechulusPushResource struct {
	client *kuma.Client
}

// NotificationTechulusPushResourceModel describes the resource data model.
type NotificationTechulusPushResourceModel struct {
	NotificationBaseModel

	APIKey        types.String `tfsdk:"api_key"`
	Title         types.String `tfsdk:"title"`
	Sound         types.String `tfsdk:"sound"`
	Channel       types.String `tfsdk:"channel"`
	TimeSensitive types.Bool   `tfsdk:"time_sensitive"`
}

// Metadata returns the metadata for the resource.
func (*NotificationTechulusPushResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_techuluspush"
}

// Schema returns the schema for the resource.
func (*NotificationTechulusPushResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "TechulusPush notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Techulus Push API key for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Notification title",
				Optional:            true,
			},
			"sound": schema.StringAttribute{
				MarkdownDescription: "Notification sound identifier",
				Optional:            true,
			},
			"channel": schema.StringAttribute{
				MarkdownDescription: "Push notification channel",
				Optional:            true,
			},
			"time_sensitive": schema.BoolAttribute{
				MarkdownDescription: "Whether the notification is time-sensitive",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		}),
	}
}

// Configure configures the TechulusPush notification resource with the API client.
func (r *NotificationTechulusPushResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new TechulusPush notification resource.
func (r *NotificationTechulusPushResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationTechulusPushResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	techulusPush := notification.TechulusPush{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		TechulusPushDetails: notification.TechulusPushDetails{
			APIKey:        data.APIKey.ValueString(),
			Title:         data.Title.ValueString(),
			Sound:         data.Sound.ValueString(),
			Channel:       data.Channel.ValueString(),
			TimeSensitive: data.TimeSensitive.ValueBool(),
		},
	}

	id, err := r.client.CreateNotification(ctx, techulusPush)
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

// Read reads the current state of the TechulusPush notification resource.
func (r *NotificationTechulusPushResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationTechulusPushResourceModel

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

	techulusPush := notification.TechulusPush{}
	err = base.As(&techulusPush)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "PushByTechulus"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(techulusPush.Name)
	data.IsActive = types.BoolValue(techulusPush.IsActive)
	data.IsDefault = types.BoolValue(techulusPush.IsDefault)
	data.ApplyExisting = types.BoolValue(techulusPush.ApplyExisting)

	if techulusPush.APIKey != "" {
		data.APIKey = types.StringValue(techulusPush.APIKey)
	}

	if techulusPush.Title != "" {
		data.Title = types.StringValue(techulusPush.Title)
	} else {
		data.Title = types.StringNull()
	}

	if techulusPush.Sound != "" {
		data.Sound = types.StringValue(techulusPush.Sound)
	} else {
		data.Sound = types.StringNull()
	}

	if techulusPush.Channel != "" {
		data.Channel = types.StringValue(techulusPush.Channel)
	} else {
		data.Channel = types.StringNull()
	}

	data.TimeSensitive = types.BoolValue(techulusPush.TimeSensitive)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the TechulusPush notification resource.
func (r *NotificationTechulusPushResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationTechulusPushResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	techulusPush := notification.TechulusPush{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		TechulusPushDetails: notification.TechulusPushDetails{
			APIKey:        data.APIKey.ValueString(),
			Title:         data.Title.ValueString(),
			Sound:         data.Sound.ValueString(),
			Channel:       data.Channel.ValueString(),
			TimeSensitive: data.TimeSensitive.ValueBool(),
		},
	}

	err := r.client.UpdateNotification(ctx, techulusPush)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the TechulusPush notification resource.
func (r *NotificationTechulusPushResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationTechulusPushResourceModel

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
func (*NotificationTechulusPushResource) ImportState(
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
