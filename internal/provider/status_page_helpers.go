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
func convertUnknownIDsToNull(ctx context.Context, publicGroupList types.List, diags *diag.Diagnostics) types.List {
	if publicGroupList.IsNull() {
		return publicGroupList
	}

	var configGroups []PublicGroupModel
	diags.Append(publicGroupList.ElementsAs(ctx, &configGroups, true)...)
	if diags.HasError() {
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

	// Convert unknown IDs to null so terraform state has known values
	groups := make([]PublicGroupModel, len(configGroups))
 // Iterate over items.
	for i, group := range configGroups {
		groups[i] = group
		if group.ID.IsUnknown() {
			groups[i].ID = types.Int64Null()
		}

		// handle monitors
		if !group.MonitorList.IsNull() {
			var mons []PublicMonitorModel
			diags.Append(group.MonitorList.ElementsAs(ctx, &mons, true)...)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":       types.Int64Type,
						"send_url": types.BoolType,
					},
				})
			}

   // Iterate over items.
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
				types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType}},
				mons,
			)
			diags.Append(d...)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":       types.Int64Type,
						"send_url": types.BoolType,
					},
				})
			}

			groups[i].MonitorList = monList
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
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":       types.Int64Type,
				"send_url": types.BoolType,
			},
		})
	}

	return groupList
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
			emptyMonList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType}}, []PublicMonitorModel{})
			diags.Append(d...)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType}})
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
