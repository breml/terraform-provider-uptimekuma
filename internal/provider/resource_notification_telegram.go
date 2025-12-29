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

var (
	_ resource.Resource                = &NotificationTelegramResource{}
	_ resource.ResourceWithImportState = &NotificationTelegramResource{}
)

// NewNotificationTelegramResource returns a new instance of the Telegram notification resource.
func NewNotificationTelegramResource() resource.Resource {
	return &NotificationTelegramResource{}
}

// NotificationTelegramResource defines the resource implementation.
type NotificationTelegramResource struct {
	client *kuma.Client
}

// NotificationTelegramResourceModel describes the resource data model.
type NotificationTelegramResourceModel struct {
	NotificationBaseModel

	BotToken          types.String `tfsdk:"bot_token"`
	ChatID            types.String `tfsdk:"chat_id"`
	ServerURL         types.String `tfsdk:"server_url"`
	SendSilently      types.Bool   `tfsdk:"send_silently"`
	ProtectContent    types.Bool   `tfsdk:"protect_content"`
	MessageThreadID   types.String `tfsdk:"message_thread_id"`
	UseTemplate       types.Bool   `tfsdk:"use_template"`
	Template          types.String `tfsdk:"template"`
	TemplateParseMode types.String `tfsdk:"template_parse_mode"`
}

// Metadata returns the metadata for the resource.
func (*NotificationTelegramResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_telegram"
}

// Schema returns the schema for the resource.
func (*NotificationTelegramResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Telegram notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"bot_token": schema.StringAttribute{
				MarkdownDescription: "Telegram bot token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"chat_id": schema.StringAttribute{
				MarkdownDescription: "Telegram chat ID",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"server_url": schema.StringAttribute{
				MarkdownDescription: "Telegram server URL (optional, defaults to official Telegram API)",
				Optional:            true,
			},
			"send_silently": schema.BoolAttribute{
				MarkdownDescription: "Send message silently",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"protect_content": schema.BoolAttribute{
				MarkdownDescription: "Protect content from forwarding",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"message_thread_id": schema.StringAttribute{
				MarkdownDescription: "Message thread ID for topics in supergroups",
				Optional:            true,
			},
			"use_template": schema.BoolAttribute{
				MarkdownDescription: "Use custom message template",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Custom message template",
				Optional:            true,
			},
			"template_parse_mode": schema.StringAttribute{
				MarkdownDescription: "Template parse mode (HTML, Markdown, MarkdownV2)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("HTML"),
				Validators: []validator.String{
					stringvalidator.OneOf("HTML", "Markdown", "MarkdownV2"),
				},
			},
		}),
	}
}

// Configure configures the Telegram notification resource with the API client.
func (r *NotificationTelegramResource) Configure(
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

// Create creates a new Telegram notification resource.
func (r *NotificationTelegramResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationTelegramResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	telegram := notification.Telegram{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		TelegramDetails: notification.TelegramDetails{
			BotToken:          data.BotToken.ValueString(),
			ChatID:            data.ChatID.ValueString(),
			ServerURL:         data.ServerURL.ValueString(),
			SendSilently:      data.SendSilently.ValueBool(),
			ProtectContent:    data.ProtectContent.ValueBool(),
			MessageThreadID:   data.MessageThreadID.ValueString(),
			UseTemplate:       data.UseTemplate.ValueBool(),
			Template:          data.Template.ValueString(),
			TemplateParseMode: data.TemplateParseMode.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, telegram)
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

// Read reads the current state of the Telegram notification resource.
func (r *NotificationTelegramResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationTelegramResourceModel

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

	telegram := notification.Telegram{}
	err = base.As(&telegram)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "telegram"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(telegram.Name)
	data.IsActive = types.BoolValue(telegram.IsActive)
	data.IsDefault = types.BoolValue(telegram.IsDefault)
	data.ApplyExisting = types.BoolValue(telegram.ApplyExisting)

	data.BotToken = types.StringValue(telegram.BotToken)
	data.ChatID = types.StringValue(telegram.ChatID)
	if telegram.ServerURL != "" {
		data.ServerURL = types.StringValue(telegram.ServerURL)
	} else {
		data.ServerURL = types.StringNull()
	}

	data.SendSilently = types.BoolValue(telegram.SendSilently)
	data.ProtectContent = types.BoolValue(telegram.ProtectContent)
	if telegram.MessageThreadID != "" {
		data.MessageThreadID = types.StringValue(telegram.MessageThreadID)
	} else {
		data.MessageThreadID = types.StringNull()
	}

	data.UseTemplate = types.BoolValue(telegram.UseTemplate)
	if telegram.Template != "" {
		data.Template = types.StringValue(telegram.Template)
	} else {
		data.Template = types.StringNull()
	}

	data.TemplateParseMode = types.StringValue(telegram.TemplateParseMode)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Telegram notification resource.
func (r *NotificationTelegramResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationTelegramResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	telegram := notification.Telegram{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		TelegramDetails: notification.TelegramDetails{
			BotToken:          data.BotToken.ValueString(),
			ChatID:            data.ChatID.ValueString(),
			ServerURL:         data.ServerURL.ValueString(),
			SendSilently:      data.SendSilently.ValueBool(),
			ProtectContent:    data.ProtectContent.ValueBool(),
			MessageThreadID:   data.MessageThreadID.ValueString(),
			UseTemplate:       data.UseTemplate.ValueBool(),
			Template:          data.Template.ValueString(),
			TemplateParseMode: data.TemplateParseMode.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, telegram)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Telegram notification resource.
func (r *NotificationTelegramResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationTelegramResourceModel

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
func (*NotificationTelegramResource) ImportState(
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
