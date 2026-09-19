package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// createdWithoutEvent reports a failed create and tells the caller whether the
// resource exists on the server nevertheless.
//
// Uptime Kuma acknowledges a create over socket.io and then broadcasts the
// update event that carries the change. When the acknowledgement arrives but
// the event does not, the client returns an error wrapping
// kuma.ErrUpdateEventTimeout together with the ID the server assigned. The
// resource does exist, so reporting a plain error would leave it out of state
// and the next apply would create a duplicate. Adopt it instead and warn.
//
// It returns false when the create genuinely failed. The error has then been
// reported and the caller must return without touching state.
func createdWithoutEvent(diags *diag.Diagnostics, err error, id int64, summary string) bool {
	if !errors.Is(err, kuma.ErrUpdateEventTimeout) || id == 0 {
		diags.AddError(summary, err.Error())

		return false
	}

	diags.AddWarning(
		"Created without confirmation",
		fmt.Sprintf(
			"Uptime Kuma created the resource with ID %d, but the event confirming it did not "+
				"arrive: %s. The resource has been written to state so that the next apply does "+
				"not create a duplicate. Run `terraform plan` to refresh it.",
			id, err,
		),
	)

	return true
}

// readWithResync fetches a cache-backed resource for a resource read.
//
// The notification, proxy and Docker host getters all serve from the client's
// local state cache. A missed update event leaves that cache stale, so a
// resource that is alive on the server is reported as not found until the next
// resync, and for these three nothing else triggers the broadcast that would
// refill the cache. Treating the miss as a deletion would drop the resource
// from state and make the next apply create a duplicate, so a resync is forced
// once before the miss is believed.
//
// The second return value reports whether the resource exists. The caller
// removes it from state when it does not, after checking diags for errors.
func readWithResync[T any](
	ctx context.Context,
	client *kuma.Client,
	id int64,
	summary string,
	get func(context.Context, int64) (T, error),
	diags *diag.Diagnostics,
) (T, bool) {
	value, err := get(ctx, id)
	if errors.Is(err, kuma.ErrNotFound) {
		resyncErr := client.Resync(ctx)
		if resyncErr != nil {
			diags.AddError("failed to resync with Uptime Kuma", resyncErr.Error())

			return value, false
		}

		value, err = get(ctx, id)
	}

	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			return value, false
		}

		diags.AddError(summary, err.Error())

		return value, false
	}

	return value, true
}
