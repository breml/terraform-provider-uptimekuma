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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationOctopushResource{}
	_ resource.ResourceWithImportState = &NotificationOctopushResource{}
)

// NewNotificationOctopushResource returns a new instance of the Octopush notification resource.
func NewNotificationOctopushResource() resource.Resource {
	return &NotificationOctopushResource{}
}

// NotificationOctopushResource defines the resource implementation.
type NotificationOctopushResource struct {
	client *kuma.Client
}

// NotificationOctopushResourceModel describes the resource data model.
type NotificationOctopushResourceModel struct {
	NotificationBaseModel

	Version       types.String `tfsdk:"version"`
	APIKey        types.String `tfsdk:"api_key"`
	Login         types.String `tfsdk:"login"`
	PhoneNumber   types.String `tfsdk:"phone_number"`
	SMSType       types.String `tfsdk:"sms_type"`
	SenderName    types.String `tfsdk:"sender_name"`
	DMLogin       types.String `tfsdk:"dm_login"`
	DMAPIKey      types.String `tfsdk:"dm_api_key"`
	DMPhoneNumber types.String `tfsdk:"dm_phone_number"`
	DMSMSType     types.String `tfsdk:"dm_sms_type"`
	DMSenderName  types.String `tfsdk:"dm_sender_name"`
}

// Metadata returns the metadata for the resource.
func (*NotificationOctopushResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_octopush"
}

// Schema returns the schema for the resource.
func (*NotificationOctopushResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Octopush notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"version": schema.StringAttribute{
				MarkdownDescription: "Octopush API version (\"1\" for Direct Mail or \"2\" for standard API)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("2"),
				Validators: []validator.String{
					stringvalidator.OneOf("1", "2"),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for Octopush V2 API authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"login": schema.StringAttribute{
				MarkdownDescription: "Login/username for Octopush V2 API authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"phone_number": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number for Octopush V2 API",
				Optional:            true,
			},
			"sms_type": schema.StringAttribute{
				MarkdownDescription: "SMS type for Octopush V2 API (e.g., \"sms_premium\", \"sms_low_cost\")",
				Optional:            true,
			},
			"sender_name": schema.StringAttribute{
				MarkdownDescription: "Sender name for Octopush V2 API",
				Optional:            true,
			},
			"dm_login": schema.StringAttribute{
				MarkdownDescription: "Login/username for Octopush V1 (Direct Mail) API",
				Optional:            true,
				Sensitive:           true,
			},
			"dm_api_key": schema.StringAttribute{
				MarkdownDescription: "API key for Octopush V1 (Direct Mail) API",
				Optional:            true,
				Sensitive:           true,
			},
			"dm_phone_number": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number for Octopush V1 API",
				Optional:            true,
			},
			"dm_sms_type": schema.StringAttribute{
				MarkdownDescription: "SMS type for Octopush V1 API",
				Optional:            true,
			},
			"dm_sender_name": schema.StringAttribute{
				MarkdownDescription: "Sender name for Octopush V1 API",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Octopush notification resource with the API client.
func (r *NotificationOctopushResource) Configure(
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

// Create creates a new Octopush notification resource.
func (r *NotificationOctopushResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationOctopushResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	octopush := notification.Octopush{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OctopushDetails: notification.OctopushDetails{
			Version:       data.Version.ValueString(),
			APIKey:        data.APIKey.ValueString(),
			Login:         data.Login.ValueString(),
			PhoneNumber:   data.PhoneNumber.ValueString(),
			SMSType:       data.SMSType.ValueString(),
			SenderName:    data.SenderName.ValueString(),
			DMLogin:       data.DMLogin.ValueString(),
			DMAPIKey:      data.DMAPIKey.ValueString(),
			DMPhoneNumber: data.DMPhoneNumber.ValueString(),
			DMSMSType:     data.DMSMSType.ValueString(),
			DMSenderName:  data.DMSenderName.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, octopush)
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

// Read reads the current state of the Octopush notification resource.
func (r *NotificationOctopushResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationOctopushResourceModel

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

	octopush := notification.Octopush{}
	err = base.As(&octopush)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "octopush"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(octopush.Name)
	data.IsActive = types.BoolValue(octopush.IsActive)
	data.IsDefault = types.BoolValue(octopush.IsDefault)
	data.ApplyExisting = types.BoolValue(octopush.ApplyExisting)

	data.Version = types.StringValue(octopush.Version)
	if octopush.APIKey != "" {
		data.APIKey = types.StringValue(octopush.APIKey)
	}

	if octopush.Login != "" {
		data.Login = types.StringValue(octopush.Login)
	}

	if octopush.PhoneNumber != "" {
		data.PhoneNumber = types.StringValue(octopush.PhoneNumber)
	}

	if octopush.SMSType != "" {
		data.SMSType = types.StringValue(octopush.SMSType)
	}

	if octopush.SenderName != "" {
		data.SenderName = types.StringValue(octopush.SenderName)
	}

	if octopush.DMLogin != "" {
		data.DMLogin = types.StringValue(octopush.DMLogin)
	}

	if octopush.DMAPIKey != "" {
		data.DMAPIKey = types.StringValue(octopush.DMAPIKey)
	}

	if octopush.DMPhoneNumber != "" {
		data.DMPhoneNumber = types.StringValue(octopush.DMPhoneNumber)
	}

	if octopush.DMSMSType != "" {
		data.DMSMSType = types.StringValue(octopush.DMSMSType)
	}

	if octopush.DMSenderName != "" {
		data.DMSenderName = types.StringValue(octopush.DMSenderName)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Octopush notification resource.
func (r *NotificationOctopushResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationOctopushResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	octopush := notification.Octopush{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OctopushDetails: notification.OctopushDetails{
			Version:       data.Version.ValueString(),
			APIKey:        data.APIKey.ValueString(),
			Login:         data.Login.ValueString(),
			PhoneNumber:   data.PhoneNumber.ValueString(),
			SMSType:       data.SMSType.ValueString(),
			SenderName:    data.SenderName.ValueString(),
			DMLogin:       data.DMLogin.ValueString(),
			DMAPIKey:      data.DMAPIKey.ValueString(),
			DMPhoneNumber: data.DMPhoneNumber.ValueString(),
			DMSMSType:     data.DMSMSType.ValueString(),
			DMSenderName:  data.DMSenderName.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, octopush)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Octopush notification resource.
func (r *NotificationOctopushResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationOctopushResourceModel

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
func (*NotificationOctopushResource) ImportState(
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
