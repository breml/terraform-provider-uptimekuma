package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var _ datasource.DataSource = &NotificationOpenWaDataSource{}

// NewNotificationOpenWaDataSource returns a new instance of the OpenWA notification data source.
func NewNotificationOpenWaDataSource() datasource.DataSource {
	return &NotificationOpenWaDataSource{}
}

// NotificationOpenWaDataSource manages OpenWA notification data source operations.
type NotificationOpenWaDataSource struct {
	client *kuma.Client
}

// NotificationOpenWaDataSourceModel describes the data model for OpenWA notification data source.
type NotificationOpenWaDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*NotificationOpenWaDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_openwa"
}

// Schema returns the schema for the data source.
func (*NotificationOpenWaDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get OpenWA notification information by ID or name",
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
func (d *NotificationOpenWaDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *NotificationOpenWaDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationOpenWaDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !validateNotificationDataSourceInput(resp, data.ID, data.Name) {
		return
	}

	// Attempt to read by ID if provided.
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		d.readByID(ctx, &data, resp)
		return
	}

	// Attempt to read by name if ID not provided.
	d.readByName(ctx, &data, resp)
}

func (d *NotificationOpenWaDataSource) readByID(
	ctx context.Context,
	data *NotificationOpenWaDataSourceModel,
	resp *datasource.ReadResponse,
) {
	notif, found := readNotificationByID(
		ctx,
		d.client,
		data.ID.ValueInt64(),
		(notification.OpenWaDetails{}).Type(),
		"an OpenWA notification",
		&resp.Diagnostics,
	)
	if !found {
		return
	}

	if !dataSourceNameMatches(&resp.Diagnostics, "Notification", data.ID.ValueInt64(), data.Name, notif.Name) {
		return
	}

	data.Name = types.StringValue(notif.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *NotificationOpenWaDataSource) readByName(
	ctx context.Context,
	data *NotificationOpenWaDataSourceModel,
	resp *datasource.ReadResponse,
) {
	id, ok := findNotificationByName(
		ctx,
		d.client,
		data.Name.ValueString(),
		notification.OpenWaDetails{}.Type(),
		&resp.Diagnostics,
	)
	if !ok {
		return
	}

	data.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
