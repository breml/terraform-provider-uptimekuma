package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/breml/go-uptime-kuma-client/monitor"
)

// populateHTTPMonitorBaseFields populates the shared HTTP monitor base fields.
func populateHTTPMonitorBaseFields(
	_ context.Context,
	httpMonitor *monitor.HTTP,
	data any,
	_ *diag.Diagnostics,
) {
	switch m := data.(type) {
	case *MonitorHTTPResourceModel:
		populateHTTPBaseFieldsForHTTP(httpMonitor, m)
	case *MonitorHTTPJSONQueryResourceModel:
		populateHTTPBaseFieldsForJSONQuery(httpMonitor, m)
	case *MonitorHTTPKeywordResourceModel:
		populateHTTPBaseFieldsForKeyword(httpMonitor, m)
	}
}

// populateHTTPBaseFieldsForHTTP populates base fields for HTTP monitor.
func populateHTTPBaseFieldsForHTTP(httpMonitor *monitor.HTTP, m *MonitorHTTPResourceModel) {
	m.Name = types.StringValue(httpMonitor.Name)
	if httpMonitor.Description != nil {
		m.Description = types.StringValue(*httpMonitor.Description)
	} else {
		m.Description = types.StringNull()
	}
	m.Interval = types.Int64Value(httpMonitor.Interval)
	m.RetryInterval = types.Int64Value(httpMonitor.RetryInterval)
	m.ResendInterval = types.Int64Value(httpMonitor.ResendInterval)
	m.MaxRetries = types.Int64Value(httpMonitor.MaxRetries)
	m.UpsideDown = types.BoolValue(httpMonitor.UpsideDown)
	m.Active = types.BoolValue(httpMonitor.IsActive)
	m.URL = types.StringValue(httpMonitor.URL)
	m.Timeout = types.Int64Value(httpMonitor.Timeout)
	m.Method = types.StringValue(httpMonitor.Method)
	m.ExpiryNotification = types.BoolValue(httpMonitor.ExpiryNotification)
	m.IgnoreTLS = types.BoolValue(httpMonitor.IgnoreTLS)
	m.MaxRedirects = types.Int64Value(int64(httpMonitor.MaxRedirects))
	m.HTTPBodyEncoding = types.StringValue(httpMonitor.HTTPBodyEncoding)
	m.Body = stringOrNull(httpMonitor.Body)
	m.Headers = stringOrNull(httpMonitor.Headers)
	m.AuthMethod = types.StringValue(string(httpMonitor.AuthMethod))
	m.BasicAuthUser = stringOrNull(httpMonitor.BasicAuthUser)
	m.BasicAuthPass = stringOrNull(httpMonitor.BasicAuthPass)
	m.AuthDomain = stringOrNull(httpMonitor.AuthDomain)
	m.AuthWorkstation = stringOrNull(httpMonitor.AuthWorkstation)
	m.TLSCert = stringOrNull(httpMonitor.TLSCert)
	m.TLSKey = stringOrNull(httpMonitor.TLSKey)
	m.TLSCa = stringOrNull(httpMonitor.TLSCa)
	m.OAuthAuthMethod = stringOrNull(httpMonitor.OAuthAuthMethod)
	m.OAuthTokenURL = stringOrNull(httpMonitor.OAuthTokenURL)
	m.OAuthClientID = stringOrNull(httpMonitor.OAuthClientID)
	m.OAuthClientSecret = stringOrNull(httpMonitor.OAuthClientSecret)
	m.OAuthScopes = stringOrNull(httpMonitor.OAuthScopes)
}

