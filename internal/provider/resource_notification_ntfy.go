package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationNtfyResource{}
	_ resource.ResourceWithImportState = &NotificationNtfyResource{}
)

func isValidURL(value string) bool {
	u, err := url.Parse(value)
	if err != nil {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}

type urlValidator struct{}

func (v urlValidator) Description(_ context.Context) string {
	return "string must be a valid URL with http:// or https:// scheme"
}

func (v urlValidator) MarkdownDescription(_ context.Context) string {
	return "string must be a valid URL with `http://` or `https://` scheme"
}

func (v urlValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if !isValidURL(value) {
		resp.Diagnostics.Append(
			diag.NewAttributeErrorDiagnostic(
				req.Path,
				"Invalid URL",
				fmt.Sprintf("Attribute must be a valid URL with http:// or https:// scheme, got: %s", value),
			),
		)
	}
}

func validateURL() validator.String {
	return urlValidator{}
}

func NewNotificationNtfyResource() resource.Resource {
	return &NotificationNtfyResource{}
}

// NotificationNtfyResource defines the resource implementation.
type NotificationNtfyResource struct {
	client *kuma.Client
}

// NotificationNtfyResourceModel describes the resource data model.
type NotificationNtfyResourceModel struct {
	NotificationBaseModel

	AccessToken          types.String `tfsdk:"access_token"`
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	Icon                 types.String `tfsdk:"icon"`
	Password             types.String `tfsdk:"password"`
	Priority             types.Int64  `tfsdk:"priority"`
	ServerURL            types.String `tfsdk:"server_url"`
	Topic                types.String `tfsdk:"topic"`
	Username             types.String `tfsdk:"username"`
}

func (r *NotificationNtfyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_ntfy"
}

func (r *NotificationNtfyResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				Optional: true,
			},
			"authentication_method": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("none"),
				Validators: []validator.String{
					stringvalidator.OneOf("none", "usernamePassword", "accessToken"),
				},
			},
			"icon": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"priority": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(5),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.AtMost(5),
				},
			},
			"server_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("https://ntfy.sh"),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validateURL(),
				},
			},
			"topic": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
		}),
	}
}

func (r *NotificationNtfyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
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

func (r *NotificationNtfyResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationNtfyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ntfy := notification.Ntfy{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		NtfyDetails: notification.NtfyDetails{
			AuthenticationMethod: data.AuthenticationMethod.ValueString(),
			Priority:             data.Priority.ValueInt64(),
			ServerURL:            data.ServerURL.ValueString(),
			Topic:                data.Topic.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, ntfy)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationNtfyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationNtfyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, err := r.client.GetNotification(ctx, id)
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	ntfy := notification.Ntfy{}
	err = base.As(&ntfy)
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "ntfy"`, err.Error())
		return
	}

	// Base properties
	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(ntfy.Name)
	data.IsActive = types.BoolValue(ntfy.IsActive)
	data.IsDefault = types.BoolValue(ntfy.IsDefault)
	data.ApplyExisting = types.BoolValue(ntfy.ApplyExisting)

	data.AuthenticationMethod = types.StringValue(ntfy.AuthenticationMethod)
	data.Priority = types.Int64Value(ntfy.Priority)
	data.ServerURL = types.StringValue(ntfy.ServerURL)
	data.Topic = types.StringValue(ntfy.Topic)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationNtfyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationNtfyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ntfy := notification.Ntfy{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		NtfyDetails: notification.NtfyDetails{
			AuthenticationMethod: data.AuthenticationMethod.ValueString(),
			Priority:             data.Priority.ValueInt64(),
			ServerURL:            data.ServerURL.ValueString(),
			Topic:                data.Topic.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, ntfy)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationNtfyResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationNtfyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}
}

func (r *NotificationNtfyResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
