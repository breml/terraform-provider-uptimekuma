package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ resource.Resource                   = &NotificationAmootSMSResource{}
	_ resource.ResourceWithImportState    = &NotificationAmootSMSResource{}
	_ resource.ResourceWithValidateConfig = &NotificationAmootSMSResource{}
)

// NewNotificationAmootSMSResource returns a new instance of the Amoot SMS notification resource.
func NewNotificationAmootSMSResource() resource.Resource {
	return &NotificationAmootSMSResource{}
}

// NotificationAmootSMSResource defines the resource implementation.
type NotificationAmootSMSResource struct {
	client *kuma.Client
}

// NotificationAmootSMSResourceModel describes the resource data model.
type NotificationAmootSMSResourceModel struct {
	NotificationBaseModel

	APIToken          types.String `tfsdk:"api_token"`
	APITokenWO        types.String `tfsdk:"api_token_wo"`
	APITokenWOVersion types.String `tfsdk:"api_token_wo_version"`
	Mobiles           types.String `tfsdk:"mobiles"`
	UsePattern        types.Bool   `tfsdk:"use_pattern"`
	PatternCodeID     types.Int64  `tfsdk:"pattern_code_id"`
	UseOwnLine        types.Bool   `tfsdk:"use_own_line"`
	LineNumber        types.String `tfsdk:"line_number"`
}

// Metadata returns the metadata for the resource.
func (*NotificationAmootSMSResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_amootsms"
}

// Schema returns the schema for the resource.
func (*NotificationAmootSMSResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Amoot SMS notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "The Amoot SMS API token.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"mobiles": schema.StringAttribute{
				MarkdownDescription: "The comma-separated recipient numbers. " +
					"Uptime Kuma trims every entry and drops empty ones.",
				Required: true,
				Validators: []validator.String{
					// Uptime Kuma saves a list without any entry, but then fails every send.
					stringvalidator.RegexMatches(
						regexp.MustCompile(`[^,\s]`),
						"must contain at least one recipient number",
					),
				},
			},
			"use_pattern": schema.BoolAttribute{
				MarkdownDescription: "Send the message as the values of a pattern registered with Amoot SMS " +
					"instead of as plain text. Requires `pattern_code_id`.",
				Optional: true,
			},
			"pattern_code_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the pattern registered with Amoot SMS. " +
					"Required when `use_pattern` is `true`, ignored otherwise.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"use_own_line": schema.BoolAttribute{
				MarkdownDescription: "Send pattern messages from `line_number`. " +
					"Only used when `use_pattern` is `true`.",
				Optional: true,
			},
			"line_number": schema.StringAttribute{
				MarkdownDescription: "The sender line, for example `public`, the value the Uptime Kuma web UI " +
					"suggests. Required unless `use_pattern` is `true`. With `use_pattern`, required when " +
					"`use_own_line` is `true` and ignored otherwise.",
				Optional: true,
				Validators: []validator.String{
					// Uptime Kuma only treats an empty line as missing, a blank one reaches the gateway.
					stringvalidator.RegexMatches(regexp.MustCompile(`\S`), "must not be blank"),
				},
			},
		}),
	}

	resp.Schema.Attributes = withWriteOnlySecret(
		resp.Schema.Attributes,
		"api_token",
		"The Amoot SMS API token",
	)
}

// Configure configures the Amoot SMS notification resource with the API client.
func (r *NotificationAmootSMSResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// ValidateConfig validates the configuration for the Amoot SMS notification resource.
//
// Uptime Kuma checks these requirements only when it sends a notification, so a broken configuration
// would apply cleanly and then fail on the first alert. AlsoRequires cannot express them, because they
// depend on the values of use_pattern and use_own_line rather than on them being set at all.
func (*NotificationAmootSMSResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config NotificationAmootSMSResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	validateAmootSMSSendMode(&config, resp)
}

// validateAmootSMSSendMode reports a value missing for the send mode selected by use_pattern and
// use_own_line. A value only known after apply is neither missing nor checked against.
func validateAmootSMSSendMode(config *NotificationAmootSMSResourceModel, resp *resource.ValidateConfigResponse) {
	if config.UsePattern.IsUnknown() {
		return
	}

	usePattern := config.UsePattern.ValueBool()

	if usePattern && config.PatternCodeID.IsNull() {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			path.Root("pattern_code_id"),
			"Missing Attribute Configuration",
			"When \"use_pattern\" is set to true, the attribute \"pattern_code_id\" must be set.",
		))
	}

	if (!usePattern || config.UseOwnLine.ValueBool()) && config.LineNumber.IsNull() {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			path.Root("line_number"),
			"Missing Attribute Configuration",
			"The attribute \"line_number\" must be set, unless \"use_pattern\" is set to true "+
				"and \"use_own_line\" is not.",
		))
	}
}

