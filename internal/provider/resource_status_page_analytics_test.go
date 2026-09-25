package provider

import (
	"slices"
	"testing"

	"github.com/breml/go-uptime-kuma-client/statuspage"
)

func TestStatusPageAnalyticsTypes(t *testing.T) {
	t.Parallel()

	allowed := analyticsTypes()

	if len(allowed) == 0 {
		t.Fatal("analyticsTypes() returned an empty allowlist")
	}

	for _, analyticsType := range allowed {
		if !statuspage.ValidAnalyticsType(&analyticsType) {
			t.Errorf("analyticsTypes() offers %q, which the client rejects", analyticsType)
		}
	}

	if !slices.Contains(allowed, statuspage.AnalyticsTypeRybbit()) {
		t.Errorf("analyticsTypes() = %v, want it to contain %q", allowed, statuspage.AnalyticsTypeRybbit())
	}
}
