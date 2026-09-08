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
	_ resource.Resource                   = &NotificationNtfyResource{}
	_ resource.ResourceWithImportState    = &NotificationNtfyResource{}
	_ resource.ResourceWithValidateConfig = &NotificationNtfyResource{}
)

// ntfyAuthenticationMethodDefault is the schema default of the authentication_method attribute.
const ntfyAuthenticationMethodDefault = "none"

func isValidURL(value string) bool {
	u, err := url.Parse(value)
	if err != nil {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}

type urlValidator struct{}

// Description returns a plain text description of the validator's behavior.
func (urlValidator) Description(_ context.Context) string {
	return "string must be a valid URL with http:// or https:// scheme"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (urlValidator) MarkdownDescription(_ context.Context) string {
	return "string must be a valid URL with `http://` or `https://` scheme"
}

// ValidateString checks that the provided string value is a valid URL.
func (urlValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
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

// NewNotificationNtfyResource returns a new instance of the ntfy notification resource.
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
	Call                 types.String `tfsdk:"call"`
	CustomMessage        types.String `tfsdk:"custom_message"`
	CustomTitle          types.String `tfsdk:"custom_title"`
	Icon                 types.String `tfsdk:"icon"`
	Password             types.String `tfsdk:"password"`
	Priority             types.Int64  `tfsdk:"priority"`
	PriorityDown         types.Int64  `tfsdk:"priority_down"`
	ServerURL            types.String `tfsdk:"server_url"`
	Topic                types.String `tfsdk:"topic"`
	UseTemplate          types.Bool   `tfsdk:"use_template"`
	Username             types.String `tfsdk:"username"`
}

// Metadata returns the metadata for the resource.
func (*NotificationNtfyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_ntfy"
}

// Schema returns the schema for the resource.
func (*NotificationNtfyResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "ntfy notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				MarkdownDescription: "ntfy access token, used when `authentication_method` is `accessToken`.",
				Optional:            true,
				Sensitive:           true,
			},
			"authentication_method": schema.StringAttribute{
				MarkdownDescription: "Authentication method used to talk to the ntfy server. " +
					"One of `none`, `usernamePassword` or `accessToken`. Defaults to `none`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(ntfyAuthenticationMethodDefault),
				Validators: []validator.String{
					stringvalidator.OneOf("none", "usernamePassword", "accessToken"),
				},
			},
			"call": schema.StringAttribute{
				MarkdownDescription: "Phone number to call for every notification, or `yes` to call the " +
					"account's verified number (ntfy `X-Call` header).",
				Optional: true,
			},
			"custom_message": schema.StringAttribute{
				MarkdownDescription: "Custom message template. Only used if `use_template` is `true`.",
				Optional:            true,
			},
			"custom_title": schema.StringAttribute{
				MarkdownDescription: "Custom title template. Only used if `use_template` is `true`.",
				Optional:            true,
			},
			"icon": schema.StringAttribute{
				MarkdownDescription: "URL of the icon attached to the notification.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "ntfy password, used when `authentication_method` is `usernamePassword`.",
				Optional:            true,
				Sensitive:           true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Priority of the notification, between 1 and 5. Used for all " +
					"notifications and as the basis for the down priority when `priority_down` is not set. " +
					"Defaults to `5`.",
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(5),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.AtMost(5),
				},
			},
			"priority_down": schema.Int64Attribute{
				MarkdownDescription: "Priority of the notification sent when a monitor goes down, between 1 and 5. " +
					"If not set, Uptime Kuma uses `priority` + 1, capped at 5.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.AtMost(5),
				},
			},
			"server_url": schema.StringAttribute{
				MarkdownDescription: "URL of the ntfy server. Defaults to `https://ntfy.sh`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("https://ntfy.sh"),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validateURL(),
				},
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "ntfy topic the notifications are published to.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"use_template": schema.BoolAttribute{
				MarkdownDescription: "Use `custom_title` and `custom_message` instead of the default " +
					"notification title and message.",
				Optional: true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "ntfy username, used when `authentication_method` is `usernamePassword`.",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the ntfy notification resource with the API client.
