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
	_ resource.Resource                = &NotificationSevenioResource{}
	_ resource.ResourceWithImportState = &NotificationSevenioResource{}
)

// NewNotificationSevenioResource returns a new instance of the Sevenio notification resource.
func NewNotificationSevenioResource() resource.Resource {
	return &NotificationSevenioResource{}
}

// NotificationSevenioResource defines the resource implementation.
type NotificationSevenioResource struct {
	client *kuma.Client
}

// NotificationSevenioResourceModel describes the resource data model.
type NotificationSevenioResourceModel struct {
	NotificationBaseModel

	APIKey types.String `tfsdk:"api_key"`
	Sender types.String `tfsdk:"sender"`
	To     types.String `tfsdk:"to"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSevenioResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_sevenio"
}

// Schema returns the schema for the resource.
func (*NotificationSevenioResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Seven.io SMS notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Seven.io API key for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"sender": schema.StringAttribute{
				MarkdownDescription: "Sender name or phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Sevenio notification resource with the API client.
func (r *NotificationSevenioResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Sevenio notification resource.
func (r *NotificationSevenioResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSevenioResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sevenio := notification.SevenIO{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SevenIODetails: notification.SevenIODetails{
			APIKey: data.APIKey.ValueString(),
			Sender: data.Sender.ValueString(),
			To:     data.To.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, sevenio)
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

// Read reads the current state of the Sevenio notification resource.
func (r *NotificationSevenioResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSevenioResourceModel

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

	sevenio := notification.SevenIO{}
	err = base.As(&sevenio)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "sevenio"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(sevenio.Name)
	data.IsActive = types.BoolValue(sevenio.IsActive)
	data.IsDefault = types.BoolValue(sevenio.IsDefault)
	data.ApplyExisting = types.BoolValue(sevenio.ApplyExisting)

	if sevenio.APIKey != "" {
		data.APIKey = types.StringValue(sevenio.APIKey)
	}

	data.Sender = types.StringValue(sevenio.Sender)
	data.To = types.StringValue(sevenio.To)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Sevenio notification resource.
func (r *NotificationSevenioResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSevenioResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sevenio := notification.SevenIO{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SevenIODetails: notification.SevenIODetails{
			APIKey: data.APIKey.ValueString(),
			Sender: data.Sender.ValueString(),
			To:     data.To.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, sevenio)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Sevenio notification resource.
func (r *NotificationSevenioResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSevenioResourceModel

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
func (*NotificationSevenioResource) ImportState(
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
