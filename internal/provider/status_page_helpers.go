package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/breml/go-uptime-kuma-client/statuspage"
)

// buildPublicGroupListFromSaved constructs a types.List value for public_group_list
// from the savedGroups returned by the API. It appends any diagnostics to the
// provided diags pointer.
func buildPublicGroupListFromSaved(ctx context.Context, saved []statuspage.PublicGroup, diags *diag.Diagnostics) types.List {
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
	for i, g := range saved {
		groups[i] = PublicGroupModel{}
		groups[i].ID = types.Int64Value(g.ID)
		groups[i].Name = types.StringValue(g.Name)
		groups[i].Weight = types.Int64Value(int64(g.Weight))

		if len(g.MonitorList) > 0 {
			monitors := make([]PublicMonitorModel, len(g.MonitorList))
			for j, m := range g.MonitorList {
				monitors[j] = PublicMonitorModel{ID: types.Int64Value(m.ID)}
				if m.SendURL != nil {
					monitors[j].SendURL = types.BoolValue(*m.SendURL)
				} else {
					monitors[j].SendURL = types.BoolNull()
				}
			}

			monList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType}}, monitors)
			diags.Append(d...)
			if diags.HasError() {
				return types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType}})
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
		return types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.Int64Type, "send_url": types.BoolType}})
	}

	return groupList
}
