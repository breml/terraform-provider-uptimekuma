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
	_ resource.Resource                = &NotificationGrafanaOncallResource{}
	_ resource.ResourceWithImportState = &NotificationGrafanaOncallResource{}
)

// NewNotificationGrafanaOncallResource returns a new instance of the Grafana OnCall notification resource.
func NewNotificationGrafanaOncallResource() resource.Resource {
	return &NotificationGrafanaOncallResource{}
}

// NotificationGrafanaOncallResource defines the resource implementation.
type NotificationGrafanaOncallResource struct {
	client *kuma.Client
}

// NotificationGrafanaOncallResourceModel describes the resource data model.
type NotificationGrafanaOncallResourceModel struct {
	NotificationBaseModel

	GrafanaOncallURL types.String `tfsdk:"grafana_oncall_url"`
}

// Metadata returns the metadata for the resource.
func (*NotificationGrafanaOncallResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_grafanaoncall"
}

// Schema returns the schema for the resource.
func (*NotificationGrafanaOncallResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Grafana OnCall notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"grafana_oncall_url": schema.StringAttribute{
				MarkdownDescription: "Grafana OnCall webhook URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Grafana OnCall notification resource with the API client.
func (r *NotificationGrafanaOncallResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Grafana OnCall notification resource.
func (r *NotificationGrafanaOncallResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationGrafanaOncallResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	grafanaOncall := notification.GrafanaOncall{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GrafanaOncallDetails: notification.GrafanaOncallDetails{
			GrafanaOncallURL: data.GrafanaOncallURL.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, grafanaOncall)
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

// Read reads the current state of the Grafana OnCall notification resource.
func (r *NotificationGrafanaOncallResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationGrafanaOncallResourceModel

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

	grafanaOncall := notification.GrafanaOncall{}
	err = base.As(&grafanaOncall)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "GrafanaOncall"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(grafanaOncall.Name)
	data.IsActive = types.BoolValue(grafanaOncall.IsActive)
	data.IsDefault = types.BoolValue(grafanaOncall.IsDefault)
	data.ApplyExisting = types.BoolValue(grafanaOncall.ApplyExisting)

	data.GrafanaOncallURL = types.StringValue(grafanaOncall.GrafanaOncallURL)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Grafana OnCall notification resource.
func (r *NotificationGrafanaOncallResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationGrafanaOncallResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	grafanaOncall := notification.GrafanaOncall{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GrafanaOncallDetails: notification.GrafanaOncallDetails{
			GrafanaOncallURL: data.GrafanaOncallURL.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, grafanaOncall)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Grafana OnCall notification resource.
func (r *NotificationGrafanaOncallResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationGrafanaOncallResourceModel

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
func (*NotificationGrafanaOncallResource) ImportState(
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
