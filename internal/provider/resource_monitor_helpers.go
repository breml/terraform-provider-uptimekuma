package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

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
