package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &NotificationDataSource{}

// NewNotificationDataSource returns a new instance of the notification data source.
func NewNotificationDataSource() datasource.DataSource {
	return &NotificationDataSource{}
}

// NotificationDataSource manages notification data source operations.
type NotificationDataSource struct {
	client *kuma.Client
}

// NotificationDataSourceModel describes the data model for notification data source.
type NotificationDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

// Metadata returns the metadata for the data source.
func (*NotificationDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification"
}

// Schema returns the schema for the data source.
func (*NotificationDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get notification information by ID or name",
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
			"type": schema.StringAttribute{
				MarkdownDescription: "Notification type",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *NotificationDataSource) Configure(
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
func (d *NotificationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NotificationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		notification, err := d.client.GetNotification(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("failed to read notification", err.Error())
			return
		}

		data.Name = types.StringValue(notification.Name)
		data.Type = types.StringValue(notification.Type())
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		notifications := d.client.GetNotifications(ctx)

		var found *struct {
			ID   int64
			Name string
			Type string
		}

		for _, notif := range notifications {
			if notif.Name == data.Name.ValueString() {
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple notifications found",
						fmt.Sprintf(
							"Multiple notifications with name '%s' found. Please use 'id' to specify the notification uniquely.",
							data.Name.ValueString(),
						),
					)
					return
				}

				found = &struct {
					ID   int64
					Name string
					Type string
				}{
					ID:   notif.GetID(),
					Name: notif.Name,
					Type: notif.Type(),
				}
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"Notification not found",
				fmt.Sprintf("No notification with name '%s' found.", data.Name.ValueString()),
			)
			return
		}

		data.ID = types.Int64Value(found.ID)
		data.Type = types.StringValue(found.Type)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
