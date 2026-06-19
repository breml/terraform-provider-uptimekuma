package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationJiraServiceManagementResource{}
	_ resource.ResourceWithImportState = &NotificationJiraServiceManagementResource{}
)

// NewNotificationJiraServiceManagementResource returns a new instance of the Jira Service Management
// notification resource.
func NewNotificationJiraServiceManagementResource() resource.Resource {
	return &NotificationJiraServiceManagementResource{}
}

// NotificationJiraServiceManagementResource defines the resource implementation.
type NotificationJiraServiceManagementResource struct {
	client *kuma.Client
}

// NotificationJiraServiceManagementResourceModel describes the resource data model.
type NotificationJiraServiceManagementResourceModel struct {
	NotificationBaseModel

	CloudID  types.String `tfsdk:"cloud_id"`
	Email    types.String `tfsdk:"email"`
	APIToken types.String `tfsdk:"api_token"`
	Priority types.Int64  `tfsdk:"priority"`
}

// Metadata returns the metadata for the resource.
func (*NotificationJiraServiceManagementResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_jiraservicemanagement"
}

// Schema returns the schema for the resource.
func (*NotificationJiraServiceManagementResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Jira Service Management notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"cloud_id": schema.StringAttribute{
				MarkdownDescription: "Atlassian site cloud ID",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Atlassian account email",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "Atlassian API token",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Alert priority (1-5)",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3),
				Validators: []validator.Int64{
					int64validator.Between(1, 5),
				},
			},
		}),
	}
}

// Configure configures the Jira Service Management notification resource with the API client.
func (r *NotificationJiraServiceManagementResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Jira Service Management notification resource.
func (r *NotificationJiraServiceManagementResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationJiraServiceManagementResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jiraServiceManagement := notification.JiraServiceManagement{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		JiraServiceManagementDetails: notification.JiraServiceManagementDetails{
			CloudID:  data.CloudID.ValueString(),
			Email:    data.Email.ValueString(),
			APIToken: data.APIToken.ValueString(),
			Priority: int(data.Priority.ValueInt64()),
		},
	}

	id, err := r.client.CreateNotification(ctx, jiraServiceManagement)
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

// Read reads the current state of the Jira Service Management notification resource.
func (r *NotificationJiraServiceManagementResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationJiraServiceManagementResourceModel

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

	jiraServiceManagement := notification.JiraServiceManagement{}
	err = base.As(&jiraServiceManagement)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			`failed to convert notification to type "jiraservicemanagement"`, err.Error(),
		)
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(jiraServiceManagement.Name)
	data.IsActive = types.BoolValue(jiraServiceManagement.IsActive)
	data.IsDefault = types.BoolValue(jiraServiceManagement.IsDefault)
	data.ApplyExisting = types.BoolValue(jiraServiceManagement.ApplyExisting)

	data.CloudID = types.StringValue(jiraServiceManagement.CloudID)
	data.Email = types.StringValue(jiraServiceManagement.Email)
	data.APIToken = types.StringValue(jiraServiceManagement.APIToken)
	data.Priority = types.Int64Value(int64(jiraServiceManagement.Priority))

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Jira Service Management notification resource.
func (r *NotificationJiraServiceManagementResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationJiraServiceManagementResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jiraServiceManagement := notification.JiraServiceManagement{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		JiraServiceManagementDetails: notification.JiraServiceManagementDetails{
			CloudID:  data.CloudID.ValueString(),
			Email:    data.Email.ValueString(),
			APIToken: data.APIToken.ValueString(),
			Priority: int(data.Priority.ValueInt64()),
		},
	}

	err := r.client.UpdateNotification(ctx, jiraServiceManagement)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Jira Service Management notification resource.
func (r *NotificationJiraServiceManagementResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationJiraServiceManagementResourceModel

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
func (*NotificationJiraServiceManagementResource) ImportState(
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
