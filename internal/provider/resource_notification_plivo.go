package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                   = &NotificationPlivoResource{}
	_ resource.ResourceWithImportState    = &NotificationPlivoResource{}
	_ resource.ResourceWithValidateConfig = &NotificationPlivoResource{}
)

// NewNotificationPlivoResource returns a new instance of the Plivo notification resource.
func NewNotificationPlivoResource() resource.Resource {
	return &NotificationPlivoResource{}
}

// NotificationPlivoResource defines the resource implementation.
type NotificationPlivoResource struct {
	client *kuma.Client
}

// NotificationPlivoResourceModel describes the resource data model.
type NotificationPlivoResourceModel struct {
	NotificationBaseModel

	AuthID      types.String `tfsdk:"auth_id"`
	AuthToken   types.String `tfsdk:"auth_token"`
	FromNumber  types.String `tfsdk:"from_number"`
	ToNumber    types.String `tfsdk:"to_number"`
	MessageType types.String `tfsdk:"message_type"`
	AnswerURL   types.String `tfsdk:"answer_url"`
}

// Metadata returns the metadata for the resource.
func (*NotificationPlivoResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_plivo"
}

// Schema returns the schema for the resource.
func (*NotificationPlivoResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Plivo notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"auth_id": schema.StringAttribute{
				MarkdownDescription: "The Plivo auth ID used to authenticate requests.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "The Plivo auth token used to authenticate requests.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"from_number": schema.StringAttribute{
				MarkdownDescription: "The Plivo sender phone number.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to_number": schema.StringAttribute{
				MarkdownDescription: "The Plivo recipient phone number.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"message_type": schema.StringAttribute{
				MarkdownDescription: "How the alert is delivered, either `sms` or `call`. " +
					"If unset, Uptime Kuma delivers the alert as an SMS.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						notification.PlivoMessageTypeSMS.String(),
						notification.PlivoMessageTypeCall.String(),
					),
				},
			},
			"answer_url": schema.StringAttribute{
				MarkdownDescription: "The absolute URL Plivo fetches with an HTTP GET to obtain the Plivo XML " +
					"driving the call. It is only used and only required when `message_type` is `call`. " +
					"Uptime Kuma sets the alert text as the `message` query parameter, replacing any `message` " +
					"parameter already present.",
				Optional: true,
				Validators: []validator.String{
					validateURL(),
				},
			},
		}),
	}
}

// Configure configures the Plivo notification resource with the API client.
func (r *NotificationPlivoResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// ValidateConfig validates the configuration for the Plivo notification resource.
//
// answer_url is only consulted by Uptime Kuma when the alert is delivered as a call, where it is
// mandatory. stringvalidator.AlsoRequires cannot express either half of that, because both depend
// on the value of message_type rather than on it being set at all.
func (*NotificationPlivoResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config NotificationPlivoResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	validatePlivoAnswerURL(&config, resp)
}

// validatePlivoAnswerURL reports an answer_url that does not match the configured message type.
func validatePlivoAnswerURL(config *NotificationPlivoResourceModel, resp *resource.ValidateConfigResponse) {
	if config.MessageType.IsUnknown() {
		// The value is only known after apply, so there is nothing to validate against.
		return
	}

	// A null message_type means the server default, which is an SMS.
	messageType := notification.PlivoMessageTypeSMS.String()
	if !config.MessageType.IsNull() {
		messageType = config.MessageType.ValueString()
	}

	switch messageType {
	case notification.PlivoMessageTypeCall.String():
		if config.AnswerURL.IsNull() {
			resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				path.Root("answer_url"),
				"Missing Attribute Configuration",
				"When \"message_type\" is set to \"call\", the attribute \"answer_url\" must be set. "+
					"Plivo fetches it to obtain the Plivo XML driving the call.",
			))
		}

	case notification.PlivoMessageTypeSMS.String():
		if !config.AnswerURL.IsNull() {
			resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				path.Root("answer_url"),
				"Invalid Attribute Combination",
				"The attribute \"answer_url\" is only used when \"message_type\" is set to \"call\" "+
					"and must not be set otherwise.",
			))
		}

	default:
		// An unexpected message type is left to the schema validator.
	}
}

// Create creates a new Plivo notification resource.
func (r *NotificationPlivoResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationPlivoResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	plivo := plivoFromModel(&data)

	id, err := r.client.CreateNotification(ctx, plivo)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Plivo notification resource.
func (r *NotificationPlivoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationPlivoResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.PlivoDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)

		return
	}

	plivo := notification.Plivo{}
	err := base.As(&plivo)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.PlivoDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(plivo.Name)
	data.IsActive = types.BoolValue(plivo.IsActive)
	data.IsDefault = types.BoolValue(plivo.IsDefault)
	data.ApplyExisting = types.BoolValue(plivo.ApplyExisting)

	data.AuthID = types.StringValue(plivo.AuthID)
	data.AuthToken = types.StringValue(plivo.AuthToken)
	data.FromNumber = types.StringValue(plivo.FromNumber)
	data.ToNumber = types.StringValue(plivo.ToNumber)

	// plivoMessageType is omitempty on the wire, so an unset message type comes back as the empty
	// string. Mapping it to an empty string instead of null would produce a perpetual diff.
	if plivo.MessageType == "" {
		data.MessageType = types.StringNull()
	} else {
		data.MessageType = types.StringValue(plivo.MessageType.String())
	}

	data.AnswerURL = ptrToTypes(plivo.AnswerURL)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Plivo notification resource.
func (r *NotificationPlivoResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationPlivoResourceModel

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

	plivo := plivoFromModel(&data)
	plivo.ID = id

	err := r.client.UpdateNotification(ctx, plivo)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Plivo notification resource.
func (r *NotificationPlivoResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationPlivoResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			return
		}

		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*NotificationPlivoResource) ImportState(
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

// plivoFromModel builds a Plivo notification from the resource model.
//
// plivoMessageType and plivoAnswerUrl are omitempty on the wire, so a null message_type maps to the
// zero value and a null answer_url to nil, which keeps them out of the request entirely.
func plivoFromModel(data *NotificationPlivoResourceModel) notification.Plivo {
	return notification.Plivo{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		PlivoDetails: notification.PlivoDetails{
			AuthID:      data.AuthID.ValueString(),
			AuthToken:   data.AuthToken.ValueString(),
			FromNumber:  data.FromNumber.ValueString(),
			ToNumber:    data.ToNumber.ValueString(),
			MessageType: notification.PlivoMessageType(data.MessageType.ValueString()),
			AnswerURL:   strToPtr(data.AnswerURL),
		},
	}
}
