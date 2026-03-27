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
	_ resource.Resource                = &NotificationSerwersmsResource{}
	_ resource.ResourceWithImportState = &NotificationSerwersmsResource{}
)

// NewNotificationSerwersmsResource returns a new instance of the SerwerSMS notification resource.
func NewNotificationSerwersmsResource() resource.Resource {
	return &NotificationSerwersmsResource{}
}

// NotificationSerwersmsResource defines the resource implementation.
type NotificationSerwersmsResource struct {
	client *kuma.Client
}

// NotificationSerwersmsResourceModel describes the resource data model.
type NotificationSerwersmsResourceModel struct {
	NotificationBaseModel

	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	PhoneNumber types.String `tfsdk:"phone_number"`
	SenderName  types.String `tfsdk:"sender_name"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSerwersmsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_serwersms"
}

// Schema returns the schema for the resource.
func (*NotificationSerwersmsResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SerwerSMS notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "SerwerSMS account username",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "SerwerSMS account password",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"phone_number": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"sender_name": schema.StringAttribute{
				MarkdownDescription: "Sender name or identifier",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the SerwerSMS notification resource with the API client.
func (r *NotificationSerwersmsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new SerwerSMS notification resource.
func (r *NotificationSerwersmsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSerwersmsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serwersms := notification.SerwerSMS{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SerwerSMSDetails: notification.SerwerSMSDetails{
			Username:    data.Username.ValueString(),
			Password:    data.Password.ValueString(),
			PhoneNumber: data.PhoneNumber.ValueString(),
			SenderName:  data.SenderName.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, serwersms)
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

// Read reads the current state of the SerwerSMS notification resource.
func (r *NotificationSerwersmsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSerwersmsResourceModel

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

	serwersms := notification.SerwerSMS{}
	err = base.As(&serwersms)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "serwersms"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(serwersms.Name)
	data.IsActive = types.BoolValue(serwersms.IsActive)
	data.IsDefault = types.BoolValue(serwersms.IsDefault)
	data.ApplyExisting = types.BoolValue(serwersms.ApplyExisting)

	data.Username = types.StringValue(serwersms.Username)
	data.Password = types.StringValue(serwersms.Password)
	data.PhoneNumber = types.StringValue(serwersms.PhoneNumber)

	if serwersms.SenderName != "" {
		data.SenderName = types.StringValue(serwersms.SenderName)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SerwerSMS notification resource.
func (r *NotificationSerwersmsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSerwersmsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serwersms := notification.SerwerSMS{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SerwerSMSDetails: notification.SerwerSMSDetails{
			Username:    data.Username.ValueString(),
			Password:    data.Password.ValueString(),
			PhoneNumber: data.PhoneNumber.ValueString(),
			SenderName:  data.SenderName.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, serwersms)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SerwerSMS notification resource.
func (r *NotificationSerwersmsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSerwersmsResourceModel

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
func (*NotificationSerwersmsResource) ImportState(
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
