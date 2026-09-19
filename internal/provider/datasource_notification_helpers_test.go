package provider

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/breml/go-uptime-kuma-client/notification"
)

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
