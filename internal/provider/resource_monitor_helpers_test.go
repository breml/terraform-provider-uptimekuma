package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/breml/go-uptime-kuma-client/monitor"
)

// fakeMonitorLister stands in for *kuma.Client. monitorExists uses nothing else
// of it, which is why it takes the monitorLister interface.
type fakeMonitorLister struct {
	monitors []monitor.Base
	err      error

	calls int
}

func (f *fakeMonitorLister) GetMonitors(context.Context) ([]monitor.Base, error) {
	f.calls++

	if f.err != nil {
		return nil, f.err
	}

	return f.monitors, nil
}

func TestMonitorExists(t *testing.T) {
	t.Parallel()

	// The zero-valued monitor is deliberate: data.ID.ValueInt64() is 0 for a null ID,
	// so a cache entry carrying ID 0 must not make such a read report present.
	listed := []monitor.Base{{ID: 0}, {ID: 7}, {ID: 9}}

	tests := map[string]struct {
		client    *fakeMonitorLister
		id        int64
		want      bool
		wantError string
	}{
		"a listed monitor is present": {
			client: &fakeMonitorLister{monitors: listed},
			id:     7,
			want:   true,
		},
		"an unlisted monitor is absent": {
			client: &fakeMonitorLister{monitors: listed},
			id:     8,
		},
		"an empty list reports every monitor absent": {
			client: &fakeMonitorLister{},
			id:     7,
		},
		// findWithResync requires a failure to arrive as an error and never as an
		// absence: only an absence is worth a resync, and only an absence may cost a
		// monitor its place in state.
		"a list that will not answer is an error, not an absence": {
			client: &fakeMonitorLister{err: errors.New("socket closed")},
			id:     7,
			// The wrapping is part of the contract: removeOnServerMiss prints this
			// error beside the read error it is confirming, and "socket closed" alone
			// would not say which of the two failed.
			wantError: "list monitors: socket closed",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := monitorExists(test.client, test.id)(t.Context())

			if test.wantError != "" {
				if err == nil {
					t.Fatalf("want an error, got present %t", got)
				}

				if !strings.Contains(err.Error(), test.wantError) {
					t.Errorf("want %q in the error, got %q", test.wantError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("want no error, got %v", err)
				}

				if got != test.want {
					t.Errorf("want present %t, got %t", test.want, got)
				}
			}

			// existsWithResync calls the closure once per look; a getter that retried
			// internally would defeat the resync it is wrapped in. Asserted on the
			// error path too, which the early return used to skip.
			if test.client.calls != 1 {
				t.Errorf("want 1 list call, got %d", test.client.calls)
			}
		})
	}
}

func TestMonitorTypeDrifted(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		managed     string
		actual      string
		wantDrifted bool
	}{
		"a monitor that is still its own type has not drifted": {
			managed: "http",
			actual:  "http",
		},
		// Base.UnmarshalJSON leaves internalType empty for a record that carried no
		// type, which says nothing about drift and must not cost the monitor its state.
		"a record with no type has not drifted": {
			managed: "http",
			actual:  "",
		},
		"a monitor the server reports as another type has drifted": {
			managed:     "http",
			actual:      "ping",
			wantDrifted: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := readResponse(t)

			got := monitorTypeDrifted(t.Context(), 3, test.managed, test.actual, resp)
			if got != test.wantDrifted {
				t.Fatalf("want drifted %t, got %t", test.wantDrifted, got)
			}

			// It never errors: the read succeeded, and the replacement Terraform plans
			// is the answer, not a failure.
			if resp.Diagnostics.HasError() {
				t.Errorf("want no error, got %v", resp.Diagnostics.Errors())
			}

			if removed := resp.State.Raw.IsNull(); removed != test.wantDrifted {
				t.Errorf("want removed %t, got %t", test.wantDrifted, removed)
			}

			if !test.wantDrifted {
				if len(resp.Diagnostics) != 0 {
					t.Errorf("want no diagnostics, got %v", resp.Diagnostics)
				}

				return
			}

			// The removal must not be silent: without a diagnostic the resource leaves
			// state with nothing but a TF_LOG line to explain it, and the monitor now at
			// that ID stops being managed by anything.
			if resp.Diagnostics.WarningsCount() != 1 {
				t.Fatalf("want 1 warning, got %v", resp.Diagnostics.Warnings())
			}

			detail := resp.Diagnostics.Warnings()[0].Detail()
			for _, want := range []string{test.actual, test.managed, "duplicate"} {
				if !strings.Contains(detail, want) {
					t.Errorf("want %q in the warning, got %q", want, detail)
				}
			}
		})
	}
}
