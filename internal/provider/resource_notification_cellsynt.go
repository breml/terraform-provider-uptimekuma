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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationCellsyntResource{}
	_ resource.ResourceWithImportState = &NotificationCellsyntResource{}
)

// NewNotificationCellsyntResource returns a new instance of the Cellsynt notification resource.
func NewNotificationCellsyntResource() resource.Resource {
	return &NotificationCellsyntResource{}
}

// NotificationCellsyntResource defines the resource implementation.
type NotificationCellsyntResource struct {
	client *kuma.Client
}

// NotificationCellsyntResourceModel describes the resource data model.
type NotificationCellsyntResourceModel struct {
	NotificationBaseModel

	Login          types.String `tfsdk:"login"`
	Password       types.String `tfsdk:"password"`
	Destination    types.String `tfsdk:"destination"`
	Originator     types.String `tfsdk:"originator"`
	OriginatorType types.String `tfsdk:"originator_type"`
	AllowLongSMS   types.Bool   `tfsdk:"allow_long_sms"`
}

// Metadata returns the metadata for the resource.
func (*NotificationCellsyntResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_cellsynt"
}

// Schema returns the schema for the resource.
func (*NotificationCellsyntResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Cellsynt SMS notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"login": schema.StringAttribute{
				MarkdownDescription: "Cellsynt account username",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Cellsynt account password",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"destination": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"originator": schema.StringAttribute{
				MarkdownDescription: "Sender name or phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"originator_type": schema.StringAttribute{
				MarkdownDescription: "Type of originator (Numeric or Alphanumeric)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Numeric"),
				Validators: []validator.String{
					stringvalidator.OneOf("Numeric", "Alphanumeric"),
				},
			},
			"allow_long_sms": schema.BoolAttribute{
				MarkdownDescription: "Allow sending SMS messages longer than 160 characters",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		}),
	}
}

// Configure configures the Cellsynt notification resource with the API client.
func (r *NotificationCellsyntResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Cellsynt notification resource.
func (r *NotificationCellsyntResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationCellsyntResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cellsynt := notification.Cellsynt{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		CellsyntDetails: notification.CellsyntDetails{
			Login:          data.Login.ValueString(),
			Password:       data.Password.ValueString(),
			Destination:    data.Destination.ValueString(),
			Originator:     data.Originator.ValueString(),
			OriginatorType: data.OriginatorType.ValueString(),
			AllowLongSMS:   data.AllowLongSMS.ValueBool(),
		},
	}

	id, err := r.client.CreateNotification(ctx, cellsynt)
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

// Read reads the current state of the Cellsynt notification resource.
func (r *NotificationCellsyntResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationCellsyntResourceModel

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

	cellsynt := notification.Cellsynt{}
	err = base.As(&cellsynt)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "cellsynt"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(cellsynt.Name)
	data.IsActive = types.BoolValue(cellsynt.IsActive)
	data.IsDefault = types.BoolValue(cellsynt.IsDefault)
	data.ApplyExisting = types.BoolValue(cellsynt.ApplyExisting)

	data.Login = types.StringValue(cellsynt.Login)
	data.Password = types.StringValue(cellsynt.Password)
	data.Destination = types.StringValue(cellsynt.Destination)
	data.Originator = types.StringValue(cellsynt.Originator)
	data.OriginatorType = types.StringValue(cellsynt.OriginatorType)
	data.AllowLongSMS = types.BoolValue(cellsynt.AllowLongSMS)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Cellsynt notification resource.
func (r *NotificationCellsyntResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationCellsyntResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cellsynt := notification.Cellsynt{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		CellsyntDetails: notification.CellsyntDetails{
			Login:          data.Login.ValueString(),
			Password:       data.Password.ValueString(),
			Destination:    data.Destination.ValueString(),
			Originator:     data.Originator.ValueString(),
			OriginatorType: data.OriginatorType.ValueString(),
			AllowLongSMS:   data.AllowLongSMS.ValueBool(),
		},
	}

	err := r.client.UpdateNotification(ctx, cellsynt)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Cellsynt notification resource.
func (r *NotificationCellsyntResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationCellsyntResourceModel

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
func (*NotificationCellsyntResource) ImportState(
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
