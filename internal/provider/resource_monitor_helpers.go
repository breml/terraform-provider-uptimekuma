package provider

import (
	"errors"
	"strings"

	kuma "github.com/breml/go-uptime-kuma-client"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// isNotFoundError checks whether an error from the kuma client indicates
// that the requested resource was not found.
//
// Resources that are looked up via a cached list (tags, notifications,
// proxies, docker hosts) return kuma.ErrNotFound.
//
// Resources fetched directly from the server (monitors, status pages,
// maintenance) may return a server-side error from the Uptime Kuma
// backend when the resource no longer exists.  Known patterns:
//   - Monitors: "Cannot read properties of null (reading 'id')"
//   - Status pages: "No slug?"
func isNotFoundError(err error) bool {
	if errors.Is(err, kuma.ErrNotFound) {
		return true
	}

	msg := err.Error()

	return strings.Contains(msg, "Cannot read properties of null") ||
		strings.Contains(msg, "No slug?")
}

// strToPtr converts a Terraform string type to a pointer to string.
// Returns nil if the value is null or unknown.
func strToPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return v.ValueStringPointer()
}

// ptrToTypes converts a pointer to string to a Terraform string type.
// Returns StringNull() if the pointer is nil.
func ptrToTypes(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}

	return types.StringValue(*v)
}

// float64ToPtr converts a Terraform float64 type to a pointer to float64.
// Returns nil if the value is null or unknown.
func float64ToPtr(v types.Float64) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return v.ValueFloat64Pointer()
}

// int64ToPtr converts a Terraform int64 type to a pointer to int64.
// Returns nil if the value is null or unknown. For columns the Uptime Kuma
// server stores as nullable, nil round-trips as SQL NULL, which is how the
// check is told to apply its own fallback.
func int64ToPtr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return v.ValueInt64Pointer()
}

// int64PtrToTypes converts a pointer to int64 to a Terraform int64 type.
// Returns Int64Null() if the pointer is nil.
func int64PtrToTypes(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}

	return types.Int64Value(*v)
}

// boolToPtr converts a Terraform bool type to a pointer to bool.
// Returns nil if the value is null or unknown.
func boolToPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return v.ValueBoolPointer()
}

// boolPtrToTypes converts a pointer to bool to a Terraform bool type.
// Returns BoolNull() if the pointer is nil.
func boolPtrToTypes(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}

	return types.BoolValue(*v)
}
