package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// The socket.io list events that feed the client's state cache, named as
// Client.MissingReadyEvents reports them.
//
// Uptime Kuma resends monitorList, notificationList and statusPageList on every
// resync; the rest are best-effort, see reportMiss.
//
// Only the best-effort ones can actually be reported missing. MissingReadyEvents
// is filled from the optional lists alone, and Client.New fails outright when a
// required list never arrives, so notificationListEvent and statusPageListEvent
// can never appear there: the helpers take them so every caller reads alike, and
// so a future client that narrows the required set is covered. For those two the
// live branch is the session token one.
//
// monitorList has no constant here because nothing takes it. monitorExists is
// what reads that list, through GetMonitors, and it is handed to
// removeOnServerMiss, which consults neither MissingReadyEvents nor the session
// token. See that helper for why.
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
// It returns false when the create failed, and also for the sentinel with a
// zero ID - a resource that exists but handed back no handle to adopt. Upstream
// reports that combination as impossible, because the server does not omit the
// ID from an ack it reports as successful; it is guarded against rather than
// relied on. The error has then been reported and the caller must return
// without touching state.
//
// A nil error is a create that simply worked, and returns true without saying
// anything, so a caller may hand it every outcome.
func createdWithoutEvent(diags *diag.Diagnostics, err error, id int64, summary string) bool {
	if err == nil {
		return true
	}

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

// updatedWithoutEvent reports a failed update and tells the caller whether the
// server applied it nevertheless.
//
// It is createdWithoutEvent for an edit. The client acknowledges the command
// and then waits for the update event that carries the change; when only the
// event is lost it returns an error wrapping kuma.ErrUpdateEventTimeout, which
// its write methods document as "was updated". Reporting that as a plain error
// would leave Terraform holding the prior values while the server holds the
// new ones, and say "failed to update" about a write that landed, so the
// caller is told to carry on to resp.State.Set instead and the mismatch is
// warned about.
//
// Unlike a create there is no ID to adopt: the resource is already in state.
//
// It returns false when the update genuinely failed. The error has then been
// reported and the caller must return without touching state. A nil error is an
// update that simply worked and returns true in silence, so a caller may hand it
// every outcome, as deletedWithoutEvent does.
func updatedWithoutEvent(diags *diag.Diagnostics, err error, summary string) bool {
	if err == nil || updateLanded(diags, err) {
		return true
	}

	diags.AddError(summary, err.Error())

	return false
}

// updateLanded reports whether a failed write nevertheless took effect on the
// server, and warns about it when it did.
//
// It is updatedWithoutEvent for a caller that reports the failure itself, such
// as one that wraps the error for a caller of its own, and is what keeps the
// two paths saying the same thing about a lost event. It also carries the
// monitor pause and resume writes, which the client documents as "was paused"
// and "was resumed"; the warning therefore speaks of a change rather than
// naming the operation.
//
// The warning describes what the server did, not what the provider will do
// next. Returning true only clears the caller to reach resp.State.Set - it
// cannot make it get there, and several callers have further work that may fail
// first - so promising that state already holds the planned values would
// misdirect whoever is reading the diagnostics.
func updateLanded(diags *diag.Diagnostics, err error) bool {
	if !errors.Is(err, kuma.ErrUpdateEventTimeout) {
		return false
	}

	diags.AddWarning(
		"Applied without confirmation",
		fmt.Sprintf(
			"Uptime Kuma applied the change, but the event confirming it did not arrive: %s. The "+
				"server holds the planned values. If this apply reports an error after this "+
				"warning, Terraform state may still hold the previous ones; run `terraform plan` "+
				"to reconcile.",
			err,
		),
	)

	return true
}

// deletedWithoutEvent reports a failed delete, unless the resource is gone and
// only the event confirming it was lost.
//
// The client's delete methods document an error wrapping
// kuma.ErrUpdateEventTimeout as "was deleted". Reporting it would fail the
// apply and keep the resource in state although the server no longer has it,
// leaving a state entry whose only cure is another destroy. Warning instead
// lets the framework drop it from state, which is what actually happened.
//
// It reports nothing when err is nil, so a caller can hand it every outcome.
//
// It returns false only when the delete genuinely failed, matching its two
// siblings. A Delete whose call is its last statement may ignore that, and most
// do; one with work left must guard on it rather than on
// resp.Diagnostics.HasError(), which would also fire for an unrelated
// diagnostic already on the response.
func deletedWithoutEvent(diags *diag.Diagnostics, err error, summary string) bool {
	switch {
	case err == nil:
	case !errors.Is(err, kuma.ErrUpdateEventTimeout):
		diags.AddError(summary, err.Error())

		return false

	default:
		diags.AddWarning(
			"Deleted without confirmation",
			fmt.Sprintf(
				"Uptime Kuma deleted the resource, but the event confirming it did not arrive: "+
					"%s. It has been removed from state, because the server no longer has it.",
				err,
			),
		)
	}

	return true
}

// resyncCache rebuilds the client's state cache, so that a lookup which matched
// nothing can be retried against fresh data.
//
// It reports whether that retry is worth making. A client without a session
// token cannot resync at all: the server hands one out only to a login it
// performed for a client that asked, and the provider's username and password
// are optional. That is a limitation of the configuration, not a failure of
// this read, so it returns false with a nil error and says nothing, leaving the
// caller to report the miss it already has. removeOnMiss and reportMiss both
// name the missing token themselves when they come to describe that miss, and a
// warning here would only duplicate them. A resync that was attempted and did
// fail returns the error instead.
func resyncCache(ctx context.Context, client resyncer) (bool, error) {
	if client.SessionToken() == "" {
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
// reports it with reportMiss. Both explain a miss the resync could not rule
// out, so nothing is reported from here.
func readWithResync[T any](
	ctx context.Context,
	client resyncer,
	id int64,
	get func(context.Context, int64) (T, error),
) (value T, found bool, err error) {
	value, err = get(ctx, id)
	if errors.Is(err, kuma.ErrNotFound) {
		retry, resyncErr := resyncCache(ctx, client)
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
) (value T, found bool, err error) {
	value, found, err = find(ctx)
	if err != nil || found {
		return value, found, err
	}

	retry, resyncErr := resyncCache(ctx, client)
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
//
// A provider holding no session token could not resync before the miss was
// believed either, which removeOnMiss treats the same way. A data source has no
// state to protect, so this stays the error it always was - but it says which
// of the two it is, so that a resource and a data source reading the same cache
// no longer give different accounts of the same lookup.
func reportMiss(
	diags *diag.Diagnostics,
	client resyncer,
	listEvent string,
	summary string,
	detail string,
) {
	switch {
	case slices.Contains(client.MissingReadyEvents(), listEvent):
		diags.AddError(
			summary,
			fmt.Sprintf(
				"%s Note that Uptime Kuma never sent its %s to the provider, so this may be a "+
					"lost socket.io event - usually a reverse proxy dropping it, or a server "+
					"version that does not emit that list - rather than a resource that does "+
					"not exist.",
				detail, listEvent,
			),
		)

	case client.SessionToken() == "":
		diags.AddError(
			summary,
			fmt.Sprintf(
				"%s Note that the provider is configured without credentials, holds no session "+
					"token and therefore could not ask Uptime Kuma to resend its %s before "+
					"reporting this. A cache that is one update behind looks exactly like a "+
					"resource that does not exist, so this may be a lookup made too early "+
					"rather than a genuine miss. Set username and password on the provider to "+
					"tell the two apart.",
				detail, listEvent,
			),
		)

	default:
		diags.AddError(summary, detail)
	}
}

// removeOnMiss drops a resource from Terraform state for a lookup that matched
// nothing, which is how a resource read reports that it was deleted outside
// Terraform.
//
// It refuses to do so whenever the miss is not evidence of a deletion, and
// reports an error instead, leaving state alone. Removing the resource would
// otherwise be a silent deletion of state that a later apply recreates as a
// duplicate. name is the resource in the words of that error, e.g. "proxy".
//
// There are two such cases. The list behind the lookup may never have arrived,
// for the reason reportMiss explains. Or the provider may hold no session
// token, in which case readWithResync could not resync before believing the
// miss: the cache is one list behind whenever a write's broadcast is
// outstanding, and without the resync that refreshes it a live resource and a
// deleted one look exactly alike.
func removeOnMiss(
	ctx context.Context,
	client resyncer,
	listEvent string,
	name string,
	resp *resource.ReadResponse,
) {
	switch {
	case slices.Contains(client.MissingReadyEvents(), listEvent):
		resp.Diagnostics.AddError(
			fmt.Sprintf("Could not determine whether the %s still exists", name),
			fmt.Sprintf(
				"Uptime Kuma never sent its %s to the provider, so the %s cannot be found in the "+
					"provider's cache and the provider cannot tell whether it was deleted. It "+
					"has been left in state rather than removed, because removing it would make "+
					"the next apply create a duplicate. This is usually a reverse proxy dropping "+
					"the socket.io event, or a server version that does not emit that list.",
				listEvent, name,
			),
		)

	case client.SessionToken() == "":
		resp.Diagnostics.AddError(
			fmt.Sprintf("Could not determine whether the %s still exists", name),
			fmt.Sprintf(
				"The %s is not in the provider's cache, but the provider is configured without "+
					"credentials, holds no session token and therefore cannot ask Uptime Kuma to "+
					"resend its %s. A cache that is one update behind reports a %s that is alive "+
					"exactly like one that was deleted, so it has been left in state rather than "+
					"removed, because removing it would make the next apply create a duplicate. "+
					"Set username and password on the provider so the provider can tell the two "+
					"apart, or remove the %s from state with `terraform state rm` if it really "+
					"is gone.",
				name, listEvent, name, name,
			),
		)

	default:
		resp.State.RemoveResource(ctx)
	}
}

// suspectedNotFound reports whether err might mean the resource is gone, for a
// getter that asks the Uptime Kuma server rather than the client's state cache -
// GetMonitor, GetMonitorAs, GetMaintenance and GetStatusPage.
//
// It is a suspicion and never a verdict. Those getters have no sentinel to
// offer: the server answers a command it could not carry out with a message
// string, which the client relays verbatim as "<command>: <msg>", so the only
// thing left to match on is prose the provider does not own. Two of the strings
// below are the server's:
//
//   - "Cannot read properties of null" is what Node.js throws when the server
//     dereferences a monitor row its query did not find. It is a generic
//     TypeError, so the server can just as well emit it for a fault on a monitor
//     that exists - a null column in a joined record, a partially migrated row.
//   - "No slug?" is the status page handler refusing an unknown slug.
//
// Neither is guaranteed to arrive as prose at all. The ack carries a msgI18n
// flag that marks msg as an untranslated i18n key, and only the client's login
// path looks at it, so a server that localises its errors would defeat both
// matches.
//
// The other three are the client's own words for an ack that reported success
// and then carried no resource, which is the server saying the row is not there.
// As the client's own text they are stable, but a suspicion is still all they
// are: the match is kind-agnostic, so a monitor read that failed for some other
// reason and happened to carry the status page or maintenance wording would
// match just as well.
//
// So callers must confirm the suspicion before acting on it. removeOnServerMiss
// does, and is what every resource read on a server-backed getter uses;
// reportMiss's server-getter counterpart does not exist because no data source
// needs one today. data_source_monitor_globalping.go is the one place that acts
// on an unconfirmed suspicion, and says so.
//
// A nil error is not a suspicion and reports false, so that a caller which
// reached here without a failure - these Read blocks are clones of one another and
// a future clone is the likely way it happens - has removeOnServerMiss report a
// nonsense error rather than crash the provider on a nil dereference.
//
// A kuma.ErrNotFound branch is deliberately absent. No caller of this predicate
// can produce that sentinel: it only ever sees GetMonitor, GetMonitorAs,
// GetMaintenance and GetStatusPage, none of which return it. The five getters
// that do - GetNotification, GetProxy, GetDockerHost, GetTag and GetMonitorTags -
// never reach here, so testing for it would be dead code dressed up as a
// guarantee.
func suspectedNotFound(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	return strings.Contains(msg, "Cannot read properties of null") ||
		strings.Contains(msg, "No slug?") ||
		strings.Contains(msg, "monitor not found in response") ||
		strings.Contains(msg, "maintenance not found in response") ||
		strings.Contains(msg, "config not found in response")
}

// existsWithResync reports whether a resource is in one of the lists the client
// caches, resyncing once before it accepts that it is not.
//
// It is findWithResync for a lookup that wants no value back, only presence, and
// exists purely so the callers of removeOnServerMiss can hand over a
// func(context.Context) (bool, error) instead of threading a value they would
// throw away. Everything findWithResync documents applies, including that exists
// must report a failure of the list getter as an error and never as an absence.
func existsWithResync(
	ctx context.Context,
	client resyncer,
	exists func(context.Context) (bool, error),
) (bool, error) {
	_, found, err := findWithResync(ctx, client, func(ctx context.Context) (struct{}, bool, error) {
		present, err := exists(ctx)

		return struct{}{}, present, err
	})

	return found, err
}

// removeOnServerMiss handles a failed resource read from a getter that asks the
// Uptime Kuma server, dropping the resource from Terraform state only for an
// error that a second lookup agrees means it was deleted.
//
// err is the failure to account for and must not be nil - a nil one is reported
// as an error naming no cause rather than crashing the provider, see
// suspectedNotFound. name is the resource in the words of the diagnostics, e.g.
// "HTTP monitor". exists looks the resource up in the list the client caches for
// its kind, and is run through existsWithResync so the list is refreshed once
// before an absence is accepted.
//
// The list is a veto, not a corroboration, and that is the whole difference from
// removeOnMiss. There the only evidence is a cache miss, so the burden of proof
// is on the miss and an unverifiable one is refused. Here the server itself has
// already answered that the resource is gone, and the list is asked only whether
// it can overturn that:
//
//   - present: the read failed on a resource that exists, so this was a
//     server-side fault and not a deletion. Report it and leave state alone.
//     This is the case removeOnMiss cannot express and the reason this helper
//     exists.
//   - absent: believe the server and remove the resource, which is how a read
//     reports a deletion made outside Terraform.
//   - the lookup itself failed: no verdict either way, so report both failures
//     and keep the resource. This is the one outcome that differs from simply
//     believing err, and it takes a suspicious read error together with a list
//     getter that will not answer.
//
// Because the burden runs that way, this helper deliberately consults neither
// MissingReadyEvents nor the session token, both of which make removeOnMiss
// refuse. A client without credentials holds no session token and so cannot
// resync at all, but its GetMonitor answer was no less authoritative for that,
// and refusing would leave every such provider unable to detect an external
// deletion at all. An unrefreshed list does not condemn the resource - err
// already did - it merely fails to save it.
//
// That is also the limit of what the confirmation is worth, and it differs by
// list. GetMonitors emits getMonitorList, so a monitor absence is an answer the
// server took part in. GetMaintenances emits nothing, and maintenanceList is a
// best-effort ready event that resyncReadyEvents drops for good once it was
// missing at connect time: against a server that never sends it the cache stays
// empty, every window reads as absent, and the veto is a permanent no-op. The
// server's verdict then stands unchallenged, which is what master did for every
// resource. statusPageList is a required ready event, so Client.New fails
// outright when it never arrives and the status page path has no such hole.
//
// Two things it cannot rule out, one in each direction:
//
//   - Absent when the resource lives. The client rebuilds each list from one
//     whole-list broadcast and applies it last-writer-wins, so a snapshot taken
//     before a resource was created can land after it and drop it - see
//     https://github.com/breml/terraform-provider-uptimekuma/issues/430. The
//     resync narrows that window rather than closing it, which is no worse than
//     believing err unchecked.
//   - Present when the resource is gone. findWithResync returns on a hit without
//     resyncing, and the lists are filled asynchronously, so a deletion whose
//     broadcast has not been applied yet reads as present and the read is
//     reported instead of removing the resource. The next refresh clears it once
//     the broadcast lands; if it never does - deleted under another account, a
//     direct database change, a proxy dropping the event - the detail below names
//     terraform state rm, because nothing else will get past it.
//
// Both are deliberate. A false absence is #409 and costs a duplicate resource; a
// false presence costs an error, so the veto is worth having even unrefreshed.
//
// Removing is silent, as it is in removeOnMiss: Terraform reports the drift
// itself, and a confirmed deletion is not news.
func removeOnServerMiss(
	ctx context.Context,
	client resyncer,
	err error,
	exists func(context.Context) (bool, error),
	name string,
	resp *resource.ReadResponse,
) {
	summary := fmt.Sprintf("failed to read %s", name)

	if !suspectedNotFound(err) {
		resp.Diagnostics.AddError(summary, err.Error())

		return
	}

	present, existsErr := existsWithResync(ctx, client, exists)
	if existsErr != nil {
		resp.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"%s. The error may mean the %s no longer exists, but checking whether it is still "+
					"listed failed as well: %s. It has been left in state rather than removed, "+
					"because removing one that does exist would make the next apply create a "+
					"duplicate.",
				err, name, existsErr,
			),
		)

		return
	}

	if present {
		resp.Diagnostics.AddError(
			summary,
			fmt.Sprintf(
				"%s. Uptime Kuma still lists this %s, so the read failed on one that exists rather "+
					"than on one that was deleted - usually a server-side fault such as a null "+
					"column in a record it joins. It has been left in state, because removing it "+
					"would make the next apply create a duplicate. If Uptime Kuma really no longer "+
					"has it, the provider's list of %ss is one broadcast behind; rerun to pick up "+
					"the refreshed list, or remove it from state with `terraform state rm` if the "+
					"broadcast never arrives.",
				err, name, name,
			),
		)

		return
	}

	resp.State.RemoveResource(ctx)
}
