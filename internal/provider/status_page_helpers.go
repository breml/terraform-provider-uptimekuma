package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/breml/go-uptime-kuma-client/statuspage"
)

// convertUnknownIDsToNull converts unknown group and monitor IDs to null values.
// This ensures all computed values are known in Terraform state after Create/Update.
// Unknown values from Terraform planning are replaced with explicit nulls.
func convertUnknownIDsToNull(ctx context.Context, publicGroupList types.List, diags *diag.Diagnostics) types.List {
	// Handle null input gracefully by returning it unchanged.
	if publicGroupList.IsNull() {
		return publicGroupList
	}

	// Deserialize the public group list from Terraform types.
	configGroups := deserializeGroupsForConversion(ctx, publicGroupList, diags)
	if diags.HasError() {
		return nullGroupList()
	}

	// Convert unknown values in groups and monitors to explicit nulls.
	groups := convertGroupsUnknownToNull(ctx, configGroups, diags)
	if diags.HasError() {
		return nullGroupList()
	}

	// Rebuild the list with nulls instead of unknowns.
	return buildGroupListFromModels(ctx, groups, diags)
}

// deserializeGroupsForConversion extracts PublicGroupModel values from a Terraform list.
// It handles deserialization errors by appending diagnostics.
func deserializeGroupsForConversion(
	ctx context.Context,
	publicGroupList types.List,
	diags *diag.Diagnostics,
) []PublicGroupModel {
	var configGroups []PublicGroupModel
	// Convert Terraform list to Go slice, preserving unknown values.
	diags.Append(publicGroupList.ElementsAs(ctx, &configGroups, true)...)
	return configGroups
}

// convertGroupsUnknownToNull converts unknown IDs in groups and their monitors to nulls.
// Iterates through each group and monitor to replace unknown values.
func convertGroupsUnknownToNull(
	ctx context.Context,
	configGroups []PublicGroupModel,
	diags *diag.Diagnostics,
) []PublicGroupModel {
	// Create a new slice with updated group values.
	groups := make([]PublicGroupModel, len(configGroups))

	for i, group := range configGroups {
		groups[i] = group
		if group.ID.IsUnknown() {
			groups[i].ID = types.Int64Null()
		}

		if !group.MonitorList.IsNull() {
			monList := convertMonitorListUnknownToNull(ctx, group.MonitorList, diags)
			if diags.HasError() {
				return groups
			}

			groups[i].MonitorList = monList
		}
	}

	return groups
}

func convertMonitorListUnknownToNull(ctx context.Context, monitorList types.List, diags *diag.Diagnostics) types.List {
	var mons []PublicMonitorModel
	diags.Append(monitorList.ElementsAs(ctx, &mons, true)...)
	if diags.HasError() {
		return nullMonitorList()
	}

	for j := range mons {
		if mons[j].ID.IsUnknown() {
			mons[j].ID = types.Int64Null()
		}

		if mons[j].SendURL.IsUnknown() {
			mons[j].SendURL = types.BoolNull()
		}
	}

	monList, d := types.ListValueFrom(
		ctx,
		monitorListAttrType(),
		mons,
	)
	diags.Append(d...)
	if diags.HasError() {
		return nullMonitorList()
	}

	return monList
}

func buildGroupListFromModels(ctx context.Context, groups []PublicGroupModel, diags *diag.Diagnostics) types.List {
	groupList, d := types.ListValueFrom(ctx, groupListAttrType(), groups)
	diags.Append(d...)
	if diags.HasError() {
		return nullMonitorList()
	}

	return groupList
}

func nullGroupList() types.List {
	return types.ListNull(groupListAttrType())
}

func nullMonitorList() types.List {
	return types.ListNull(monitorListAttrType())
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

func monitorListAttrType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":       types.Int64Type,
			"send_url": types.BoolType,
		},
	}
}

// buildPublicGroupListFromSaved constructs a types.List value for public_group_list
// from the savedGroups returned by the API. It appends any diagnostics to the
// provided diags pointer.
func buildPublicGroupListFromSaved(
	ctx context.Context,
	saved []statuspage.PublicGroup,
	diags *diag.Diagnostics,
) types.List {
	if len(saved) == 0 {
		return types.ListNull(types.ObjectType{
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
		})
	}

	groups := make([]PublicGroupModel, len(saved))
	// Iterate over items.
	for i, g := range saved {
		groups[i] = PublicGroupModel{}
		groups[i].ID = types.Int64Value(g.ID)
		groups[i].Name = types.StringValue(g.Name)
		groups[i].Weight = types.Int64Value(int64(g.Weight))

		if len(g.MonitorList) > 0 {
			monitors := make([]PublicMonitorModel, len(g.MonitorList))
			// Iterate over items.
			for j, m := range g.MonitorList {
				monitors[j] = PublicMonitorModel{ID: types.Int64Value(m.ID)}
				if m.SendURL != nil {
					monitors[j].SendURL = types.BoolValue(*m.SendURL)
				} else {
					monitors[j].SendURL = types.BoolNull()
				}
			}

			monList, d := types.ListValueFrom(
				ctx,
				types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType}},
				monitors,
			)
			diags.Append(d...)
			if diags.HasError() {
				return types.ListNull(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType},
					},
				)
			}

			groups[i].MonitorList = monList
		} else {
			// No monitors returned by server - set to empty list for clarity
			emptyMonList, d := types.ListValueFrom(
				ctx,
				types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType}},
				[]PublicMonitorModel{},
			)
			diags.Append(d...)
			if diags.HasError() {
				return types.ListNull(
					types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType}},
				)
			}

			groups[i].MonitorList = emptyMonList
		}
	}

	groupList, d := types.ListValueFrom(ctx, types.ObjectType{
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
	}, groups)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(
			types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType}},
		)
	}

	return groupList
}
