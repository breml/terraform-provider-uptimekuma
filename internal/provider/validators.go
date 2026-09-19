package provider

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// wholeNumberValidator rejects fractional float64 values.
//
// Some Uptime Kuma checks round a stored float to whole seconds. Terraform
// writes the configured value to state, so a fractional value leaves state
// disagreeing with the server and produces a plan that never converges.
// Rejecting the value at plan time surfaces that as a clear error instead.
type wholeNumberValidator struct{}

// Description returns a plain text description of the validator's behavior.
func (wholeNumberValidator) Description(_ context.Context) string {
	return "value must be a whole number"
}

// MarkdownDescription returns a markdown description of the validator's behavior.
func (v wholeNumberValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateFloat64 reports an error if the configured value has a fractional part.
func (v wholeNumberValidator) ValidateFloat64(
	ctx context.Context,
	req validator.Float64Request,
	resp *validator.Float64Response,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueFloat64()
	if value == math.Trunc(value) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Attribute Value",
		fmt.Sprintf(
			"Attribute %s %s, got: %v. Uptime Kuma rounds this value to whole seconds, so a "+
				"fractional value would leave Terraform state permanently out of sync with the "+
				"server. Use %v or %v instead.",
			req.Path,
			v.Description(ctx),
			value,
			math.Floor(value),
			math.Ceil(value),
		),
	)
}

// wholeNumber returns a validator that rejects fractional float64 values.
func wholeNumber() validator.Float64 {
	return wholeNumberValidator{}
}
