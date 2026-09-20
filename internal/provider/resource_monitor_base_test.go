package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// fakePauser stands in for *kuma.Client. The active-state helpers use nothing
// else of it, which is why they take the monitorPauser interface.
type fakePauser struct {
	err error

	paused  []int64
	resumed []int64
}

func (f *fakePauser) PauseMonitor(_ context.Context, monitorID int64) error {
	f.paused = append(f.paused, monitorID)

	return f.err
}

func (f *fakePauser) ResumeMonitor(_ context.Context, monitorID int64) error {
	f.resumed = append(f.resumed, monitorID)

	return f.err
}

func TestHandleMonitorActiveStateCreate(t *testing.T) {
	t.Parallel()

	const monitorID = 3

	landed := fmt.Errorf("pause monitor %d: %w", monitorID, kuma.ErrUpdateEventTimeout)

	tests := map[string]struct {
		active     types.Bool
		err        error
		wantPauses int
		wantErr    bool
		wantWarns  int
	}{
		// Uptime Kuma creates every monitor active, so only a configured false
		// is worth an event.
		"an active monitor is not paused": {
			active: types.BoolValue(true),
		},
		"an unset active is not paused": {
			active: types.BoolNull(),
		},
		"an unknown active is not paused": {
			active: types.BoolUnknown(),
		},
		"an inactive monitor is paused": {
			active:     types.BoolValue(false),
			wantPauses: 1,
		},
		"a failed pause is reported to the caller": {
			active:     types.BoolValue(false),
			err:        errors.New("server said no"),
			wantPauses: 1,
			wantErr:    true,
		},
		// The server acked the pause and only the broadcast was lost, so the
		// monitor really is paused: reporting a failure would make the caller
		// fail an apply that worked.
		"a pause that landed without its event is a success": {
			active:     types.BoolValue(false),
			err:        landed,
			wantPauses: 1,
			wantWarns:  1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			client := &fakePauser{err: test.err}

			err := handleMonitorActiveStateCreate(t.Context(), client, monitorID, test.active, &diags)

			if (err != nil) != test.wantErr {
				t.Fatalf("want error %v, got %v", test.wantErr, err)
			}

			if len(client.paused) != test.wantPauses {
				t.Errorf("want %d pauses, got %v", test.wantPauses, client.paused)
			}

			if len(client.resumed) != 0 {
				t.Errorf("want no resumes, got %v", client.resumed)
			}

			// The failure travels in the return value so the caller can persist
			// the new monitor ID before surfacing it; only the landed case adds
			// anything here.
			if diags.HasError() {
				t.Errorf("want no error diagnostic, got %v", diags.Errors())
			}

			if diags.WarningsCount() != test.wantWarns {
				t.Errorf("want %d warnings, got %v", test.wantWarns, diags.Warnings())
			}

			if test.wantErr && !strings.Contains(err.Error(), "failed to pause monitor 3") {
				t.Errorf("want the monitor named in %q", err)
			}
		})
	}
}

func TestHandleMonitorActiveStateUpdate(t *testing.T) {
	t.Parallel()

	const monitorID = 3

	landed := fmt.Errorf("resume monitor %d: %w", monitorID, kuma.ErrUpdateEventTimeout)

	tests := map[string]struct {
		oldActive   types.Bool
		newActive   types.Bool
		err         error
		wantPauses  int
		wantResumes int
		wantErrs    int
		wantWarns   int
	}{
		"an unchanged active state sends nothing": {
			oldActive: types.BoolValue(true),
			newActive: types.BoolValue(true),
		},
		"an unset new active sends nothing": {
			oldActive: types.BoolValue(true),
			newActive: types.BoolNull(),
		},
		"an unknown new active sends nothing": {
			oldActive: types.BoolValue(true),
			newActive: types.BoolUnknown(),
		},
		"switching off pauses": {
			oldActive:  types.BoolValue(true),
			newActive:  types.BoolValue(false),
			wantPauses: 1,
		},
		"switching on resumes": {
			oldActive:   types.BoolValue(false),
			newActive:   types.BoolValue(true),
			wantResumes: 1,
		},
		// A null prior state reads as inactive, so a plan of true is a change.
		"an unset prior state counts as inactive": {
			oldActive:   types.BoolNull(),
			newActive:   types.BoolValue(true),
			wantResumes: 1,
		},
		"a genuine failure is an error": {
			oldActive:   types.BoolValue(false),
			newActive:   types.BoolValue(true),
			err:         errors.New("server said no"),
			wantResumes: 1,
			wantErrs:    1,
		},
		"a resume that landed without its event is a warning": {
			oldActive:   types.BoolValue(false),
			newActive:   types.BoolValue(true),
			err:         landed,
			wantResumes: 1,
			wantWarns:   1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			client := &fakePauser{err: test.err}

			handleMonitorActiveStateUpdate(
				t.Context(), client, monitorID, test.oldActive, test.newActive, &diags,
			)

			if len(client.paused) != test.wantPauses {
				t.Errorf("want %d pauses, got %v", test.wantPauses, client.paused)
			}

			if len(client.resumed) != test.wantResumes {
				t.Errorf("want %d resumes, got %v", test.wantResumes, client.resumed)
			}

			if diags.ErrorsCount() != test.wantErrs {
				t.Errorf("want %d errors, got %v", test.wantErrs, diags.Errors())
			}

			if diags.WarningsCount() != test.wantWarns {
				t.Errorf("want %d warnings, got %v", test.wantWarns, diags.Warnings())
			}

			if test.wantErrs > 0 {
				want := "failed to update active state for monitor 3"
				if got := diags.Errors()[0].Summary(); got != want {
					t.Errorf("want summary %q, got %q", want, got)
				}
			}
		})
	}
}
