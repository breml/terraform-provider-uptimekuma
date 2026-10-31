package provider

import (
	"context"
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
	_ resource.Resource                = &NotificationTurboSMTPResource{}
	_ resource.ResourceWithImportState = &NotificationTurboSMTPResource{}
)

// NewNotificationTurboSMTPResource returns a new instance of the TurboSMTP notification resource.
func NewNotificationTurboSMTPResource() resource.Resource {
	return &NotificationTurboSMTPResource{}
}

// NotificationTurboSMTPResource defines the resource implementation.
type NotificationTurboSMTPResource struct {
	client *kuma.Client
}

// NotificationTurboSMTPResourceModel describes the resource data model.
type NotificationTurboSMTPResourceModel struct {
	NotificationBaseModel

	ConsumerKey             types.String `tfsdk:"consumer_key"`
	ConsumerSecret          types.String `tfsdk:"consumer_secret"`
	ConsumerSecretWO        types.String `tfsdk:"consumer_secret_wo"`
	ConsumerSecretWOVersion types.String `tfsdk:"consumer_secret_wo_version"`
	Region                  types.String `tfsdk:"region"`
	FromEmail               types.String `tfsdk:"from_email"`
	ToEmail                 types.String `tfsdk:"to_email"`
	CcEmail                 types.String `tfsdk:"cc_email"`
	BccEmail                types.String `tfsdk:"bcc_email"`
	Subject                 types.String `tfsdk:"subject"`
}

// Metadata returns the metadata for the resource.
func (*NotificationTurboSMTPResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_turbosmtp"
}

// Schema returns the schema for the resource.
func (*NotificationTurboSMTPResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "TurboSMTP notification resource. The alert is sent as an email through the " +
			"TurboSMTP HTTP API. Requires Uptime Kuma 2.5.1 or later.",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"consumer_key": schema.StringAttribute{
				MarkdownDescription: "The TurboSMTP consumer key identifying the API account.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"consumer_secret": schema.StringAttribute{
				MarkdownDescription: "The TurboSMTP consumer secret used to authenticate requests.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The TurboSMTP API region, either `us` or `eu`. " +
					"If unset, Uptime Kuma sends the alert through the US API host.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						notification.TurboSMTPRegionUS.String(),
						notification.TurboSMTPRegionEU.String(),
					),
				},
			},
			"from_email": schema.StringAttribute{
				MarkdownDescription: "The sender email address.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to_email": schema.StringAttribute{
				MarkdownDescription: "A comma-separated list of recipient email addresses. Unlike `cc_email` " +
					"and `bcc_email`, it is passed on verbatim, so the whitespace around an address is not trimmed.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cc_email": schema.StringAttribute{
				MarkdownDescription: "A comma-separated list of CC email addresses.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"bcc_email": schema.StringAttribute{
				MarkdownDescription: "A comma-separated list of BCC email addresses.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "The subject of the email. " +
					"If unset, Uptime Kuma uses `Notification from Your Uptime Kuma`.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}

	resp.Schema.Attributes = withWriteOnlySecret(
		resp.Schema.Attributes,
		"consumer_secret",
		"The TurboSMTP consumer secret used to authenticate requests",
	)
}

// Configure configures the TurboSMTP notification resource with the API client.
func (r *NotificationTurboSMTPResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new TurboSMTP notification resource.
func (r *NotificationTurboSMTPResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationTurboSMTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	turboSMTP := turboSMTPFromModel(&data)
	turboSMTP.ConsumerSecret = writeOnlyOrPlan(
		ctx, req.Config, "consumer_secret", data.ConsumerSecret, &resp.Diagnostics,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateNotification(ctx, turboSMTP)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the TurboSMTP notification resource.
func (r *NotificationTurboSMTPResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationTurboSMTPResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.TurboSMTPDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	turboSMTP := notification.TurboSMTP{}
	err := base.As(&turboSMTP)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.TurboSMTPDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(turboSMTP.Name)
	data.IsDefault = types.BoolValue(turboSMTP.IsDefault)
	data.ApplyExisting = types.BoolValue(turboSMTP.ApplyExisting)

	data.ConsumerKey = types.StringValue(turboSMTP.ConsumerKey)

	if data.ConsumerSecretWOVersion.IsNull() {
		data.ConsumerSecret = types.StringValue(turboSMTP.ConsumerSecret)
	}

	// turbosmtpRegion is omitempty on the wire, so an unset region comes back as the empty string.
	// Mapping it to an empty string instead of null would produce a perpetual diff.
	if turboSMTP.Region == "" {
		data.Region = types.StringNull()
	} else {
		data.Region = types.StringValue(turboSMTP.Region.String())
	}

	data.FromEmail = types.StringValue(turboSMTP.FromEmail)
	data.ToEmail = types.StringValue(turboSMTP.ToEmail)

	// The optional fields are *string, so omitempty drops them only when they are nil, not when they
	// point at the empty string. A field cleared in the Uptime Kuma UI comes back as "", and keeping
	// that in state against a null configuration would be a perpetual diff.
	data.CcEmail = stringOrNullPtr(turboSMTP.CcEmail)
	data.BccEmail = stringOrNullPtr(turboSMTP.BccEmail)
	data.Subject = stringOrNullPtr(turboSMTP.Subject)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the TurboSMTP notification resource.
func (r *NotificationTurboSMTPResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationTurboSMTPResourceModel

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

	turboSMTP := turboSMTPFromModel(&data)
	turboSMTP.ID = id
	turboSMTP.ConsumerSecret = writeOnlyOrPlan(
		ctx, req.Config, "consumer_secret", data.ConsumerSecret, &resp.Diagnostics,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNotification(ctx, turboSMTP)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the TurboSMTP notification resource.
func (r *NotificationTurboSMTPResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationTurboSMTPResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationTurboSMTPResource) ImportState(
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

// turboSMTPFromModel builds a TurboSMTP notification from the resource model. The caller resolves
// consumer_secret, which may be write-only and so is not in the model.
//
// turbosmtpRegion and the optional fields are omitempty on the wire, so a null region maps to the
// zero value and a null optional field to nil, which keeps them out of the request entirely.
func turboSMTPFromModel(data *NotificationTurboSMTPResourceModel) notification.TurboSMTP {
	return notification.TurboSMTP{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      true,
			Name:          data.Name.ValueString(),
		},
		TurboSMTPDetails: notification.TurboSMTPDetails{
			ConsumerKey: data.ConsumerKey.ValueString(),
			Region:      notification.TurboSMTPRegion(data.Region.ValueString()),
			FromEmail:   data.FromEmail.ValueString(),
			ToEmail:     data.ToEmail.ValueString(),
			CcEmail:     strToPtr(data.CcEmail),
			BccEmail:    strToPtr(data.BccEmail),
			Subject:     strToPtr(data.Subject),
		},
	}
}
