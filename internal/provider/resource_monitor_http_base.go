package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MonitorHTTPBaseModel struct {
	URL                 types.String `tfsdk:"url"`
	Timeout             types.Int64  `tfsdk:"timeout"`
	Method              types.String `tfsdk:"method"`
	ExpiryNotification  types.Bool   `tfsdk:"expiry_notification"`
	IgnoreTLS           types.Bool   `tfsdk:"ignore_tls"`
	MaxRedirects        types.Int64  `tfsdk:"max_redirects"`
	AcceptedStatusCodes types.List   `tfsdk:"accepted_status_codes"`
	ProxyID             types.Int64  `tfsdk:"proxy_id"`
	HTTPBodyEncoding    types.String `tfsdk:"http_body_encoding"`
	Body                types.String `tfsdk:"body"`
	Headers             types.String `tfsdk:"headers"`
	AuthMethod          types.String `tfsdk:"auth_method"`
	BasicAuthUser       types.String `tfsdk:"basic_auth_user"`
	BasicAuthPass       types.String `tfsdk:"basic_auth_pass"`
	AuthDomain          types.String `tfsdk:"auth_domain"`
	AuthWorkstation     types.String `tfsdk:"auth_workstation"`
	TLSCert             types.String `tfsdk:"tls_cert"`
	TLSKey              types.String `tfsdk:"tls_key"`
	TLSCa               types.String `tfsdk:"tls_ca"`
	OAuthAuthMethod     types.String `tfsdk:"oauth_auth_method"`
	OAuthTokenURL       types.String `tfsdk:"oauth_token_url"`
	OAuthClientID       types.String `tfsdk:"oauth_client_id"`
	OAuthClientSecret   types.String `tfsdk:"oauth_client_secret"`
	OAuthScopes         types.String `tfsdk:"oauth_scopes"`
}

func withHTTPMonitorBaseAttributes(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	attrs["url"] = schema.StringAttribute{
		MarkdownDescription: "URL to monitor",
		Required:            true,
	}

	attrs["timeout"] = schema.Int64Attribute{
		MarkdownDescription: "Request timeout in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(48),
		Validators: []validator.Int64{
			int64validator.Between(1, 3600),
		},
	}

	attrs["method"] = schema.StringAttribute{
		MarkdownDescription: "HTTP method",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("GET"),
		Validators: []validator.String{
			stringvalidator.OneOf("GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"),
		},
	}

	attrs["expiry_notification"] = schema.BoolAttribute{
		MarkdownDescription: "Enable certificate expiry notification",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	}

	attrs["ignore_tls"] = schema.BoolAttribute{
		MarkdownDescription: "Ignore TLS/SSL errors",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	}

	attrs["max_redirects"] = schema.Int64Attribute{
		MarkdownDescription: "Maximum number of redirects to follow",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(10),
		Validators: []validator.Int64{
			int64validator.Between(0, 20),
		},
	}

	attrs["accepted_status_codes"] = schema.ListAttribute{
		MarkdownDescription: "Accepted HTTP status codes (e.g., ['200-299', '301'])",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200-299")})),
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
	}

	attrs["proxy_id"] = schema.Int64Attribute{
		MarkdownDescription: "Proxy ID",
		Optional:            true,
	}

	attrs["http_body_encoding"] = schema.StringAttribute{
		MarkdownDescription: "HTTP body encoding",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("json"),
	}

	attrs["body"] = schema.StringAttribute{
		MarkdownDescription: "Request body",
		Optional:            true,
	}

	attrs["headers"] = schema.StringAttribute{
		MarkdownDescription: "Request headers (JSON format)",
		Optional:            true,
	}

	attrs["auth_method"] = schema.StringAttribute{
		MarkdownDescription: "Authentication method",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(""),
		Validators: []validator.String{
			stringvalidator.OneOf("", "basic", "ntlm", "mtls", "oauth2-cc"),
		},
	}

	attrs["basic_auth_user"] = schema.StringAttribute{
		MarkdownDescription: "Basic authentication username",
		Optional:            true,
	}

	attrs["basic_auth_pass"] = schema.StringAttribute{
		MarkdownDescription: "Basic authentication password",
		Optional:            true,
		Sensitive:           true,
	}

	attrs["auth_domain"] = schema.StringAttribute{
		MarkdownDescription: "NTLM authentication domain",
		Optional:            true,
	}

	attrs["auth_workstation"] = schema.StringAttribute{
		MarkdownDescription: "NTLM authentication workstation",
		Optional:            true,
	}

	attrs["tls_cert"] = schema.StringAttribute{
		MarkdownDescription: "TLS client certificate",
		Optional:            true,
		Sensitive:           true,
	}

	attrs["tls_key"] = schema.StringAttribute{
		MarkdownDescription: "TLS client key",
		Optional:            true,
		Sensitive:           true,
	}

	attrs["tls_ca"] = schema.StringAttribute{
		MarkdownDescription: "TLS CA certificate",
		Optional:            true,
	}

	attrs["oauth_auth_method"] = schema.StringAttribute{
		MarkdownDescription: "OAuth authentication method",
		Optional:            true,
	}

	attrs["oauth_token_url"] = schema.StringAttribute{
		MarkdownDescription: "OAuth token URL",
		Optional:            true,
	}

	attrs["oauth_client_id"] = schema.StringAttribute{
		MarkdownDescription: "OAuth client ID",
		Optional:            true,
	}

	attrs["oauth_client_secret"] = schema.StringAttribute{
		MarkdownDescription: "OAuth client secret",
		Optional:            true,
		Sensitive:           true,
	}

	attrs["oauth_scopes"] = schema.StringAttribute{
		MarkdownDescription: "OAuth scopes",
		Optional:            true,
	}

	return attrs
}
