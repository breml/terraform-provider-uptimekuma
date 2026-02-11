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
	_ resource.Resource                = &NotificationHeiiOnCallResource{}
	_ resource.ResourceWithImportState = &NotificationHeiiOnCallResource{}
)

// NewNotificationHeiiOnCallResource returns a new instance of the Heii On-Call notification
// resource.
func NewNotificationHeiiOnCallResource() resource.Resource {
	return &NotificationHeiiOnCallResource{}
}

// NotificationHeiiOnCallResource defines the resource implementation.
type NotificationHeiiOnCallResource struct {
	client *kuma.Client
}

// NotificationHeiiOnCallResourceModel describes the resource data model.
type NotificationHeiiOnCallResourceModel struct {
	NotificationBaseModel

	APIKey    types.String `tfsdk:"api_key"`
	TriggerID types.String `tfsdk:"trigger_id"`
}

// Metadata returns the metadata for the resource.
func (*NotificationHeiiOnCallResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_heiioncall"
}

// Schema returns the schema for the resource.
func (*NotificationHeiiOnCallResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Heii On-Call notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Heii On-Call API key for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"trigger_id": schema.StringAttribute{
				MarkdownDescription: "Heii On-Call trigger ID",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Heii On-Call notification resource with the API client.
func (r *NotificationHeiiOnCallResource) Configure(
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

// Create creates a new Heii On-Call notification resource.
func (r *NotificationHeiiOnCallResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationHeiiOnCallResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	heiiOnCall := notification.HeiiOnCall{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		HeiiOnCallDetails: notification.HeiiOnCallDetails{
			APIKey:    data.APIKey.ValueString(),
			TriggerID: data.TriggerID.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, heiiOnCall)
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

// Read reads the current state of the Heii On-Call notification resource.
func (r *NotificationHeiiOnCallResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationHeiiOnCallResourceModel

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

	heiiOnCall := notification.HeiiOnCall{}
	err = base.As(&heiiOnCall)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			`failed to convert notification to type "heiioncall"`,
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(heiiOnCall.Name)
	data.IsActive = types.BoolValue(heiiOnCall.IsActive)
	data.IsDefault = types.BoolValue(heiiOnCall.IsDefault)
	data.ApplyExisting = types.BoolValue(heiiOnCall.ApplyExisting)

	data.APIKey = types.StringValue(heiiOnCall.APIKey)
	data.TriggerID = types.StringValue(heiiOnCall.TriggerID)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Heii On-Call notification resource.
func (r *NotificationHeiiOnCallResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationHeiiOnCallResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	heiiOnCall := notification.HeiiOnCall{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		HeiiOnCallDetails: notification.HeiiOnCallDetails{
			APIKey:    data.APIKey.ValueString(),
			TriggerID: data.TriggerID.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, heiiOnCall)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Heii On-Call notification resource.
func (r *NotificationHeiiOnCallResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationHeiiOnCallResourceModel

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
func (*NotificationHeiiOnCallResource) ImportState(
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
