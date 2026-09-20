package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

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

// nonBlankValidator rejects strings that are empty once surrounding whitespace
// is removed.
//
// Some Uptime Kuma fields are stored trimmed, so a value consisting only of
// whitespace is equivalent to an empty one. Catching it at plan time reports the
// problem on the attribute instead of failing during apply with a marshal error
// from the client library.
type nonBlankValidator struct{}

// Description returns a plain text description of the validator's behavior.
func (nonBlankValidator) Description(_ context.Context) string {
	return "value must not be empty or consist only of whitespace"
}

// MarkdownDescription returns a markdown description of the validator's behavior.
func (v nonBlankValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString reports an error if the configured value is blank.
func (v nonBlankValidator) ValidateString(
	ctx context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if strings.TrimSpace(req.ConfigValue.ValueString()) != "" {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Attribute Value",
		fmt.Sprintf(
			"Attribute %s %s, got: %q.",
			req.Path,
			v.Description(ctx),
			req.ConfigValue.ValueString(),
		),
	)
}

// nonBlank returns a validator that rejects blank strings.
func nonBlank() validator.String {
	return nonBlankValidator{}
}

// jsonObjectValidator rejects strings that do not decode to a JSON object.
//
// Uptime Kuma parses such a value with `JSON.parse` and merges the result into
// the outgoing request, so a scalar, an array or a syntax error is only
// noticed when the notification actually fires. Rejecting it at plan time
// reports the typo on the attribute instead.
type jsonObjectValidator struct{}

// Description returns a plain text description of the validator's behavior.
func (jsonObjectValidator) Description(_ context.Context) string {
	return "value must be a JSON object"
}

// MarkdownDescription returns a markdown description of the validator's behavior.
func (v jsonObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString reports an error if the configured value is not a JSON object.
func (v jsonObjectValidator) ValidateString(
	ctx context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var object map[string]any

	err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &object)
	if err == nil && object != nil {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Attribute Value",
		fmt.Sprintf(
			"Attribute %s %s, got: %q.",
			req.Path,
			v.Description(ctx),
			req.ConfigValue.ValueString(),
		),
	)
}

// jsonObject returns a validator that rejects strings which do not decode to a
// JSON object.
func jsonObject() validator.String {
	return jsonObjectValidator{}
}
