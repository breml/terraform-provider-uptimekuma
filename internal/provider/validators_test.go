package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWholeNumberValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     types.Float64
		wantError bool
	}{
		"whole":             {value: types.Float64Value(48), wantError: false},
		"whole negative":    {value: types.Float64Value(-3), wantError: false},
		"zero":              {value: types.Float64Value(0), wantError: false},
		"fractional":        {value: types.Float64Value(2.5), wantError: true},
		"barely fractional": {value: types.Float64Value(2.0000001), wantError: true},
		"null":              {value: types.Float64Null(), wantError: false},
		"unknown":           {value: types.Float64Unknown(), wantError: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.Float64Response{}
			wholeNumber().ValidateFloat64(
				t.Context(),
				validator.Float64Request{ConfigValue: test.value},
				resp,
			)

			if got := resp.Diagnostics.HasError(); got != test.wantError {
				t.Errorf("HasError() = %v, want %v (%v)", got, test.wantError, resp.Diagnostics)
			}
		})
	}
}

func TestFloat64ToPtr(t *testing.T) {
	t.Parallel()

	if got := float64ToPtr(types.Float64Null()); got != nil {
		t.Errorf("float64ToPtr(null) = %v, want nil", *got)
	}

	if got := float64ToPtr(types.Float64Unknown()); got != nil {
		t.Errorf("float64ToPtr(unknown) = %v, want nil", *got)
	}

	got := float64ToPtr(types.Float64Value(2.5))
	if got == nil {
		t.Fatal("float64ToPtr(2.5) = nil, want pointer to 2.5")
	}

	if *got != 2.5 {
		t.Errorf("float64ToPtr(2.5) = %v, want 2.5", *got)
	}
}

// systemServiceNameValidators returns the validators the System Service
// monitor resource declares for system_service_name.
func systemServiceNameValidators(t *testing.T) []validator.String {
	t.Helper()

	resp := &resource.SchemaResponse{}
	(&MonitorSystemServiceResource{}).Schema(t.Context(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["system_service_name"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("system_service_name is %T, want schema.StringAttribute", resp.Schema.Attributes["system_service_name"])
	}

	return attr.Validators
}

func TestMonitorSystemServiceNameValidation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     string
		wantError bool
	}{
		"systemd unit":      {value: "nginx.service", wantError: false},
		"templated unit":    {value: "sshd@0.service", wantError: false},
		"hyphen":            {value: "user-session.service", wantError: false},
		"underscore":        {value: "my_service", wantError: false},
		"windows scm name":  {value: "Spooler", wantError: false},
		"digits":            {value: "svc42", wantError: false},
		"empty":             {value: "", wantError: true},
		"space":             {value: "my service", wantError: true},
		"leading space":     {value: " nginx.service", wantError: true},
		"slash":             {value: "nginx/1", wantError: true},
		"trailing newline":  {value: "nginx.service\n", wantError: true},
		"shell metacharact": {value: "nginx;reboot", wantError: true},
	}

	validators := systemServiceNameValidators(t)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.StringResponse{}
			for _, v := range validators {
				v.ValidateString(
					t.Context(),
					validator.StringRequest{ConfigValue: types.StringValue(test.value)},
					resp,
				)
			}

			if got := resp.Diagnostics.HasError(); got != test.wantError {
				t.Errorf("HasError() = %v, want %v (%v)", got, test.wantError, resp.Diagnostics)
			}
		})
	}
}

func TestNonBlankValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     types.String
		wantError bool
	}{
		"value":             {value: types.StringValue("api-server"), wantError: false},
		"inner space":       {value: types.StringValue("worker 1"), wantError: false},
		"surrounded by pad": {value: types.StringValue("  api-server  "), wantError: false},
		"null":              {value: types.StringNull(), wantError: false},
		"unknown":           {value: types.StringUnknown(), wantError: false},
		"empty":             {value: types.StringValue(""), wantError: true},
		"spaces only":       {value: types.StringValue("   "), wantError: true},
		"tab only":          {value: types.StringValue("\t"), wantError: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.StringResponse{}
			nonBlank().ValidateString(
				t.Context(),
				validator.StringRequest{ConfigValue: test.value},
				resp,
			)

			if got := resp.Diagnostics.HasError(); got != test.wantError {
				t.Errorf("HasError() = %v, want %v (%v)", got, test.wantError, resp.Diagnostics)
			}
		})
	}
}

