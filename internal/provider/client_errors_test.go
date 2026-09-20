package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// fakeResyncer stands in for *kuma.Client. The helpers under test use nothing
// else of it, which is why they take the resyncer interface.
type fakeResyncer struct {
	token   string
	missing []string
	err     error

	resyncs int
}

func (f *fakeResyncer) Resync(context.Context) error {
	f.resyncs++

	return f.err
}

func (f *fakeResyncer) SessionToken() string { return f.token }

func (f *fakeResyncer) MissingReadyEvents() []string { return f.missing }

// loggedIn is the ordinary client: it holds a session token and the server sent
// every list.
func loggedIn() *fakeResyncer { return &fakeResyncer{token: "session-token"} }

func TestResyncCacheWithoutSessionTokenSkipsTheResyncInSilence(t *testing.T) {
	t.Parallel()

	client := &fakeResyncer{token: ""}

	retry, err := resyncCache(t.Context(), client)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}

	if retry {
		t.Error("want retry false, got true")
	}

	if client.resyncs != 0 {
		t.Errorf("want no resync attempt, got %d", client.resyncs)
	}
}

func TestResyncCacheReportsAFailedResync(t *testing.T) {
	t.Parallel()

	client := &fakeResyncer{token: "session-token", err: errors.New("socket closed")}

	retry, err := resyncCache(t.Context(), client)

	if err == nil {
		t.Fatal("want an error, got nil")
	}

	if !strings.Contains(err.Error(), "socket closed") {
		t.Errorf("want the cause in the error, got %q", err)
	}

	if retry {
		t.Error("want retry false, got true")
	}
}

// getter returns a get func that serves ids from the given set, counting calls.
// The second call sees appears as well, which is how a resync that refilled the
// cache is simulated.
func getter(calls *int, has map[int64]string, appears map[int64]string) func(context.Context, int64) (string, error) {
	return func(_ context.Context, id int64) (string, error) {
		*calls++

		if *calls > 1 {
			if value, ok := appears[id]; ok {
				return value, nil
			}
		}

		if value, ok := has[id]; ok {
			return value, nil
		}

		return "", kuma.ErrNotFound
	}
}

func TestReadWithResync(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		client      *fakeResyncer
		has         map[int64]string
		appears     map[int64]string
		wantValue   string
		wantFound   bool
		wantErr     bool
		wantResyncs int
		wantGets    int
	}{
		"a hit does not resync": {
			client:      loggedIn(),
			has:         map[int64]string{1: "proxy"},
			wantValue:   "proxy",
			wantFound:   true,
			wantResyncs: 0,
			wantGets:    1,
		},
		"a miss resyncs once and finds it": {
			client:      loggedIn(),
			appears:     map[int64]string{1: "proxy"},
			wantValue:   "proxy",
			wantFound:   true,
			wantResyncs: 1,
			wantGets:    2,
		},
		"a miss that survives the resync is a miss, with no error": {
			client:      loggedIn(),
			wantFound:   false,
			wantResyncs: 1,
			wantGets:    2,
		},
		"a miss without a session token skips the resync and stays a miss": {
			client:      &fakeResyncer{token: ""},
			wantFound:   false,
			wantResyncs: 0,
			wantGets:    1,
		},
		"a failed resync is an error": {
			client:      &fakeResyncer{token: "session-token", err: errors.New("socket closed")},
			wantFound:   false,
			wantErr:     true,
			wantResyncs: 1,
			wantGets:    1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var gets int

			value, found, err := readWithResync(
				t.Context(), test.client, 1, getter(&gets, test.has, test.appears),
			)

			if (err != nil) != test.wantErr {
				t.Fatalf("want error %v, got %v", test.wantErr, err)
			}

			if found != test.wantFound {
				t.Errorf("want found %v, got %v", test.wantFound, found)
			}

			if found && value != test.wantValue {
				t.Errorf("want value %q, got %q", test.wantValue, value)
			}

			if test.client.resyncs != test.wantResyncs {
				t.Errorf("want %d resyncs, got %d", test.wantResyncs, test.client.resyncs)
			}

			if gets != test.wantGets {
				t.Errorf("want %d gets, got %d", test.wantGets, gets)
			}
		})
	}
}

