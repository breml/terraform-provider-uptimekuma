package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// The socket.io list events that feed the client's state cache, named as
// Client.MissingReadyEvents reports them.
//
// Uptime Kuma resends monitorList, notificationList and statusPageList on every
// resync; the rest are best-effort, see reportMiss.
const (
	notificationListEvent = "notificationList"
	statusPageListEvent   = "statusPageList"
	maintenanceListEvent  = "maintenanceList"
	proxyListEvent        = "proxyList"
	dockerHostListEvent   = "dockerHostList"
)

// resyncer is the part of *kuma.Client the cache helpers below need. It exists
// so they can be exercised without a socket.io server, see client_errors_test.go.
type resyncer interface {
	// Resync asks the server to resend the lists that feed the state cache.
	Resync(ctx context.Context) error
	// SessionToken is empty for a client that never logged in, which cannot
	// resync at all.
	SessionToken() string
	// MissingReadyEvents names the best-effort lists the server never sent.
	MissingReadyEvents() []string
}

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

// resyncCache rebuilds the client's state cache, so that a lookup which matched
// nothing can be retried against fresh data.
//
// It reports whether that retry is worth making. A client without a session
// token cannot resync at all: the server hands one out only to a login it
// performed for a client that asked, and the provider's username and password
// are optional. That is a limitation of the configuration, not a failure of
// this read, so it warns and returns false with a nil error, leaving the caller
// to report the miss it already has. A resync that was attempted and did fail
// returns the error instead.
func resyncCache(ctx context.Context, client resyncer, diags *diag.Diagnostics) (bool, error) {
	if client.SessionToken() == "" {
		diags.AddWarning(
			"Could not refresh the Uptime Kuma cache",
			"The provider is configured without credentials, so it holds no session token and "+
				"cannot ask the server to resend its lists. A resource created moments ago may "+
				"therefore be reported as missing. Set username and password on the provider to "+
				"enable the refresh.",
		)

		return false, nil
	}

	err := client.Resync(ctx)
	if err != nil {
		return false, fmt.Errorf("resync with Uptime Kuma: %w", err)
	}

	return true, nil
}

// readWithResync fetches a cache-backed resource by ID, for a resource or a
// data source read.
//
// It wraps a getter of the form func(ctx, id) (T, error) that serves from the
// client's local state cache and reports a miss as kuma.ErrNotFound - the
// notification, proxy and Docker host getters. Such a cache goes stale whenever
// a write's broadcast is outstanding: the client confirms creates and deletes
// against the cache, but an edit changes no property it could be checked
// against and so waits for the broadcast by name alone, a create whose
// broadcast never arrived is adopted through createdWithoutEvent with the cache
// still lacking it, and a status page write waits for no broadcast at all. With
// several writes in flight on the one connection a provider holds - Terraform
// runs the operations of an apply concurrently - the cache can therefore be one
// list behind. Believing that miss would drop a live resource from state, or
// fail the very apply that created it, so a resync is forced once before it is
// believed.
//
// found reports whether the resource exists, and is meaningful only when the
// error is nil. A resource read passes a miss to removeOnMiss; a data source
// reports it with reportMiss.
func readWithResync[T any](
	ctx context.Context,
	client resyncer,
	id int64,
	get func(context.Context, int64) (T, error),
	diags *diag.Diagnostics,
) (value T, found bool, err error) {
	value, err = get(ctx, id)
	if errors.Is(err, kuma.ErrNotFound) {
		retry, resyncErr := resyncCache(ctx, client, diags)
		if resyncErr != nil {
			return value, false, resyncErr
		}

		if !retry {
			return value, false, nil
		}

		value, err = get(ctx, id)
	}

	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			return value, false, nil
		}

		return value, false, err
	}

	return value, true, nil
}

// findWithResync looks a cache-backed resource up through a caller-supplied
// predicate, and resyncs once before a miss is believed.
//
// It is readWithResync for a getter that has no func(ctx, id) (T, error) form
// to ask - a whole-list getter, searched by whatever key suits the caller,
// including an ID. The same stale cache is behind both, see readWithResync for
// why it goes stale.
//
// find is called at most twice, and must report a failure as an error rather
// than as a miss: only a miss is worth a resync, and only a miss may be
// reported to the user as a resource that does not exist.
func findWithResync[T any](
	ctx context.Context,
	client resyncer,
	find func(context.Context) (T, bool, error),
	diags *diag.Diagnostics,
) (value T, found bool, err error) {
	value, found, err = find(ctx)
	if err != nil || found {
		return value, found, err
	}

	retry, resyncErr := resyncCache(ctx, client, diags)
	if resyncErr != nil {
		return value, false, resyncErr
	}

	if !retry {
		return value, false, nil
	}

	return find(ctx)
}

// reportMiss reports a cache-backed lookup that matched nothing, for a data
// source read.
//
// summary and detail are what to say when the resource is genuinely absent.
// They are not the whole story, because a successful resync does not refresh
// every list: the client waits for monitorList, notificationList and
// statusPageList, but gives the rest only a short grace period, and drops from
// that wait any list the server did not send when the connection was made. A
// miss in one of those best-effort lists can therefore mean the server never
// sent the list at all - a reverse proxy dropping the socket.io event, or a
// version that does not emit it - and flatly reporting the resource as absent
// would send the reader looking in the wrong place. listEvent names the list
// behind the lookup, so that case can be told apart and said out loud.
func reportMiss(
	diags *diag.Diagnostics,
	client resyncer,
	listEvent string,
	summary string,
	detail string,
) {
	if !slices.Contains(client.MissingReadyEvents(), listEvent) {
		diags.AddError(summary, detail)

		return
	}

	diags.AddError(
		summary,
		fmt.Sprintf(
			"%s Note that Uptime Kuma never sent its %s to the provider, so this may be a lost "+
				"socket.io event - usually a reverse proxy dropping it, or a server version that "+
				"does not emit that list - rather than a resource that does not exist.",
			detail, listEvent,
		),
	)
}

// removeOnMiss drops a resource from Terraform state for a lookup that matched
// nothing, which is how a resource read reports that it was deleted outside
// Terraform.
//
// It refuses to do so when the list behind the lookup never arrived, for the
// reason reportMiss explains: removing the resource would be a silent deletion
// of state that a later apply recreates as a duplicate, on no better evidence
// than a socket.io event the server never sent. It reports an error instead,
// leaving state alone. name is the resource in the words of that error, e.g.
// "proxy".
func removeOnMiss(
	ctx context.Context,
	client resyncer,
	listEvent string,
	name string,
	resp *resource.ReadResponse,
) {
	if !slices.Contains(client.MissingReadyEvents(), listEvent) {
		resp.State.RemoveResource(ctx)

		return
	}

	resp.Diagnostics.AddError(
		fmt.Sprintf("Could not determine whether the %s still exists", name),
		fmt.Sprintf(
			"Uptime Kuma never sent its %s to the provider, so the %s cannot be found in the "+
				"provider's cache and the provider cannot tell whether it was deleted. It has "+
				"been left in state rather than removed, because removing it would make the next "+
				"apply create a duplicate. This is usually a reverse proxy dropping the socket.io "+
				"event, or a server version that does not emit that list.",
			listEvent, name,
		),
	)
}
