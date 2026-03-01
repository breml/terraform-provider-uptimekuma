package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &NotificationPagerDutyDataSource{}

// NewNotificationPagerDutyDataSource returns a new instance of the PagerDuty notification data source.
func NewNotificationPagerDutyDataSource() datasource.DataSource {
	return &NotificationPagerDutyDataSource{}
}

// NotificationPagerDutyDataSource manages PagerDuty notification data source operations.
type NotificationPagerDutyDataSource struct {
	client *kuma.Client
}

// NotificationPagerDutyDataSourceModel describes the data model for PagerDuty notification data source.
type NotificationPagerDutyDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*NotificationPagerDutyDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_pagerduty"
}

// Schema returns the schema for the data source.
func (*NotificationPagerDutyDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get PagerDuty notification information by ID or name",
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
func (d *NotificationPagerDutyDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *NotificationPagerDutyDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationPagerDutyDataSourceModel

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

func (d *NotificationPagerDutyDataSource) readByID(
	ctx context.Context,
	data *NotificationPagerDutyDataSourceModel,
	resp *datasource.ReadResponse,
) {
	notification, err := d.client.GetNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	if notification.Type() != "PagerDuty" {
		resp.Diagnostics.AddError("Incorrect notification type", "Notification is not a PagerDuty notification")
		return
	}

	data.Name = types.StringValue(notification.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *NotificationPagerDutyDataSource) readByName(
	ctx context.Context,
	data *NotificationPagerDutyDataSourceModel,
	resp *datasource.ReadResponse,
) {
	id, ok := findNotificationByName(ctx, d.client, data.Name.ValueString(), "PagerDuty", &resp.Diagnostics)
	if !ok {
		return
	}

	data.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
