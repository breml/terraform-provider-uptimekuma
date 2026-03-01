package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &NotificationGTXMessagingDataSource{}

// NewNotificationGTXMessagingDataSource returns a new instance of the GTX Messaging notification data source.
func NewNotificationGTXMessagingDataSource() datasource.DataSource {
	return &NotificationGTXMessagingDataSource{}
}

// NotificationGTXMessagingDataSource manages GTX Messaging notification data source operations.
type NotificationGTXMessagingDataSource struct {
	client *kuma.Client
}

// NotificationGTXMessagingDataSourceModel describes the data model for GTX Messaging notification data source.
type NotificationGTXMessagingDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*NotificationGTXMessagingDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_gtxmessaging"
}

// Schema returns the schema for the data source.
func (*NotificationGTXMessagingDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get GTX Messaging notification information by ID or name",
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
func (d *NotificationGTXMessagingDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *NotificationGTXMessagingDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationGTXMessagingDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !validateNotificationDataSourceInput(resp, data.ID, data.Name) {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		d.readByID(ctx, &data, resp)
		return
	}

	d.readByName(ctx, &data, resp)
}

func (d *NotificationGTXMessagingDataSource) readByID(
	ctx context.Context,
	data *NotificationGTXMessagingDataSourceModel,
	resp *datasource.ReadResponse,
) {
	notification, err := d.client.GetNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	if notification.Type() != "gtxmessaging" {
		resp.Diagnostics.AddError("Incorrect notification type", "Notification is not a GTX Messaging notification")
		return
	}

	data.Name = types.StringValue(notification.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *NotificationGTXMessagingDataSource) readByName(
	ctx context.Context,
	data *NotificationGTXMessagingDataSourceModel,
	resp *datasource.ReadResponse,
) {
	id, ok := findNotificationByName(ctx, d.client, data.Name.ValueString(), "gtxmessaging", &resp.Diagnostics)
	if !ok {
		return
	}

	data.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
