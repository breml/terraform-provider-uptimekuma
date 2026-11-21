package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/maintenance"
)

// fakeMaintenanceLister stands in for *kuma.Client. maintenanceExists uses nothing
// else of it, which is why it takes the maintenanceLister interface.
type fakeMaintenanceLister struct {
	windows []maintenance.Maintenance
	err     error

	calls int
}

// GetMaintenances returns windows and err together, as the client does for a
// list it could decode only in part.
func (f *fakeMaintenanceLister) GetMaintenances(context.Context) ([]maintenance.Maintenance, error) {
	f.calls++

	return f.windows, f.err
}

// undecodable is the error the client returns alongside the windows it could
// decode when it could not decode window 8.
func undecodable() error {
	return fmt.Errorf(
		"get maintenances: %w",
		errors.Join(fmt.Errorf("maintenance 8: %w: bad date", kuma.ErrMaintenanceDecode)),
	)
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
		// GetMaintenances returns the windows that decoded alongside an error
		// naming those that did not, and a hit among them is an answer.
		"a listed window is present although another could not be decoded": {
			client: &fakeMaintenanceLister{windows: listed, err: undecodable()},
			id:     7,
			want:   true,
		},
		"a listed window is present whatever the error": {
			client: &fakeMaintenanceLister{windows: listed, err: errors.New("socket closed")},
			id:     9,
			want:   true,
		},
		// The window left out may be the one asked about, so its absence from
		// the rest is no verdict.
		"an unlisted window next to an undecodable one is an error, not an absence": {
			client: &fakeMaintenanceLister{windows: listed, err: undecodable()},
			id:     8,
			wantError: "the windows named here could not be decoded and window 8 may be among them; " +
				"if it really is gone, `terraform state rm` removes it: get maintenances: " +
				"maintenance 8: maintenance window could not be decoded",
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
