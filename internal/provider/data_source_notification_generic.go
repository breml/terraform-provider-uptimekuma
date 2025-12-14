package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &NotificationGenericDataSource{}

func NewNotificationGenericDataSource() datasource.DataSource {
	return &NotificationGenericDataSource{}
}

type NotificationGenericDataSource struct {
	client *kuma.Client
}

type NotificationGenericDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

func (d *NotificationGenericDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_generic"
}

func (d *NotificationGenericDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get generic notification information by ID or name",
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
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (d *NotificationGenericDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NotificationGenericDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NotificationGenericDataSourceModel

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

		notificationType := ""
		if !data.Type.IsNull() && !data.Type.IsUnknown() {
			notificationType = data.Type.ValueString()
		}

		for i := range notifications {
			if notifications[i].Name == data.Name.ValueString() {
				if notificationType != "" && notifications[i].Type() != notificationType {
					continue
				}
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple notifications found",
						fmt.Sprintf("Multiple notifications with name '%s' found. Please use 'id' to specify the notification uniquely.", data.Name.ValueString()),
					)
					return
				}
				found = &struct {
					ID   int64
					Name string
					Type string
				}{
					ID:   notifications[i].GetID(),
					Name: notifications[i].Name,
					Type: notifications[i].Type(),
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
		data.Name = types.StringValue(found.Name)
		data.Type = types.StringValue(found.Type)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
