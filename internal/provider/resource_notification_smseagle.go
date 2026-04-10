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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationSMSEagleResource{}
	_ resource.ResourceWithImportState = &NotificationSMSEagleResource{}
)

// NewNotificationSMSEagleResource returns a new instance of the SMSEagle notification resource.
func NewNotificationSMSEagleResource() resource.Resource {
	return &NotificationSMSEagleResource{}
}

// NotificationSMSEagleResource defines the resource implementation.
type NotificationSMSEagleResource struct {
	client *kuma.Client
}

// NotificationSMSEagleResourceModel describes the resource data model.
type NotificationSMSEagleResourceModel struct {
	NotificationBaseModel

	URL              types.String `tfsdk:"url"`
	Token            types.String `tfsdk:"token"`
	RecipientType    types.String `tfsdk:"recipient_type"`
	Recipient        types.String `tfsdk:"recipient"`
	RecipientTo      types.String `tfsdk:"recipient_to"`
	RecipientContact types.String `tfsdk:"recipient_contact"`
	RecipientGroup   types.String `tfsdk:"recipient_group"`
	MsgType          types.String `tfsdk:"msg_type"`
	Priority         types.Int64  `tfsdk:"priority"`
	Encoding         types.Bool   `tfsdk:"encoding"`
	Duration         types.Int64  `tfsdk:"duration"`
	TtsModel         types.Int64  `tfsdk:"tts_model"`
	APIType          types.String `tfsdk:"api_type"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSMSEagleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_smseagle"
}

// Schema returns the schema for the resource.
func (*NotificationSMSEagleResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SMSEagle notification resource. SMSEagle is an SMS gateway solution " +
			"that supports SMS, MMS, ring calls, and text-to-speech calls.",
		Attributes: withNotificationBaseAttributes(smseagleSchemaAttributes()),
	}
}

func smseagleSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			MarkdownDescription: "SMSEagle device URL (e.g., https://192.168.1.100)",
			Required:            true,
			Sensitive:           true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"token": schema.StringAttribute{
			MarkdownDescription: "API access token for authentication",
			Required:            true,
			Sensitive:           true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"recipient_type": schema.StringAttribute{
			MarkdownDescription: "Recipient type: `smseagle-to` (phone number), " +
				"`smseagle-contact` (contact ID), or `smseagle-group` (group ID)",
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					"smseagle-to",
					"smseagle-contact",
					"smseagle-group",
				),
			},
		},
		"recipient": schema.StringAttribute{
			MarkdownDescription: "Recipient identifier for API v1 (phone number, contact ID, or group ID)",
			Optional:            true,
		},
		"recipient_to": schema.StringAttribute{
			MarkdownDescription: "Recipient phone number(s) for API v2 (comma-separated for multiple)",
			Optional:            true,
		},
		"recipient_contact": schema.StringAttribute{
			MarkdownDescription: "Contact recipient ID(s) for API v2 (comma-separated for multiple)",
			Optional:            true,
		},
		"recipient_group": schema.StringAttribute{
			MarkdownDescription: "Group recipient ID(s) for API v2 (comma-separated for multiple)",
			Optional:            true,
		},
		"msg_type": schema.StringAttribute{
			MarkdownDescription: "Message type: `smseagle-sms`, `smseagle-ring`, " +
				"`smseagle-tts`, or `smseagle-tts-advanced`",
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("smseagle-sms"),
			Validators: []validator.String{
				stringvalidator.OneOf(
					"smseagle-sms",
					"smseagle-ring",
					"smseagle-tts",
					"smseagle-tts-advanced",
				),
			},
		},
		"priority": schema.Int64Attribute{
			MarkdownDescription: "Message priority level (0-9, where 9 is highest)",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(0),
			Validators: []validator.Int64{
				int64validator.Between(0, 9),
			},
		},
		"encoding": schema.BoolAttribute{
			MarkdownDescription: "Enable unicode encoding for non-ASCII characters",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"duration": schema.Int64Attribute{
			MarkdownDescription: "Duration for voice calls in seconds",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(10),
		},
		"tts_model": schema.Int64Attribute{
			MarkdownDescription: "TTS voice ID/model for text-to-speech calls",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(1),
		},
		"api_type": schema.StringAttribute{
			MarkdownDescription: "API version: `smseagle-apiv1` or `smseagle-apiv2`",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("smseagle-apiv2"),
			Validators: []validator.String{
				stringvalidator.OneOf(
					"smseagle-apiv1",
					"smseagle-apiv2",
				),
			},
		},
	}
}

// Configure configures the SMSEagle notification resource with the API client.
func (r *NotificationSMSEagleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new SMSEagle notification resource.
func (r *NotificationSMSEagleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSMSEagleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smseagle := notification.SMSEagle{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMSEagleDetails: notification.SMSEagleDetails{
			URL:              data.URL.ValueString(),
			Token:            data.Token.ValueString(),
			RecipientType:    data.RecipientType.ValueString(),
			Recipient:        data.Recipient.ValueString(),
			RecipientTo:      data.RecipientTo.ValueString(),
			RecipientContact: data.RecipientContact.ValueString(),
			RecipientGroup:   data.RecipientGroup.ValueString(),
			MsgType:          data.MsgType.ValueString(),
			Priority:         int(data.Priority.ValueInt64()),
			Encoding:         data.Encoding.ValueBool(),
			Duration:         int(data.Duration.ValueInt64()),
			TtsModel:         int(data.TtsModel.ValueInt64()),
			APIType:          data.APIType.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, smseagle)
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

// Read reads the current state of the SMSEagle notification resource.
func (r *NotificationSMSEagleResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSMSEagleResourceModel

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

	smseagle := notification.SMSEagle{}
	err = base.As(&smseagle)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "SMSEagle"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(smseagle.Name)
	data.IsActive = types.BoolValue(smseagle.IsActive)
	data.IsDefault = types.BoolValue(smseagle.IsDefault)
	data.ApplyExisting = types.BoolValue(smseagle.ApplyExisting)

	if smseagle.URL != "" {
		data.URL = types.StringValue(smseagle.URL)
	}

	if smseagle.Token != "" {
		data.Token = types.StringValue(smseagle.Token)
	}

	data.RecipientType = types.StringValue(smseagle.RecipientType)

	if smseagle.Recipient != "" {
		data.Recipient = types.StringValue(smseagle.Recipient)
	} else {
		data.Recipient = types.StringNull()
	}

	if smseagle.RecipientTo != "" {
		data.RecipientTo = types.StringValue(smseagle.RecipientTo)
	} else {
		data.RecipientTo = types.StringNull()
	}

	if smseagle.RecipientContact != "" {
		data.RecipientContact = types.StringValue(smseagle.RecipientContact)
	} else {
		data.RecipientContact = types.StringNull()
	}

	if smseagle.RecipientGroup != "" {
		data.RecipientGroup = types.StringValue(smseagle.RecipientGroup)
	} else {
		data.RecipientGroup = types.StringNull()
	}

	data.MsgType = types.StringValue(smseagle.MsgType)
	data.Priority = types.Int64Value(int64(smseagle.Priority))
	data.Encoding = types.BoolValue(smseagle.Encoding)
	data.Duration = types.Int64Value(int64(smseagle.Duration))
	data.TtsModel = types.Int64Value(int64(smseagle.TtsModel))
	data.APIType = types.StringValue(smseagle.APIType)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SMSEagle notification resource.
func (r *NotificationSMSEagleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSMSEagleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smseagle := notification.SMSEagle{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMSEagleDetails: notification.SMSEagleDetails{
			URL:              data.URL.ValueString(),
			Token:            data.Token.ValueString(),
			RecipientType:    data.RecipientType.ValueString(),
			Recipient:        data.Recipient.ValueString(),
			RecipientTo:      data.RecipientTo.ValueString(),
			RecipientContact: data.RecipientContact.ValueString(),
			RecipientGroup:   data.RecipientGroup.ValueString(),
			MsgType:          data.MsgType.ValueString(),
			Priority:         int(data.Priority.ValueInt64()),
			Encoding:         data.Encoding.ValueBool(),
			Duration:         int(data.Duration.ValueInt64()),
			TtsModel:         int(data.TtsModel.ValueInt64()),
			APIType:          data.APIType.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, smseagle)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SMSEagle notification resource.
func (r *NotificationSMSEagleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSMSEagleResourceModel

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
func (*NotificationSMSEagleResource) ImportState(
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
