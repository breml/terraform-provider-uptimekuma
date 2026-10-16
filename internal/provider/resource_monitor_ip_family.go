package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/breml/go-uptime-kuma-client/monitor"
)

// ipFamilyAuto is the ip_family value that lets Uptime Kuma select the IP
// family automatically. The empty string is accepted as a synonym.
const ipFamilyAuto = "auto"

// ipFamilyAttribute returns the ip_family attribute of the HTTP, keyword and
// json-query monitors, the monitor types Uptime Kuma applies the setting to.
func ipFamilyAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "IP protocol version used to connect to the target. One of `auto` " +
			"(default, selects the family automatically using Happy Eyeballs), `ipv4` or `ipv6`. " +
			"An empty string is accepted as a synonym for `auto`.",
		Optional: true,
		Computed: true,
		Default:  stringdefault.StaticString(ipFamilyAuto),
		Validators: []validator.String{
			stringvalidator.OneOf("", ipFamilyAuto, string(monitor.IPFamilyIPv4), string(monitor.IPFamilyIPv6)),
		},
	}
}

// ipFamilyToAPI converts the ip_family attribute to the client's IP family,
// mapping both `auto` and the empty string to the automatic selection.
func ipFamilyToAPI(value types.String) monitor.IPFamily {
	if value.ValueString() == ipFamilyAuto {
		return monitor.IPFamilyAuto
	}

	return monitor.IPFamily(value.ValueString())
}

// ipFamilyFromAPI converts the client's IP family to the ip_family attribute.
// The server cannot tell `auto` from the empty string, so for the automatic
// selection the prior value is kept when it is one of the two, which avoids a
// perpetual diff for a configuration using the empty string. Without such a
// prior value, e.g. on import or in a data source, `auto` is returned.
func ipFamilyFromAPI(apiValue monitor.IPFamily, prior types.String) types.String {
	if apiValue != monitor.IPFamilyAuto {
		return types.StringValue(string(apiValue))
	}

	if !prior.IsNull() && !prior.IsUnknown() && ipFamilyToAPI(prior) == monitor.IPFamilyAuto {
		return prior
	}

	return types.StringValue(ipFamilyAuto)
}