func (r *NotificationNtfyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// ValidateConfig validates the configuration for the ntfy notification resource.
//
// buildNtfyDetails sends every credential to Uptime Kuma regardless of the selected
// authentication method, and Uptime Kuma then uses only the credentials belonging to that
// method. Credentials for the other methods therefore have no effect at all, which is
// reported here instead of being silently ignored.
func (*NotificationNtfyResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config NotificationNtfyResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	validateNtfyAuthentication(&config, resp)
}

// validateNtfyAuthentication reports the credentials that do not match the configured
// authentication method.
func validateNtfyAuthentication(config *NotificationNtfyResourceModel, resp *resource.ValidateConfigResponse) {
	if config.AuthenticationMethod.IsUnknown() {
		// The value is only known after apply, so there is nothing to validate against.
		return
	}

	// Schema defaults are not applied to the configuration during validation, so a null value
	// here is the schema default and has to be resolved by hand.
	authenticationMethod := ntfyAuthenticationMethodDefault
	if !config.AuthenticationMethod.IsNull() {
		authenticationMethod = config.AuthenticationMethod.ValueString()
	}

	switch authenticationMethod {
	case "accessToken":
		validateNtfyFieldSet(resp, config.AccessToken, path.Root("access_token"), "accessToken")
		validateNtfyFieldNotSet(resp, config.Username, path.Root("username"), "accessToken")
		validateNtfyFieldNotSet(resp, config.Password, path.Root("password"), "accessToken")

	case "usernamePassword":
		validateNtfyFieldSet(resp, config.Username, path.Root("username"), "usernamePassword")
		validateNtfyFieldSet(resp, config.Password, path.Root("password"), "usernamePassword")
		validateNtfyFieldNotSet(resp, config.AccessToken, path.Root("access_token"), "usernamePassword")

	case "none":
		validateNtfyFieldNotSet(resp, config.AccessToken, path.Root("access_token"), "none")
		validateNtfyFieldNotSet(resp, config.Username, path.Root("username"), "none")
		validateNtfyFieldNotSet(resp, config.Password, path.Root("password"), "none")

	default:
		// If the authentication method has an unexpected value, assume other validators handle it.
		return
	}
}

// validateNtfyFieldSet reports an error if field is not set for the given authentication method.
//
// An empty string counts as not set. An unknown value is accepted, because its content is only
// available after apply.
func validateNtfyFieldSet(
	resp *resource.ValidateConfigResponse,
	field types.String,
	fieldPath path.Path,
	authenticationMethod string,
) {
	if field.IsUnknown() {
		return
	}

	if field.IsNull() || field.ValueString() == "" {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			fieldPath,
			"Missing Attribute for ntfy Authentication Method",
			fmt.Sprintf(
				"With %q set to %q, the attribute %q must be set to a non-empty value.",
				"authentication_method", authenticationMethod, fieldPath.String(),
			),
		))
	}
}

// validateNtfyFieldNotSet reports an error if field is set for the given authentication method.
//
// Any non-null value counts as set, including an empty string and a value that is only known
// after apply: the attribute is present in the configuration either way.
func validateNtfyFieldNotSet(
	resp *resource.ValidateConfigResponse,
	field types.String,
	fieldPath path.Path,
	authenticationMethod string,
) {
	if !field.IsNull() {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			fieldPath,
			"Invalid Attribute for ntfy Authentication Method",
			fmt.Sprintf(
				"With %q set to %q, the attribute %q must not be set. Either remove it or change "+
					"%q to the method the credential belongs to.",
				"authentication_method", authenticationMethod, fieldPath.String(), "authentication_method",
			),
		))
	}
}

// buildNtfyDetails builds the ntfy specific part of the notification from the resource model.
//
// All credentials are sent unconditionally; Uptime Kuma picks the ones belonging to the
// configured authentication method. A null priority_down becomes 0, which relies on the
// omitempty tag of notification.NtfyDetails.PriorityDown to keep it out of the payload.
func buildNtfyDetails(data *NotificationNtfyResourceModel) notification.NtfyDetails {
	return notification.NtfyDetails{
		AccessToken:          data.AccessToken.ValueString(),
		AuthenticationMethod: data.AuthenticationMethod.ValueString(),
		Call:                 strToPtr(data.Call),
		CustomMessage:        strToPtr(data.CustomMessage),
		CustomTitle:          strToPtr(data.CustomTitle),
		Icon:                 data.Icon.ValueString(),
		Password:             data.Password.ValueString(),
		Priority:             data.Priority.ValueInt64(),
		PriorityDown:         data.PriorityDown.ValueInt64(),
		ServerURL:            data.ServerURL.ValueString(),
		Topic:                data.Topic.ValueString(),
		UseTemplate:          boolToPtr(data.UseTemplate),
		Username:             data.Username.ValueString(),
	}
}

// updateNtfyFields populates the resource model with the values returned by Uptime Kuma.
//
// The credentials are set to null if Uptime Kuma returns an empty value, so that Terraform does
// not detect a difference between an unset attribute and an empty string. icon keeps an empty
// string that is present in the configuration, because Uptime Kuma returns it verbatim. The
// pointer valued attributes carry the distinction themselves.
func updateNtfyFields(data *NotificationNtfyResourceModel, ntfy *notification.Ntfy) {
	data.AuthenticationMethod = types.StringValue(ntfy.AuthenticationMethod)
	data.Priority = types.Int64Value(ntfy.Priority)
	data.ServerURL = types.StringValue(ntfy.ServerURL)
	data.Topic = types.StringValue(ntfy.Topic)

	data.AccessToken = stringOrNull(ntfy.AccessToken)
	data.Icon = stringOrNullPreserveEmpty(ntfy.Icon, data.Icon)
	data.Password = stringOrNull(ntfy.Password)
	data.Username = stringOrNull(ntfy.Username)

	data.Call = ptrToTypes(ntfy.Call)
	data.CustomMessage = ptrToTypes(ntfy.CustomMessage)
	data.CustomTitle = ptrToTypes(ntfy.CustomTitle)
	data.UseTemplate = boolPtrToTypes(ntfy.UseTemplate)

	if ntfy.PriorityDown != 0 {
		data.PriorityDown = types.Int64Value(ntfy.PriorityDown)
	} else {
		data.PriorityDown = types.Int64Null()
	}
}

// Create creates a new ntfy notification resource.
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
		NtfyDetails: buildNtfyDetails(&data),
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

// Read reads the current state of the ntfy notification resource.
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

	updateNtfyFields(&data, &ntfy)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the ntfy notification resource.
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
		NtfyDetails: buildNtfyDetails(&data),
	}

	err := r.client.UpdateNotification(ctx, ntfy)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the ntfy notification resource.
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
		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*NotificationNtfyResource) ImportState(
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
