package provider

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/breml/go-uptime-kuma-client/maintenance"
)

// The fixtures below carry a timezone other than UTC, so that a read path that
// converted a floating date through In(loc) would show up as a wrong offset.
const testMaintenanceTimezone = "Europe/Zurich"

// cest is the offset of Europe/Zurich in summer, without depending on tzdata.
func cest() *time.Location {
	return time.FixedZone("CEST", 2*60*60)
}

func TestPopulateModelFromMaintenanceSingle(t *testing.T) {
	t.Parallel()

	floatingStart := maintenance.NewFloatingDateTime(time.Date(2030, 6, 29, 13, 39, 0, 0, time.UTC))
	floatingEnd := maintenance.NewFloatingDateTime(time.Date(2030, 6, 29, 15, 9, 0, 0, time.UTC))
	offsetStart := maintenance.NewDateTime(time.Date(2030, 6, 29, 13, 39, 0, 0, cest()))
	offsetEnd := maintenance.NewDateTime(time.Date(2030, 6, 29, 15, 9, 0, 0, cest()))
	utcStart := maintenance.NewDateTime(time.Date(2025, 12, 31, 10, 0, 0, 0, time.UTC))
	utcEnd := maintenance.NewDateTime(time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC))

	// Every case starts from these, so a date the server does not return must be
	// seen to be replaced by null rather than kept.
	const (
		priorStart = "prior-start"
		priorEnd   = "prior-end"
	)

	tests := map[string]struct {
		dateRange []maintenance.DateTime
		wantStart types.String
		wantEnd   types.String
	}{
		"floating range is read as its wall-clock digits with Z": {
			dateRange: []maintenance.DateTime{floatingStart, floatingEnd},
			wantStart: types.StringValue("2030-06-29T13:39:00Z"),
			wantEnd:   types.StringValue("2030-06-29T15:09:00Z"),
		},
		"RFC 3339 range keeps its non-UTC offset": {
			dateRange: []maintenance.DateTime{offsetStart, offsetEnd},
			wantStart: types.StringValue("2030-06-29T13:39:00+02:00"),
			wantEnd:   types.StringValue("2030-06-29T15:09:00+02:00"),
		},
		"RFC 3339 range in UTC": {
			dateRange: []maintenance.DateTime{utcStart, utcEnd},
			wantStart: types.StringValue("2025-12-31T10:00:00Z"),
			wantEnd:   types.StringValue("2025-12-31T12:00:00Z"),
		},
		"a one-element range has a null end": {
			dateRange: []maintenance.DateTime{utcStart},
			wantStart: types.StringValue("2025-12-31T10:00:00Z"),
			wantEnd:   types.StringNull(),
		},
		"an empty range is null": {
			dateRange: nil,
			wantStart: types.StringNull(),
			wantEnd:   types.StringNull(),
		},
		"absent values are null rather than year 1": {
			dateRange: make([]maintenance.DateTime, 2),
			wantStart: types.StringNull(),
			wantEnd:   types.StringNull(),
		},
		"an absent start is null, the end is read": {
			dateRange: []maintenance.DateTime{{}, utcEnd},
			wantStart: types.StringNull(),
			wantEnd:   types.StringValue("2025-12-31T12:00:00Z"),
		},
		"a present start with an absent end": {
			dateRange: []maintenance.DateTime{floatingStart, {}},
			wantStart: types.StringValue("2030-06-29T13:39:00Z"),
			wantEnd:   types.StringNull(),
		},
		// Only an absent date, decoded from null or "", is zero. A year-1 date the
		// server stores is a date like any other and is read as stored.
		"a stored year-1 date is present": {
			dateRange: []maintenance.DateTime{
				maintenance.NewFloatingDateTime(time.Time{}),
				maintenance.NewDateTime(time.Time{}),
			},
			wantStart: types.StringValue("0001-01-01T00:00:00Z"),
			wantEnd:   types.StringValue("0001-01-01T00:00:00Z"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data := MaintenanceResourceModel{
				StartDate: types.StringValue(priorStart),
				EndDate:   types.StringValue(priorEnd),
			}

			(&MaintenanceResource{}).populateModelFromMaintenanceSingle(
				&maintenance.Maintenance{
					Strategy:       "single",
					Timezone:       testMaintenanceTimezone,
					TimezoneOption: testMaintenanceTimezone,
					DateRange:      tc.dateRange,
				},
				&data,
			)

			if !data.StartDate.Equal(tc.wantStart) {
				t.Errorf("start_date = %s, want %s", data.StartDate, tc.wantStart)
			}

			if !data.EndDate.Equal(tc.wantEnd) {
				t.Errorf("end_date = %s, want %s", data.EndDate, tc.wantEnd)
			}
		})
	}
}