func TestReadWithResyncDoesNotResyncOnATransportError(t *testing.T) {
	t.Parallel()

	client := loggedIn()
	transportErr := errors.New("connection reset")

	_, found, err := readWithResync(
		t.Context(), client, 1,
		func(context.Context, int64) (string, error) { return "", transportErr },
	)

	if !errors.Is(err, transportErr) {
		t.Fatalf("want the transport error back, got %v", err)
	}

	if found {
		t.Error("want found false, got true")
	}

	// Only kuma.ErrNotFound means the cache may be stale; anything else is a
	// failure the resync cannot fix.
	if client.resyncs != 0 {
		t.Errorf("want no resync, got %d", client.resyncs)
	}
}

func TestFindWithResync(t *testing.T) {
	t.Parallel()

	findErr := errors.New("list unavailable")

	tests := map[string]struct {
		client      *fakeResyncer
		matchOn     int // the find call that matches; 0 means never
		findErr     error
		wantFound   bool
		wantErr     bool
		wantResyncs int
		wantFinds   int
	}{
		"a hit does not resync": {
			client:      loggedIn(),
			matchOn:     1,
			wantFound:   true,
			wantResyncs: 0,
			wantFinds:   1,
		},
		"a miss resyncs once and finds it": {
			client:      loggedIn(),
			matchOn:     2,
			wantFound:   true,
			wantResyncs: 1,
			wantFinds:   2,
		},
		"a miss that survives the resync is a miss, with no error": {
			client:      loggedIn(),
			wantFound:   false,
			wantResyncs: 1,
			wantFinds:   2,
		},
		"a failing find is an error, not a miss, and is not resynced": {
			client:      loggedIn(),
			findErr:     findErr,
			wantErr:     true,
			wantResyncs: 0,
			wantFinds:   1,
		},
		"a miss without a session token skips the resync and stays a miss": {
			client:      &fakeResyncer{token: ""},
			wantFound:   false,
			wantResyncs: 0,
			wantFinds:   1,
		},
		"a failed resync is an error": {
			client:      &fakeResyncer{token: "session-token", err: errors.New("socket closed")},
			wantErr:     true,
			wantResyncs: 1,
			wantFinds:   1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var finds int

			_, found, err := findWithResync(t.Context(), test.client, func(context.Context) (string, bool, error) {
				finds++

				if test.findErr != nil {
					return "", false, test.findErr
				}

				return "match", finds == test.matchOn, nil
			})

			if (err != nil) != test.wantErr {
				t.Fatalf("want error %v, got %v", test.wantErr, err)
			}

			if found != test.wantFound {
				t.Errorf("want found %v, got %v", test.wantFound, found)
			}

			if test.client.resyncs != test.wantResyncs {
				t.Errorf("want %d resyncs, got %d", test.wantResyncs, test.client.resyncs)
			}

			if finds != test.wantFinds {
				t.Errorf("want %d finds, got %d", test.wantFinds, finds)
			}
		})
	}
}

func TestReportMiss(t *testing.T) {
	t.Parallel()

	const (
		summary = "Proxy not found"
		detail  = "No proxy with ID 3 found."
	)

	t.Run("a refreshed list means the resource really is absent", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics

		reportMiss(&diags, loggedIn(), proxyListEvent, summary, detail)

		if diags.ErrorsCount() != 1 {
			t.Fatalf("want 1 error, got %d", diags.ErrorsCount())
		}

		if got := diags.Errors()[0].Detail(); got != detail {
			t.Errorf("want detail %q, got %q", detail, got)
		}
	})

	t.Run("a list the server never sent makes the miss inconclusive", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics

		client := &fakeResyncer{token: "session-token", missing: []string{proxyListEvent}}

		reportMiss(&diags, client, proxyListEvent, summary, detail)

		if diags.ErrorsCount() != 1 {
			t.Fatalf("want 1 error, got %d", diags.ErrorsCount())
		}

		got := diags.Errors()[0].Detail()

		// The plain message still leads, so a caller matching on it - and the
		// acceptance tests that do - still sees what it expects.
		if !strings.HasPrefix(got, detail) {
			t.Errorf("want detail to start with %q, got %q", detail, got)
		}

		if !strings.Contains(got, proxyListEvent) {
			t.Errorf("want the missing list named in %q", got)
		}
	})

	t.Run("a different list missing does not change the message", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics

		client := &fakeResyncer{token: "session-token", missing: []string{dockerHostListEvent}}

		reportMiss(&diags, client, proxyListEvent, summary, detail)

		if got := diags.Errors()[0].Detail(); got != detail {
			t.Errorf("want detail %q, got %q", detail, got)
		}
	})
}

