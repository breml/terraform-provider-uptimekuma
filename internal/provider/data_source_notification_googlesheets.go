package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var _ datasource.DataSource = &NotificationGoogleSheetsDataSource{}

// NewNotificationGoogleSheetsDataSource returns a new instance of the Google Sheets notification data source.
func NewNotificationGoogleSheetsDataSource() datasource.DataSource {
	return &NotificationGoogleSheetsDataSource{}
}

// NotificationGoogleSheetsDataSource manages Google Sheets notification data source operations.
type NotificationGoogleSheetsDataSource struct {
	client *kuma.Client
}

// NotificationGoogleSheetsDataSourceModel describes the data model for Google Sheets notification data source.
type NotificationGoogleSheetsDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*NotificationGoogleSheetsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_googlesheets"
}

// Schema returns the schema for the data source.
func (*NotificationGoogleSheetsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Google Sheets notification information by ID or name",
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
func (d *NotificationGoogleSheetsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *NotificationGoogleSheetsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationGoogleSheetsDataSourceModel

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

func (d *NotificationGoogleSheetsDataSource) readByID(
	ctx context.Context,
	data *NotificationGoogleSheetsDataSourceModel,
	resp *datasource.ReadResponse,
) {
	notif, found := readNotificationWithResync(ctx, d.client, data.ID.ValueInt64(), &resp.Diagnostics)
	if !found {
		return
	}

	if notif.Type() != (notification.GoogleSheetsDetails{}).Type() {
		resp.Diagnostics.AddError(
			"incorrect notification type",
			fmt.Sprintf(
				"notification with ID %d has type %q, expected %q",
				data.ID.ValueInt64(),
				notif.Type(),
				notification.GoogleSheetsDetails{}.Type(),
			),
		)
		return
	}

	data.Name = types.StringValue(notif.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *NotificationGoogleSheetsDataSource) readByName(
	ctx context.Context,
	data *NotificationGoogleSheetsDataSourceModel,
	resp *datasource.ReadResponse,
) {
	id, ok := findNotificationByName(
		ctx,
		d.client,
		data.Name.ValueString(),
		notification.GoogleSheetsDetails{}.Type(),
		&resp.Diagnostics,
	)
	if !ok {
		return
	}

	data.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
