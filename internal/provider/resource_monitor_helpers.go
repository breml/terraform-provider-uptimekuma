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