// Create creates a new Amoot SMS notification resource.
func (r *NotificationAmootSMSResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationAmootSMSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	amootSMS := amootSMSFromModel(&data)
	amootSMS.APIToken = writeOnlyOrPlan(ctx, req.Config, "api_token", data.APIToken, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateNotification(ctx, amootSMS)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create notification") {
		return
	}

	data.ID = types.Int64Value(id)

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Amoot SMS notification resource.
func (r *NotificationAmootSMSResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationAmootSMSResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, found := readNotification(ctx, r.client, id, notification.AmootSMSDetails{}.Type(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		removeOnMiss(ctx, r.client, notificationListEvent, "notification", resp)

		return
	}

	amootSMS := notification.AmootSMS{}
	err := base.As(&amootSMS)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to convert notification to type %q", notification.AmootSMSDetails{}.Type()),
			err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(amootSMS.Name)
	data.IsDefault = types.BoolValue(amootSMS.IsDefault)
	data.ApplyExisting = applyExistingFromState(data.ApplyExisting)

	if data.APITokenWOVersion.IsNull() {
		data.APIToken = types.StringValue(amootSMS.APIToken)
	}

	data.Mobiles = types.StringValue(amootSMS.Mobiles)
	data.UsePattern = boolPtrToTypes(amootSMS.UsePattern)
	data.PatternCodeID = int64PtrToTypes(amootSMS.PatternCodeID)
	data.UseOwnLine = boolPtrToTypes(amootSMS.UseOwnLine)

	// The server keeps whatever string it was given, and an empty one decodes to a non-nil pointer.
	// Treat it as unset, which is also how the server treats it when sending, so that it does not show
	// up as a diff against a null line_number.
	if amootSMS.LineNumber == nil || *amootSMS.LineNumber == "" {
		data.LineNumber = types.StringNull()
	} else {
		data.LineNumber = types.StringValue(*amootSMS.LineNumber)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Amoot SMS notification resource.
func (r *NotificationAmootSMSResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationAmootSMSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	applyExisting := applyExistingOnUpdate(ctx, req, &resp.Diagnostics)
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

	amootSMS := amootSMSFromModel(&data)
	amootSMS.ApplyExisting = applyExisting
	amootSMS.ID = id
	amootSMS.APIToken = writeOnlyOrPlan(ctx, req.Config, "api_token", data.APIToken, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateNotification(ctx, amootSMS)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update notification") {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Amoot SMS notification resource.
func (r *NotificationAmootSMSResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationAmootSMSResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete notification")
}

// ImportState imports an existing resource by ID.
func (*NotificationAmootSMSResource) ImportState(
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

// amootSMSFromModel builds an Amoot SMS notification from the resource model. The caller resolves
// api_token, which may be write-only and so is not in the model.
//
// The optional fields are omitempty on the wire, so a null attribute maps to nil, which keeps it out
// of the request entirely.
func amootSMSFromModel(data *NotificationAmootSMSResourceModel) notification.AmootSMS {
	return notification.AmootSMS{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      true,
			Name:          data.Name.ValueString(),
		},
		AmootSMSDetails: notification.AmootSMSDetails{
			Mobiles:       data.Mobiles.ValueString(),
			UsePattern:    boolToPtr(data.UsePattern),
			PatternCodeID: int64ToPtr(data.PatternCodeID),
			UseOwnLine:    boolToPtr(data.UseOwnLine),
			LineNumber:    strToPtr(data.LineNumber),
		},
	}
}