// pm2ProcessNameValidators returns the validators the PM2 monitor resource
// declares for process_name.
func pm2ProcessNameValidators(t *testing.T) []validator.String {
	t.Helper()

	resp := &resource.SchemaResponse{}
	(&MonitorPM2Resource{}).Schema(t.Context(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["process_name"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("process_name is %T, want schema.StringAttribute", resp.Schema.Attributes["process_name"])
	}

	return attr.Validators
}

func TestMonitorPM2ProcessNameValidation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     string
		wantError bool
	}{
		"process name":     {value: "api-server", wantError: false},
		"numeric pm2 id":   {value: "0", wantError: false},
		"inner space":      {value: "worker 1", wantError: false},
		"padded":           {value: "  api-server  ", wantError: false},
		"unicode":          {value: "wörker", wantError: false},
		"slash":            {value: "apps/api", wantError: false},
		"empty":            {value: "", wantError: true},
		"whitespace only":  {value: "   ", wantError: true},
		"bell":             {value: "api\aserver", wantError: true},
		"newline":          {value: "api\nserver", wantError: true},
		"trailing newline": {value: "api-server\n", wantError: true},
		"delete":           {value: "api\x7fserver", wantError: true},
	}

	validators := pm2ProcessNameValidators(t)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.StringResponse{}
			for _, v := range validators {
				v.ValidateString(
					t.Context(),
					validator.StringRequest{ConfigValue: types.StringValue(test.value)},
					resp,
				)
			}

			if got := resp.Diagnostics.HasError(); got != test.wantError {
				t.Errorf("HasError() = %v, want %v (%v)", got, test.wantError, resp.Diagnostics)
			}
		})
	}
}

func TestJSONObjectValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     types.String
		wantError bool
	}{
		"object":            {value: types.StringValue(`{"X-Custom-Header": "Additional Header"}`), wantError: false},
		"empty object":      {value: types.StringValue(`{}`), wantError: false},
		"nested object":     {value: types.StringValue(`{"a": {"b": 1}}`), wantError: false},
		"surrounded by ws":  {value: types.StringValue("  {\"a\": \"b\"}\n"), wantError: false},
		"null":              {value: types.StringNull(), wantError: false},
		"unknown":           {value: types.StringUnknown(), wantError: false},
		"array":             {value: types.StringValue(`["a", "b"]`), wantError: true},
		"string":            {value: types.StringValue(`"a"`), wantError: true},
		"number":            {value: types.StringValue(`1`), wantError: true},
		"json null":         {value: types.StringValue(`null`), wantError: true},
		"empty":             {value: types.StringValue(``), wantError: true},
		"trailing comma":    {value: types.StringValue(`{"a": "b",}`), wantError: true},
		"unquoted key":      {value: types.StringValue(`{a: "b"}`), wantError: true},
		"two objects":       {value: types.StringValue(`{"a": "b"}{"c": "d"}`), wantError: true},
		"header list":       {value: types.StringValue("X-Custom-Header: Additional Header"), wantError: true},
		"non-string values": {value: types.StringValue(`{"a": 1, "b": true}`), wantError: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.StringResponse{}
			jsonObject().ValidateString(
				t.Context(),
				validator.StringRequest{ConfigValue: test.value},
				resp,
			)

			if got := resp.Diagnostics.HasError(); got != test.wantError {
				t.Errorf("HasError() = %v, want %v (%v)", got, test.wantError, resp.Diagnostics)
			}
		})
	}
}

// kafkaProducerTimeoutValidators returns the validators the Kafka Producer
// monitor resource declares for timeout.
func kafkaProducerTimeoutValidators(t *testing.T) []validator.Float64 {
	t.Helper()

	resp := &resource.SchemaResponse{}
	(&MonitorKafkaProducerResource{}).Schema(t.Context(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["timeout"].(schema.Float64Attribute)
	if !ok {
		t.Fatalf("timeout is %T, want schema.Float64Attribute", resp.Schema.Attributes["timeout"])
	}

	return attr.Validators
}

// TestMonitorKafkaProducerTimeoutValidation pins the plan-time floor. Uptime Kuma
// reads a timeout of 0 or less as a request to fall back to 80% of the interval
// rather than as a timeout, so allowing it would silently apply the fallback
// instead of the configured value. There is no strict-greater-than float64
// validator, so the floor is the 0.1 step the web UI uses for non-ping monitors.
func TestMonitorKafkaProducerTimeoutValidation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     types.Float64
		wantError bool
	}{
		"floor":          {value: types.Float64Value(0.1), wantError: false},
		"whole second":   {value: types.Float64Value(1), wantError: false},
		"fractional":     {value: types.Float64Value(2.5), wantError: false},
		"above ui clamp": {value: types.Float64Value(3600), wantError: false},
		"below floor":    {value: types.Float64Value(0.09), wantError: true},
		"zero":           {value: types.Float64Value(0), wantError: true},
		"negative":       {value: types.Float64Value(-1), wantError: true},
		"null":           {value: types.Float64Null(), wantError: false},
		"unknown":        {value: types.Float64Unknown(), wantError: false},
	}

	validators := kafkaProducerTimeoutValidators(t)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.Float64Response{}
			for _, v := range validators {
				v.ValidateFloat64(
					t.Context(),
					validator.Float64Request{ConfigValue: test.value},
					resp,
				)
			}

			if got := resp.Diagnostics.HasError(); got != test.wantError {
				t.Errorf("HasError() = %v, want %v (%v)", got, test.wantError, resp.Diagnostics)
			}
		})
	}
}
