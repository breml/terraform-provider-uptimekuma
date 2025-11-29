package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// Ensure UptimeKumaProvider satisfies various provider interfaces.
var (
	_ provider.Provider = &UptimeKumaProvider{}
)

// UptimeKumaProvider defines the provider implementation.
type UptimeKumaProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// UptimeKumaProviderModel describes the provider data model.
type UptimeKumaProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *UptimeKumaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "uptimekuma"
	resp.Version = p.version
}

func (p *UptimeKumaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Uptime Kuma endpoint",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Uptime Kuma username",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Uptime Kuma password",
				Optional:            true,
			},
		},
	}
}

func (p *UptimeKumaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data UptimeKumaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Endpoint.IsNull() {
		resp.Diagnostics.AddError("endpoint nil", "endpoint is required")
	}

	if data.Endpoint.IsNull() {
		resp.Diagnostics.AddError("username nil", "username is required")
	}

	if data.Endpoint.IsNull() {
		resp.Diagnostics.AddError("password nil", "password is required")
	}

	// Uptime Kuma client configuration for data sources and resources
	// We can not use the context from Terraform, since it gets cancelled too early.
	client, err := kuma.New(context.Background(), data.Endpoint.ValueString(), data.Username.ValueString(), data.Password.ValueString(), kuma.WithLogLevel(kuma.LogLevelDebug))
	if err != nil {
		resp.Diagnostics.AddError("failed to create client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *UptimeKumaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNotificationResource,
		NewNotificationNtfyResource,
		NewNotificationSlackResource,
		NewNotificationTeamsResource,
		NewMonitorHTTPResource,
		NewMonitorHTTPKeywordResource,
		NewMonitorGroupResource,
		NewMonitorPingResource,
		NewMonitorDNSResource,
		NewMonitorPushResource,
		NewMonitorRealBrowserResource,
	}
}

func (p *UptimeKumaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UptimeKumaProvider{
			version: version,
		}
	}
}
