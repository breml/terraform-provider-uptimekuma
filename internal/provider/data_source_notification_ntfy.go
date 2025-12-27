package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &NotificationNtfyDataSource{}

// NewNotificationNtfyDataSource returns a new instance of the ntfy notification data source.
func NewNotificationNtfyDataSource() datasource.DataSource {
	return &NotificationNtfyDataSource{}
}

// NotificationNtfyDataSource manages ntfy notification data source operations.
type NotificationNtfyDataSource struct {
	client *kuma.Client
}

// NotificationNtfyDataSourceModel describes the data model for ntfy notification data source.
type NotificationNtfyDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (*NotificationNtfyDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_ntfy"
}

func (*NotificationNtfyDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get ntfy notification information by ID or name",
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

// Configure configures the ntfy notification data source with the API client.
func (d *NotificationNtfyDataSource) Configure(
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

func (d *NotificationNtfyDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationNtfyDataSourceModel

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

		if notification.Type() != "ntfy" {
			resp.Diagnostics.AddError("Incorrect notification type", "Notification is not an ntfy notification")
			return
		}

		data.Name = types.StringValue(notification.Name)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		notifications := d.client.GetNotifications(ctx)

		var found *struct {
			ID   int64
			Name string
		}

		for i := range notifications {
			if notifications[i].Name == data.Name.ValueString() && notifications[i].Type() == "ntfy" {
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple notifications found",
						fmt.Sprintf(
							"Multiple ntfy notifications with name '%s' found. Please use 'id' to specify the notification uniquely.",
							data.Name.ValueString(),
						),
					)
					return
				}

				found = &struct {
					ID   int64
					Name string
				}{
					ID:   notifications[i].GetID(),
					Name: notifications[i].Name,
				}
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"Notification not found",
				fmt.Sprintf("No ntfy notification with name '%s' found.", data.Name.ValueString()),
			)
			return
		}

		data.ID = types.Int64Value(found.ID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
