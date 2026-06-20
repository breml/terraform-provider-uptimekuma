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
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

const maxDefaultAPIURL = "https://platform-api.max.ru"

var (
	_ resource.Resource                = &NotificationMaxResource{}
	_ resource.ResourceWithImportState = &NotificationMaxResource{}
)

// NewNotificationMaxResource returns a new instance of the MAX messenger notification resource.
func NewNotificationMaxResource() resource.Resource {
	return &NotificationMaxResource{}
}

// NotificationMaxResource defines the resource implementation.
type NotificationMaxResource struct {
	client *kuma.Client
}

// NotificationMaxResourceModel describes the resource data model.
type NotificationMaxResourceModel struct {
	NotificationBaseModel

	APIURL         types.String `tfsdk:"api_url"`
	BotToken       types.String `tfsdk:"bot_token"`
	ChatID         types.String `tfsdk:"chat_id"`
	UseTemplate    types.Bool   `tfsdk:"use_template"`
	Template       types.String `tfsdk:"template"`
	TemplateFormat types.String `tfsdk:"template_format"`
}

// Metadata returns the metadata for the resource.
func (*NotificationMaxResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_max"
}

// Schema returns the schema for the resource.
func (*NotificationMaxResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MAX messenger notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: "MAX messenger API base URL",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(maxDefaultAPIURL),
			},
			"bot_token": schema.StringAttribute{
				MarkdownDescription: "MAX messenger bot token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"chat_id": schema.StringAttribute{
				MarkdownDescription: "MAX messenger chat ID to send notifications to",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"use_template": schema.BoolAttribute{
				MarkdownDescription: "Enable custom message template",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Custom message template for notifications",
				Optional:            true,
			},
			"template_format": schema.StringAttribute{
				MarkdownDescription: "Format of the custom message template (e.g. `markdown`)",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the MAX messenger notification resource with the API client.
func (r *NotificationMaxResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new MAX messenger notification resource.
func (r *NotificationMaxResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationMaxResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	maxNotification := maxFromModel(&data)

	id, err := r.client.CreateNotification(ctx, maxNotification)
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

// Read reads the current state of the MAX messenger notification resource.
func (r *NotificationMaxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationMaxResourceModel

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

	maxNotification := notification.Max{}
	err = base.As(&maxNotification)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "max"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(maxNotification.Name)
	data.IsActive = types.BoolValue(maxNotification.IsActive)
	data.IsDefault = types.BoolValue(maxNotification.IsDefault)
	data.ApplyExisting = types.BoolValue(maxNotification.ApplyExisting)

	data.APIURL = types.StringValue(maxNotification.APIURL)
	data.BotToken = types.StringValue(maxNotification.BotToken)
	data.ChatID = types.StringValue(maxNotification.ChatID)
	data.UseTemplate = types.BoolValue(maxNotification.UseTemplate)

	if maxNotification.Template != "" {
		data.Template = types.StringValue(maxNotification.Template)
	} else {
		data.Template = types.StringNull()
	}

	if maxNotification.TemplateFormat != "" {
		data.TemplateFormat = types.StringValue(maxNotification.TemplateFormat)
	} else {
		data.TemplateFormat = types.StringNull()
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the MAX messenger notification resource.
func (r *NotificationMaxResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationMaxResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	maxNotification := maxFromModel(&data)
	maxNotification.ID = data.ID.ValueInt64()

	err := r.client.UpdateNotification(ctx, maxNotification)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the MAX messenger notification resource.
func (r *NotificationMaxResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationMaxResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			return
		}

		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*NotificationMaxResource) ImportState(
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

// maxFromModel builds a MAX messenger notification from the resource model.
func maxFromModel(data *NotificationMaxResourceModel) notification.Max {
	return notification.Max{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		MaxDetails: notification.MaxDetails{
			APIURL:         data.APIURL.ValueString(),
			BotToken:       data.BotToken.ValueString(),
			ChatID:         data.ChatID.ValueString(),
			UseTemplate:    data.UseTemplate.ValueBool(),
			Template:       data.Template.ValueString(),
			TemplateFormat: data.TemplateFormat.ValueString(),
		},
	}
}