// readResponse builds a ReadResponse holding a populated single-attribute
// state, so that a removal is visible as the state going null.
func readResponse(t *testing.T) *resource.ReadResponse {
	t.Helper()

	stateSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{Computed: true},
		},
	}

	return &resource.ReadResponse{
		State: tfsdk.State{
			Schema: stateSchema,
			Raw: tftypes.NewValue(
				tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.Number}},
				map[string]tftypes.Value{"id": tftypes.NewValue(tftypes.Number, 3)},
			),
		},
	}
}

func TestRemoveOnMissRemovesTheResourceWhenTheListWasRefreshed(t *testing.T) {
	t.Parallel()

	resp := readResponse(t)

	removeOnMiss(t.Context(), loggedIn(), proxyListEvent, "proxy", resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("want no error, got %v", resp.Diagnostics.Errors())
	}

	if !resp.State.Raw.IsNull() {
		t.Error("want the resource removed from state")
	}
}

func TestRemoveOnMissKeepsStateWhenTheServerNeverSentTheList(t *testing.T) {
	t.Parallel()

	resp := readResponse(t)
	client := &fakeResyncer{token: "session-token", missing: []string{proxyListEvent}}

	removeOnMiss(t.Context(), client, proxyListEvent, "proxy", resp)

	// Removing it here would be a silent state deletion on the evidence of an
	// event the server never sent, and the next apply would create a duplicate.
	if resp.State.Raw.IsNull() {
		t.Error("want the resource left in state")
	}

	if resp.Diagnostics.ErrorsCount() != 1 {
		t.Fatalf("want 1 error, got %d", resp.Diagnostics.ErrorsCount())
	}

	if got := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(got, proxyListEvent) {
		t.Errorf("want the missing list named in %q", got)
	}
}

func TestUpdatedWithoutEvent(t *testing.T) {
	t.Parallel()

	const summary = "failed to update notification"

	t.Run("a genuine failure is reported and stops the caller", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics

		if updatedWithoutEvent(&diags, errors.New("server said no"), summary) {
			t.Error("want false for a failed update, got true")
		}

		if diags.ErrorsCount() != 1 {
			t.Fatalf("want 1 error, got %d", diags.ErrorsCount())
		}

		if got := diags.Errors()[0].Summary(); got != summary {
			t.Errorf("want summary %q, got %q", summary, got)
		}
	})

	t.Run("a lost update event is a success with a warning", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics

		err := fmt.Errorf("edit notification: %w", kuma.ErrUpdateEventTimeout)

		// The server applied the write, so the caller must go on to write the
		// planned values to state rather than leave the prior ones there.
		if !updatedWithoutEvent(&diags, err, summary) {
			t.Error("want true for an update that landed, got false")
		}

		if diags.HasError() {
			t.Errorf("want no error diagnostic, got %v", diags.Errors())
		}

		if diags.WarningsCount() != 1 {
			t.Fatalf("want 1 warning, got %d", diags.WarningsCount())
		}
	})
}

