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
	_ resource.Resource                = &NotificationMatrixResource{}
	_ resource.ResourceWithImportState = &NotificationMatrixResource{}
)

// NewNotificationMatrixResource returns a new instance of the Matrix notification resource.
func NewNotificationMatrixResource() resource.Resource {
	return &NotificationMatrixResource{}
}

// NotificationMatrixResource defines the resource implementation.
type NotificationMatrixResource struct {
	client *kuma.Client
}

// NotificationMatrixResourceModel describes the resource data model.
type NotificationMatrixResourceModel struct {
	NotificationBaseModel

	HomeserverURL  types.String `tfsdk:"homeserver_url"`
	InternalRoomID types.String `tfsdk:"internal_room_id"`
	AccessToken    types.String `tfsdk:"access_token"`
}

// Metadata returns the metadata for the resource.
func (*NotificationMatrixResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_matrix"
}

// Schema returns the schema for the resource.
func (*NotificationMatrixResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Matrix notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"homeserver_url": schema.StringAttribute{
				MarkdownDescription: "Matrix homeserver URL",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"internal_room_id": schema.StringAttribute{
				MarkdownDescription: "Matrix internal room ID",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Matrix access token for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Matrix notification resource with the API client.
func (r *NotificationMatrixResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Matrix notification resource.
func (r *NotificationMatrixResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationMatrixResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	matrix := notification.Matrix{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		MatrixDetails: notification.MatrixDetails{
			HomeserverURL:  data.HomeserverURL.ValueString(),
			InternalRoomID: data.InternalRoomID.ValueString(),
			AccessToken:    data.AccessToken.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, matrix)
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

// Read reads the current state of the Matrix notification resource.
func (r *NotificationMatrixResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationMatrixResourceModel

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

	matrix := notification.Matrix{}
	err = base.As(&matrix)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "matrix"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(matrix.Name)
	data.IsActive = types.BoolValue(matrix.IsActive)
	data.IsDefault = types.BoolValue(matrix.IsDefault)
	data.ApplyExisting = types.BoolValue(matrix.ApplyExisting)

	data.HomeserverURL = types.StringValue(matrix.HomeserverURL)
	data.InternalRoomID = types.StringValue(matrix.InternalRoomID)
	data.AccessToken = types.StringValue(matrix.AccessToken)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Matrix notification resource.
func (r *NotificationMatrixResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationMatrixResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	matrix := notification.Matrix{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		MatrixDetails: notification.MatrixDetails{
			HomeserverURL:  data.HomeserverURL.ValueString(),
			InternalRoomID: data.InternalRoomID.ValueString(),
			AccessToken:    data.AccessToken.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, matrix)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Matrix notification resource.
func (r *NotificationMatrixResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationMatrixResourceModel

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
func (*NotificationMatrixResource) ImportState(
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
