package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/breml/go-uptime-kuma-client/statuspage"
)

// fakeStatusPageLister stands in for *kuma.Client. statusPageExists uses nothing
// else of it, which is why it takes the statusPageLister interface.
type fakeStatusPageLister struct {
	pages map[int64]statuspage.StatusPage
	err   error

	calls int
}

func (f *fakeStatusPageLister) GetStatusPages(context.Context) (map[int64]statuspage.StatusPage, error) {
	f.calls++

	if f.err != nil {
		return nil, f.err
	}

	return f.pages, nil
}

func TestStatusPageExists(t *testing.T) {
	t.Parallel()

	// Keyed by ID, matched by slug: the one shape where the two can be confused. The
	// key 7 is deliberately also a plausible slug, so a lookup that matched keys
	// instead of values would show up as "7" being present.
	listed := map[int64]statuspage.StatusPage{
		7: {Slug: "other"},
		9: {Slug: "wanted"},
	}

	tests := map[string]struct {
		client    *fakeStatusPageLister
		slug      string
		want      bool
		wantError string
	}{
		"a listed page is present": {
			client: &fakeStatusPageLister{pages: listed},
			slug:   "wanted",
			want:   true,
		},
		"the page under the lowest key is found too": {
			client: &fakeStatusPageLister{pages: listed},
			slug:   "other",
			want:   true,
		},
		"an unlisted slug is absent": {
			client: &fakeStatusPageLister{pages: listed},
			slug:   "missing",
		},
		"a key is not a slug": {
			client: &fakeStatusPageLister{pages: listed},
			slug:   "7",
		},
		"the empty slug matches nothing": {
			client: &fakeStatusPageLister{pages: listed},
			slug:   "",
		},
		"an empty list reports every page absent": {
			client: &fakeStatusPageLister{},
			slug:   "wanted",
		},
		// findWithResync requires a failure to arrive as an error and never as an
		// absence: only an absence is worth a resync, and only an absence may cost a
		// page its place in state.
		"a list that will not answer is an error, not an absence": {
			client:    &fakeStatusPageLister{err: errors.New("socket closed")},
			slug:      "wanted",
			wantError: "list status pages: socket closed",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := statusPageExists(test.client, test.slug)(t.Context())

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
