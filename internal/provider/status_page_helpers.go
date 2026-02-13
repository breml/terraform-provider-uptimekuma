package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/breml/go-uptime-kuma-client/statuspage"
)

// mergeGroupIDsIntoPlan preserves the plan's public_group_list values and only
// injects server-assigned group IDs from the SaveStatusPage response. This
// prevents perpetual diffs caused by the server omitting optional fields
// (e.g. sendUrl when false) that would otherwise be stored as null in state
// while the config has an explicit value.
func mergeGroupIDsIntoPlan(
	ctx context.Context,
	planGroupList types.List,
	savedGroups []statuspage.PublicGroup,
	diags *diag.Diagnostics,
) types.List {
	var planModels []PublicGroupModel

	diags.Append(planGroupList.ElementsAs(ctx, &planModels, true)...)
	if diags.HasError() {
		return nullGroupList()
	}

	for i := range planModels {
		if planModels[i].ID.IsUnknown() || planModels[i].ID.IsNull() {
			if i < len(savedGroups) {
				planModels[i].ID = types.Int64Value(savedGroups[i].ID)
			} else {
				planModels[i].ID = types.Int64Null()
			}
		}
	}

	return buildGroupListFromModels(ctx, planModels, diags)
}

func buildGroupListFromModels(ctx context.Context, groups []PublicGroupModel, diags *diag.Diagnostics) types.List {
	groupList, d := types.ListValueFrom(ctx, groupListAttrType(), groups)
	diags.Append(d...)
	if diags.HasError() {
		return nullGroupList()
	}

	return groupList
}

func nullGroupList() types.List {
	return types.ListNull(groupListAttrType())
}

func groupListAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":     types.Int64Type,
			"name":   types.StringType,
			"weight": types.Int64Type,
			"monitor_list": types.ListType{ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":       types.Int64Type,
					"send_url": types.BoolType,
				},
			}},
		},
	}
}
