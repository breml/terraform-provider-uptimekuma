package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/breml/go-uptime-kuma-client/monitor"
)

// populateHTTPMonitorBaseFields populates the shared HTTP monitor base fields.
// This dispatches to type-specific implementations based on the monitor model type.
// Used during Read operations to extract API data into Terraform model structures.
func populateHTTPMonitorBaseFields(
	_ context.Context,
	httpMonitor *monitor.HTTP,
	data any,
	_ *diag.Diagnostics,
) {
	// Handle different HTTP monitor types with their specific models.
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
// Extracts HTTP-specific fields from the API response into the model.
// Handles base monitor fields and all HTTP configuration options.
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
// Extracts common HTTP fields from API response for JSON Query specific model.
// Similar to HTTP monitor population but uses JSON Query specific types.
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
// Extracts common HTTP fields from API response for Keyword specific model.
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
// This handles fields that can be null or unknown in the API response.
func populateHTTPMonitorOptionalFields(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	data any,
	diags *diag.Diagnostics,
) {
	// Dispatch to type-specific handler based on monitor model type.
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
// Handles proxy, parent group, accepted status codes, and notification IDs.
// Converts null API values to Terraform null types appropriately.
func populateOptionalFieldsForHTTP(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	m *MonitorHTTPResourceModel,
	diags *diag.Diagnostics,
) {
	// Set parent monitor group if present.
	if httpMonitor.Parent != nil {
		m.Parent = types.Int64Value(*httpMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	// Set proxy if configured.
	if httpMonitor.ProxyID != nil {
		m.ProxyID = types.Int64Value(*httpMonitor.ProxyID)
	} else {
		m.ProxyID = types.Int64Null()
	}

	// Convert accepted status codes list if non-empty.
	if len(httpMonitor.AcceptedStatusCodes) > 0 {
		statusCodes, d := types.ListValueFrom(ctx, types.StringType, httpMonitor.AcceptedStatusCodes)
		diags.Append(d...)
		m.AcceptedStatusCodes = statusCodes
	}

	// Convert notification IDs list if present.
	if len(httpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, httpMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// populateOptionalFieldsForJSONQuery populates optional fields for HTTP JSON Query monitor.
// Includes parent group, proxy, status codes, and notification configuration.
// The parent field identifies the parent monitor group for organization purposes.
// The proxy field specifies the proxy server to use for the monitor connection.
// Status codes and notifications are converted to Terraform list types for state management.
func populateOptionalFieldsForJSONQuery(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	m *MonitorHTTPJSONQueryResourceModel,
	diags *diag.Diagnostics,
) {
	// Set parent monitor group if configured in API response.
	if httpMonitor.Parent != nil {
		m.Parent = types.Int64Value(*httpMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	// Set proxy ID if the monitor uses a proxy.
	if httpMonitor.ProxyID != nil {
		m.ProxyID = types.Int64Value(*httpMonitor.ProxyID)
	} else {
		m.ProxyID = types.Int64Null()
	}

	// Convert accepted status codes list if present.
	if len(httpMonitor.AcceptedStatusCodes) > 0 {
		statusCodes, d := types.ListValueFrom(ctx, types.StringType, httpMonitor.AcceptedStatusCodes)
		diags.Append(d...)
		m.AcceptedStatusCodes = statusCodes
	}

	// Convert notification IDs list if present.
	if len(httpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, httpMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// populateOptionalFieldsForKeyword populates optional fields for HTTP Keyword monitor.
// Handles parent group, proxy, status codes, and notification configuration.
// This function follows the same pattern as JSON Query to ensure consistency.
// Parent and proxy fields are optional and may be null in the API response.
// Lists are properly converted to Terraform types for accurate state representation.
func populateOptionalFieldsForKeyword(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	m *MonitorHTTPKeywordResourceModel,
	diags *diag.Diagnostics,
) {
	// Set parent monitor group if configured in API response.
	if httpMonitor.Parent != nil {
		m.Parent = types.Int64Value(*httpMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	// Set proxy ID if the monitor uses a proxy.
	if httpMonitor.ProxyID != nil {
		m.ProxyID = types.Int64Value(*httpMonitor.ProxyID)
	} else {
		m.ProxyID = types.Int64Null()
	}

	// Convert accepted status codes list if present.
	if len(httpMonitor.AcceptedStatusCodes) > 0 {
		statusCodes, d := types.ListValueFrom(ctx, types.StringType, httpMonitor.AcceptedStatusCodes)
		diags.Append(d...)
		m.AcceptedStatusCodes = statusCodes
	}

	// Convert notification IDs list if present.
	if len(httpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, httpMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}
