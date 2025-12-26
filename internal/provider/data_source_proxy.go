package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &ProxyDataSource{}

func NewProxyDataSource() datasource.DataSource {
	return &ProxyDataSource{}
}

type ProxyDataSource struct {
	client *kuma.Client
}

type ProxyDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Protocol types.String `tfsdk:"protocol"`
}

func (d *ProxyDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_proxy"
}

func (d *ProxyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get proxy information by ID",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Proxy identifier",
				Required:            true,
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "Proxy host",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Proxy port",
				Computed:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Proxy protocol (http, https, socks5)",
				Computed:            true,
			},
		},
	}
}

func (d *ProxyDataSource) Configure(
	ctx context.Context,
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

func (d *ProxyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProxyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy, err := d.client.GetProxy(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read proxy", err.Error())
		return
	}

	data.Host = types.StringValue(proxy.Host)
	data.Port = types.Int64Value(int64(proxy.Port))
	data.Protocol = types.StringValue(proxy.Protocol)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
