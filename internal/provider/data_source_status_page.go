package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/statuspage"
)

var _ datasource.DataSource = &StatusPageDataSource{}

// NewStatusPageDataSource returns a new instance of the status page data source.
func NewStatusPageDataSource() datasource.DataSource {
	return &StatusPageDataSource{}
}

// StatusPageDataSource manages status page data source operations.
type StatusPageDataSource struct {
	client *kuma.Client
}

// StatusPageDataSourceModel describes the data model for status page data source.
type StatusPageDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Slug  types.String `tfsdk:"slug"`
	Title types.String `tfsdk:"title"`
}

// Metadata returns the metadata for the data source.
func (*StatusPageDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

// Schema returns the schema for the data source.
func (*StatusPageDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get status page information by ID or slug",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Status page identifier",
				Optional:            true,
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Status page slug (unique identifier)",
				Optional:            true,
				Computed:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Status page title",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *StatusPageDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *StatusPageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StatusPageDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Slug.IsNull() && !data.Slug.IsUnknown() {
		statusPage, err := d.client.GetStatusPage(ctx, data.Slug.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to read status page", err.Error())
			return
		}

		data.ID = types.Int64Value(statusPage.ID)
		data.Title = types.StringValue(statusPage.Title)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Attempt to read by ID if provided.
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		// The list serves from the state cache, so a resync is forced once
		// before a miss is believed, see findWithResync.
		match, found, err := findWithResync(ctx, d.client, func(ctx context.Context) (statuspage.StatusPage, bool, error) {
			statusPages, listErr := d.client.GetStatusPages(ctx)
			if listErr != nil {
				return statuspage.StatusPage{}, false, fmt.Errorf("get status pages: %w", listErr)
			}

			sp, ok := statusPages[data.ID.ValueInt64()]

			return sp, ok, nil
		})
		if err != nil {
			resp.Diagnostics.AddError("failed to read status pages", err.Error())

			return
		}

		if !found {
			reportMiss(
				&resp.Diagnostics, d.client, statusPageListEvent,
				"failed to read status page",
				fmt.Sprintf("No status page with ID %d found.", data.ID.ValueInt64()),
			)

			return
		}

		data.Slug = types.StringValue(match.Slug)
		data.Title = types.StringValue(match.Title)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

		return
	}

	resp.Diagnostics.AddError(
		// Error if neither ID nor name provided.
		"Missing query parameters",
		"Either 'id' or 'slug' must be specified.",
	)
}
