// Package provider implements the Uptime Kuma Terraform provider.
// This file provides base HTTP monitor schema and utilities.
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

// MonitorHTTPBaseModel describes the base data model for HTTP-based monitor types.
// This includes network config, authentication (Basic, NTLM, OAuth), and TLS settings.
type MonitorHTTPBaseModel struct {
	URL                 types.String `tfsdk:"url"`                   // HTTP(S) endpoint URL to monitor.
	Timeout             types.Int64  `tfsdk:"timeout"`               // Request timeout in seconds.
	Method              types.String `tfsdk:"method"`                // HTTP method (GET, POST, etc).
	ExpiryNotification  types.Bool   `tfsdk:"expiry_notification"`   // Notify on certificate expiry.
	IgnoreTLS           types.Bool   `tfsdk:"ignore_tls"`            // Skip TLS/SSL certificate validation.
	MaxRedirects        types.Int64  `tfsdk:"max_redirects"`         // Maximum HTTP redirects to follow.
	AcceptedStatusCodes types.List   `tfsdk:"accepted_status_codes"` // HTTP status codes to treat as success.
	ProxyID             types.Int64  `tfsdk:"proxy_id"`              // Optional proxy ID for routing requests.
	HTTPBodyEncoding    types.String `tfsdk:"http_body_encoding"`    // Encoding for request body.
	Body                types.String `tfsdk:"body"`                  // Request body for POST/PUT methods.
	Headers             types.String `tfsdk:"headers"`               // Custom HTTP headers as JSON.
	AuthMethod          types.String `tfsdk:"auth_method"`           // Authentication method (basic, digest, ntlm, oauth).
	BasicAuthUser       types.String `tfsdk:"basic_auth_user"`       // Basic auth username.
	BasicAuthPass       types.String `tfsdk:"basic_auth_pass"`       // Basic auth password.
	AuthDomain          types.String `tfsdk:"auth_domain"`           // Domain for NTLM authentication.
	AuthWorkstation     types.String `tfsdk:"auth_workstation"`      // Workstation for NTLM authentication.
	TLSCert             types.String `tfsdk:"tls_cert"`              // Client TLS certificate in PEM format.
	TLSKey              types.String `tfsdk:"tls_key"`               // Client TLS key in PEM format.
	TLSCa               types.String `tfsdk:"tls_ca"`                // CA certificate for server verification.
	OAuthAuthMethod     types.String `tfsdk:"oauth_auth_method"`     // OAuth authentication method.
	OAuthTokenURL       types.String `tfsdk:"oauth_token_url"`       // OAuth token endpoint URL.
	OAuthClientID       types.String `tfsdk:"oauth_client_id"`       // OAuth client ID.
	OAuthClientSecret   types.String `tfsdk:"oauth_client_secret"`   // OAuth client secret.
	OAuthScopes         types.String `tfsdk:"oauth_scopes"`          // OAuth scopes to request.
}

// withHTTPMonitorBaseAttributes adds HTTP-specific schema attributes to the provided attribute map.
// This includes URL, timeout, method, authentication, TLS, and OAuth configuration options.
func withHTTPMonitorBaseAttributes(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	attrs["url"] = httpURLAttribute()
	attrs["timeout"] = httpTimeoutAttribute()
	attrs["method"] = httpMethodAttribute()
	attrs["expiry_notification"] = httpExpiryNotificationAttribute()
	attrs["ignore_tls"] = httpIgnoreTLSAttribute()
	attrs["max_redirects"] = httpMaxRedirectsAttribute()
	attrs["accepted_status_codes"] = httpAcceptedStatusCodesAttribute()
	attrs["proxy_id"] = httpProxyIDAttribute()
	attrs["http_body_encoding"] = httpBodyEncodingAttribute()
	attrs["body"] = httpBodyAttribute()
	attrs["headers"] = httpHeadersAttribute()
	attrs["auth_method"] = httpAuthMethodAttribute()
	attrs["basic_auth_user"] = httpBasicAuthUserAttribute()
	attrs["basic_auth_pass"] = httpBasicAuthPassAttribute()
	attrs["auth_domain"] = httpAuthDomainAttribute()
	attrs["auth_workstation"] = httpAuthWorkstationAttribute()
	attrs["tls_cert"] = httpTLSCertAttribute()
	attrs["tls_key"] = httpTLSKeyAttribute()
	attrs["tls_ca"] = httpTLSCAAttribute()
	attrs["oauth_auth_method"] = httpOAuthAuthMethodAttribute()
	attrs["oauth_token_url"] = httpOAuthTokenURLAttribute()
	attrs["oauth_client_id"] = httpOAuthClientIDAttribute()
	attrs["oauth_client_secret"] = httpOAuthClientSecretAttribute()
	attrs["oauth_scopes"] = httpOAuthScopesAttribute()
	return attrs
}

func httpURLAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "URL to monitor",
		Required:            true,
	}
}

func httpTimeoutAttribute() schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: "Request timeout in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(48),
		Validators: []validator.Int64{
			int64validator.Between(1, 3600),
		},
	}
}

func httpMethodAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "HTTP method",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("GET"),
		Validators: []validator.String{
			stringvalidator.OneOf("GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"),
		},
	}
}

func httpExpiryNotificationAttribute() schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: "Enable certificate expiry notification",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	}
}

func httpIgnoreTLSAttribute() schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: "Ignore TLS/SSL errors",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	}
}

func httpMaxRedirectsAttribute() schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: "Maximum number of redirects to follow",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(10),
		Validators: []validator.Int64{
			int64validator.Between(0, 20),
		},
	}
}

func httpAcceptedStatusCodesAttribute() schema.ListAttribute {
	return schema.ListAttribute{
		MarkdownDescription: "Accepted HTTP status codes (e.g., ['200-299', '301'])",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Default: listdefault.StaticValue(
			types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200-299")}),
		),
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
	}
}

func httpProxyIDAttribute() schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: "Proxy ID",
		Optional:            true,
	}
}

func httpBodyEncodingAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "HTTP body encoding",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("json"),
	}
}

func httpBodyAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Request body",
		Optional:            true,
	}
}

func httpHeadersAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Request headers (JSON format)",
		Optional:            true,
	}
}

func httpAuthMethodAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Authentication method",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(""),
		Validators: []validator.String{
			stringvalidator.OneOf("", "basic", "ntlm", "mtls", "oauth2-cc"),
		},
	}
}

func httpBasicAuthUserAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Basic authentication username",
		Optional:            true,
	}
}

func httpBasicAuthPassAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Basic authentication password",
		Optional:            true,
		Sensitive:           true,
	}
}

func httpAuthDomainAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "NTLM authentication domain",
		Optional:            true,
	}
}

func httpAuthWorkstationAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "NTLM authentication workstation",
		Optional:            true,
	}
}

func httpTLSCertAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "TLS client certificate",
		Optional:            true,
		Sensitive:           true,
	}
}

func httpTLSKeyAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "TLS client key",
		Optional:            true,
		Sensitive:           true,
	}
}

func httpTLSCAAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "TLS CA certificate",
		Optional:            true,
	}
}

func httpOAuthAuthMethodAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "OAuth authentication method",
		Optional:            true,
	}
}

func httpOAuthTokenURLAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "OAuth token URL",
		Optional:            true,
	}
}

func httpOAuthClientIDAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "OAuth client ID",
		Optional:            true,
	}
}

func httpOAuthClientSecretAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "OAuth client secret",
		Optional:            true,
		Sensitive:           true,
	}
}

func httpOAuthScopesAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "OAuth scopes",
		Optional:            true,
	}
}
