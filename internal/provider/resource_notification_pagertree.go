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
	_ resource.Resource                = &NotificationPagerTreeResource{}
	_ resource.ResourceWithImportState = &NotificationPagerTreeResource{}
)

// NewNotificationPagerTreeResource returns a new instance of the PagerTree notification resource.
func NewNotificationPagerTreeResource() resource.Resource {
	return &NotificationPagerTreeResource{}
}

// NotificationPagerTreeResource defines the resource implementation.
type NotificationPagerTreeResource struct {
	client *kuma.Client
}

// NotificationPagerTreeResourceModel describes the resource data model.
type NotificationPagerTreeResourceModel struct {
	NotificationBaseModel

	IntegrationURL types.String `tfsdk:"integration_url"`
	Urgency        types.String `tfsdk:"urgency"`
	AutoResolve    types.String `tfsdk:"auto_resolve"`
}

// Metadata returns the metadata for the resource.
func (*NotificationPagerTreeResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_pagertree"
}

// Schema returns the schema for the resource.
func (*NotificationPagerTreeResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PagerTree notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"integration_url": schema.StringAttribute{
				MarkdownDescription: "PagerTree integration endpoint URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"urgency": schema.StringAttribute{
				MarkdownDescription: "Urgency level of the alert (e.g., high, medium, low)",
				Optional:            true,
			},
			"auto_resolve": schema.StringAttribute{
				MarkdownDescription: "Auto-resolve alerts (use 'resolve' to enable)",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the PagerTree notification resource with the API client.
func (r *NotificationPagerTreeResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new PagerTree notification resource.
func (r *NotificationPagerTreeResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationPagerTreeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pagertree := notification.PagerTree{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PagerTreeDetails: notification.PagerTreeDetails{
			IntegrationURL: data.IntegrationURL.ValueString(),
			Urgency:        data.Urgency.ValueString(),
			AutoResolve:    data.AutoResolve.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, pagertree)
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

// Read reads the current state of the PagerTree notification resource.
func (r *NotificationPagerTreeResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationPagerTreeResourceModel

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

	pagertree := notification.PagerTree{}
	err = base.As(&pagertree)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "pagertree"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(pagertree.Name)
	data.IsActive = types.BoolValue(pagertree.IsActive)
	data.IsDefault = types.BoolValue(pagertree.IsDefault)
	data.ApplyExisting = types.BoolValue(pagertree.ApplyExisting)

	if pagertree.IntegrationURL != "" {
		data.IntegrationURL = types.StringValue(pagertree.IntegrationURL)
	} else {
		data.IntegrationURL = types.StringNull()
	}

	if pagertree.Urgency != "" {
		data.Urgency = types.StringValue(pagertree.Urgency)
	} else {
		data.Urgency = types.StringNull()
	}

	if pagertree.AutoResolve != "" {
		data.AutoResolve = types.StringValue(pagertree.AutoResolve)
	} else {
		data.AutoResolve = types.StringNull()
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the PagerTree notification resource.
func (r *NotificationPagerTreeResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationPagerTreeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pagertree := notification.PagerTree{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PagerTreeDetails: notification.PagerTreeDetails{
			IntegrationURL: data.IntegrationURL.ValueString(),
			Urgency:        data.Urgency.ValueString(),
			AutoResolve:    data.AutoResolve.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, pagertree)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the PagerTree notification resource.
func (r *NotificationPagerTreeResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationPagerTreeResourceModel

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
func (*NotificationPagerTreeResource) ImportState(
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
