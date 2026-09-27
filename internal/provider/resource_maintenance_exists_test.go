package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/breml/go-uptime-kuma-client/maintenance"
)

// fakeMaintenanceLister stands in for *kuma.Client. maintenanceExists uses nothing
// else of it, which is why it takes the maintenanceLister interface.
type fakeMaintenanceLister struct {
	windows []maintenance.Maintenance
	err     error

	calls int
}

func (f *fakeMaintenanceLister) GetMaintenances(context.Context) ([]maintenance.Maintenance, error) {
	f.calls++

	if f.err != nil {
		return nil, f.err
	}

	return f.windows, nil
}

func TestMaintenanceExists(t *testing.T) {
	t.Parallel()

	// The zero-valued window is deliberate: data.ID.ValueInt64() is 0 for a null ID,
	// so a cache entry carrying ID 0 must not make such a read report present.
	listed := []maintenance.Maintenance{{ID: 0}, {ID: 7}, {ID: 9}}

	tests := map[string]struct {
		client    *fakeMaintenanceLister
		id        int64
		want      bool
		wantError string
	}{
		"a listed window is present": {
			client: &fakeMaintenanceLister{windows: listed},
			id:     7,
			want:   true,
		},
		"a window listed last is still found": {
			client: &fakeMaintenanceLister{windows: listed},
			id:     9,
			want:   true,
		},
		"an unlisted window is absent": {
			client: &fakeMaintenanceLister{windows: listed},
			id:     8,
		},
		// maintenanceList is a best-effort ready event the resync cannot recover, so
		// this is the shape a server that never sends it produces - and the one that
		// makes the confirmation a no-op. See removeOnServerMiss.
		"an empty list reports every window absent": {
			client: &fakeMaintenanceLister{},
			id:     7,
		},
		// findWithResync requires a failure to arrive as an error and never as an
		// absence: only an absence is worth a resync, and only an absence may cost a
		// window its place in state.
		"a list that will not answer is an error, not an absence": {
			client:    &fakeMaintenanceLister{err: errors.New("socket closed")},
			id:        7,
			wantError: "list maintenance windows: socket closed",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := maintenanceExists(test.client, test.id)(t.Context())

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
			// internally would defeat the resync it is wrapped in.
			if test.client.calls != 1 {
				t.Errorf("want 1 list call, got %d", test.client.calls)
			}
		})
	}
}
