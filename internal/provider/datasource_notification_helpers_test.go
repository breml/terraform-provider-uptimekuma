package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

// fakeNotificationClient stands in for *kuma.Client behind notificationGetter.
// It models the state cache the real getters serve from: cached is what the
// cache holds now, late is what only a resync brings in.
type fakeNotificationClient struct {
	*fakeResyncer

	cached []notification.Base
	late   []notification.Base
	// err is what GetNotification fails with instead of looking anything up.
	err error
}

func (f *fakeNotificationClient) GetNotifications(context.Context) []notification.Base {
	if f.resyncs == 0 {
		return f.cached
	}

	return append(slices.Clone(f.cached), f.late...)
}

func (f *fakeNotificationClient) GetNotification(ctx context.Context, id int64) (notification.Base, error) {
	if f.err != nil {
		return notification.Base{}, f.err
	}

	held := f.GetNotifications(ctx)
	for i := range held {
		if held[i].GetID() == id {
			return held[i], nil
		}
	}

	return notification.Base{}, fmt.Errorf("get notification %d: %w", id, kuma.ErrNotFound)
}

// wantDiag is the single error a lookup is expected to report. The zero value
// expects none.
type wantDiag struct {
	summary string
	// detail is matched as a prefix, because reportMiss appends to it when the
	// miss is inconclusive.
	detail string
	// inDetail must appear somewhere in the detail.
	inDetail string
}

func checkDiag(t *testing.T, diags diag.Diagnostics, want wantDiag) {
	t.Helper()

	if want.summary == "" {
		if diags.HasError() {
			t.Fatalf("want no error, got %v", diags.Errors())
		}

		return
	}

	if diags.ErrorsCount() != 1 {
		t.Fatalf("want 1 error, got %d: %v", diags.ErrorsCount(), diags.Errors())
	}

	got := diags.Errors()[0]

	if got.Summary() != want.summary {
		t.Errorf("want summary %q, got %q", want.summary, got.Summary())
	}

	if !strings.HasPrefix(got.Detail(), want.detail) {
		t.Errorf("want detail to start with %q, got %q", want.detail, got.Detail())
	}

	if !strings.Contains(got.Detail(), want.inDetail) {
		t.Errorf("want %q in detail %q", want.inDetail, got.Detail())
	}
}

// notificationBase builds a notification.Base the way the client does, from the
// server's JSON. Its type is carried in an unexported field, so it cannot be
// set with a struct literal.
func notificationBase(t *testing.T, id int64, name string, notificationType string) notification.Base {
	t.Helper()

	payload, err := json.Marshal(map[string]any{
		"id":     id,
		"name":   name,
		"active": true,
		"type":   notificationType,
		"config": `{"type":"` + notificationType + `"}`,
	})
	if err != nil {
		t.Fatalf("marshal notification: %v", err)
	}

	var base notification.Base

	err = json.Unmarshal(payload, &base)
	if err != nil {
		t.Fatalf("unmarshal notification: %v", err)
	}

	if base.Type() != notificationType {
		t.Fatalf("want type %q, got %q", notificationType, base.Type())
	}

	return base
}

