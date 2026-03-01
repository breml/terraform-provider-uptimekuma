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
	_ resource.Resource                = &NotificationAliyunsmsResource{}
	_ resource.ResourceWithImportState = &NotificationAliyunsmsResource{}
)

// NewNotificationAliyunsmsResource returns a new instance of the Aliyun SMS notification resource.
func NewNotificationAliyunsmsResource() resource.Resource {
	return &NotificationAliyunsmsResource{}
}

// NotificationAliyunsmsResource defines the resource implementation.
type NotificationAliyunsmsResource struct {
	client *kuma.Client
}

// NotificationAliyunsmsResourceModel describes the resource data model.
type NotificationAliyunsmsResourceModel struct {
	NotificationBaseModel

	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	PhoneNumber     types.String `tfsdk:"phone_number"`
	SignName        types.String `tfsdk:"sign_name"`
	TemplateCode    types.String `tfsdk:"template_code"`
}

// Metadata returns the metadata for the resource.
func (*NotificationAliyunsmsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_aliyunsms"
}

// Schema returns the schema for the resource.
func (*NotificationAliyunsmsResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Aliyun SMS notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"access_key_id": schema.StringAttribute{
				MarkdownDescription: "Aliyun access key ID",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"secret_access_key": schema.StringAttribute{
				MarkdownDescription: "Aliyun secret access key",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"phone_number": schema.StringAttribute{
				MarkdownDescription: "Phone number to send SMS to",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"sign_name": schema.StringAttribute{
				MarkdownDescription: "Aliyun SMS sign name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"template_code": schema.StringAttribute{
				MarkdownDescription: "Aliyun SMS template code",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Aliyun SMS notification resource with the API client.
func (r *NotificationAliyunsmsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Aliyun SMS notification resource.
func (r *NotificationAliyunsmsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationAliyunsmsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	aliyunsms := notification.AliyunSMS{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		AliyunSMSDetails: notification.AliyunSMSDetails{
			AccessKeyID:     data.AccessKeyID.ValueString(),
			SecretAccessKey: data.SecretAccessKey.ValueString(),
			PhoneNumber:     data.PhoneNumber.ValueString(),
			SignName:        data.SignName.ValueString(),
			TemplateCode:    data.TemplateCode.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, aliyunsms)
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

// Read reads the current state of the Aliyun SMS notification resource.
func (r *NotificationAliyunsmsResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationAliyunsmsResourceModel

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

	aliyunsms := notification.AliyunSMS{}
	err = base.As(&aliyunsms)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "AliyunSMS"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(aliyunsms.Name)
	data.IsActive = types.BoolValue(aliyunsms.IsActive)
	data.IsDefault = types.BoolValue(aliyunsms.IsDefault)
	data.ApplyExisting = types.BoolValue(aliyunsms.ApplyExisting)

	data.AccessKeyID = types.StringValue(aliyunsms.AccessKeyID)
	data.SecretAccessKey = types.StringValue(aliyunsms.SecretAccessKey)
	data.PhoneNumber = types.StringValue(aliyunsms.PhoneNumber)
	data.SignName = types.StringValue(aliyunsms.SignName)
	data.TemplateCode = types.StringValue(aliyunsms.TemplateCode)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Aliyun SMS notification resource.
func (r *NotificationAliyunsmsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationAliyunsmsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	aliyunsms := notification.AliyunSMS{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		AliyunSMSDetails: notification.AliyunSMSDetails{
			AccessKeyID:     data.AccessKeyID.ValueString(),
			SecretAccessKey: data.SecretAccessKey.ValueString(),
			PhoneNumber:     data.PhoneNumber.ValueString(),
			SignName:        data.SignName.ValueString(),
			TemplateCode:    data.TemplateCode.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, aliyunsms)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Aliyun SMS notification resource.
func (r *NotificationAliyunsmsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationAliyunsmsResourceModel

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
func (*NotificationAliyunsmsResource) ImportState(
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
