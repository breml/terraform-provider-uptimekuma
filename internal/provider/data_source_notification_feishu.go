package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &NotificationFeishuDataSource{}

func NewNotificationFeishuDataSource() datasource.DataSource {
	return &NotificationFeishuDataSource{}
}

type NotificationFeishuDataSource struct {
	client *kuma.Client
}

type NotificationFeishuDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *NotificationFeishuDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_feishu"
}

func (d *NotificationFeishuDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Feishu notification information by ID or name",
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

func (d *NotificationFeishuDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *kuma.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *NotificationFeishuDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NotificationFeishuDataSourceModel

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

		if notification.Type() != "Feishu" {
			resp.Diagnostics.AddError("Incorrect notification type", "Notification is not a Feishu notification")
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
			if notifications[i].Name == data.Name.ValueString() && notifications[i].Type() == "Feishu" {
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple notifications found",
						fmt.Sprintf("Multiple Feishu notifications with name '%s' found. Please use 'id' to specify the notification uniquely.", data.Name.ValueString()),
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
				fmt.Sprintf("No Feishu notification with name '%s' found.", data.Name.ValueString()),
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
