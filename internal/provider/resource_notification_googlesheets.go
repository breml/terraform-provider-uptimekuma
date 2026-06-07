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
	_ resource.Resource                = &NotificationGoogleSheetsResource{}
	_ resource.ResourceWithImportState = &NotificationGoogleSheetsResource{}
)

// NewNotificationGoogleSheetsResource returns a new instance of the Google Sheets notification resource.
func NewNotificationGoogleSheetsResource() resource.Resource {
	return &NotificationGoogleSheetsResource{}
}

// NotificationGoogleSheetsResource defines the resource implementation.
type NotificationGoogleSheetsResource struct {
	client *kuma.Client
}

// NotificationGoogleSheetsResourceModel describes the resource data model.
type NotificationGoogleSheetsResourceModel struct {
	NotificationBaseModel

	WebhookURL types.String `tfsdk:"webhook_url"`
}

// Metadata returns the metadata for the resource.
func (*NotificationGoogleSheetsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_googlesheets"
}

// Schema returns the schema for the resource.
func (*NotificationGoogleSheetsResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Google Sheets notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "Deployed Google Apps Script web app URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Google Sheets notification resource with the API client.
func (r *NotificationGoogleSheetsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Google Sheets notification resource.
func (r *NotificationGoogleSheetsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationGoogleSheetsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	googleSheets := notification.GoogleSheets{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GoogleSheetsDetails: notification.GoogleSheetsDetails{
			WebhookURL: data.WebhookURL.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, googleSheets)
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

// Read reads the current state of the Google Sheets notification resource.
func (r *NotificationGoogleSheetsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationGoogleSheetsResourceModel

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

	googleSheets := notification.GoogleSheets{}
	err = base.As(&googleSheets)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "GoogleSheets"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(googleSheets.Name)
	data.IsActive = types.BoolValue(googleSheets.IsActive)
	data.IsDefault = types.BoolValue(googleSheets.IsDefault)
	data.ApplyExisting = types.BoolValue(googleSheets.ApplyExisting)

	data.WebhookURL = types.StringValue(googleSheets.WebhookURL)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Google Sheets notification resource.
func (r *NotificationGoogleSheetsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationGoogleSheetsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	googleSheets := notification.GoogleSheets{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GoogleSheetsDetails: notification.GoogleSheetsDetails{
			WebhookURL: data.WebhookURL.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, googleSheets)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Google Sheets notification resource.
func (r *NotificationGoogleSheetsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationGoogleSheetsResourceModel

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
func (*NotificationGoogleSheetsResource) ImportState(
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
