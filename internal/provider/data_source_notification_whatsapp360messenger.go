package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &NotificationWhatsapp360messengerDataSource{}

// NewNotificationWhatsapp360messengerDataSource returns a new instance of the WhatsApp 360messenger
// notification data source.
func NewNotificationWhatsapp360messengerDataSource() datasource.DataSource {
	return &NotificationWhatsapp360messengerDataSource{}
}

// NotificationWhatsapp360messengerDataSource manages WhatsApp 360messenger notification data source operations.
type NotificationWhatsapp360messengerDataSource struct {
	client *kuma.Client
}

// NotificationWhatsapp360messengerDataSourceModel describes the data model for the WhatsApp 360messenger
// notification data source.
type NotificationWhatsapp360messengerDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*NotificationWhatsapp360messengerDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_whatsapp360messenger"
}

// Schema returns the schema for the data source.
func (*NotificationWhatsapp360messengerDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get WhatsApp 360messenger notification information by ID or name",
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
func (d *NotificationWhatsapp360messengerDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *NotificationWhatsapp360messengerDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationWhatsapp360messengerDataSourceModel

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

func (d *NotificationWhatsapp360messengerDataSource) readByID(
	ctx context.Context,
	data *NotificationWhatsapp360messengerDataSourceModel,
	resp *datasource.ReadResponse,
) {
	notification, err := d.client.GetNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	if notification.Type() != "Whatsapp360messenger" {
		resp.Diagnostics.AddError(
			"incorrect notification type",
			fmt.Sprintf(
				"notification with ID %d has type %q, expected \"Whatsapp360messenger\"",
				data.ID.ValueInt64(),
				notification.Type(),
			),
		)
		return
	}

	data.Name = types.StringValue(notification.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *NotificationWhatsapp360messengerDataSource) readByName(
	ctx context.Context,
	data *NotificationWhatsapp360messengerDataSourceModel,
	resp *datasource.ReadResponse,
) {
	id, ok := findNotificationByName(ctx, d.client, data.Name.ValueString(), "Whatsapp360messenger", &resp.Diagnostics)
	if !ok {
		return
	}

	data.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