// populateHTTPBaseFieldsForJSONQuery populates base fields for HTTP JSON Query monitor.
func populateHTTPBaseFieldsForJSONQuery(httpMonitor *monitor.HTTP, m *MonitorHTTPJSONQueryResourceModel) {
	m.Name = types.StringValue(httpMonitor.Name)
	if httpMonitor.Description != nil {
		m.Description = types.StringValue(*httpMonitor.Description)
	} else {
		m.Description = types.StringNull()
	}
	m.Interval = types.Int64Value(httpMonitor.Interval)
	m.RetryInterval = types.Int64Value(httpMonitor.RetryInterval)
	m.ResendInterval = types.Int64Value(httpMonitor.ResendInterval)
	m.MaxRetries = types.Int64Value(httpMonitor.MaxRetries)
	m.UpsideDown = types.BoolValue(httpMonitor.UpsideDown)
	m.Active = types.BoolValue(httpMonitor.IsActive)
	m.URL = types.StringValue(httpMonitor.URL)
	m.Timeout = types.Int64Value(httpMonitor.Timeout)
	m.Method = types.StringValue(httpMonitor.Method)
	m.ExpiryNotification = types.BoolValue(httpMonitor.ExpiryNotification)
	m.IgnoreTLS = types.BoolValue(httpMonitor.IgnoreTLS)
	m.MaxRedirects = types.Int64Value(int64(httpMonitor.MaxRedirects))
	m.HTTPBodyEncoding = types.StringValue(httpMonitor.HTTPBodyEncoding)
	m.Body = stringOrNull(httpMonitor.Body)
	m.Headers = stringOrNull(httpMonitor.Headers)
	m.AuthMethod = types.StringValue(string(httpMonitor.AuthMethod))
	m.BasicAuthUser = stringOrNull(httpMonitor.BasicAuthUser)
	m.BasicAuthPass = stringOrNull(httpMonitor.BasicAuthPass)
	m.AuthDomain = stringOrNull(httpMonitor.AuthDomain)
	m.AuthWorkstation = stringOrNull(httpMonitor.AuthWorkstation)
	m.TLSCert = stringOrNull(httpMonitor.TLSCert)
	m.TLSKey = stringOrNull(httpMonitor.TLSKey)
	m.TLSCa = stringOrNull(httpMonitor.TLSCa)
	m.OAuthAuthMethod = stringOrNull(httpMonitor.OAuthAuthMethod)
	m.OAuthTokenURL = stringOrNull(httpMonitor.OAuthTokenURL)
	m.OAuthClientID = stringOrNull(httpMonitor.OAuthClientID)
	m.OAuthClientSecret = stringOrNull(httpMonitor.OAuthClientSecret)
	m.OAuthScopes = stringOrNull(httpMonitor.OAuthScopes)
}

// populateHTTPBaseFieldsForKeyword populates base fields for HTTP Keyword monitor.
func populateHTTPBaseFieldsForKeyword(httpMonitor *monitor.HTTP, m *MonitorHTTPKeywordResourceModel) {
	m.Name = types.StringValue(httpMonitor.Name)
	if httpMonitor.Description != nil {
		m.Description = types.StringValue(*httpMonitor.Description)
	} else {
		m.Description = types.StringNull()
	}
	m.Interval = types.Int64Value(httpMonitor.Interval)
	m.RetryInterval = types.Int64Value(httpMonitor.RetryInterval)
	m.ResendInterval = types.Int64Value(httpMonitor.ResendInterval)
	m.MaxRetries = types.Int64Value(httpMonitor.MaxRetries)
	m.UpsideDown = types.BoolValue(httpMonitor.UpsideDown)
	m.Active = types.BoolValue(httpMonitor.IsActive)
	m.URL = types.StringValue(httpMonitor.URL)
	m.Timeout = types.Int64Value(httpMonitor.Timeout)
	m.Method = types.StringValue(httpMonitor.Method)
	m.ExpiryNotification = types.BoolValue(httpMonitor.ExpiryNotification)
	m.IgnoreTLS = types.BoolValue(httpMonitor.IgnoreTLS)
	m.MaxRedirects = types.Int64Value(int64(httpMonitor.MaxRedirects))
	m.HTTPBodyEncoding = types.StringValue(httpMonitor.HTTPBodyEncoding)
	m.Body = stringOrNull(httpMonitor.Body)
	m.Headers = stringOrNull(httpMonitor.Headers)
	m.AuthMethod = types.StringValue(string(httpMonitor.AuthMethod))
	m.BasicAuthUser = stringOrNull(httpMonitor.BasicAuthUser)
	m.BasicAuthPass = stringOrNull(httpMonitor.BasicAuthPass)
	m.AuthDomain = stringOrNull(httpMonitor.AuthDomain)
	m.AuthWorkstation = stringOrNull(httpMonitor.AuthWorkstation)
	m.TLSCert = stringOrNull(httpMonitor.TLSCert)
	m.TLSKey = stringOrNull(httpMonitor.TLSKey)
	m.TLSCa = stringOrNull(httpMonitor.TLSCa)
	m.OAuthAuthMethod = stringOrNull(httpMonitor.OAuthAuthMethod)
	m.OAuthTokenURL = stringOrNull(httpMonitor.OAuthTokenURL)
	m.OAuthClientID = stringOrNull(httpMonitor.OAuthClientID)
	m.OAuthClientSecret = stringOrNull(httpMonitor.OAuthClientSecret)
	m.OAuthScopes = stringOrNull(httpMonitor.OAuthScopes)
}