func TestMatchNotificationsByName(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		name             string
		notificationType string
		want             []int64
	}{
		"matches on name and type together": {
			name:             "alerts",
			notificationType: "slack",
			want:             []int64{1},
		},
		"the same name under another type does not match": {
			name:             "alerts",
			notificationType: "discord",
			want:             []int64{4},
		},
		"another name under the same type does not match": {
			name:             "pages",
			notificationType: "slack",
			want:             []int64{3},
		},
		"every duplicate is reported, so the caller can tell ambiguous from absent": {
			name:             "duplicated",
			notificationType: "webhook",
			want:             []int64{5, 6},
		},
		"nothing matching is an empty result rather than a zero ID": {
			name:             "absent",
			notificationType: "slack",
			want:             nil,
		},
	}

	notifications := []notification.Base{
		notificationBase(t, 1, "alerts", "slack"),
		notificationBase(t, 3, "pages", "slack"),
		notificationBase(t, 4, "alerts", "discord"),
		notificationBase(t, 5, "duplicated", "webhook"),
		notificationBase(t, 6, "duplicated", "webhook"),
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := matchNotificationsByName(notifications, test.name, test.notificationType)

			if !slices.Equal(got, test.want) {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}

func TestReadNotificationWithResync(t *testing.T) {
	t.Parallel()

	const missDetail = "No notification with ID 7 found."

	errTransport := errors.New("connection reset")
	cached := []notification.Base{notificationBase(t, 1, "alerts", "slack")}
	late := []notification.Base{notificationBase(t, 7, "pages", "discord")}

	tests := map[string]struct {
		resyncer    *fakeResyncer
		late        []notification.Base
		err         error
		id          int64
		wantFound   bool
		wantResyncs int
		wantDiag    wantDiag
	}{
		"a cached notification is returned without a resync": {
			resyncer:  loggedIn(),
			id:        1,
			wantFound: true,
		},
		"a notification of any type is returned, the type is the caller's to check": {
			resyncer:    loggedIn(),
			late:        late,
			id:          7,
			wantFound:   true,
			wantResyncs: 1,
		},
		"a miss the resync confirms is reported as not found": {
			resyncer:    loggedIn(),
			id:          7,
			wantResyncs: 1,
			wantDiag:    wantDiag{summary: "Notification not found", detail: missDetail, inDetail: missDetail},
		},
		"a miss without a session token says the resync could not be made": {
			resyncer: &fakeResyncer{},
			id:       7,
			wantDiag: wantDiag{summary: "Notification not found", detail: missDetail, inDetail: "without credentials"},
		},
		"a miss in a list the server never sent names that list": {
			resyncer:    &fakeResyncer{token: "session-token", missing: []string{notificationListEvent}},
			id:          7,
			wantResyncs: 1,
			wantDiag:    wantDiag{summary: "Notification not found", detail: missDetail, inDetail: notificationListEvent},
		},
		"a failed resync is a failed read, not a miss": {
			resyncer:    &fakeResyncer{token: "session-token", err: errTransport},
			id:          7,
			wantResyncs: 1,
			wantDiag:    wantDiag{summary: "failed to read notification", inDetail: errTransport.Error()},
		},
		"a getter failure is reported without a resync": {
			resyncer: loggedIn(),
			err:      errTransport,
			id:       1,
			wantDiag: wantDiag{summary: "failed to read notification", detail: errTransport.Error()},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			client := &fakeNotificationClient{fakeResyncer: test.resyncer, cached: cached, late: test.late, err: test.err}

			got, found := readNotificationWithResync(t.Context(), client, test.id, &diags)

			if found != test.wantFound {
				t.Errorf("want found %t, got %t", test.wantFound, found)
			}

			if found && got.GetID() != test.id {
				t.Errorf("want notification %d, got %d", test.id, got.GetID())
			}

			if client.resyncs != test.wantResyncs {
				t.Errorf("want %d resyncs, got %d", test.wantResyncs, client.resyncs)
			}

			checkDiag(t, diags, test.wantDiag)
		})
	}
}

func TestReadNotificationByID(t *testing.T) {
	t.Parallel()

	cached := []notification.Base{
		notificationBase(t, 1, "alerts", "slack"),
		notificationBase(t, 2, "alerts", "discord"),
	}

	tests := map[string]struct {
		id        int64
		wantFound bool
		wantDiag  wantDiag
	}{
		"a notification of the wanted type is returned": {
			id:        1,
			wantFound: true,
		},
		"an ID that belongs to another type is rejected": {
			id: 2,
			wantDiag: wantDiag{
				summary: "Incorrect notification type",
				detail:  "Notification is not a Slack notification",
			},
		},
		"a miss is reported once, as a miss and not as a wrong type": {
			id: 3,
			wantDiag: wantDiag{
				summary: "Notification not found",
				detail:  "No notification with ID 3 found.",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			client := &fakeNotificationClient{fakeResyncer: loggedIn(), cached: cached}

			got, found := readNotificationByID(t.Context(), client, test.id, "slack", "a Slack notification", &diags)

			if found != test.wantFound {
				t.Errorf("want found %t, got %t", test.wantFound, found)
			}

			if found && got.GetID() != test.id {
				t.Errorf("want notification %d, got %d", test.id, got.GetID())
			}

			checkDiag(t, diags, test.wantDiag)
		})
	}
}