func TestDeletedWithoutEvent(t *testing.T) {
	t.Parallel()

	const summary = "failed to delete notification"

	t.Run("a successful delete reports nothing", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics

		if !deletedWithoutEvent(&diags, nil, summary) {
			t.Error("want true for a delete that succeeded, got false")
		}

		if len(diags) != 0 {
			t.Errorf("want no diagnostics, got %v", diags)
		}
	})

	t.Run("a genuine failure is reported as an error", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics

		if deletedWithoutEvent(&diags, errors.New("server said no"), summary) {
			t.Error("want false for a delete that failed, got true")
		}

		if diags.ErrorsCount() != 1 {
			t.Fatalf("want 1 error, got %d", diags.ErrorsCount())
		}

		if got := diags.Errors()[0].Summary(); got != summary {
			t.Errorf("want summary %q, got %q", summary, got)
		}
	})

	t.Run("a lost update event leaves the delete a success", func(t *testing.T) {
		t.Parallel()

		var diags diag.Diagnostics

		err := fmt.Errorf("delete notification: %w", kuma.ErrUpdateEventTimeout)

		if !deletedWithoutEvent(&diags, err, summary) {
			t.Error("want true for a delete that landed, got false")
		}

		// An error here would fail the apply and keep a resource in state that
		// the server no longer has.
		if diags.HasError() {
			t.Errorf("want no error diagnostic, got %v", diags.Errors())
		}

		if diags.WarningsCount() != 1 {
			t.Fatalf("want 1 warning, got %d", diags.WarningsCount())
		}

		// The wording is the contract: it must say the resource is gone, not
		// repeat what updateLanded says about an update.
		if got := diags.Warnings()[0].Summary(); got != "Deleted without confirmation" {
			t.Errorf("want the delete wording, got %q", got)
		}
	})
}

func TestRemoveOnMissKeepsStateWithoutASessionToken(t *testing.T) {
	t.Parallel()

	resp := readResponse(t)
	client := &fakeResyncer{token: ""}

	removeOnMiss(t.Context(), client, notificationListEvent, "notification", resp)

	// Without a token readWithResync could not resync, so the miss may be
	// nothing more than a cache that is one broadcast behind.
	if resp.State.Raw.IsNull() {
		t.Error("want the resource left in state")
	}

	if resp.Diagnostics.ErrorsCount() != 1 {
		t.Fatalf("want 1 error, got %d", resp.Diagnostics.ErrorsCount())
	}

	if got := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(got, "username and password") {
		t.Errorf("want the fix named in %q", got)
	}
}

func TestCreatedWithoutEvent(t *testing.T) {
	t.Parallel()

	const summary = "failed to create notification"

	timeout := fmt.Errorf("add notification: %w", kuma.ErrUpdateEventTimeout)

	tests := map[string]struct {
		err       error
		id        int64
		wantAdopt bool
		wantErrs  int
		wantWarns int
	}{
		"a successful create reports nothing": {
			err:       nil,
			id:        7,
			wantAdopt: true,
		},
		"a genuine failure is an error": {
			err:      errors.New("server said no"),
			wantErrs: 1,
		},
		"a lost create event adopts the ID the server assigned": {
			err:       timeout,
			id:        7,
			wantAdopt: true,
			wantWarns: 1,
		},
		// Upstream reports this combination as impossible: the server does not
		// omit the ID from an ack it calls successful. Guarded against anyway,
		// because adopting ID 0 would write a resource Terraform can never read
		// back.
		"a lost create event with no ID cannot be adopted": {
			err:      timeout,
			id:       0,
			wantErrs: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			if got := createdWithoutEvent(&diags, test.err, test.id, summary); got != test.wantAdopt {
				t.Errorf("want adopt %v, got %v", test.wantAdopt, got)
			}

			if diags.ErrorsCount() != test.wantErrs {
				t.Errorf("want %d errors, got %v", test.wantErrs, diags.Errors())
			}

			if diags.WarningsCount() != test.wantWarns {
				t.Errorf("want %d warnings, got %v", test.wantWarns, diags.Warnings())
			}
		})
	}
}

