package provider

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/maintenance"
)

func TestMaintenanceCreated(t *testing.T) {
	t.Parallel()

	// The client's errors for a window it created but could not decode, for one
	// whose update event never arrived, and for a read-back that failed.
	decodeErr := fmt.Errorf(
		"create maintenance: maintenance 5: %w: bad date", kuma.ErrMaintenanceDecode,
	)
	timeoutErr := fmt.Errorf("create maintenance: %w", kuma.ErrUpdateEventTimeout)
	readBackErr := errors.New("get maintenance 5: socket closed")

	sentWithID := func() *maintenance.Maintenance { return &maintenance.Maintenance{ID: 5, Title: "sent"} }

	tests := map[string]struct {
		sent    *maintenance.Maintenance
		created *maintenance.Maintenance
		err     error
		// wantID is the ID of the window returned for adoption; zero expects none.
		wantID      int64
		wantSummary string
		// wantWarning is the summary of the one warning expected, and
		// wantInWarning what its detail must contain.
		wantWarning   string
		wantInWarning []string
	}{
		"a create that worked": {
			sent:    &maintenance.Maintenance{},
			created: &maintenance.Maintenance{ID: 5},
			wantID:  5,
		},
		// The window exists, so reporting an error would orphan it and make the
		// next apply create a duplicate.
		"created but not decodable is adopted with a warning": {
			sent:          sentWithID(),
			created:       sentWithID(),
			err:           decodeErr,
			wantID:        5,
			wantWarning:   "Maintenance created but could not be read back",
			wantInWarning: []string{"ID 5", decodeErr.Error(), "fixed in Uptime Kuma"},
		},
		// GetMaintenance returns nil on any failure after the server acknowledged
		// the create; the ID CreateMaintenance set on what was sent is the window.
		"a failed read-back is adopted from what was sent": {
			sent:          sentWithID(),
			err:           readBackErr,
			wantID:        5,
			wantWarning:   "Maintenance created but could not be read back",
			wantInWarning: []string{"ID 5", readBackErr.Error(), "computed ones null"},
		},
		"a decode error without an ID is a failure": {
			sent:        &maintenance.Maintenance{},
			created:     &maintenance.Maintenance{},
			err:         decodeErr,
			wantSummary: "failed to create maintenance",
		},
		"an error before the server assigned an ID is a failure": {
			sent:        &maintenance.Maintenance{},
			err:         errors.New("create maintenance: addMaintenance: invalid"),
			wantSummary: "failed to create maintenance",
		},
		"a lost update event is still createdWithoutEvent's": {
			sent:          sentWithID(),
			created:       sentWithID(),
			err:           timeoutErr,
			wantID:        5,
			wantWarning:   "Created without confirmation",
			wantInWarning: []string{"ID 5"},
		},
		"a lost update event without an ID is a failure": {
			sent:        &maintenance.Maintenance{},
			err:         timeoutErr,
			wantSummary: "failed to create maintenance",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			got, ok := maintenanceCreated(&diags, test.sent, test.created, test.err)
			if ok != (test.wantID != 0) {
				t.Errorf("want ok %t, got %t", test.wantID != 0, ok)
			}

			if ok && (got == nil || got.ID != test.wantID) {
				t.Errorf("want window %d to adopt, got %v", test.wantID, got)
			}

			checkSummary(t, "error", diags.Errors(), test.wantSummary)
			checkSummary(t, "warning", diags.Warnings(), test.wantWarning)

			for _, want := range test.wantInWarning {
				if !strings.Contains(diags.Warnings()[0].Detail(), want) {
					t.Errorf("want %q in the warning detail, got %q", want, diags.Warnings()[0].Detail())
				}
			}
		})
	}
}

// checkSummary expects exactly one diagnostic with summary want, or none when
// want is empty.
func checkSummary(t *testing.T, kind string, diags diag.Diagnostics, want string) {
	t.Helper()

	if want == "" {
		if len(diags) != 0 {
			t.Errorf("want no %s, got %v", kind, diags)
		}

		return
	}

	if len(diags) != 1 || diags[0].Summary() != want {
		t.Fatalf("want one %s %q, got %v", kind, want, diags)
	}
}

// modelFields returns the attributes of data by field name, so a test covers
// an attribute added later without being told about it.
func modelFields(t *testing.T, data MaintenanceResourceModel) map[string]attr.Value {
	t.Helper()

	fields := map[string]attr.Value{}
	value := reflect.ValueOf(data)

	for i := range value.NumField() {
		field, ok := reflect.TypeAssert[attr.Value](value.Field(i))
		if !ok {
			t.Fatalf("field %s is not an attr.Value", value.Type().Field(i).Name)
		}

		fields[value.Type().Field(i).Name] = field
	}

	return fields
}