func TestPopulateMaintenanceTimeslotList(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		timeslot  maintenance.Timeslot
		wantStart types.String
		wantEnd   types.String
	}{
		"floating dates are read as their wall-clock digits with Z": {
			timeslot: maintenance.Timeslot{
				StartDate: maintenance.NewFloatingDateTime(time.Date(2030, 6, 29, 13, 39, 0, 0, time.UTC)),
				EndDate:   maintenance.NewFloatingDateTime(time.Date(2030, 6, 29, 15, 9, 0, 0, time.UTC)),
			},
			wantStart: types.StringValue("2030-06-29T13:39:00Z"),
			wantEnd:   types.StringValue("2030-06-29T15:09:00Z"),
		},
		"dates with an offset are formatted as before": {
			timeslot: maintenance.Timeslot{
				StartDate: maintenance.NewDateTime(time.Date(2030, 6, 29, 13, 39, 0, 0, cest())),
				EndDate:   maintenance.NewDateTime(time.Date(2030, 6, 29, 15, 9, 0, 0, time.UTC)),
			},
			wantStart: types.StringValue("2030-06-29T13:39:00+02:00"),
			wantEnd:   types.StringValue("2030-06-29T15:09:00Z"),
		},
		"absent dates are null rather than year 1": {
			timeslot:  maintenance.Timeslot{},
			wantStart: types.StringNull(),
			wantEnd:   types.StringNull(),
		},
		"an absent end is null, the start is kept": {
			timeslot: maintenance.Timeslot{
				StartDate: maintenance.NewFloatingDateTime(time.Date(2030, 6, 29, 13, 39, 0, 0, time.UTC)),
			},
			wantStart: types.StringValue("2030-06-29T13:39:00Z"),
			wantEnd:   types.StringNull(),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var (
				data  MaintenanceResourceModel
				diags diag.Diagnostics
			)

			(&MaintenanceResource{}).populateMaintenanceTimeslotList(
				t.Context(),
				&maintenance.Maintenance{
					Timezone:       testMaintenanceTimezone,
					TimezoneOption: testMaintenanceTimezone,
					TimeslotList:   []maintenance.Timeslot{tc.timeslot},
				},
				&data,
				&diags,
			)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			timeslot := firstTimeslot(t, data)

			if !timeslot.StartDate.Equal(tc.wantStart) {
				t.Errorf("start_date = %s, want %s", timeslot.StartDate, tc.wantStart)
			}

			if !timeslot.EndDate.Equal(tc.wantEnd) {
				t.Errorf("end_date = %s, want %s", timeslot.EndDate, tc.wantEnd)
			}
		})
	}
}

func firstTimeslot(t *testing.T, data MaintenanceResourceModel) TimeslotModel {
	t.Helper()

	var timeslots []TimeslotModel

	diags := data.TimeslotList.ElementsAs(t.Context(), &timeslots, false)
	if diags.HasError() {
		t.Fatalf("decode timeslot_list: %v", diags)
	}

	if len(timeslots) != 1 {
		t.Fatalf("timeslot_list has %d elements, want 1", len(timeslots))
	}

	return timeslots[0]
}

