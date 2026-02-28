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
	_ resource.Resource                = &NotificationLunaseaResource{}
	_ resource.ResourceWithImportState = &NotificationLunaseaResource{}
)

// NewNotificationLunaseaResource returns a new instance of the Lunasea notification resource.
func NewNotificationLunaseaResource() resource.Resource {
	return &NotificationLunaseaResource{}
}

// NotificationLunaseaResource defines the resource implementation.
type NotificationLunaseaResource struct {
	client *kuma.Client
}

// NotificationLunaseaResourceModel describes the resource data model.
type NotificationLunaseaResourceModel struct {
	NotificationBaseModel

	Target        types.String `tfsdk:"target"`
	LunaSeaUserID types.String `tfsdk:"lunasea_user_id"`
	Device        types.String `tfsdk:"device"`
}

// Metadata returns the metadata for the resource.
func (*NotificationLunaseaResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_lunasea"
}

// Schema returns the schema for the resource.
func (*NotificationLunaseaResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lunasea notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"target": schema.StringAttribute{
				MarkdownDescription: "Target type: \"user\" or \"device\"",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("user", "device"),
				},
			},
			"lunasea_user_id": schema.StringAttribute{
				MarkdownDescription: "Lunasea user ID (required when target is \"user\")",
				Optional:            true,
				Sensitive:           true,
			},
			"device": schema.StringAttribute{
				MarkdownDescription: "Lunasea device ID (required when target is \"device\")",
				Optional:            true,
				Sensitive:           true,
			},
		}),
	}
}

// Configure configures the Lunasea notification resource with the API client.
func (r *NotificationLunaseaResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Lunasea notification resource.
func (r *NotificationLunaseaResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationLunaseaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	target := data.Target.ValueString()

	if target == "user" && data.LunaSeaUserID.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Field",
			"lunasea_user_id is required when target is 'user'",
		)
		return
	}

	if target == "device" && data.Device.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Field",
			"device is required when target is 'device'",
		)
		return
	}

	lunasea := notification.LunaSea{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		LunaSeaDetails: notification.LunaSeaDetails{
			Target:        target,
			LunaSeaUserID: data.LunaSeaUserID.ValueString(),
			Device:        data.Device.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, lunasea)
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

// Read reads the current state of the Lunasea notification resource.
func (r *NotificationLunaseaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationLunaseaResourceModel

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

	lunasea := notification.LunaSea{}
	err = base.As(&lunasea)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "lunasea"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(lunasea.Name)
	data.IsActive = types.BoolValue(lunasea.IsActive)
	data.IsDefault = types.BoolValue(lunasea.IsDefault)
	data.ApplyExisting = types.BoolValue(lunasea.ApplyExisting)

	data.Target = types.StringValue(lunasea.Target)
	if lunasea.LunaSeaUserID != "" {
		data.LunaSeaUserID = types.StringValue(lunasea.LunaSeaUserID)
	} else {
		data.LunaSeaUserID = types.StringNull()
	}

	if lunasea.Device != "" {
		data.Device = types.StringValue(lunasea.Device)
	} else {
		data.Device = types.StringNull()
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Lunasea notification resource.
func (r *NotificationLunaseaResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationLunaseaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	target := data.Target.ValueString()

	if target == "user" && data.LunaSeaUserID.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Field",
			"lunasea_user_id is required when target is 'user'",
		)
		return
	}

	if target == "device" && data.Device.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Field",
			"device is required when target is 'device'",
		)
		return
	}

	lunasea := notification.LunaSea{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		LunaSeaDetails: notification.LunaSeaDetails{
			Target:        target,
			LunaSeaUserID: data.LunaSeaUserID.ValueString(),
			Device:        data.Device.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, lunasea)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Lunasea notification resource.
func (r *NotificationLunaseaResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationLunaseaResourceModel

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
func (*NotificationLunaseaResource) ImportState(
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
