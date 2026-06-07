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

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationFluxerResource{}
	_ resource.ResourceWithImportState = &NotificationFluxerResource{}
)

// NewNotificationFluxerResource returns a new instance of the Fluxer notification resource.
func NewNotificationFluxerResource() resource.Resource {
	return &NotificationFluxerResource{}
}

// NotificationFluxerResource defines the resource implementation.
type NotificationFluxerResource struct {
	client *kuma.Client
}

// NotificationFluxerResourceModel describes the resource data model.
type NotificationFluxerResourceModel struct {
	NotificationBaseModel

	WebhookURL         types.String `tfsdk:"webhook_url"`
	Username           types.String `tfsdk:"username"`
	PrefixMessage      types.String `tfsdk:"prefix_message"`
	DisableURL         types.Bool   `tfsdk:"disable_url"`
	UseMessageTemplate types.Bool   `tfsdk:"use_message_template"`
	MessageFormat      types.String `tfsdk:"message_format"`
	MessageTemplate    types.String `tfsdk:"message_template"`
}

// Metadata returns the metadata for the resource.
func (*NotificationFluxerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_fluxer"
}

// Schema returns the schema for the resource.
func (*NotificationFluxerResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fluxer notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "The Fluxer webhook URL to which notifications will be sent.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username to display as the sender of the Fluxer notification.",
				Optional:            true,
			},
			"prefix_message": schema.StringAttribute{
				MarkdownDescription: "A message to prefix to all Fluxer notifications.",
				Optional:            true,
			},
			"disable_url": schema.BoolAttribute{
				MarkdownDescription: "If true, disables including the monitor URL in the Fluxer notification.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"use_message_template": schema.BoolAttribute{
				MarkdownDescription: "If true, use a custom message template for Fluxer notifications.",
				Optional:            true,
			},
			"message_format": schema.StringAttribute{
				MarkdownDescription: "The format of the Fluxer notification message.",
				Optional:            true,
			},
			"message_template": schema.StringAttribute{
				MarkdownDescription: "The custom message template for Fluxer notifications.",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Fluxer notification resource with the API client.
func (r *NotificationFluxerResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Fluxer notification resource.
func (r *NotificationFluxerResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationFluxerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	fluxer := fluxerFromModel(&data)

	id, err := r.client.CreateNotification(ctx, fluxer)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Fluxer notification resource.
func (r *NotificationFluxerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationFluxerResourceModel

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

	fluxer := notification.Fluxer{}
	err = base.As(&fluxer)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "fluxer"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(fluxer.Name)
	data.IsActive = types.BoolValue(fluxer.IsActive)
	data.IsDefault = types.BoolValue(fluxer.IsDefault)
	data.ApplyExisting = types.BoolValue(fluxer.ApplyExisting)

	data.WebhookURL = types.StringValue(fluxer.WebhookURL)
	if fluxer.Username != "" {
		data.Username = types.StringValue(fluxer.Username)
	} else {
		data.Username = types.StringNull()
	}

	if fluxer.PrefixMessage != "" {
		data.PrefixMessage = types.StringValue(fluxer.PrefixMessage)
	} else {
		data.PrefixMessage = types.StringNull()
	}

	data.DisableURL = types.BoolValue(fluxer.DisableURL)
	data.UseMessageTemplate = boolPtrToTypes(fluxer.UseMessageTemplate)
	data.MessageFormat = ptrToTypes(fluxer.MessageFormat)
	data.MessageTemplate = ptrToTypes(fluxer.MessageTemplate)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Fluxer notification resource.
func (r *NotificationFluxerResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationFluxerResourceModel

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

	fluxer := fluxerFromModel(&data)
	fluxer.ID = id

	err := r.client.UpdateNotification(ctx, fluxer)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Fluxer notification resource.
func (r *NotificationFluxerResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationFluxerResourceModel

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
func (*NotificationFluxerResource) ImportState(
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

// fluxerFromModel builds a Fluxer notification from the resource model.
func fluxerFromModel(data *NotificationFluxerResourceModel) notification.Fluxer {
	return notification.Fluxer{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		FluxerDetails: notification.FluxerDetails{
			WebhookURL:         data.WebhookURL.ValueString(),
			Username:           data.Username.ValueString(),
			PrefixMessage:      data.PrefixMessage.ValueString(),
			DisableURL:         data.DisableURL.ValueBool(),
			UseMessageTemplate: boolToPtr(data.UseMessageTemplate),
			MessageFormat:      strToPtr(data.MessageFormat),
			MessageTemplate:    strToPtr(data.MessageTemplate),
		},
	}
}
