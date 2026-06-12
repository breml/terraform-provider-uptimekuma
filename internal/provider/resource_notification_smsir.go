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
	_ resource.Resource                = &NotificationSMSIRResource{}
	_ resource.ResourceWithImportState = &NotificationSMSIRResource{}
)

// NewNotificationSMSIRResource returns a new instance of the SMS.ir notification resource.
func NewNotificationSMSIRResource() resource.Resource {
	return &NotificationSMSIRResource{}
}

// NotificationSMSIRResource defines the resource implementation.
type NotificationSMSIRResource struct {
	client *kuma.Client
}

// NotificationSMSIRResourceModel describes the resource data model.
type NotificationSMSIRResourceModel struct {
	NotificationBaseModel

	APIKey   types.String `tfsdk:"api_key"`
	Number   types.String `tfsdk:"number"`
	Template types.String `tfsdk:"template"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSMSIRResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_smsir"
}

// Schema returns the schema for the resource.
func (*NotificationSMSIRResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SMS.ir notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The SMS.ir API key used to send notifications.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"number": schema.StringAttribute{
				MarkdownDescription: "The recipient phone number(s). Multiple numbers can be " +
					"specified as a comma-separated list.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "The pre-approved SMS.ir template ID used for sending the message.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the SMS.ir notification resource with the API client.
func (r *NotificationSMSIRResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new SMS.ir notification resource.
func (r *NotificationSMSIRResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSMSIRResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smsir := smsirFromModel(&data)

	id, err := r.client.CreateNotification(ctx, smsir)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the SMS.ir notification resource.
func (r *NotificationSMSIRResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationSMSIRResourceModel

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

	smsir := notification.SMSIR{}
	err = base.As(&smsir)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "SMSIR"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(smsir.Name)
	data.IsActive = types.BoolValue(smsir.IsActive)
	data.IsDefault = types.BoolValue(smsir.IsDefault)
	data.ApplyExisting = types.BoolValue(smsir.ApplyExisting)

	data.APIKey = types.StringValue(smsir.APIKey)
	data.Number = types.StringValue(smsir.Number)
	data.Template = types.StringValue(smsir.Template)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SMS.ir notification resource.
func (r *NotificationSMSIRResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSMSIRResourceModel

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

	smsir := smsirFromModel(&data)
	smsir.ID = id

	err := r.client.UpdateNotification(ctx, smsir)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SMS.ir notification resource.
func (r *NotificationSMSIRResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSMSIRResourceModel

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
func (*NotificationSMSIRResource) ImportState(
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

// smsirFromModel builds an SMSIR notification from the resource model.
func smsirFromModel(data *NotificationSMSIRResourceModel) notification.SMSIR {
	return notification.SMSIR{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMSIRDetails: notification.SMSIRDetails{
			APIKey:   data.APIKey.ValueString(),
			Number:   data.Number.ValueString(),
			Template: data.Template.ValueString(),
		},
	}
}
