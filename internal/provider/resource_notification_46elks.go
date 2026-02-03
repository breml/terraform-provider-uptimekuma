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
	_ resource.Resource                = &Notification46ElksResource{}
	_ resource.ResourceWithImportState = &Notification46ElksResource{}
)

// NewNotification46ElksResource returns a new instance of the 46elks notification resource.
func NewNotification46ElksResource() resource.Resource {
	return &Notification46ElksResource{}
}

// Notification46ElksResource defines the resource implementation.
type Notification46ElksResource struct {
	client *kuma.Client
}

// Notification46ElksResourceModel describes the resource data model.
type Notification46ElksResourceModel struct {
	NotificationBaseModel

	Username   types.String `tfsdk:"username"`
	AuthToken  types.String `tfsdk:"auth_token"`
	FromNumber types.String `tfsdk:"from_number"`
	ToNumber   types.String `tfsdk:"to_number"`
}

// Metadata returns the metadata for the resource.
func (*Notification46ElksResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_46elks"
}

// Schema returns the schema for the resource.
func (*Notification46ElksResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "46elks notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "46elks username",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "46elks authentication token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"from_number": schema.StringAttribute{
				MarkdownDescription: "46elks phone number to send from",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to_number": schema.StringAttribute{
				MarkdownDescription: "46elks phone number to send to",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the 46elks notification resource with the API client.
func (r *Notification46ElksResource) Configure(
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

// Create creates a new 46elks notification resource.
func (r *Notification46ElksResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data Notification46ElksResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	elks := notification.FortySixElks{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		FortySixElksDetails: notification.FortySixElksDetails{
			Username:   data.Username.ValueString(),
			AuthToken:  data.AuthToken.ValueString(),
			FromNumber: data.FromNumber.ValueString(),
			ToNumber:   data.ToNumber.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, elks)
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

// Read reads the current state of the 46elks notification resource.
func (r *Notification46ElksResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data Notification46ElksResourceModel

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

	elks := notification.FortySixElks{}
	err = base.As(&elks)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "46elks"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(elks.Name)
	data.IsActive = types.BoolValue(elks.IsActive)
	data.IsDefault = types.BoolValue(elks.IsDefault)
	data.ApplyExisting = types.BoolValue(elks.ApplyExisting)

	data.Username = types.StringValue(elks.Username)
	data.AuthToken = types.StringValue(elks.AuthToken)
	data.FromNumber = types.StringValue(elks.FromNumber)
	data.ToNumber = types.StringValue(elks.ToNumber)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the 46elks notification resource.
func (r *Notification46ElksResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data Notification46ElksResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	elks := notification.FortySixElks{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		FortySixElksDetails: notification.FortySixElksDetails{
			Username:   data.Username.ValueString(),
			AuthToken:  data.AuthToken.ValueString(),
			FromNumber: data.FromNumber.ValueString(),
			ToNumber:   data.ToNumber.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, elks)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the 46elks notification resource.
func (r *Notification46ElksResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data Notification46ElksResourceModel

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
func (*Notification46ElksResource) ImportState(
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
