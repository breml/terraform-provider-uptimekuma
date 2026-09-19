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

var _ datasource.DataSource = &NotificationSMSIRDataSource{}

// NewNotificationSMSIRDataSource returns a new instance of the SMS.ir notification data source.
func NewNotificationSMSIRDataSource() datasource.DataSource {
	return &NotificationSMSIRDataSource{}
}

// NotificationSMSIRDataSource manages SMS.ir notification data source operations.
type NotificationSMSIRDataSource struct {
	client *kuma.Client
}

// NotificationSMSIRDataSourceModel describes the data model for SMS.ir notification data source.
type NotificationSMSIRDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*NotificationSMSIRDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_smsir"
}

// Schema returns the schema for the data source.
func (*NotificationSMSIRDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get SMS.ir notification information by ID or name",
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
func (d *NotificationSMSIRDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *NotificationSMSIRDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationSMSIRDataSourceModel

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

func (d *NotificationSMSIRDataSource) readByID(
	ctx context.Context,
	data *NotificationSMSIRDataSourceModel,
	resp *datasource.ReadResponse,
) {
	notif, err := d.client.GetNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	if notif.Type() != (notification.SMSIRDetails{}).Type() {
		resp.Diagnostics.AddError(
			"incorrect notification type",
			fmt.Sprintf(
				"notification with ID %d has type %q, expected %q",
				data.ID.ValueInt64(),
				notif.Type(),
				notification.SMSIRDetails{}.Type(),
			),
		)
		return
	}

	data.Name = types.StringValue(notif.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *NotificationSMSIRDataSource) readByName(
	ctx context.Context,
	data *NotificationSMSIRDataSourceModel,
	resp *datasource.ReadResponse,
) {
	id, ok := findNotificationByName(
		ctx,
		d.client,
		data.Name.ValueString(),
		notification.SMSIRDetails{}.Type(),
		&resp.Diagnostics,
	)
	if !ok {
		return
	}

	data.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
