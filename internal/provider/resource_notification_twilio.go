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
	_ resource.Resource                = &NotificationTwilioResource{}
	_ resource.ResourceWithImportState = &NotificationTwilioResource{}
)

// NewNotificationTwilioResource returns a new instance of the Twilio notification resource.
func NewNotificationTwilioResource() resource.Resource {
	return &NotificationTwilioResource{}
}

// NotificationTwilioResource defines the resource implementation.
type NotificationTwilioResource struct {
	client *kuma.Client
}

// NotificationTwilioResourceModel describes the resource data model.
type NotificationTwilioResourceModel struct {
	NotificationBaseModel

	AccountSID types.String `tfsdk:"account_sid"`
	APIKey     types.String `tfsdk:"api_key"`
	AuthToken  types.String `tfsdk:"auth_token"`
	ToNumber   types.String `tfsdk:"to_number"`
	FromNumber types.String `tfsdk:"from_number"`
}

// Metadata returns the metadata for the resource.
func (*NotificationTwilioResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_twilio"
}

// Schema returns the schema for the resource.
func (*NotificationTwilioResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Twilio notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"account_sid": schema.StringAttribute{
				MarkdownDescription: "Twilio account SID",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Twilio API key",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "Twilio auth token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"to_number": schema.StringAttribute{
				MarkdownDescription: "Twilio recipient phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"from_number": schema.StringAttribute{
				MarkdownDescription: "Twilio sender phone number",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the Twilio notification resource with the API client.
func (r *NotificationTwilioResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Twilio notification resource.
func (r *NotificationTwilioResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationTwilioResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	twilio := notification.Twilio{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		TwilioDetails: notification.TwilioDetails{
			AccountSID: data.AccountSID.ValueString(),
			APIKey:     data.APIKey.ValueString(),
			AuthToken:  data.AuthToken.ValueString(),
			ToNumber:   data.ToNumber.ValueString(),
			FromNumber: data.FromNumber.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, twilio)
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

// Read reads the current state of the Twilio notification resource.
func (r *NotificationTwilioResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationTwilioResourceModel

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

	twilio := notification.Twilio{}
	err = base.As(&twilio)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "twilio"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(twilio.Name)
	data.IsActive = types.BoolValue(twilio.IsActive)
	data.IsDefault = types.BoolValue(twilio.IsDefault)
	data.ApplyExisting = types.BoolValue(twilio.ApplyExisting)

	data.AccountSID = types.StringValue(twilio.AccountSID)
	if twilio.APIKey != "" {
		data.APIKey = types.StringValue(twilio.APIKey)
	} else {
		data.APIKey = types.StringNull()
	}

	data.AuthToken = types.StringValue(twilio.AuthToken)
	data.ToNumber = types.StringValue(twilio.ToNumber)
	data.FromNumber = types.StringValue(twilio.FromNumber)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Twilio notification resource.
func (r *NotificationTwilioResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationTwilioResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	twilio := notification.Twilio{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		TwilioDetails: notification.TwilioDetails{
			AccountSID: data.AccountSID.ValueString(),
			APIKey:     data.APIKey.ValueString(),
			AuthToken:  data.AuthToken.ValueString(),
			ToNumber:   data.ToNumber.ValueString(),
			FromNumber: data.FromNumber.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, twilio)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Twilio notification resource.
func (r *NotificationTwilioResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationTwilioResourceModel

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
func (*NotificationTwilioResource) ImportState(
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
