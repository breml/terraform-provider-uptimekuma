package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var _ datasource.DataSource = &NotificationSMSGatewayDataSource{}

// NewNotificationSMSGatewayDataSource returns a new instance of the SMS Gateway notification data source.
func NewNotificationSMSGatewayDataSource() datasource.DataSource {
	return &NotificationSMSGatewayDataSource{}
}

// NotificationSMSGatewayDataSource manages SMS Gateway notification data source operations.
type NotificationSMSGatewayDataSource struct {
	client *kuma.Client
}

// NotificationSMSGatewayDataSourceModel describes the data model for SMS Gateway notification data source.
type NotificationSMSGatewayDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*NotificationSMSGatewayDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_smsgateway"
}

// Schema returns the schema for the data source.
func (*NotificationSMSGatewayDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get SMS Gateway notification information by ID or name",
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
func (d *NotificationSMSGatewayDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *NotificationSMSGatewayDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationSMSGatewayDataSourceModel

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

func (d *NotificationSMSGatewayDataSource) readByID(
	ctx context.Context,
	data *NotificationSMSGatewayDataSourceModel,
	resp *datasource.ReadResponse,
) {
	notif, found := readNotificationByID(
		ctx,
		d.client,
		data.ID.ValueInt64(),
		(notification.SMSGatewayDetails{}).Type(),
		"an SMS Gateway notification",
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

func (d *NotificationSMSGatewayDataSource) readByName(
	ctx context.Context,
	data *NotificationSMSGatewayDataSourceModel,
	resp *datasource.ReadResponse,
) {
	id, ok := findNotificationByName(
		ctx,
		d.client,
		data.Name.ValueString(),
		notification.SMSGatewayDetails{}.Type(),
		&resp.Diagnostics,
	)
	if !ok {
		return
	}

	data.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
