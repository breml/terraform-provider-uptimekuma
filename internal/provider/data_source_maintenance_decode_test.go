package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/breml/go-uptime-kuma-client/maintenance"
)

// maintenanceListing is one answer of GetMaintenances: the windows that decoded
// and the error that comes with them.
type maintenanceListing struct {
	windows []maintenance.Maintenance
	err     error
}

// fakeMaintenanceListResyncer stands in for *kuma.Client behind
// maintenanceListResyncer. It answers with before until a resync was
// attempted and with after from then on.
type fakeMaintenanceListResyncer struct {
	*fakeResyncer

	before maintenanceListing
	after  maintenanceListing

	looks int
}

func (f *fakeMaintenanceListResyncer) GetMaintenances(context.Context) ([]maintenance.Maintenance, error) {
	f.looks++

	if f.resyncs == 0 {
		return f.before.windows, f.before.err
	}

	return f.after.windows, f.after.err
}

func TestFindMaintenanceByTitle(t *testing.T) {
	t.Parallel()

	wanted := maintenance.Maintenance{ID: 7, Title: "wanted"}
	other := maintenance.Maintenance{ID: 9, Title: "other"}

	tests := map[string]struct {
		before maintenanceListing
		after  maintenanceListing
		// wantID is the window found; zero expects none.
		wantID   int64
		wantDiag wantDiag
		// wantWarning must appear in the detail of the one warning expected; empty
		// expects none.
		wantWarning string
		wantLooks   int
	}{
		"found without a decode error": {
			before:    maintenanceListing{windows: []maintenance.Maintenance{other, wanted}},
			wantID:    7,
			wantLooks: 1,
		},
		// The window looked for decoded, so the one that did not is no reason to
		// fail the lookup, but the match is unique only among the decoded ones.
		"found next to an undecodable window": {
			before: maintenanceListing{windows: []maintenance.Maintenance{wanted}, err: undecodable()},
			wantID: 7,
			wantWarning: "The maintenance window with title 'wanted' is unique only among the windows that " +
				"could be decoded: " + undecodable().Error() + ". " + maintenanceDecodeAdvice,
			wantLooks: 1,
		},
		// A miss next to an undecodable window is still a miss worth a resync,
		// which may bring the window in.
		"found after a resync that followed a partial list": {
			before:    maintenanceListing{windows: []maintenance.Maintenance{other}, err: undecodable()},
			after:     maintenanceListing{windows: []maintenance.Maintenance{other, wanted}},
			wantID:    7,
			wantLooks: 2,
		},
		"a miss names the undecodable windows": {
			before: maintenanceListing{windows: []maintenance.Maintenance{other}, err: undecodable()},
			after:  maintenanceListing{windows: []maintenance.Maintenance{other}, err: undecodable()},
			wantDiag: wantDiag{
				summary: "Maintenance not found",
				detail: "No maintenance window with title 'wanted' found. Some maintenance windows " +
					"could not be decoded and could be the one looked for: ",
				inDetail: undecodable().Error() + ". " + maintenanceDecodeAdvice +
					" Looking the window up by id does not help.",
			},
			wantLooks: 2,
		},
		// Only the look the outcome rests on is reported.
		"a miss after a resync that decoded everything is a plain miss": {
			before: maintenanceListing{windows: []maintenance.Maintenance{other}, err: undecodable()},
			after:  maintenanceListing{windows: []maintenance.Maintenance{other}},
			wantDiag: wantDiag{
				summary: "Maintenance not found",
				detail:  "No maintenance window with title 'wanted' found.",
				exact:   true,
			},
			wantLooks: 2,
		},
		// Any other failure is no answer at all and so not worth a resync.
		"another error fails the lookup": {
			before: maintenanceListing{windows: []maintenance.Maintenance{wanted}, err: errors.New("socket closed")},
			wantDiag: wantDiag{
				summary: "failed to read maintenances",
				detail:  "get maintenances: socket closed",
				exact:   true,
			},
			wantLooks: 1,
		},
		"an ambiguous title next to an undecodable window": {
			before: maintenanceListing{
				windows: []maintenance.Maintenance{wanted, {ID: 11, Title: "wanted"}},
				err:     undecodable(),
			},
			wantDiag:  wantDiag{summary: "Multiple maintenances found"},
			wantLooks: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := &fakeMaintenanceListResyncer{
				fakeResyncer: loggedIn(),
				before:       test.before,
				after:        test.after,
			}

			var diags diag.Diagnostics

			got, ok := findMaintenanceByTitle(t.Context(), client, "wanted", &diags)

			checkDiag(t, diags, test.wantDiag)

			if ok != (test.wantID != 0) {
				t.Errorf("want found %t, got %t", test.wantID != 0, ok)
			}

			if got.ID != test.wantID {
				t.Errorf("want window %d, got %d", test.wantID, got.ID)
			}

			checkOneDiag(t, "warning", diags.Warnings(), test.wantWarning)

			if client.looks != test.wantLooks {
				t.Errorf("want %d looks, got %d", test.wantLooks, client.looks)
			}
		})
	}
}

func TestListMaintenances(t *testing.T) {
	t.Parallel()

	listed := []maintenance.Maintenance{{ID: 7}, {ID: 9}}

	tests := map[string]struct {
		client      *fakeMaintenanceLister
		wantOK      bool
		wantCount   int
		wantError   string
		wantWarning string
	}{
		"every window decoded": {
			client:    &fakeMaintenanceLister{windows: listed},
			wantOK:    true,
			wantCount: 2,
		},
		"an undecodable window is left out with a warning": {
			client:      &fakeMaintenanceLister{windows: listed, err: undecodable()},
			wantOK:      true,
			wantCount:   2,
			wantWarning: undecodable().Error() + ". " + maintenanceDecodeAdvice,
		},
		"another error fails the read": {
			client:    &fakeMaintenanceLister{windows: listed, err: errors.New("socket closed")},
			wantError: "socket closed",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			got, ok := listMaintenances(t.Context(), test.client, &diags)

			if ok != test.wantOK {
				t.Errorf("want ok %t, got %t", test.wantOK, ok)
			}

			if len(got) != test.wantCount {
				t.Errorf("want %d windows, got %d", test.wantCount, len(got))
			}

			checkOneDiag(t, "error", diags.Errors(), test.wantError)
			checkOneDiag(t, "warning", diags.Warnings(), test.wantWarning)
		})
	}
}

// checkOneDiag expects exactly one diagnostic whose detail contains want, or
// none when want is empty.
func checkOneDiag(t *testing.T, kind string, diags diag.Diagnostics, want string) {
	t.Helper()

	if want == "" {
		if len(diags) != 0 {
			t.Errorf("want no %s, got %v", kind, diags)
		}

		return
	}

	if len(diags) != 1 {
		t.Fatalf("want 1 %s, got %d: %v", kind, len(diags), diags)
	}

	if !strings.Contains(diags[0].Detail(), want) {
		t.Errorf("want %q in the %s detail, got %q", want, kind, diags[0].Detail())
	}
}
