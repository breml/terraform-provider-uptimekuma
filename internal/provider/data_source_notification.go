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

var _ datasource.DataSource = &NotificationDataSource{}

// NewNotificationDataSource returns a new instance of the notification data source.
func NewNotificationDataSource() datasource.DataSource {
	return &NotificationDataSource{}
}

// NotificationDataSource manages notification data source operations.
type NotificationDataSource struct {
	client *kuma.Client
}

// NotificationDataSourceModel describes the data model for notification data source.
type NotificationDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

// Metadata returns the metadata for the data source.
func (*NotificationDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification"
}

// Schema returns the schema for the data source.
func (*NotificationDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get notification information by ID or name",
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
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *NotificationDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *NotificationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NotificationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Attempt to read by ID if provided.
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		d.readByID(ctx, &data, resp)
		return
	}

	// Attempt to read by name if ID not provided.
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		d.readByName(ctx, &data, resp)
		return
	}

	resp.Diagnostics.AddError(
		// Error if neither ID nor name provided.
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}

func (d *NotificationDataSource) readByID(
	ctx context.Context,
	data *NotificationDataSourceModel,
	resp *datasource.ReadResponse,
) {
	notif, found := readNotificationWithResync(ctx, d.client, data.ID.ValueInt64(), &resp.Diagnostics)
	if !found {
		return
	}

	data.Name = types.StringValue(notif.Name)
	data.Type = types.StringValue(notif.Type())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *NotificationDataSource) readByName(
	ctx context.Context,
	data *NotificationDataSourceModel,
	resp *datasource.ReadResponse,
) {
	// The getter serves from the state cache, so a resync is forced once
	// before a miss is believed, see findWithResync.
	matches, found := findWithResync(ctx, d.client, func(ctx context.Context) ([]notification.Base, bool) {
		var matched []notification.Base

		notifications := d.client.GetNotifications(ctx)
		for _, notif := range notifications {
			if notif.Name == data.Name.ValueString() {
				matched = append(matched, notif)
			}
		}

		return matched, len(matched) > 0
	}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Error if no matching item found.
	if !found {
		resp.Diagnostics.AddError(
			"Notification not found",
			fmt.Sprintf("No notification with name '%s' found.", data.Name.ValueString()),
		)

		return
	}

	// Error if multiple matches found.
	if len(matches) > 1 {
		resp.Diagnostics.AddError(
			"Multiple notifications found",
			fmt.Sprintf(
				"Multiple notifications with name '%s' found. Please use 'id' to specify the notification uniquely.",
				data.Name.ValueString(),
			),
		)

		return
	}

	data.ID = types.Int64Value(matches[0].GetID())
	data.Type = types.StringValue(matches[0].Type())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
