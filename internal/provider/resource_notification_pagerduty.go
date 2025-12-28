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
	_ resource.Resource                = &NotificationPagerDutyResource{}
	_ resource.ResourceWithImportState = &NotificationPagerDutyResource{}
)

// NewNotificationPagerDutyResource returns a new instance of the PagerDuty notification resource.
func NewNotificationPagerDutyResource() resource.Resource {
	return &NotificationPagerDutyResource{}
}

// NotificationPagerDutyResource defines the resource implementation.
type NotificationPagerDutyResource struct {
	client *kuma.Client
}

// NotificationPagerDutyResourceModel describes the resource data model.
type NotificationPagerDutyResourceModel struct {
	NotificationBaseModel

	IntegrationURL types.String `tfsdk:"integration_url"`
	IntegrationKey types.String `tfsdk:"integration_key"`
	Priority       types.String `tfsdk:"priority"`
	AutoResolve    types.String `tfsdk:"auto_resolve"`
}

// Metadata returns the metadata for the resource.
func (*NotificationPagerDutyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_pagerduty"
}

// Schema returns the schema for the resource.
func (*NotificationPagerDutyResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PagerDuty notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"integration_url": schema.StringAttribute{
				MarkdownDescription: "PagerDuty integration URL",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"integration_key": schema.StringAttribute{
				MarkdownDescription: "PagerDuty integration key",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"priority": schema.StringAttribute{
				MarkdownDescription: "PagerDuty incident priority",
				Optional:            true,
			},
			"auto_resolve": schema.StringAttribute{
				MarkdownDescription: "Auto-resolve incidents",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the PagerDuty notification resource with the API client.
func (r *NotificationPagerDutyResource) Configure(
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

// Create creates a new PagerDuty notification resource.
func (r *NotificationPagerDutyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationPagerDutyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pagerduty := notification.PagerDuty{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PagerDutyDetails: notification.PagerDutyDetails{
			IntegrationURL: data.IntegrationURL.ValueString(),
			IntegrationKey: data.IntegrationKey.ValueString(),
			Priority:       data.Priority.ValueString(),
			AutoResolve:    data.AutoResolve.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, pagerduty)
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

// Read reads the current state of the PagerDuty notification resource.
func (r *NotificationPagerDutyResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationPagerDutyResourceModel

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

	pagerduty := notification.PagerDuty{}
	err = base.As(&pagerduty)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "pagerduty"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(pagerduty.Name)
	data.IsActive = types.BoolValue(pagerduty.IsActive)
	data.IsDefault = types.BoolValue(pagerduty.IsDefault)
	data.ApplyExisting = types.BoolValue(pagerduty.ApplyExisting)

	data.IntegrationURL = types.StringValue(pagerduty.IntegrationURL)
	if pagerduty.IntegrationKey != "" {
		data.IntegrationKey = types.StringValue(pagerduty.IntegrationKey)
	}

	if pagerduty.Priority != "" {
		data.Priority = types.StringValue(pagerduty.Priority)
	}

	if pagerduty.AutoResolve != "" {
		data.AutoResolve = types.StringValue(pagerduty.AutoResolve)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the PagerDuty notification resource.
func (r *NotificationPagerDutyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationPagerDutyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pagerduty := notification.PagerDuty{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PagerDutyDetails: notification.PagerDutyDetails{
			IntegrationURL: data.IntegrationURL.ValueString(),
			IntegrationKey: data.IntegrationKey.ValueString(),
			Priority:       data.Priority.ValueString(),
			AutoResolve:    data.AutoResolve.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, pagerduty)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the PagerDuty notification resource.
func (r *NotificationPagerDutyResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationPagerDutyResourceModel

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
func (*NotificationPagerDutyResource) ImportState(
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
