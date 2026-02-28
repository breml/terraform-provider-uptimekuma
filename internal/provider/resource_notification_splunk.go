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
	_ resource.Resource                = &NotificationSplunkResource{}
	_ resource.ResourceWithImportState = &NotificationSplunkResource{}
)

// NewNotificationSplunkResource returns a new instance of the Splunk notification resource.
func NewNotificationSplunkResource() resource.Resource {
	return &NotificationSplunkResource{}
}

// NotificationSplunkResource defines the resource implementation.
type NotificationSplunkResource struct {
	client *kuma.Client
}

// NotificationSplunkResourceModel describes the resource data model.
type NotificationSplunkResourceModel struct {
	NotificationBaseModel

	RestURL        types.String `tfsdk:"rest_url"`
	Severity       types.String `tfsdk:"severity"`
	AutoResolve    types.String `tfsdk:"auto_resolve"`
	IntegrationKey types.String `tfsdk:"integration_key"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSplunkResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_splunk"
}

// Schema returns the schema for the resource.
func (*NotificationSplunkResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Splunk notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"rest_url": schema.StringAttribute{
				MarkdownDescription: "Splunk On-Call REST API URL endpoint",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"severity": schema.StringAttribute{
				MarkdownDescription: "Alert severity level for triggered events",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"auto_resolve": schema.StringAttribute{
				MarkdownDescription: "Action to take when a monitor recovers",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"integration_key": schema.StringAttribute{
				MarkdownDescription: "Splunk On-Call routing key for alert routing",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Splunk notification resource with the API client.
func (r *NotificationSplunkResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Splunk notification resource.
func (r *NotificationSplunkResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSplunkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	splunk := notification.Splunk{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SplunkDetails: notification.SplunkDetails{
			RestURL:        data.RestURL.ValueString(),
			Severity:       data.Severity.ValueString(),
			AutoResolve:    data.AutoResolve.ValueString(),
			IntegrationKey: data.IntegrationKey.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, splunk)
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

// Read reads the current state of the Splunk notification resource.
func (r *NotificationSplunkResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSplunkResourceModel

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

	splunk := notification.Splunk{}
	err = base.As(&splunk)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "splunk"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(splunk.Name)
	data.IsActive = types.BoolValue(splunk.IsActive)
	data.IsDefault = types.BoolValue(splunk.IsDefault)
	data.ApplyExisting = types.BoolValue(splunk.ApplyExisting)

	data.RestURL = types.StringValue(splunk.RestURL)
	data.Severity = types.StringValue(splunk.Severity)
	data.AutoResolve = types.StringValue(splunk.AutoResolve)
	data.IntegrationKey = types.StringValue(splunk.IntegrationKey)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Splunk notification resource.
func (r *NotificationSplunkResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSplunkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	splunk := notification.Splunk{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SplunkDetails: notification.SplunkDetails{
			RestURL:        data.RestURL.ValueString(),
			Severity:       data.Severity.ValueString(),
			AutoResolve:    data.AutoResolve.ValueString(),
			IntegrationKey: data.IntegrationKey.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, splunk)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Splunk notification resource.
func (r *NotificationSplunkResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSplunkResourceModel

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
func (*NotificationSplunkResource) ImportState(
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
