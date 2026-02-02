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
	_ resource.Resource                = &NotificationWAHAResource{}
	_ resource.ResourceWithImportState = &NotificationWAHAResource{}
)

// NewNotificationWAHAResource returns a new instance of the WAHA notification resource.
func NewNotificationWAHAResource() resource.Resource {
	return &NotificationWAHAResource{}
}

// NotificationWAHAResource defines the resource implementation.
type NotificationWAHAResource struct {
	client *kuma.Client
}

// NotificationWAHAResourceModel describes the resource data model.
type NotificationWAHAResourceModel struct {
	NotificationBaseModel

	APIURL  types.String `tfsdk:"api_url"`
	Session types.String `tfsdk:"session"`
	ChatID  types.String `tfsdk:"chat_id"`
	APIKey  types.String `tfsdk:"api_key"`
}

// Metadata returns the metadata for the resource.
func (*NotificationWAHAResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_waha"
}

// Schema returns the schema for the resource.
func (*NotificationWAHAResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "WAHA (WhatsApp HTTP API) notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: "WAHA API endpoint URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"session": schema.StringAttribute{
				MarkdownDescription: "WAHA session name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"chat_id": schema.StringAttribute{
				MarkdownDescription: "Recipient chat ID (typically a phone number)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "WAHA API key for authentication",
				Optional:            true,
				Sensitive:           true,
			},
		}),
	}
}

// Configure configures the WAHA notification resource with the API client.
func (r *NotificationWAHAResource) Configure(
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

// Create creates a new WAHA notification resource.
func (r *NotificationWAHAResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationWAHAResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	waha := notification.WAHA{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WAHADetails: notification.WAHADetails{
			APIURL:  data.APIURL.ValueString(),
			Session: data.Session.ValueString(),
			ChatID:  data.ChatID.ValueString(),
			APIKey:  data.APIKey.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, waha)
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

// Read reads the current state of the WAHA notification resource.
func (r *NotificationWAHAResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationWAHAResourceModel

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

	waha := notification.WAHA{}
	err = base.As(&waha)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "waha"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(waha.Name)
	data.IsActive = types.BoolValue(waha.IsActive)
	data.IsDefault = types.BoolValue(waha.IsDefault)
	data.ApplyExisting = types.BoolValue(waha.ApplyExisting)

	data.APIURL = types.StringValue(waha.APIURL)
	data.Session = types.StringValue(waha.Session)
	data.ChatID = types.StringValue(waha.ChatID)
	if waha.APIKey != "" {
		data.APIKey = types.StringValue(waha.APIKey)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the WAHA notification resource.
func (r *NotificationWAHAResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationWAHAResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	waha := notification.WAHA{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WAHADetails: notification.WAHADetails{
			APIURL:  data.APIURL.ValueString(),
			Session: data.Session.ValueString(),
			ChatID:  data.ChatID.ValueString(),
			APIKey:  data.APIKey.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, waha)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the WAHA notification resource.
func (r *NotificationWAHAResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationWAHAResourceModel

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
func (*NotificationWAHAResource) ImportState(
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
