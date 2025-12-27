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
	_ resource.Resource                = &NotificationAppriseResource{}
	_ resource.ResourceWithImportState = &NotificationAppriseResource{}
)

// NewNotificationAppriseResource returns a new instance of the Apprise notification resource.
func NewNotificationAppriseResource() resource.Resource {
	return &NotificationAppriseResource{}
}

// NotificationAppriseResource defines the resource implementation.
type NotificationAppriseResource struct {
	client *kuma.Client
}

// NotificationAppriseResourceModel describes the resource data model.
type NotificationAppriseResourceModel struct {
	NotificationBaseModel

	AppriseURL types.String `tfsdk:"apprise_url"`
	Title      types.String `tfsdk:"title"`
}

// Metadata returns the metadata for the resource.
func (*NotificationAppriseResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_apprise"
}

// Schema returns the schema for the resource.
func (*NotificationAppriseResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Apprise notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"apprise_url": schema.StringAttribute{
				MarkdownDescription: "Apprise URL or comma-separated list of URLs",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Notification title",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Apprise notification resource with the API client.
func (r *NotificationAppriseResource) Configure(
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

// Create creates a new Apprise notification resource.
func (r *NotificationAppriseResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationAppriseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apprise := notification.Apprise{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		AppriseDetails: notification.AppriseDetails{
			AppriseURL: data.AppriseURL.ValueString(),
			Title:      data.Title.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, apprise)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Apprise notification resource.
func (r *NotificationAppriseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationAppriseResourceModel

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

	apprise := notification.Apprise{}
	err = base.As(&apprise)
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "apprise"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(apprise.Name)
	data.IsActive = types.BoolValue(apprise.IsActive)
	data.IsDefault = types.BoolValue(apprise.IsDefault)
	data.ApplyExisting = types.BoolValue(apprise.ApplyExisting)

	data.AppriseURL = types.StringValue(apprise.AppriseURL)
	if apprise.Title != "" {
		data.Title = types.StringValue(apprise.Title)
	} else {
		data.Title = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Apprise notification resource.
func (r *NotificationAppriseResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationAppriseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apprise := notification.Apprise{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		AppriseDetails: notification.AppriseDetails{
			AppriseURL: data.AppriseURL.ValueString(),
			Title:      data.Title.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, apprise)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Apprise notification resource.
func (r *NotificationAppriseResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationAppriseResourceModel

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
func (*NotificationAppriseResource) ImportState(
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
