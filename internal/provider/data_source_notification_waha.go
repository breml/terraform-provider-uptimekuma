package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &NotificationWAHADataSource{}

// NewNotificationWAHADataSource returns a new instance of the WAHA notification data source.
func NewNotificationWAHADataSource() datasource.DataSource {
	return &NotificationWAHADataSource{}
}

// NotificationWAHADataSource manages WAHA notification data source operations.
type NotificationWAHADataSource struct {
	client *kuma.Client
}

// NotificationWAHADataSourceModel describes the data model for WAHA notification data source.
type NotificationWAHADataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*NotificationWAHADataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_waha"
}

// Schema returns the schema for the data source.
func (*NotificationWAHADataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get WAHA notification information by ID or name",
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
func (d *NotificationWAHADataSource) Configure(
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
func (d *NotificationWAHADataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationWAHADataSourceModel

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

func (d *NotificationWAHADataSource) readByID(
	ctx context.Context,
	data *NotificationWAHADataSourceModel,
	resp *datasource.ReadResponse,
) {
	notification, err := d.client.GetNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	if notification.Type() != "waha" {
		resp.Diagnostics.AddError("Incorrect notification type", "Notification is not a WAHA notification")
		return
	}

	data.Name = types.StringValue(notification.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *NotificationWAHADataSource) readByName(
	ctx context.Context,
	data *NotificationWAHADataSourceModel,
	resp *datasource.ReadResponse,
) {
	id, ok := findNotificationByName(ctx, d.client, data.Name.ValueString(), "waha", &resp.Diagnostics)
	if !ok {
		return
	}

	data.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
