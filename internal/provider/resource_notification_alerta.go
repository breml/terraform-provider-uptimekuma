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
	_ resource.Resource                = &NotificationAlertaResource{}
	_ resource.ResourceWithImportState = &NotificationAlertaResource{}
)

// NewNotificationAlertaResource returns a new instance of the Alerta notification resource.
func NewNotificationAlertaResource() resource.Resource {
	return &NotificationAlertaResource{}
}

// NotificationAlertaResource defines the resource implementation.
type NotificationAlertaResource struct {
	client *kuma.Client
}

// NotificationAlertaResourceModel describes the resource data model.
type NotificationAlertaResourceModel struct {
	NotificationBaseModel

	APIEndpoint  types.String `tfsdk:"api_endpoint"`
	APIKey       types.String `tfsdk:"api_key"`
	Environment  types.String `tfsdk:"environment"`
	AlertState   types.String `tfsdk:"alert_state"`
	RecoverState types.String `tfsdk:"recover_state"`
}

// Metadata returns the metadata for the resource.
func (*NotificationAlertaResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_alerta"
}

// Schema returns the schema for the resource.
func (*NotificationAlertaResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Alerta notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_endpoint": schema.StringAttribute{
				MarkdownDescription: "Alerta API endpoint",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Alerta API key",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "Alerta environment",
				Optional:            true,
			},
			"alert_state": schema.StringAttribute{
				MarkdownDescription: "Alert state for incoming alerts",
				Optional:            true,
			},
			"recover_state": schema.StringAttribute{
				MarkdownDescription: "Alert state for recovered alerts",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Alerta notification resource with the API client.
func (r *NotificationAlertaResource) Configure(
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

// Create creates a new Alerta notification resource.
func (r *NotificationAlertaResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationAlertaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	alerta := notification.Alerta{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		AlertaDetails: notification.AlertaDetails{
			APIEndpoint:  data.APIEndpoint.ValueString(),
			APIKey:       data.APIKey.ValueString(),
			Environment:  data.Environment.ValueString(),
			AlertState:   data.AlertState.ValueString(),
			RecoverState: data.RecoverState.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, alerta)
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

// Read reads the current state of the Alerta notification resource.
func (r *NotificationAlertaResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationAlertaResourceModel

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

	alerta := notification.Alerta{}
	err = base.As(&alerta)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "alerta"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(alerta.Name)
	data.IsActive = types.BoolValue(alerta.IsActive)
	data.IsDefault = types.BoolValue(alerta.IsDefault)
	data.ApplyExisting = types.BoolValue(alerta.ApplyExisting)

	data.APIEndpoint = types.StringValue(alerta.APIEndpoint)
	data.APIKey = types.StringValue(alerta.APIKey)
	if alerta.Environment != "" {
		data.Environment = types.StringValue(alerta.Environment)
	}

	if alerta.AlertState != "" {
		data.AlertState = types.StringValue(alerta.AlertState)
	}

	if alerta.RecoverState != "" {
		data.RecoverState = types.StringValue(alerta.RecoverState)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Alerta notification resource.
func (r *NotificationAlertaResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationAlertaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	alerta := notification.Alerta{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		AlertaDetails: notification.AlertaDetails{
			APIEndpoint:  data.APIEndpoint.ValueString(),
			APIKey:       data.APIKey.ValueString(),
			Environment:  data.Environment.ValueString(),
			AlertState:   data.AlertState.ValueString(),
			RecoverState: data.RecoverState.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, alerta)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Alerta notification resource.
func (r *NotificationAlertaResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationAlertaResourceModel

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
func (*NotificationAlertaResource) ImportState(
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
