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
	_ resource.Resource                = &NotificationEvolutionResource{}
	_ resource.ResourceWithImportState = &NotificationEvolutionResource{}
)

// NewNotificationEvolutionResource returns a new instance of the Evolution notification resource.
func NewNotificationEvolutionResource() resource.Resource {
	return &NotificationEvolutionResource{}
}

// NotificationEvolutionResource defines the resource implementation.
type NotificationEvolutionResource struct {
	client *kuma.Client
}

// NotificationEvolutionResourceModel describes the resource data model.
type NotificationEvolutionResourceModel struct {
	NotificationBaseModel

	APIURL       types.String `tfsdk:"api_url"`
	InstanceName types.String `tfsdk:"instance_name"`
	AuthToken    types.String `tfsdk:"auth_token"`
	Recipient    types.String `tfsdk:"recipient"`
}

// Metadata returns the metadata for the resource.
func (*NotificationEvolutionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_evolution"
}

// Schema returns the schema for the resource.
func (*NotificationEvolutionResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Evolution API notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Evolution API URL endpoint",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"instance_name": schema.StringAttribute{
				MarkdownDescription: "Evolution API instance name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "Evolution API authentication token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"recipient": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number for WhatsApp messages",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Evolution notification resource with the API client.
func (r *NotificationEvolutionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Evolution notification resource.
func (r *NotificationEvolutionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationEvolutionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	evolution := notification.Evolution{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		EvolutionDetails: notification.EvolutionDetails{
			APIURL:       data.APIURL.ValueString(),
			InstanceName: data.InstanceName.ValueString(),
			AuthToken:    data.AuthToken.ValueString(),
			Recipient:    data.Recipient.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, evolution)
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

// Read reads the current state of the Evolution notification resource.
func (r *NotificationEvolutionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationEvolutionResourceModel

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

	evolution := notification.Evolution{}
	err = base.As(&evolution)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "EvolutionApi"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(evolution.Name)
	data.IsActive = types.BoolValue(evolution.IsActive)
	data.IsDefault = types.BoolValue(evolution.IsDefault)
	data.ApplyExisting = types.BoolValue(evolution.ApplyExisting)

	data.APIURL = types.StringValue(evolution.APIURL)
	data.InstanceName = types.StringValue(evolution.InstanceName)
	data.AuthToken = types.StringValue(evolution.AuthToken)
	data.Recipient = types.StringValue(evolution.Recipient)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Evolution notification resource.
func (r *NotificationEvolutionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationEvolutionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	evolution := notification.Evolution{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		EvolutionDetails: notification.EvolutionDetails{
			APIURL:       data.APIURL.ValueString(),
			InstanceName: data.InstanceName.ValueString(),
			AuthToken:    data.AuthToken.ValueString(),
			Recipient:    data.Recipient.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, evolution)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Evolution notification resource.
func (r *NotificationEvolutionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationEvolutionResourceModel

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
func (*NotificationEvolutionResource) ImportState(
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
