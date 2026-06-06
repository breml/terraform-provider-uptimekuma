package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/breml/go-uptime-kuma-client/monitor"
)

// MonitorConditionModel describes a single assertion clause evaluated against a
// monitor's parsed result (query result, MQTT payload, SNMP value, etc.).
// Conditions are chained together using the AndOr field.
type MonitorConditionModel struct {
	Variable types.String `tfsdk:"variable"`
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
	AndOr    types.String `tfsdk:"and_or"`
}

// conditionsAttribute returns the schema attribute used to expose the optional
// list of monitor conditions on the monitor types that support them.
func conditionsAttribute() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Optional list of assertion clauses evaluated against the monitor result. " +
			"Each condition is chained with the previous one using `and_or`.",
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"variable": schema.StringAttribute{
					MarkdownDescription: "Name of the field to test against (monitor-type specific).",
					Required:            true,
				},
				"operator": schema.StringAttribute{
					MarkdownDescription: "Comparison operator (e.g. `==`, `!=`, `<`, `>`, `contains`).",
					Required:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Value to compare against.",
					Required:            true,
				},
				"and_or": schema.StringAttribute{
					MarkdownDescription: "Chains this condition with the previous one. Valid values: `and`, `or`.",
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString("and"),
					Validators: []validator.String{
						stringvalidator.OneOf("and", "or"),
					},
				},
			},
		},
	}
}

// conditionObjectType returns the Terraform object type for a single condition.
func conditionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"variable": types.StringType,
			"operator": types.StringType,
			"value":    types.StringType,
			"and_or":   types.StringType,
		},
	}
}

// nullConditionList returns a typed null list for the conditions attribute.
func nullConditionList() types.List {
	return types.ListNull(conditionObjectType())
}

// buildConditions converts the Terraform conditions list into the client
// representation. A null, unknown or empty list yields nil.
func buildConditions(ctx context.Context, list types.List, diags *diag.Diagnostics) []monitor.Condition {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var models []MonitorConditionModel
	d := list.ElementsAs(ctx, &models, false)
	diags.Append(d...)
	if d.HasError() || len(models) == 0 {
		return nil
	}

	conditions := make([]monitor.Condition, 0, len(models))
	for _, m := range models {
		conditions = append(conditions, monitor.Condition{
			Variable: m.Variable.ValueString(),
			Operator: m.Operator.ValueString(),
			Value:    m.Value.ValueString(),
			AndOr:    monitor.ConditionOperator(m.AndOr.ValueString()),
		})
	}

	return conditions
}

// populateConditions converts the client conditions into a Terraform list.
// An empty slice yields a typed null list to avoid perpetual diffs.
func populateConditions(ctx context.Context, conditions []monitor.Condition, diags *diag.Diagnostics) types.List {
	if len(conditions) == 0 {
		return nullConditionList()
	}

	models := make([]MonitorConditionModel, 0, len(conditions))
	for _, c := range conditions {
		andOr := string(c.AndOr)
		if andOr == "" {
			andOr = "and"
		}

		models = append(models, MonitorConditionModel{
			Variable: types.StringValue(c.Variable),
			Operator: types.StringValue(c.Operator),
			Value:    types.StringValue(c.Value),
			AndOr:    types.StringValue(andOr),
		})
	}

	list, d := types.ListValueFrom(ctx, conditionObjectType(), models)
	diags.Append(d...)
	if diags.HasError() {
		return nullConditionList()
	}

	return list
}
