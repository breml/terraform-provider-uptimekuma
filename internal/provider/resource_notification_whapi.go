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
	_ resource.Resource                = &NotificationWhapiResource{}
	_ resource.ResourceWithImportState = &NotificationWhapiResource{}
)

// NewNotificationWhapiResource returns a new instance of the Whapi notification resource.
func NewNotificationWhapiResource() resource.Resource {
	return &NotificationWhapiResource{}
}

// NotificationWhapiResource defines the resource implementation.
type NotificationWhapiResource struct {
	client *kuma.Client
}

// NotificationWhapiResourceModel describes the resource data model.
type NotificationWhapiResourceModel struct {
	NotificationBaseModel

	APIURL    types.String `tfsdk:"api_url"`
	AuthToken types.String `tfsdk:"auth_token"`
	Recipient types.String `tfsdk:"recipient"`
}

// Metadata returns the metadata for the resource.
func (*NotificationWhapiResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_whapi"
}

// Schema returns the schema for the resource.
func (*NotificationWhapiResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Whapi (WhatsApp API) notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Whapi API endpoint URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "Whapi API authorization token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"recipient": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Whapi notification resource with the API client.
func (r *NotificationWhapiResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Whapi notification resource.
func (r *NotificationWhapiResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationWhapiResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	whapi := notification.Whapi{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WhapiDetails: notification.WhapiDetails{
			APIURL:    data.APIURL.ValueString(),
			AuthToken: data.AuthToken.ValueString(),
			Recipient: data.Recipient.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, whapi)
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

// Read reads the current state of the Whapi notification resource.
func (r *NotificationWhapiResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationWhapiResourceModel

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

	whapi := notification.Whapi{}
	err = base.As(&whapi)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "whapi"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(whapi.Name)
	data.IsActive = types.BoolValue(whapi.IsActive)
	data.IsDefault = types.BoolValue(whapi.IsDefault)
	data.ApplyExisting = types.BoolValue(whapi.ApplyExisting)

	data.APIURL = types.StringValue(whapi.APIURL)
	data.AuthToken = types.StringValue(whapi.AuthToken)
	data.Recipient = types.StringValue(whapi.Recipient)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Whapi notification resource.
func (r *NotificationWhapiResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationWhapiResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	whapi := notification.Whapi{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		WhapiDetails: notification.WhapiDetails{
			APIURL:    data.APIURL.ValueString(),
			AuthToken: data.AuthToken.ValueString(),
			Recipient: data.Recipient.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, whapi)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Whapi notification resource.
func (r *NotificationWhapiResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationWhapiResourceModel

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
func (*NotificationWhapiResource) ImportState(
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
