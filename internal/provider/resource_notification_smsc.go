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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationSMSCResource{}
	_ resource.ResourceWithImportState = &NotificationSMSCResource{}
)

// NewNotificationSMSCResource returns a new instance of the SMSC notification resource.
func NewNotificationSMSCResource() resource.Resource {
	return &NotificationSMSCResource{}
}

// NotificationSMSCResource defines the resource implementation.
type NotificationSMSCResource struct {
	client *kuma.Client
}

// NotificationSMSCResourceModel describes the resource data model.
type NotificationSMSCResourceModel struct {
	NotificationBaseModel

	Login      types.String `tfsdk:"login"`
	Password   types.String `tfsdk:"password"`
	ToNumber   types.String `tfsdk:"to_number"`
	SenderName types.String `tfsdk:"sender_name"`
	Translit   types.String `tfsdk:"translit"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSMSCResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_smsc"
}

// Schema returns the schema for the resource.
func (*NotificationSMSCResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SMSC SMS notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"login": schema.StringAttribute{
				MarkdownDescription: "SMSC account login/username",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "SMSC account password",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to_number": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"sender_name": schema.StringAttribute{
				MarkdownDescription: "Sender name or identifier",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"translit": schema.StringAttribute{
				MarkdownDescription: "Transliterate non-ASCII characters (0 = disabled, 1 = enabled)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("0"),
				Validators: []validator.String{
					stringvalidator.OneOf("0", "1"),
				},
			},
		}),
	}
}

// Configure configures the SMSC notification resource with the API client.
func (r *NotificationSMSCResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new SMSC notification resource.
func (r *NotificationSMSCResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationSMSCResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smsc := notification.SMSC{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMSCDetails: notification.SMSCDetails{
			Login:      data.Login.ValueString(),
			Password:   data.Password.ValueString(),
			ToNumber:   data.ToNumber.ValueString(),
			SenderName: data.SenderName.ValueString(),
			Translit:   data.Translit.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, smsc)
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

// Read reads the current state of the SMSC notification resource.
func (r *NotificationSMSCResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationSMSCResourceModel

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

	smsc := notification.SMSC{}
	err = base.As(&smsc)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "smsc"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(smsc.Name)
	data.IsActive = types.BoolValue(smsc.IsActive)
	data.IsDefault = types.BoolValue(smsc.IsDefault)
	data.ApplyExisting = types.BoolValue(smsc.ApplyExisting)

	data.Login = types.StringValue(smsc.Login)
	data.Password = types.StringValue(smsc.Password)
	data.ToNumber = types.StringValue(smsc.ToNumber)
	data.SenderName = types.StringValue(smsc.SenderName)
	data.Translit = types.StringValue(smsc.Translit)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SMSC notification resource.
func (r *NotificationSMSCResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSMSCResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	smsc := notification.SMSC{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SMSCDetails: notification.SMSCDetails{
			Login:      data.Login.ValueString(),
			Password:   data.Password.ValueString(),
			ToNumber:   data.ToNumber.ValueString(),
			SenderName: data.SenderName.ValueString(),
			Translit:   data.Translit.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, smsc)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SMSC notification resource.
func (r *NotificationSMSCResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSMSCResourceModel

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
func (*NotificationSMSCResource) ImportState(
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
