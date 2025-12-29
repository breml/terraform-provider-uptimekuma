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
	_ resource.Resource                = &NotificationBrevoResource{}
	_ resource.ResourceWithImportState = &NotificationBrevoResource{}
)

// NewNotificationBrevoResource returns a new instance of the Brevo notification resource.
func NewNotificationBrevoResource() resource.Resource {
	return &NotificationBrevoResource{}
}

// NotificationBrevoResource defines the resource implementation.
type NotificationBrevoResource struct {
	client *kuma.Client
}

// NotificationBrevoResourceModel describes the resource data model.
type NotificationBrevoResourceModel struct {
	NotificationBaseModel

	APIKey    types.String `tfsdk:"api_key"`
	ToEmail   types.String `tfsdk:"to_email"`
	FromEmail types.String `tfsdk:"from_email"`
	FromName  types.String `tfsdk:"from_name"`
	Subject   types.String `tfsdk:"subject"`
	CCEmail   types.String `tfsdk:"cc_email"`
	BCCEmail  types.String `tfsdk:"bcc_email"`
}

// Metadata returns the metadata for the resource.
func (*NotificationBrevoResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_brevo"
}

// Schema returns the schema for the resource.
func (*NotificationBrevoResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Brevo notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Brevo API key for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to_email": schema.StringAttribute{
				MarkdownDescription: "Recipient email address",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"from_email": schema.StringAttribute{
				MarkdownDescription: "Sender email address",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"from_name": schema.StringAttribute{
				MarkdownDescription: "Sender name",
				Optional:            true,
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "Email subject line",
				Optional:            true,
			},
			"cc_email": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of CC email addresses",
				Optional:            true,
			},
			"bcc_email": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of BCC email addresses",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Brevo notification resource with the API client.
func (r *NotificationBrevoResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
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

// Create creates a new Brevo notification resource.
func (r *NotificationBrevoResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationBrevoResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	brevo := notification.Brevo{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		BrevoDetails: notification.BrevoDetails{
			APIKey:    data.APIKey.ValueString(),
			ToEmail:   data.ToEmail.ValueString(),
			FromEmail: data.FromEmail.ValueString(),
			FromName:  data.FromName.ValueString(),
			Subject:   data.Subject.ValueString(),
			CCEmail:   data.CCEmail.ValueString(),
			BCCEmail:  data.BCCEmail.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, brevo)
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

// Read reads the current state of the Brevo notification resource.
func (r *NotificationBrevoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationBrevoResourceModel

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

	brevo := notification.Brevo{}
	err = base.As(&brevo)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "brevo"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(brevo.Name)
	data.IsActive = types.BoolValue(brevo.IsActive)
	data.IsDefault = types.BoolValue(brevo.IsDefault)
	data.ApplyExisting = types.BoolValue(brevo.ApplyExisting)

	data.APIKey = types.StringValue(brevo.APIKey)
	data.ToEmail = types.StringValue(brevo.ToEmail)
	data.FromEmail = types.StringValue(brevo.FromEmail)

	if brevo.FromName != "" {
		data.FromName = types.StringValue(brevo.FromName)
	} else {
		data.FromName = types.StringNull()
	}

	if brevo.Subject != "" {
		data.Subject = types.StringValue(brevo.Subject)
	} else {
		data.Subject = types.StringNull()
	}

	if brevo.CCEmail != "" {
		data.CCEmail = types.StringValue(brevo.CCEmail)
	} else {
		data.CCEmail = types.StringNull()
	}

	if brevo.BCCEmail != "" {
		data.BCCEmail = types.StringValue(brevo.BCCEmail)
	} else {
		data.BCCEmail = types.StringNull()
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Brevo notification resource.
func (r *NotificationBrevoResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationBrevoResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	brevo := notification.Brevo{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		BrevoDetails: notification.BrevoDetails{
			APIKey:    data.APIKey.ValueString(),
			ToEmail:   data.ToEmail.ValueString(),
			FromEmail: data.FromEmail.ValueString(),
			FromName:  data.FromName.ValueString(),
			Subject:   data.Subject.ValueString(),
			CCEmail:   data.CCEmail.ValueString(),
			BCCEmail:  data.BCCEmail.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, brevo)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Brevo notification resource.
func (r *NotificationBrevoResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationBrevoResourceModel

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
func (*NotificationBrevoResource) ImportState(
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