// TestPopulateModelFromMaintenanceUIPayload decodes a window the way the
// Uptime Kuma UI stores it and maps it as Read does.
func TestPopulateModelFromMaintenanceUIPayload(t *testing.T) {
	t.Parallel()

	const payload = `{"strategy":"single","timezoneOption":"Europe/Zurich","timezone":"Europe/Zurich",` +
		`"dateRange":["2026-06-29T13:39","2026-06-29T15:09"],` +
		`"timeslotList":[{"startDate":"2026-06-29T13:39","endDate":"2026-06-29T15:09"}]}`

	var m maintenance.Maintenance

	err := json.Unmarshal([]byte(payload), &m)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	var (
		data  MaintenanceResourceModel
		diags diag.Diagnostics
	)

	(&MaintenanceResource{}).populateModelFromMaintenance(t.Context(), &m, &data, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	timeslot := firstTimeslot(t, data)

	got := map[string]types.String{
		"start_date":                 data.StartDate,
		"end_date":                   data.EndDate,
		"timeslot_list.0.start_date": timeslot.StartDate,
		"timeslot_list.0.end_date":   timeslot.EndDate,
		"timezone":                   data.Timezone,
	}

	want := map[string]string{
		"start_date":                 "2026-06-29T13:39:00Z",
		"end_date":                   "2026-06-29T15:09:00Z",
		"timeslot_list.0.start_date": "2026-06-29T13:39:00Z",
		"timeslot_list.0.end_date":   "2026-06-29T15:09:00Z",
		"timezone":                   testMaintenanceTimezone,
	}

	for attr, value := range want {
		if !got[attr].Equal(types.StringValue(value)) {
			t.Errorf("%s = %s, want %q", attr, got[attr], value)
		}
	}
}

// TestPopulateMaintenanceFromModelSingle shows that a date the provider writes
// keeps its offset and its text, and reads back as configured.
func TestPopulateMaintenanceFromModelSingle(t *testing.T) {
	t.Parallel()

	for _, date := range []string{"2030-06-29T13:39:00Z", "2030-06-29T13:39:00+02:00"} {
		t.Run(date, func(t *testing.T) {
			t.Parallel()

			r := &MaintenanceResource{}
			planned := MaintenanceResourceModel{
				StartDate: types.StringValue(date),
				EndDate:   types.StringValue(date),
				Timezone:  types.StringValue(testMaintenanceTimezone),
			}

			var m maintenance.Maintenance

			err := r.populateMaintenanceFromModelSingle(&planned, &m)
			if err != nil {
				t.Fatalf("populate: %v", err)
			}

			if len(m.DateRange) != 2 {
				t.Fatalf("date range has %d elements, want 2", len(m.DateRange))
			}

			for i, d := range m.DateRange {
				if d.IsZero() || d.Floating() {
					t.Fatalf("date %d = %v, want a present date with an offset", i, d)
				}

				encoded, err := json.Marshal(d)
				if err != nil {
					t.Fatalf("encode date %d: %v", i, err)
				}

				if string(encoded) != fmt.Sprintf("%q", date) {
					t.Errorf("date %d encodes as %s, want %q", i, encoded, date)
				}
			}

			var read MaintenanceResourceModel

			r.populateModelFromMaintenanceSingle(&m, &read)

			if read.StartDate.ValueString() != date || read.EndDate.ValueString() != date {
				t.Errorf("read back as %s / %s, want %q", read.StartDate, read.EndDate, date)
			}
		})
	}
}

// describeDateRange renders a date range so two can be compared by kind and
// value, absent elements included.
func describeDateRange(dateRange []maintenance.DateTime) []string {
	described := make([]string, len(dateRange))

	for i, d := range dateRange {
		switch {
		case d.IsZero():
			described[i] = "absent"

		case d.Floating():
			described[i] = "floating " + d.String()

		default:
			described[i] = "offset " + d.String()
		}
	}

	return described
}

func TestPreserveStoredMaintenanceDates(t *testing.T) {
	t.Parallel()

	floating := func(hour, minute int) maintenance.DateTime {
		return maintenance.NewFloatingDateTime(time.Date(2030, 6, 29, hour, minute, 0, 0, time.UTC))
	}

	offset := func(hour, minute int, loc *time.Location) maintenance.DateTime {
		return maintenance.NewDateTime(time.Date(2030, 6, 29, hour, minute, 0, 0, loc))
	}

	single := func(stored ...maintenance.DateTime) *maintenance.Maintenance {
		return &maintenance.Maintenance{Strategy: "single", DateRange: stored}
	}

	// What Update builds from a plan with start_date 13:39Z and end_date 15:09Z.
	plannedSingle := func() *maintenance.Maintenance {
		return single(offset(13, 39, time.UTC), offset(15, 9, time.UTC))
	}

	plannedOf := func(strategy string) func() *maintenance.Maintenance {
		return func() *maintenance.Maintenance {
			return &maintenance.Maintenance{Strategy: strategy, DateRange: make([]maintenance.DateTime, 2)}
		}
	}

	storedOf := func(strategy string, dateRange ...maintenance.DateTime) *maintenance.Maintenance {
		return &maintenance.Maintenance{Strategy: strategy, DateRange: dateRange}
	}

	tests := map[string]struct {
		planned func() *maintenance.Maintenance
		stored  *maintenance.Maintenance
		want    []string
	}{
		// The stored digits differ from the planned ones throughout, so that
		// sending the stored date instead of the planned one fails.
		"single, stored floating: sent floating with the planned digits": {
			planned: plannedSingle,
			stored:  single(floating(10, 0), floating(11, 0)),
			want:    []string{"floating 2030-06-29T13:39", "floating 2030-06-29T15:09"},
		},
		// A floating date reads back with Z, which Terraform would reject as
		// not being the planned value.
		"single, stored floating, planned with a non-zero offset: sent as planned": {
			planned: func() *maintenance.Maintenance {
				return single(offset(15, 39, cest()), offset(17, 9, cest()))
			},
			stored: single(floating(10, 0), floating(11, 0)),
			want:   []string{"offset 2030-06-29T15:39:00+02:00", "offset 2030-06-29T17:09:00+02:00"},
		},
		"single, stored with an offset: sent as planned": {
			planned: plannedSingle,
			stored:  single(offset(10, 0, cest()), offset(11, 0, cest())),
			want:    []string{"offset 2030-06-29T13:39:00Z", "offset 2030-06-29T15:09:00Z"},
		},
		"single, stored absent: sent as planned": {
			planned: plannedSingle,
			stored:  single(maintenance.DateTime{}, maintenance.DateTime{}),
			want:    []string{"offset 2030-06-29T13:39:00Z", "offset 2030-06-29T15:09:00Z"},
		},
		"single, stored one floating element: only the start floating": {
			planned: plannedSingle,
			stored:  single(floating(10, 0)),
			want:    []string{"floating 2030-06-29T13:39", "offset 2030-06-29T15:09:00Z"},
		},
		"single, stored floating start and offset end: each on its own": {
			planned: plannedSingle,
			stored:  single(floating(10, 0), offset(11, 0, time.UTC)),
			want:    []string{"floating 2030-06-29T13:39", "offset 2030-06-29T15:09:00Z"},
		},
		"single, changed from cron with a floating range: sent floating": {
			planned: plannedSingle,
			stored:  storedOf("cron", floating(10, 0), floating(11, 0)),
			want:    []string{"floating 2030-06-29T13:39", "floating 2030-06-29T15:09"},
		},
		// Validation requires both dates for single, so this only pins that an
		// absent planned date is never turned into a floating year-1 date.
		"single, planned absent, stored floating: stays absent": {
			planned: func() *maintenance.Maintenance {
				return single(maintenance.DateTime{}, offset(15, 9, time.UTC))
			},
			stored: single(floating(10, 0), floating(11, 0)),
			want:   []string{"absent", "floating 2030-06-29T15:09"},
		},
		"single, nothing stored: sent as planned": {
			planned: plannedSingle,
			stored:  nil,
			want:    []string{"offset 2030-06-29T13:39:00Z", "offset 2030-06-29T15:09:00Z"},
		},
		"cron, same strategy, stored floating range: carried through": {
			planned: plannedOf("cron"),
			stored:  storedOf("cron", floating(0, 0), floating(23, 59)),
			want:    []string{"floating 2030-06-29T00:00", "floating 2030-06-29T23:59"},
		},
		"recurring-weekday, same strategy, stored range with an offset: carried through": {
			planned: plannedOf("recurring-weekday"),
			stored:  storedOf("recurring-weekday", offset(0, 0, cest()), offset(23, 59, cest())),
			want:    []string{"offset 2030-06-29T00:00:00+02:00", "offset 2030-06-29T23:59:00+02:00"},
		},
		"recurring-interval, same strategy, stored one element: normalised to two": {
			planned: plannedOf("recurring-interval"),
			stored:  storedOf("recurring-interval", floating(0, 0)),
			want:    []string{"floating 2030-06-29T00:00", "absent"},
		},
		"recurring-day-of-month, same strategy, stored absent: two absent": {
			planned: plannedOf("recurring-day-of-month"),
			stored:  storedOf("recurring-day-of-month", maintenance.DateTime{}, maintenance.DateTime{}),
			want:    []string{"absent", "absent"},
		},
		"cron, same strategy, stored empty: two absent": {
			planned: plannedOf("cron"),
			stored:  storedOf("cron"),
			want:    []string{"absent", "absent"},
		},
		"cron, same strategy, stored absent: carried through": {
			planned: plannedOf("cron"),
			stored:  storedOf("cron", maintenance.DateTime{}, floating(23, 59)),
			want:    []string{"absent", "floating 2030-06-29T23:59"},
		},
		// The effective date range belongs to the window, not to one schedule.
		"recurring-weekday, changed from cron: carried through": {
			planned: plannedOf("recurring-weekday"),
			stored:  storedOf("cron", floating(0, 0), floating(23, 59)),
			want:    []string{"floating 2030-06-29T00:00", "floating 2030-06-29T23:59"},
		},
		"cron, changed from recurring-interval: carried through": {
			planned: plannedOf("cron"),
			stored:  storedOf("recurring-interval", offset(0, 0, cest()), floating(23, 59)),
			want:    []string{"offset 2030-06-29T00:00:00+02:00", "floating 2030-06-29T23:59"},
		},
		"recurring-day-of-month, changed from manual: sent as built": {
			planned: plannedOf("recurring-day-of-month"),
			stored:  storedOf("manual", floating(0, 0), floating(23, 59)),
			want:    []string{"absent", "absent"},
		},
		// The range of a single window is the window itself, not an effective
		// date range.
		"cron, changed from single: sent as built": {
			planned: plannedOf("cron"),
			stored:  single(floating(13, 39), floating(15, 9)),
			want:    []string{"absent", "absent"},
		},
		"manual, same strategy with a stored range: sent as built": {
			planned: plannedOf("manual"),
			stored:  storedOf("manual", floating(0, 0), floating(23, 59)),
			want:    []string{"absent", "absent"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			planned := tc.planned()
			preserveStoredMaintenanceDates(planned, tc.stored)

			got := describeDateRange(planned.DateRange)
			if fmt.Sprint(got) != fmt.Sprint(tc.want) {
				t.Errorf("date range = %q, want %q", got, tc.want)
			}
		})
	}
}