func TestFindNotificationByName(t *testing.T) {
	t.Parallel()

	errTransport := errors.New("connection reset")
	cached := []notification.Base{
		notificationBase(t, 1, "alerts", "slack"),
		notificationBase(t, 4, "alerts", "discord"),
		notificationBase(t, 5, "duplicated", "slack"),
		notificationBase(t, 6, "duplicated", "slack"),
		notificationBase(t, 8, "elsewhere", "discord"),
	}
	late := []notification.Base{notificationBase(t, 7, "pages", "slack")}

	tests := map[string]struct {
		resyncer    *fakeResyncer
		name        string
		wantID      int64
		wantResyncs int
		wantDiag    wantDiag
	}{
		"a unique name resolves to its ID, whatever other types share it": {
			resyncer: loggedIn(),
			name:     "alerts",
			wantID:   1,
		},
		"a name the cache is one update behind on is found after a resync": {
			resyncer:    loggedIn(),
			name:        "pages",
			wantID:      7,
			wantResyncs: 1,
		},
		"a name nothing carries is reported as not found": {
			resyncer:    loggedIn(),
			name:        "absent",
			wantResyncs: 1,
			wantDiag: wantDiag{
				summary: "Notification not found",
				detail:  "No slack notification with name 'absent' found.",
			},
		},
		"a name only another type carries is not found": {
			resyncer:    loggedIn(),
			name:        "elsewhere",
			wantResyncs: 1,
			wantDiag: wantDiag{
				summary: "Notification not found",
				detail:  "No slack notification with name 'elsewhere' found.",
			},
		},
		"an ambiguous name is rejected without a resync": {
			resyncer: loggedIn(),
			name:     "duplicated",
			wantDiag: wantDiag{
				summary:  "Multiple notifications found",
				detail:   "Multiple slack notifications with name 'duplicated' found.",
				inDetail: "use 'id'",
			},
		},
		"a failed resync is a failed read, not a miss": {
			resyncer:    &fakeResyncer{token: "session-token", err: errTransport},
			name:        "absent",
			wantResyncs: 1,
			wantDiag:    wantDiag{summary: "failed to read notifications", inDetail: errTransport.Error()},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			client := &fakeNotificationClient{fakeResyncer: test.resyncer, cached: cached, late: late}

			id, found := findNotificationByName(t.Context(), client, test.name, "slack", &diags)

			if found != (test.wantID != 0) {
				t.Errorf("want found %t, got %t", test.wantID != 0, found)
			}

			if id != test.wantID {
				t.Errorf("want ID %d, got %d", test.wantID, id)
			}

			if client.resyncs != test.wantResyncs {
				t.Errorf("want %d resyncs, got %d", test.wantResyncs, client.resyncs)
			}

			checkDiag(t, diags, test.wantDiag)
		})
	}
}

func TestValidateNotificationDataSourceInput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		id   types.Int64
		name types.String
		want bool
	}{
		"an id alone is enough": {
			id:   types.Int64Value(1),
			name: types.StringNull(),
			want: true,
		},
		"a name alone is enough": {
			id:   types.Int64Null(),
			name: types.StringValue("alerts"),
			want: true,
		},
		"both may be set": {
			id:   types.Int64Value(1),
			name: types.StringValue("alerts"),
			want: true,
		},
		"an unknown id falls back to the name": {
			id:   types.Int64Unknown(),
			name: types.StringValue("alerts"),
			want: true,
		},
		"neither set is rejected": {
			id:   types.Int64Null(),
			name: types.StringNull(),
		},
		"neither known is rejected": {
			id:   types.Int64Unknown(),
			name: types.StringUnknown(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &datasource.ReadResponse{}

			got := validateNotificationDataSourceInput(resp, test.id, test.name)

			if got != test.want {
				t.Errorf("want %t, got %t", test.want, got)
			}

			var want wantDiag
			if !test.want {
				want = wantDiag{summary: "Missing query parameters", detail: "Either 'id' or 'name' must be specified."}
			}

			checkDiag(t, resp.Diagnostics, want)
		})
	}
}
