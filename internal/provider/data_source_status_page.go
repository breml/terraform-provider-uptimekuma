package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &StatusPageDataSource{}

func NewStatusPageDataSource() datasource.DataSource {
	return &StatusPageDataSource{}
}

type StatusPageDataSource struct {
	client *kuma.Client
}

type StatusPageDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Slug  types.String `tfsdk:"slug"`
	Title types.String `tfsdk:"title"`
}

func (d *StatusPageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (d *StatusPageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *StatusPageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *kuma.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

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

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		statusPages, err := d.client.GetStatusPages(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read status pages", err.Error())
			return
		}
		for id, sp := range statusPages {
			if id == data.ID.ValueInt64() {
				data.Slug = types.StringValue(sp.Slug)
				data.Title = types.StringValue(sp.Title)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
		}
		resp.Diagnostics.AddError("failed to read status page", "Status page not found")
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'slug' must be specified.",
	)
}
