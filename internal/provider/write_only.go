package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// writeOnlySuffix names the write-only alternative of a secret attribute.
	writeOnlySuffix = "_wo"

	// writeOnlyVersionSuffix names the stored stand-in whose change triggers an update that
	// sends the write-only value.
	writeOnlyVersionSuffix = "_wo_version"
)

// withWriteOnlySecret adds the write-only alternative <name>_wo and its trigger
// <name>_wo_version for the secret string attribute name, which attrs must already contain.
//
// Terraform and OpenTofu never store a write-only value in plan or state, so a change to it
// alone produces no diff; <name>_wo_version is the stored stand-in whose change triggers the
// update. The two require each other. The existing attribute stays the stored alternative: a
// required one becomes optional with exactly one of it and <name>_wo set, an optional one
// conflicts with <name>_wo.
//
// description describes the secret, for example "SMTP password for authentication"; it is the
// start of the <name>_wo description.
//
// A missing or non-string attribute is a programming error in the schema and panics.
func withWriteOnlySecret(
	attrs map[string]schema.Attribute,
	name string,
	description string,
) map[string]schema.Attribute {
	woName := name + writeOnlySuffix
	versionName := name + writeOnlyVersionSuffix

	stored, ok := attrs[name].(schema.StringAttribute)
	if !ok {
		panic(fmt.Sprintf("withWriteOnlySecret: %q is not a string attribute of the schema", name))
	}

	// The write-only value is held to the same constraints as the stored one.
	woValidators := append([]validator.String{}, stored.Validators...)

	if stored.Required {
		stored.Required = false
		stored.Optional = true
		stored.Validators = append(stored.Validators, stringvalidator.ExactlyOneOf(path.MatchRoot(woName)))
	} else {
		stored.Validators = append(stored.Validators, stringvalidator.ConflictsWith(path.MatchRoot(woName)))
	}

	// A computed attribute without a default would otherwise plan its prior value, which Read
	// no longer refreshes while the write-only variant is in use.
	if stored.Computed && stored.Default == nil {
		stored.PlanModifiers = append(stored.PlanModifiers, nullWhileWriteOnly{versionName: versionName})
	}

	stored.MarkdownDescription += fmt.Sprintf(" Stored in state; use `%s` to keep it out of state.", woName)
	attrs[name] = stored

	attrs[woName] = schema.StringAttribute{
		MarkdownDescription: fmt.Sprintf(
			"%s, write-only: sent to Uptime Kuma on create and update but never stored in plan or "+
				"state. Requires Terraform 1.11 or OpenTofu 1.11 or later. Conflicts with `%s`; requires "+
				"`%s`. An imported resource has no `%s` yet, so `terraform import` may store the server's "+
				"value in `%s` until the next apply with `%s` set removes it from state.",
			description, name, versionName, versionName, name, woName,
		),
		Optional:   true,
		Sensitive:  true,
		WriteOnly:  true,
		Validators: append(woValidators, stringvalidator.AlsoRequires(path.MatchRoot(versionName))),
	}

	attrs[versionName] = schema.StringAttribute{
		MarkdownDescription: fmt.Sprintf(
			"Any value that changes whenever `%s` changes, for example a counter or a hash of the "+
				"secret's encrypted source. A change triggers an update that sends the current `%s`.",
			woName, woName,
		),
		Optional: true,
		Validators: []validator.String{
			stringvalidator.AlsoRequires(path.MatchRoot(woName)),
		},
	}

	return attrs
}

// writeOnlyOrPlan returns the secret to send to Uptime Kuma: <name>_wo, which only the
// configuration carries, or else the stored attribute's value from the plan.
func writeOnlyOrPlan(
	ctx context.Context,
	config tfsdk.Config,
	name string,
	plan types.String,
	diags *diag.Diagnostics,
) string {
	return writeOnlyOrPlanValue(ctx, config, name, plan, diags).ValueString()
}

// writeOnlyOrPlanValue is writeOnlyOrPlan for an API field that tells an unset secret apart
// from an empty one: the result is null when neither <name>_wo nor the stored attribute is set.
func writeOnlyOrPlanValue(
	ctx context.Context,
	config tfsdk.Config,
	name string,
	plan types.String,
	diags *diag.Diagnostics,
) types.String {
	var writeOnly types.String

	diags.Append(config.GetAttribute(ctx, path.Root(name+writeOnlySuffix), &writeOnly)...)

	if !writeOnly.IsNull() && !writeOnly.IsUnknown() {
		return writeOnly
	}

	return plan
}

// nullWhileWriteOnly plans a computed secret attribute as null while its write-only variant is
// in use, which a set <name>_wo_version signals. Read does not refresh the attribute then, so
// any other planned value would not match the state after apply.
type nullWhileWriteOnly struct {
	versionName string
}

// Description implements planmodifier.String.
func (m nullWhileWriteOnly) Description(_ context.Context) string {
	return fmt.Sprintf("The value is null while %s is set.", m.versionName)
}

// MarkdownDescription implements planmodifier.String.
func (m nullWhileWriteOnly) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyString implements planmodifier.String.
func (m nullWhileWriteOnly) PlanModifyString(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	if !req.ConfigValue.IsNull() {
		return
	}

	var version types.String

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(m.versionName), &version)...)

	if version.IsNull() {
		return
	}

	resp.PlanValue = types.StringNull()
}
