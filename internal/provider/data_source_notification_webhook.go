package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &NotificationWebhookDataSource{}

// NewNotificationWebhookDataSource returns a new instance of the Webhook notification data source.
func NewNotificationWebhookDataSource() datasource.DataSource {
	return &NotificationWebhookDataSource{}
}

// NotificationWebhookDataSource manages Webhook notification data source operations.
type NotificationWebhookDataSource struct {
	client *kuma.Client
}

// NotificationWebhookDataSourceModel describes the data model for Webhook notification data source.
type NotificationWebhookDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*NotificationWebhookDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_webhook"
}

// Schema returns the schema for the data source.
func (*NotificationWebhookDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Webhook notification information by ID or name",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Notification identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Notification name",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *NotificationWebhookDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	d.client = client
}

// Read reads the current state of the data source.
func (d *NotificationWebhookDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationWebhookDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

 // Attempt to read by ID if provided.
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		notification, err := d.client.GetNotification(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("failed to read notification", err.Error())
			return
		}

		if notification.Type() != "webhook" {
			resp.Diagnostics.AddError("Incorrect notification type", "Notification is not a Webhook notification")
			return
		}

		data.Name = types.StringValue(notification.Name)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

 // Attempt to read by name if ID not provided.
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		notifications := d.client.GetNotifications(ctx)

		var found *struct {
			ID   int64
			Name string
		}

		for i := range notifications {
			if notifications[i].Name == data.Name.ValueString() && notifications[i].Type() == "webhook" {
    // Error if multiple matches found.
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple notifications found",
						fmt.Sprintf(
							"Multiple Webhook notifications with name '%s' found. Please use 'id' to specify the notification uniquely.",
							data.Name.ValueString(),
						),
					)
					return
				}

    // Store matched item.
				found = &struct {
					ID   int64
					Name string
				}{
					ID:   notifications[i].GetID(),
					Name: notifications[i].Name,
				}
			}
		}

  // Error if no matching item found.
		if found == nil {
			resp.Diagnostics.AddError(
				"Notification not found",
				fmt.Sprintf("No Webhook notification with name '%s' found.", data.Name.ValueString()),
			)
			return
		}

		data.ID = types.Int64Value(found.ID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
 // Error if neither ID nor name provided.
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
