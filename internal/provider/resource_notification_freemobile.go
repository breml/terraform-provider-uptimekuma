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
	_ resource.Resource                = &NotificationFreemobileResource{}
	_ resource.ResourceWithImportState = &NotificationFreemobileResource{}
)

// NewNotificationFreemobileResource returns a new instance of the Freemobile notification resource.
func NewNotificationFreemobileResource() resource.Resource {
	return &NotificationFreemobileResource{}
}

// NotificationFreemobileResource defines the resource implementation.
type NotificationFreemobileResource struct {
	client *kuma.Client
}

// NotificationFreemobileResourceModel describes the resource data model.
type NotificationFreemobileResourceModel struct {
	NotificationBaseModel

	User types.String `tfsdk:"user"`
	Pass types.String `tfsdk:"pass"`
}

// Metadata returns the metadata for the resource.
func (*NotificationFreemobileResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_freemobile"
}

// Schema returns the schema for the resource.
func (*NotificationFreemobileResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Free Mobile notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"user": schema.StringAttribute{
				MarkdownDescription: "Free Mobile user ID",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"pass": schema.StringAttribute{
				MarkdownDescription: "Free Mobile API key",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Free Mobile notification resource with the API client.
func (r *NotificationFreemobileResource) Configure(
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

// Create creates a new Free Mobile notification resource.
func (r *NotificationFreemobileResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationFreemobileResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	freemobile := notification.FreeMobile{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		FreeMobileDetails: notification.FreeMobileDetails{
			User: data.User.ValueString(),
			Pass: data.Pass.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, freemobile)
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

// Read reads the current state of the Free Mobile notification resource.
func (r *NotificationFreemobileResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationFreemobileResourceModel

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

	freemobile := notification.FreeMobile{}
	err = base.As(&freemobile)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "freemobile"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(freemobile.Name)
	data.IsActive = types.BoolValue(freemobile.IsActive)
	data.IsDefault = types.BoolValue(freemobile.IsDefault)
	data.ApplyExisting = types.BoolValue(freemobile.ApplyExisting)

	data.User = types.StringValue(freemobile.User)
	data.Pass = types.StringValue(freemobile.Pass)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Free Mobile notification resource.
func (r *NotificationFreemobileResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationFreemobileResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	freemobile := notification.FreeMobile{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		FreeMobileDetails: notification.FreeMobileDetails{
			User: data.User.ValueString(),
			Pass: data.Pass.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, freemobile)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Free Mobile notification resource.
func (r *NotificationFreemobileResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationFreemobileResourceModel

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
func (*NotificationFreemobileResource) ImportState(
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
