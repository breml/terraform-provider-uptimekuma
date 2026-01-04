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
	_ resource.Resource                = &NotificationFlashDutyResource{}
	_ resource.ResourceWithImportState = &NotificationFlashDutyResource{}
)

// NewNotificationFlashDutyResource returns a new instance of the FlashDuty notification resource.
func NewNotificationFlashDutyResource() resource.Resource {
	return &NotificationFlashDutyResource{}
}

// NotificationFlashDutyResource defines the resource implementation.
type NotificationFlashDutyResource struct {
	client *kuma.Client
}

// NotificationFlashDutyResourceModel describes the resource data model.
type NotificationFlashDutyResourceModel struct {
	NotificationBaseModel

	IntegrationKey types.String `tfsdk:"integration_key"`
	Severity       types.String `tfsdk:"severity"`
}

// Metadata returns the metadata for the resource.
func (*NotificationFlashDutyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_flashduty"
}

// Schema returns the schema for the resource.
func (*NotificationFlashDutyResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "FlashDuty notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"integration_key": schema.StringAttribute{
				MarkdownDescription: "FlashDuty integration key or webhook URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"severity": schema.StringAttribute{
				MarkdownDescription: "Alert severity level (Info, Warning, Critical, Ok)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Critical"),
			},
		}),
	}
}

// Configure configures the FlashDuty notification resource with the API client.
func (r *NotificationFlashDutyResource) Configure(
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

// Create creates a new FlashDuty notification resource.
func (r *NotificationFlashDutyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationFlashDutyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	flashduty := notification.FlashDuty{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		FlashDutyDetails: notification.FlashDutyDetails{
			IntegrationKey: data.IntegrationKey.ValueString(),
			Severity:       data.Severity.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, flashduty)
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

// Read reads the current state of the FlashDuty notification resource.
func (r *NotificationFlashDutyResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationFlashDutyResourceModel

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

	flashduty := notification.FlashDuty{}
	err = base.As(&flashduty)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "flashduty"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(flashduty.Name)
	data.IsActive = types.BoolValue(flashduty.IsActive)
	data.IsDefault = types.BoolValue(flashduty.IsDefault)
	data.ApplyExisting = types.BoolValue(flashduty.ApplyExisting)

	data.IntegrationKey = types.StringValue(flashduty.IntegrationKey)
	if flashduty.Severity != "" {
		data.Severity = types.StringValue(flashduty.Severity)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the FlashDuty notification resource.
func (r *NotificationFlashDutyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationFlashDutyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	flashduty := notification.FlashDuty{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		FlashDutyDetails: notification.FlashDutyDetails{
			IntegrationKey: data.IntegrationKey.ValueString(),
			Severity:       data.Severity.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, flashduty)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the FlashDuty notification resource.
func (r *NotificationFlashDutyResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationFlashDutyResourceModel

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
func (*NotificationFlashDutyResource) ImportState(
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
