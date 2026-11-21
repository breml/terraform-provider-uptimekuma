package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/maintenance"
)

// fakeMaintenanceGetter stands in for *kuma.Client behind maintenanceGetter.
type fakeMaintenanceGetter struct {
	window *maintenance.Maintenance
	err    error
}

func (f fakeMaintenanceGetter) GetMaintenance(context.Context, int64) (*maintenance.Maintenance, error) {
	return f.window, f.err
}

func TestStoredMaintenance(t *testing.T) {
	t.Parallel()

	decodeErr := fmt.Errorf("get maintenance 5: %w: bad date", kuma.ErrMaintenanceDecode)
	window := &maintenance.Maintenance{ID: 5}

	tests := map[string]struct {
		client      fakeMaintenanceGetter
		want        *maintenance.Maintenance
		wantOK      bool
		wantError   string
		wantWarning string
	}{
		"the stored window is returned": {
			client: fakeMaintenanceGetter{window: window},
			want:   window,
			wantOK: true,
		},
		// The update may be the repair of the window, so it goes ahead without
		// a stored window to preserve dates from.
		"an undecodable window lets the update go ahead without it": {
			client:      fakeMaintenanceGetter{err: decodeErr},
			wantOK:      true,
			wantWarning: decodeErr.Error(),
		},
		"another error fails the update": {
			client:    fakeMaintenanceGetter{err: errors.New("socket closed")},
			wantError: "socket closed",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			got, ok := storedMaintenance(t.Context(), test.client, 5, &diags)

			if ok != test.wantOK {
				t.Errorf("want ok %t, got %t", test.wantOK, ok)
			}

			if got != test.want {
				t.Errorf("want stored window %v, got %v", test.want, got)
			}

			checkOneDiag(t, "error", diags.Errors(), test.wantError)
			checkOneDiag(t, "warning", diags.Warnings(), test.wantWarning)
		})
	}
}

func TestAddMaintenanceReadBackWarning(t *testing.T) {
	t.Parallel()

	decodeErr := fmt.Errorf("get maintenance 5: %w: bad date", kuma.ErrMaintenanceDecode)
	otherErr := errors.New("get maintenance 5: socket closed")

	tests := map[string]struct {
		action      string
		err         error
		wantSummary string
		wantIn      []string
		wantNotIn   []string
	}{
		"a read-back that failed on the connection": {
			action:      "updated",
			err:         otherErr,
			wantSummary: "Maintenance updated but could not be read back",
			wantIn:      []string{"Uptime Kuma updated", "ID 5", otherErr.Error(), "filled in by the next refresh"},
			wantNotIn:   []string{"cannot be decoded", maintenanceIncompatibilityAdvice},
		},
		// The window the provider just wrote cannot be decoded, so every refresh
		// will fail, and the cause is more likely the server version than data.
		"a read-back that could not decode the window": {
			action:      "created",
			err:         decodeErr,
			wantSummary: "Maintenance created but could not be read back",
			wantIn: []string{
				"Uptime Kuma created",
				decodeErr.Error(),
				"fails until it is fixed",
				maintenanceIncompatibilityAdvice,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			addMaintenanceReadBackWarning(&diags, test.action, 5, test.err)

			checkSummary(t, "warning", diags.Warnings(), test.wantSummary)

			detail := diags.Warnings()[0].Detail()
			for _, want := range test.wantIn {
				if !strings.Contains(detail, want) {
					t.Errorf("want %q in the warning detail, got %q", want, detail)
				}
			}

			for _, unwanted := range test.wantNotIn {
				if strings.Contains(detail, unwanted) {
					t.Errorf("want no %q in the warning detail, got %q", unwanted, detail)
				}
			}
		})
	}
}

func TestMaintenanceDecodeErrorDetail(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("get maintenance 5: %w: bad date", kuma.ErrMaintenanceDecode)
	detail := maintenanceDecodeErrorDetail(5, err)

	for _, want := range []string{
		"Maintenance window 5 exists",
		err.Error(),
		maintenanceDecodeAdvice,
		"-refresh=false",
		"terraform state rm",
	} {
		if !strings.Contains(detail, want) {
			t.Errorf("want %q in the detail, got %q", want, detail)
		}
	}
}
