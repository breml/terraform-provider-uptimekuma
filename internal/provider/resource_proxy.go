package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/proxy"
)

var _ resource.Resource = &ProxyResource{}

// NewProxyResource returns a new instance of the proxy resource.
func NewProxyResource() resource.Resource {
	return &ProxyResource{}
}

// ProxyResource defines the resource implementation.
type ProxyResource struct {
	client *kuma.Client
}

// ProxyResourceModel describes the resource data model.
type ProxyResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Protocol      types.String `tfsdk:"protocol"`
	Host          types.String `tfsdk:"host"`
	Port          types.Int64  `tfsdk:"port"`
	Auth          types.Bool   `tfsdk:"auth"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	Active        types.Bool   `tfsdk:"active"`
	Default       types.Bool   `tfsdk:"default"`
	ApplyExisting types.Bool   `tfsdk:"apply_existing"`
}

// Metadata returns the metadata for the resource.
func (*ProxyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxy"
}

// Schema returns the schema for the resource.
func (*ProxyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Proxy resource for managing HTTP/HTTPS/SOCKS proxies in Uptime Kuma",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Proxy identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"protocol": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Proxy protocol (http, https, or socks5)",
				Validators: []validator.String{
					stringvalidator.OneOf("http", "https", "socks5"),
				},
			},
			"host": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Proxy server hostname or IP address",
			},
			"port": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Proxy server port",
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"auth": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether proxy authentication is required",
			},
			"username": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Username for proxy authentication",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Password for proxy authentication",
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable or disable the proxy",
			},
			"default": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Set as the default proxy",
			},
			"apply_existing": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Apply proxy to all existing monitors on creation",
			},
		},
	}
}

// Configure configures the resource with the API client.
func (r *ProxyResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = client
}

// Create creates a new resource.
func (r *ProxyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProxyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := proxy.Config{
		Protocol:      data.Protocol.ValueString(),
		Host:          data.Host.ValueString(),
		Port:          int(data.Port.ValueInt64()),
		Auth:          data.Auth.ValueBool(),
		Active:        data.Active.ValueBool(),
		Default:       data.Default.ValueBool(),
		ApplyExisting: data.ApplyExisting.ValueBool(),
	}

	if data.Auth.ValueBool() {
		if data.Username.IsNull() || data.Username.ValueString() == "" {
			resp.Diagnostics.AddError("validation error", "username is required when auth is enabled")
			return
		}

		if data.Password.IsNull() || data.Password.ValueString() == "" {
			resp.Diagnostics.AddError("validation error", "password is required when auth is enabled")
			return
		}

		p.Username = data.Username.ValueString()
		p.Password = data.Password.ValueString()
	}

	id, err := r.client.CreateProxy(ctx, p)
	if err != nil {
		resp.Diagnostics.AddError("failed to create proxy", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the resource.
func (r *ProxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProxyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.client.GetProxy(ctx, data.ID.ValueInt64())
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read proxy", err.Error())
		return
	}

	data.ID = types.Int64Value(p.ID)
	data.Protocol = types.StringValue(p.Protocol)
	data.Host = types.StringValue(p.Host)
	data.Port = types.Int64Value(int64(p.Port))
	data.Auth = types.BoolValue(p.Auth)
	data.Active = types.BoolValue(p.Active)
	data.Default = types.BoolValue(p.Default)
	// apply_existing is not returned by the API, so we preserve the state value
	// password is not set from API response to avoid storing plaintext password in state
	// password is preserved from Terraform state

	if p.Auth {
		data.Username = types.StringValue(p.Username)
		// Do not overwrite password from API response - preserve from state
	} else {
		data.Username = types.StringNull()
		data.Password = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *ProxyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProxyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := proxy.Config{
		ID:            data.ID.ValueInt64(),
		Protocol:      data.Protocol.ValueString(),
		Host:          data.Host.ValueString(),
		Port:          int(data.Port.ValueInt64()),
		Auth:          data.Auth.ValueBool(),
		Active:        data.Active.ValueBool(),
		Default:       data.Default.ValueBool(),
		ApplyExisting: data.ApplyExisting.ValueBool(),
	}

	if data.Auth.ValueBool() {
		if data.Username.IsNull() || data.Username.ValueString() == "" {
			resp.Diagnostics.AddError("validation error", "username is required when auth is enabled")
			return
		}

		if data.Password.IsNull() || data.Password.ValueString() == "" {
			resp.Diagnostics.AddError("validation error", "password is required when auth is enabled")
			return
		}

		p.Username = data.Username.ValueString()
		p.Password = data.Password.ValueString()
	}

	err := r.client.UpdateProxy(ctx, p)
	if err != nil {
		resp.Diagnostics.AddError("failed to update proxy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *ProxyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProxyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteProxy(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete proxy", err.Error())
		return
	}
}