func TestNullUnknownMaintenanceComputed(t *testing.T) {
	t.Parallel()

	known := func() MaintenanceResourceModel {
		return MaintenanceResourceModel{
			Status:           types.StringValue("scheduled"),
			TimezoneResolved: types.StringValue("Europe/Zurich"),
			TimezoneOffset:   types.StringValue("+02:00"),
		}
	}

	tests := map[string]struct {
		// unknown makes one attribute unknown; nil leaves every one known.
		unknown func(*MaintenanceResourceModel)
		field   string
	}{
		"nothing unknown": {},
		"status": {
			unknown: func(data *MaintenanceResourceModel) { data.Status = types.StringUnknown() },
			field:   "Status",
		},
		"timezone_resolved": {
			unknown: func(data *MaintenanceResourceModel) { data.TimezoneResolved = types.StringUnknown() },
			field:   "TimezoneResolved",
		},
		"timezone_offset": {
			unknown: func(data *MaintenanceResourceModel) { data.TimezoneOffset = types.StringUnknown() },
			field:   "TimezoneOffset",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data := known()
			if test.unknown != nil {
				test.unknown(&data)
			}

			nullUnknownMaintenanceComputed(&data)

			want := modelFields(t, known())

			for field, got := range modelFields(t, data) {
				if field == test.field {
					if !got.IsNull() {
						t.Errorf("want %s null, got %s", field, got)
					}

					continue
				}

				// DeepEqual, because the zero value of a list has no element
				// type and Equal never matches it.
				if !reflect.DeepEqual(got, want[field]) {
					t.Errorf("want %s kept as %s, got %s", field, want[field], got)
				}
			}
		})
	}
}

// TestPopulateModelFromCreatedAdopted builds the state of a window adopted from
// what was sent: no attribute may be left unknown, which Terraform rejects
// after a create, and every value the plan knew must be kept.
func TestPopulateModelFromCreatedAdopted(t *testing.T) {
	t.Parallel()

	timeOfDayTypes := map[string]attr.Type{
		"hours":   types.Int64Type,
		"minutes": types.Int64Type,
		"seconds": types.Int64Type,
	}

	timeOfDay := func(hours int64) types.Object {
		return types.ObjectValueMust(timeOfDayTypes, map[string]attr.Value{
			"hours":   types.Int64Value(hours),
			"minutes": types.Int64Value(0),
			"seconds": types.Int64Value(0),
		})
	}

	tests := map[string]func(*MaintenanceResourceModel){
		"single": func(plan *MaintenanceResourceModel) {
			plan.StartDate = types.StringValue("2030-06-29T13:39:00Z")
			plan.EndDate = types.StringValue("2030-06-29T15:09:00+02:00")
		},
		"cron": func(plan *MaintenanceResourceModel) {
			plan.Cron = types.StringValue("0 3 * * *")
			plan.DurationMinutes = types.Int64Value(60)
		},
		"recurring-interval": func(plan *MaintenanceResourceModel) {
			plan.IntervalDay = types.Int64Value(7)
			plan.StartTime = timeOfDay(2)
			plan.EndTime = timeOfDay(4)
		},
		"recurring-weekday": func(plan *MaintenanceResourceModel) {
			plan.Weekdays = types.ListValueMust(types.Int64Type, []attr.Value{types.Int64Value(1), types.Int64Value(3)})
			plan.StartTime = timeOfDay(2)
			plan.EndTime = timeOfDay(4)
		},
		"recurring-day-of-month": func(plan *MaintenanceResourceModel) {
			plan.DaysOfMonth = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("lastDay1")})
			plan.StartTime = timeOfDay(2)
			plan.EndTime = timeOfDay(4)
		},
		"manual": func(*MaintenanceResourceModel) {},
	}

	for strategy, configure := range tests {
		t.Run(strategy, func(t *testing.T) {
			t.Parallel()

			// The plan of a create: configured and defaulted values known, the
			// rest of the optional ones null, every computed one unknown.
			plan := MaintenanceResourceModel{
				ID:               types.Int64Unknown(),
				Title:            types.StringValue("adopted"),
				Description:      types.StringValue(""),
				Strategy:         types.StringValue(strategy),
				Active:           types.BoolValue(true),
				StartDate:        types.StringNull(),
				EndDate:          types.StringNull(),
				IntervalDay:      types.Int64Null(),
				Weekdays:         types.ListNull(types.Int64Type),
				DaysOfMonth:      types.ListNull(types.StringType),
				Cron:             types.StringUnknown(),
				DurationMinutes:  types.Int64Null(),
				StartTime:        types.ObjectNull(timeOfDayTypes),
				EndTime:          types.ObjectNull(timeOfDayTypes),
				Timezone:         types.StringValue("Europe/Zurich"),
				Status:           types.StringUnknown(),
				TimezoneResolved: types.StringUnknown(),
				TimezoneOffset:   types.StringUnknown(),
				Duration:         types.Int64Unknown(),
				TimeslotList:     types.ListUnknown(types.ObjectType{}),
			}
			configure(&plan)

			r := &MaintenanceResource{}

			var diags diag.Diagnostics

			// What Create sends, with the ID the server acknowledged.
			sent := &maintenance.Maintenance{
				Title:       plan.Title.ValueString(),
				Description: plan.Description.ValueString(),
				Strategy:    strategy,
				Active:      plan.Active.ValueBool(),
			}

			err := r.populateMaintenanceFromModel(t.Context(), &plan, sent, &diags)
			if err != nil {
				t.Fatalf("populate the window to send: %v", err)
			}

			sent.ID = 5

			data := plan
			r.populateModelFromCreated(t.Context(), sent, &data, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			planned := modelFields(t, plan)

			for field, got := range modelFields(t, data) {
				if got.IsUnknown() {
					t.Errorf("%s is unknown after the create", field)
				}

				if want := planned[field]; !want.IsUnknown() && !got.Equal(want) {
					t.Errorf("%s = %s, want the planned %s", field, got, want)
				}
			}

			if data.ID.ValueInt64() != 5 {
				t.Errorf("id = %s, want 5", data.ID)
			}
		})
	}
}
