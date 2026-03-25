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
	_ resource.Resource                = &NotificationPromoSMSResource{}
	_ resource.ResourceWithImportState = &NotificationPromoSMSResource{}
)

// NewNotificationPromoSMSResource returns a new instance of the PromoSMS notification resource.
func NewNotificationPromoSMSResource() resource.Resource {
	return &NotificationPromoSMSResource{}
}

// NotificationPromoSMSResource defines the resource implementation.
type NotificationPromoSMSResource struct {
	client *kuma.Client
}

// NotificationPromoSMSResourceModel describes the resource data model.
type NotificationPromoSMSResourceModel struct {
	NotificationBaseModel

	Login        types.String `tfsdk:"login"`
	Password     types.String `tfsdk:"password"`
	PhoneNumber  types.String `tfsdk:"phone_number"`
	SenderName   types.String `tfsdk:"sender_name"`
	SMSType      types.String `tfsdk:"sms_type"`
	AllowLongSMS types.Bool   `tfsdk:"allow_long_sms"`
}

// Metadata returns the metadata for the resource.
func (*NotificationPromoSMSResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_promosms"
}

// Schema returns the schema for the resource.
func (*NotificationPromoSMSResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PromoSMS notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"login": schema.StringAttribute{
				MarkdownDescription: "PromoSMS login/username for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "PromoSMS password for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"phone_number": schema.StringAttribute{
				MarkdownDescription: "Recipient phone number for SMS messages",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"sender_name": schema.StringAttribute{
				MarkdownDescription: "Sender name/ID displayed in the SMS",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"sms_type": schema.StringAttribute{
				MarkdownDescription: "SMS type identifier",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"allow_long_sms": schema.BoolAttribute{
				MarkdownDescription: "Allow long SMS messages (up to 639 characters instead of 159)",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		}),
	}
}

// Configure configures the PromoSMS notification resource with the API client.
func (r *NotificationPromoSMSResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new PromoSMS notification resource.
func (r *NotificationPromoSMSResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationPromoSMSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	promosms := notification.PromoSMS{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PromoSMSDetails: notification.PromoSMSDetails{
			Login:        data.Login.ValueString(),
			Password:     data.Password.ValueString(),
			PhoneNumber:  data.PhoneNumber.ValueString(),
			SenderName:   data.SenderName.ValueString(),
			SMSType:      data.SMSType.ValueString(),
			AllowLongSMS: data.AllowLongSMS.ValueBool(),
		},
	}

	id, err := r.client.CreateNotification(ctx, promosms)
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

// Read reads the current state of the PromoSMS notification resource.
func (r *NotificationPromoSMSResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationPromoSMSResourceModel

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

	promosms := notification.PromoSMS{}
	err = base.As(&promosms)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "promosms"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(promosms.Name)
	data.IsActive = types.BoolValue(promosms.IsActive)
	data.IsDefault = types.BoolValue(promosms.IsDefault)
	data.ApplyExisting = types.BoolValue(promosms.ApplyExisting)

	data.Login = types.StringValue(promosms.Login)
	data.Password = types.StringValue(promosms.Password)
	data.PhoneNumber = types.StringValue(promosms.PhoneNumber)
	data.SenderName = types.StringValue(promosms.SenderName)
	data.SMSType = types.StringValue(promosms.SMSType)
	data.AllowLongSMS = types.BoolValue(promosms.AllowLongSMS)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the PromoSMS notification resource.
func (r *NotificationPromoSMSResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationPromoSMSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	promosms := notification.PromoSMS{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PromoSMSDetails: notification.PromoSMSDetails{
			Login:        data.Login.ValueString(),
			Password:     data.Password.ValueString(),
			PhoneNumber:  data.PhoneNumber.ValueString(),
			SenderName:   data.SenderName.ValueString(),
			SMSType:      data.SMSType.ValueString(),
			AllowLongSMS: data.AllowLongSMS.ValueBool(),
		},
	}

	err := r.client.UpdateNotification(ctx, promosms)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the PromoSMS notification resource.
func (r *NotificationPromoSMSResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationPromoSMSResourceModel

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
func (*NotificationPromoSMSResource) ImportState(
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
