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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationTeltonikaResource{}
	_ resource.ResourceWithImportState = &NotificationTeltonikaResource{}
)

// NewNotificationTeltonikaResource returns a new instance of the Teltonika notification resource.
func NewNotificationTeltonikaResource() resource.Resource {
	return &NotificationTeltonikaResource{}
}

// NotificationTeltonikaResource defines the resource implementation.
type NotificationTeltonikaResource struct {
	client *kuma.Client
}

// NotificationTeltonikaResourceModel describes the resource data model.
type NotificationTeltonikaResourceModel struct {
	NotificationBaseModel

	URL         types.String `tfsdk:"url"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	Modem       types.String `tfsdk:"modem"`
	PhoneNumber types.String `tfsdk:"phone_number"`
	UnsafeTLS   types.Bool   `tfsdk:"unsafe_tls"`
}

// Metadata returns the metadata for the resource.
func (*NotificationTeltonikaResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_teltonika"
}

// Schema returns the schema for the resource.
func (*NotificationTeltonikaResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Teltonika notification resource. Uses a Teltonika RUTxxx series router as an " +
			"SMS gateway over its HTTP API. This provider is only compatible with Teltonika RutOS >= 7.14.0 devices.",
		Attributes: withNotificationBaseAttributes(teltonikaSchemaAttributes()),
	}
}

func teltonikaSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			MarkdownDescription: "Teltonika router base URL (e.g., https://192.168.1.1).",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"username": schema.StringAttribute{
			MarkdownDescription: "The router user used to authenticate against the Teltonika HTTP API.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"password": schema.StringAttribute{
			MarkdownDescription: "The password for the router user.",
			Required:            true,
			Sensitive:           true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"modem": schema.StringAttribute{
			MarkdownDescription: "The modem identifier on the router (e.g., \"1-1\").",
			Optional:            true,
		},
		"phone_number": schema.StringAttribute{
			MarkdownDescription: "The recipient phone number (e.g., \"+336xxxxxxxx\"). " +
				"Multiple recipients can be provided as a comma-separated list.",
			Required: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"unsafe_tls": schema.BoolAttribute{
			MarkdownDescription: "Disables TLS certificate validation when communicating with the router. " +
				"This is useful for routers using a self-signed certificate.",
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
	}
}

// Configure configures the Teltonika notification resource with the API client.
func (r *NotificationTeltonikaResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Teltonika notification resource.
func (r *NotificationTeltonikaResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationTeltonikaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	teltonika := teltonikaFromModel(&data)

	id, err := r.client.CreateNotification(ctx, teltonika)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Teltonika notification resource.
func (r *NotificationTeltonikaResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationTeltonikaResourceModel

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

	teltonika := notification.Teltonika{}
	err = base.As(&teltonika)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "Teltonika"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(teltonika.Name)
	data.IsActive = types.BoolValue(teltonika.IsActive)
	data.IsDefault = types.BoolValue(teltonika.IsDefault)
	data.ApplyExisting = types.BoolValue(teltonika.ApplyExisting)

	data.URL = types.StringValue(teltonika.URL)
	data.Username = types.StringValue(teltonika.Username)
	data.PhoneNumber = types.StringValue(teltonika.PhoneNumber)
	data.UnsafeTLS = types.BoolValue(teltonika.UnsafeTLS)

	if teltonika.Modem != "" {
		data.Modem = types.StringValue(teltonika.Modem)
	} else {
		data.Modem = types.StringNull()
	}

	// Uptime Kuma does not return secret fields, so only overwrite the
	// password in state if the API actually returned a non-empty value.
	if teltonika.Password != "" {
		data.Password = types.StringValue(teltonika.Password)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Teltonika notification resource.
func (r *NotificationTeltonikaResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationTeltonikaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if id == 0 {
		resp.Diagnostics.AddError(
			"Invalid resource state",
			"Cannot update notification: resource ID is missing from state. This is a provider bug.",
		)

		return
	}

	teltonika := teltonikaFromModel(&data)
	teltonika.ID = id

	err := r.client.UpdateNotification(ctx, teltonika)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Teltonika notification resource.
func (r *NotificationTeltonikaResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationTeltonikaResourceModel

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
func (*NotificationTeltonikaResource) ImportState(
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

// teltonikaFromModel builds a Teltonika notification from the resource model.
func teltonikaFromModel(data *NotificationTeltonikaResourceModel) notification.Teltonika {
	return notification.Teltonika{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		TeltonikaDetails: notification.TeltonikaDetails{
			URL:         data.URL.ValueString(),
			Username:    data.Username.ValueString(),
			Password:    data.Password.ValueString(),
			Modem:       data.Modem.ValueString(),
			PhoneNumber: data.PhoneNumber.ValueString(),
			UnsafeTLS:   data.UnsafeTLS.ValueBool(),
		},
	}
}
