package provider

import (
	"context"
	"fmt"
	"slices"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// monitorLister is the part of *kuma.Client monitorExists needs. It is a seam of
// its own rather than an addition to resyncer, so that the presence check can be
// exercised without a socket.io server and without teaching every other cache
// helper about monitors.
type monitorLister interface {
	// GetMonitors asks the server to resend its monitor list and returns what the
	// client then holds.
	GetMonitors(ctx context.Context) ([]monitor.Base, error)
}

// monitorExists builds the presence check removeOnServerMiss confirms a monitor
// deletion against.
//
// GetMonitors emits getMonitorList before reading the client's cache, so unlike
// the maintenance and status page lists it does ask the server - but the answer
// arrives as a whole-list broadcast the client applies asynchronously, so what
// this reads may still be one broadcast behind. existsWithResync forces a resync
// before an absence is believed; a hit is taken as it stands, because
// findWithResync returns on one. See removeOnServerMiss for why that asymmetry is
// the right way round.
//
// A failure of the getter is returned as an error rather than as an absence, as
// findWithResync requires: only an absence is worth a resync, and only an absence
// may cost a monitor its place in state.
func monitorExists(client monitorLister, id int64) func(context.Context) (bool, error) {
	return func(ctx context.Context) (bool, error) {
		monitors, err := client.GetMonitors(ctx)
		if err != nil {
			return false, fmt.Errorf("list monitors: %w", err)
		}

		return slices.ContainsFunc(monitors, func(mon monitor.Base) bool {
			return mon.GetID() == id
		}), nil
	}
}

// removeMonitorOnServerMiss accounts for a failed monitor read, dropping the
// monitor from Terraform state only once Uptime Kuma's monitor list agrees that
// it is gone.
//
// It is removeOnServerMiss wired to a monitor, and every monitor resource Read
// hands it the error from GetMonitorAs. name is the monitor in the words of the
// diagnostics, e.g. "HTTP monitor". See removeOnServerMiss for what each outcome
// means and why a list that cannot be refreshed never condemns the monitor.
func removeMonitorOnServerMiss(
	ctx context.Context,
	client *kuma.Client,
	err error,
	id int64,
	name string,
	resp *resource.ReadResponse,
) {
	removeOnServerMiss(ctx, client, err, monitorExists(client, id), name, resp)
}

// monitorTypeDrifted reports whether the monitor Terraform holds in state is no
// longer of the type this resource manages, and drops it from state when it is not.
//
// managed is the type this resource speaks for, which the client hardcodes per
// monitor type, e.g. monitor.HTTP.Type() is always "http". actual is the type the
// server reports, which Base.UnmarshalJSON reads off the record - so an empty
// actual is a record that carried no type and means only that there is nothing to
// compare, not that anything drifted.
//
// The removal is not the #409 mistake even though it looks like one. There the
// provider guessed from an error string that a monitor was gone; here the read
// succeeded and the server says the monitor at this ID is something else, so the
// resource this state entry speaks for really does not exist and Terraform is
// right to plan a replacement. What the removal must not be is invisible: it
// warns as well as logging, because the monitor now at that ID stops being
// managed by anything and the reader would otherwise see a resource leave state
// with no account of why. tflog.Warn is kept for the structured fields, which
// carry the two types for anyone reading provider logs.
//
// It never errors, so a caller only has to return when it reports true.
func monitorTypeDrifted(
	ctx context.Context,
	id int64,
	managed string,
	actual string,
	resp *resource.ReadResponse,
) bool {
	if actual == "" || actual == managed {
		return false
	}

	tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
		"id":            id,
		"expected_type": managed,
		"actual_type":   actual,
	})

	resp.Diagnostics.AddWarning(
		"Monitor type changed outside Terraform",
		fmt.Sprintf(
			"Monitor %d is of type %q but is managed as %q, so it was removed from state and "+
				"Terraform will plan to create a replacement. Manage it with the resource type "+
				"matching %q, or remove it from the configuration, to avoid a duplicate.",
			id, actual, managed, actual,
		),
	)

	resp.State.RemoveResource(ctx)

	return true
}

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

// float64ToPtr converts a Terraform float64 type to a pointer to float64.
// Returns nil if the value is null or unknown.
func float64ToPtr(v types.Float64) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return v.ValueFloat64Pointer()
}

// float64PtrToTypes converts a pointer to float64 to a Terraform float64 type.
// Returns Float64Null() if the pointer is nil.
func float64PtrToTypes(v *float64) types.Float64 {
	if v == nil {
		return types.Float64Null()
	}

	return types.Float64Value(*v)
}

// timeoutValueOrDefault converts the timeout returned by the client into a
// Terraform Float64 for a resource whose timeout attribute is Computed with a
// default.
//
// The monitor.timeout column is NOT NULL, so the server cannot report a null
// timeout; a nil pointer means the field was absent from the response
// altogether. Substituting the fallback, which callers pass as the schema
// default, keeps the Computed attribute consistent with the plan, but the
// substitution is logged and reported as a warning diagnostic so the unexpected
// response shape is not silent.
//
// A stored 0 is passed through unchanged rather than substituted. It is a real
// value the server can report for monitors created outside Terraform (ping
// excepted, whose timeout the server bounds to 1-300), and rewriting it on read
// would mask genuine drift.
func timeoutValueOrDefault(
	ctx context.Context,
	id int64,
	timeout *float64,
	fallback float64,
	monitorType string,
	diags *diag.Diagnostics,
) types.Float64 {
	if timeout != nil {
		return types.Float64Value(*timeout)
	}

	tflog.Warn(ctx, "monitor response carried no timeout, assuming the schema default", map[string]any{
		"id":      id,
		"type":    monitorType,
		"assumed": fallback,
	})

	diags.AddWarning(
		"Monitor returned no timeout",
		fmt.Sprintf(
			"Monitor %d (%s) returned no timeout, so the default of %g was assumed. State may "+
				"not reflect the value stored in Uptime Kuma.",
			id, monitorType, fallback,
		),
	)

	return types.Float64Value(fallback)
}

// int64ToPtr converts a Terraform int64 type to a pointer to int64.
// Returns nil if the value is null or unknown. For columns the Uptime Kuma
// server stores as nullable, nil round-trips as SQL NULL, which is how the
// check is told to apply its own fallback.
func int64ToPtr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return v.ValueInt64Pointer()
}

// int64PtrToTypes converts a pointer to int64 to a Terraform int64 type.
// Returns Int64Null() if the pointer is nil.
func int64PtrToTypes(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}

	return types.Int64Value(*v)
}

// boolToPtr converts a Terraform bool type to a pointer to bool.
// Returns nil if the value is null or unknown.
func boolToPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return v.ValueBoolPointer()
}

// boolPtrToTypes converts a pointer to bool to a Terraform bool type.
// Returns BoolNull() if the pointer is nil.
func boolPtrToTypes(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}

	return types.BoolValue(*v)
}
