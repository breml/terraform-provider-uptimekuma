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
	_ resource.Resource                = &NotificationSignalResource{}
	_ resource.ResourceWithImportState = &NotificationSignalResource{}
)

// NewNotificationSignalResource returns a new instance of the Signal notification resource.
func NewNotificationSignalResource() resource.Resource {
	return &NotificationSignalResource{}
}

// NotificationSignalResource defines the resource implementation.
type NotificationSignalResource struct {
	client *kuma.Client
}

// NotificationSignalResourceModel describes the resource data model.
type NotificationSignalResourceModel struct {
	NotificationBaseModel

	URL        types.String `tfsdk:"url"`
	Number     types.String `tfsdk:"number"`
	Recipients types.String `tfsdk:"recipients"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSignalResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_signal"
}

// Schema returns the schema for the resource.
func (*NotificationSignalResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Signal notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Signal API URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"number": schema.StringAttribute{
				MarkdownDescription: "Signal bot phone number",
				Required:            true,
			},
			"recipients": schema.StringAttribute{
				MarkdownDescription: "Recipient phone numbers (comma-separated)",
				Required:            true,
			},
		}),
	}
}

// Configure configures the Signal notification resource with the API client.
func (r *NotificationSignalResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Signal notification resource.
func (r *NotificationSignalResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSignalResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	signal := notification.Signal{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SignalDetails: notification.SignalDetails{
			URL:        data.URL.ValueString(),
			Number:     data.Number.ValueString(),
			Recipients: data.Recipients.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, signal)
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

// Read reads the current state of the Signal notification resource.
func (r *NotificationSignalResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSignalResourceModel

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

	signal := notification.Signal{}
	err = base.As(&signal)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "signal"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(signal.Name)
	data.IsActive = types.BoolValue(signal.IsActive)
	data.IsDefault = types.BoolValue(signal.IsDefault)
	data.ApplyExisting = types.BoolValue(signal.ApplyExisting)

	data.URL = types.StringValue(signal.URL)
	data.Number = types.StringValue(signal.Number)
	data.Recipients = types.StringValue(signal.Recipients)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Signal notification resource.
func (r *NotificationSignalResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSignalResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	signal := notification.Signal{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SignalDetails: notification.SignalDetails{
			URL:        data.URL.ValueString(),
			Number:     data.Number.ValueString(),
			Recipients: data.Recipients.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, signal)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Signal notification resource.
func (r *NotificationSignalResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSignalResourceModel

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
func (*NotificationSignalResource) ImportState(
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