// populateHTTPMonitorOptionalFields populates optional pointer fields and list fields.
func populateHTTPMonitorOptionalFields(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	data any,
	diags *diag.Diagnostics,
) {
	switch m := data.(type) {
	case *MonitorHTTPResourceModel:
		populateOptionalFieldsForHTTP(ctx, httpMonitor, m, diags)
	case *MonitorHTTPJSONQueryResourceModel:
		populateOptionalFieldsForJSONQuery(ctx, httpMonitor, m, diags)
	case *MonitorHTTPKeywordResourceModel:
		populateOptionalFieldsForKeyword(ctx, httpMonitor, m, diags)
	}
}

// populateOptionalFieldsForHTTP populates optional fields for HTTP monitor.
func populateOptionalFieldsForHTTP(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	m *MonitorHTTPResourceModel,
	diags *diag.Diagnostics,
) {
	if httpMonitor.Parent != nil {
		m.Parent = types.Int64Value(*httpMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}
	if httpMonitor.ProxyID != nil {
		m.ProxyID = types.Int64Value(*httpMonitor.ProxyID)
	} else {
		m.ProxyID = types.Int64Null()
	}
	if len(httpMonitor.AcceptedStatusCodes) > 0 {
		statusCodes, d := types.ListValueFrom(ctx, types.StringType, httpMonitor.AcceptedStatusCodes)
		diags.Append(d...)
		m.AcceptedStatusCodes = statusCodes
	}
	if len(httpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, httpMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// populateOptionalFieldsForJSONQuery populates optional fields for HTTP JSON Query monitor.
func populateOptionalFieldsForJSONQuery(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	m *MonitorHTTPJSONQueryResourceModel,
	diags *diag.Diagnostics,
) {
	if httpMonitor.Parent != nil {
		m.Parent = types.Int64Value(*httpMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}
	if httpMonitor.ProxyID != nil {
		m.ProxyID = types.Int64Value(*httpMonitor.ProxyID)
	} else {
		m.ProxyID = types.Int64Null()
	}
	if len(httpMonitor.AcceptedStatusCodes) > 0 {
		statusCodes, d := types.ListValueFrom(ctx, types.StringType, httpMonitor.AcceptedStatusCodes)
		diags.Append(d...)
		m.AcceptedStatusCodes = statusCodes
	}
	if len(httpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, httpMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// populateOptionalFieldsForKeyword populates optional fields for HTTP Keyword monitor.
func populateOptionalFieldsForKeyword(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	m *MonitorHTTPKeywordResourceModel,
	diags *diag.Diagnostics,
) {
	if httpMonitor.Parent != nil {
		m.Parent = types.Int64Value(*httpMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}
	if httpMonitor.ProxyID != nil {
		m.ProxyID = types.Int64Value(*httpMonitor.ProxyID)
	} else {
		m.ProxyID = types.Int64Null()
	}
	if len(httpMonitor.AcceptedStatusCodes) > 0 {
		statusCodes, d := types.ListValueFrom(ctx, types.StringType, httpMonitor.AcceptedStatusCodes)
		diags.Append(d...)
		m.AcceptedStatusCodes = statusCodes
	}
	if len(httpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, httpMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}