func TestUpdateLanded(t *testing.T) {
	t.Parallel()

	// updateLanded has a second caller of its own in
	// handleMonitorActiveStateCreate, which acts on the bool alone. Its
	// contract is therefore pinned here rather than only through
	// updatedWithoutEvent.
	tests := map[string]struct {
		err       error
		wantAdopt bool
		wantWarns int
	}{
		"a nil error did not land, because nothing failed": {
			err: nil,
		},
		"an ordinary failure did not land": {
			err: errors.New("server said no"),
		},
		"a bare sentinel landed": {
			err:       kuma.ErrUpdateEventTimeout,
			wantAdopt: true,
			wantWarns: 1,
		},
		"a wrapped sentinel landed": {
			err:       fmt.Errorf("pause monitor 3: %w", kuma.ErrUpdateEventTimeout),
			wantAdopt: true,
			wantWarns: 1,
		},
		"a joined sentinel landed": {
			err:       errors.Join(errors.New("and another thing"), kuma.ErrUpdateEventTimeout),
			wantAdopt: true,
			wantWarns: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			if got := updateLanded(&diags, test.err); got != test.wantAdopt {
				t.Errorf("want landed %v, got %v", test.wantAdopt, got)
			}

			// It warns; it never errors. The caller owns the failing path.
			if diags.HasError() {
				t.Errorf("want no error diagnostic, got %v", diags.Errors())
			}

			if diags.WarningsCount() != test.wantWarns {
				t.Fatalf("want %d warnings, got %v", test.wantWarns, diags.Warnings())
			}

			if test.wantWarns == 0 {
				return
			}

			// The warning may not promise that state already holds the planned
			// values: returning true only clears the caller to reach
			// resp.State.Set, and several callers can still fail before it.
			detail := diags.Warnings()[0].Detail()
			if strings.Contains(detail, "have been written to state") {
				t.Errorf("want no promise about state, got %q", detail)
			}

			if !strings.Contains(detail, "terraform plan") {
				t.Errorf("want the remedy named in %q", detail)
			}
		})
	}
}

func TestUpdatedWithoutEventTakesANilErrorAsSuccess(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics

	// Its sibling deletedWithoutEvent accepts every outcome, and the callers
	// are clones of one another, so this may not be the one that panics.
	if !updatedWithoutEvent(&diags, nil, "failed to update notification") {
		t.Error("want true for an update that succeeded, got false")
	}

	if len(diags) != 0 {
		t.Errorf("want no diagnostics, got %v", diags)
	}
}

func TestReportMissNamesAMissingSessionToken(t *testing.T) {
	t.Parallel()

	const (
		summary = "Notification not found"
		detail  = "No notification with ID 3 found."
	)

	var diags diag.Diagnostics

	reportMiss(&diags, &fakeResyncer{token: ""}, notificationListEvent, summary, detail)

	if diags.ErrorsCount() != 1 {
		t.Fatalf("want 1 error, got %d", diags.ErrorsCount())
	}

	// A data source has no state to protect, so this stays an error - but it
	// must not flatly claim absence the provider could not establish, or it
	// would contradict removeOnMiss about the very same lookup.
	got := diags.Errors()[0].Detail()
	if !strings.Contains(got, detail) {
		t.Errorf("want the caller's detail kept in %q", got)
	}

	if !strings.Contains(got, "without credentials") {
		t.Errorf("want the reason named in %q", got)
	}
}

func TestReadWithResyncThenRemoveOnMissReportsTheTokenOnce(t *testing.T) {
	t.Parallel()

	client := &fakeResyncer{token: ""}
	resp := readResponse(t)

	_, found, err := readWithResync(
		t.Context(), client, 1,
		func(context.Context, int64) (string, error) { return "", kuma.ErrNotFound },
	)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}

	if found {
		t.Fatal("want found false, got true")
	}

	removeOnMiss(t.Context(), client, notificationListEvent, "notification", resp)

	if resp.State.Raw.IsNull() {
		t.Error("want the resource left in state")
	}

	// The two helpers sit on the same read. Exactly one diagnostic may describe
	// the missing token, or the user reads the same paragraph twice.
	if resp.Diagnostics.ErrorsCount() != 1 {
		t.Fatalf("want 1 error, got %v", resp.Diagnostics.Errors())
	}

	if resp.Diagnostics.WarningsCount() != 0 {
		t.Errorf("want no warnings, got %v", resp.Diagnostics.Warnings())
	}
}
