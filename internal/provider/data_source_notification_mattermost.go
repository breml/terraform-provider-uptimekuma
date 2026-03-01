package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &NotificationMattermostDataSource{}

// NewNotificationMattermostDataSource returns a new instance of the Mattermost notification data source.
func NewNotificationMattermostDataSource() datasource.DataSource {
	return &NotificationMattermostDataSource{}
}

// NotificationMattermostDataSource manages Mattermost notification data source operations.
type NotificationMattermostDataSource struct {
	client *kuma.Client
}

// NotificationMattermostDataSourceModel describes the data model for Mattermost notification data source.
type NotificationMattermostDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*NotificationMattermostDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_mattermost"
}

// Schema returns the schema for the data source.
func (*NotificationMattermostDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Mattermost notification information by ID or name",
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
func (d *NotificationMattermostDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *NotificationMattermostDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data NotificationMattermostDataSourceModel

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

func (d *NotificationMattermostDataSource) readByID(
	ctx context.Context,
	data *NotificationMattermostDataSourceModel,
	resp *datasource.ReadResponse,
) {
	notification, err := d.client.GetNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	if notification.Type() != "mattermost" {
		resp.Diagnostics.AddError("Incorrect notification type", "Notification is not a Mattermost notification")
		return
	}

	data.Name = types.StringValue(notification.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *NotificationMattermostDataSource) readByName(
	ctx context.Context,
	data *NotificationMattermostDataSourceModel,
	resp *datasource.ReadResponse,
) {
	id, ok := findNotificationByName(ctx, d.client, data.Name.ValueString(), "mattermost", &resp.Diagnostics)
	if !ok {
		return
	}

	data.ID = types.Int64Value(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
