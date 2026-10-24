package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestWithWriteOnlySecretRequired(t *testing.T) {
	attrs := withWriteOnlySecret(map[string]schema.Attribute{
		"token": schema.StringAttribute{
			MarkdownDescription: "API token.",
			Required:            true,
			Sensitive:           true,
		},
	}, "token", "API token")

	stored := stringAttribute(t, attrs, "token")
	if stored.Required || !stored.Optional {
		t.Errorf("token: Required=%t Optional=%t, want optional", stored.Required, stored.Optional)
	}

	if !stored.Sensitive {
		t.Error("token lost Sensitive")
	}

	assertValidatorDescription(t, "token", stored, "token_wo")

	if !strings.Contains(stored.MarkdownDescription, "`token_wo`") {
		t.Errorf("token description %q does not point to token_wo", stored.MarkdownDescription)
	}

	assertWriteOnlyPair(t, attrs, "token")
}

func TestWithWriteOnlySecretOptional(t *testing.T) {
	attrs := withWriteOnlySecret(map[string]schema.Attribute{
		"password": schema.StringAttribute{
			MarkdownDescription: "Password.",
			Optional:            true,
			Computed:            true,
			Sensitive:           true,
			Default:             stringdefault.StaticString(""),
		},
	}, "password", "Password")

	stored := stringAttribute(t, attrs, "password")
	if stored.Required || !stored.Optional || !stored.Computed || stored.Default == nil {
		t.Errorf("password: Required=%t Optional=%t Computed=%t Default=%v, want unchanged",
			stored.Required, stored.Optional, stored.Computed, stored.Default)
	}

	if len(stored.PlanModifiers) != 0 {
		t.Errorf("password has a default, but got plan modifiers %v", stored.PlanModifiers)
	}

	assertValidatorDescription(t, "password", stored, "password_wo")
	assertWriteOnlyPair(t, attrs, "password")
}

func TestWithWriteOnlySecretComputedWithoutDefault(t *testing.T) {
	attrs := withWriteOnlySecret(map[string]schema.Attribute{
		"api_key": schema.StringAttribute{
			MarkdownDescription: "API key.",
			Optional:            true,
			Computed:            true,
			Sensitive:           true,
		},
	}, "api_key", "API key")

	stored := stringAttribute(t, attrs, "api_key")

	var found bool

	for _, m := range stored.PlanModifiers {
		if _, ok := m.(writeOnlyComputedPlan); ok {
			found = true
		}
	}

	if !found {
		t.Error("api_key lacks the writeOnlyComputedPlan plan modifier")
	}

	assertWriteOnlyPair(t, attrs, "api_key")
}

func TestWithWriteOnlySecretCopiesValidators(t *testing.T) {
	attrs := withWriteOnlySecret(map[string]schema.Attribute{
		"token": schema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(8)},
		},
	}, "token", "API token")

	assertValidatorDescription(t, "token_wo", stringAttribute(t, attrs, "token_wo"), "at least 8")
}

func TestWithWriteOnlySecretUnknownAttribute(t *testing.T) {
	for name, attrs := range map[string]map[string]schema.Attribute{
		"missing":    {},
		"not string": {"token": schema.BoolAttribute{Optional: true}},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("withWriteOnlySecret did not panic")
				}
			}()

			withWriteOnlySecret(attrs, "token", "API token")
		})
	}
}

// TestResourceSchemasValidate builds the schema of every resource, which catches a
// withWriteOnlySecret call on an unknown attribute, and validates it the way the framework
// does at runtime, which catches a write-only attribute that is also computed or required.
func TestResourceSchemasValidate(t *testing.T) {
	ctx := t.Context()

	for _, newResource := range New("test")().Resources(ctx) {
		r := newResource()

		var metaResp resource.MetadataResponse

		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "uptimekuma"}, &metaResp)

		t.Run(metaResp.TypeName, func(t *testing.T) {
			var resp resource.SchemaResponse

			r.Schema(ctx, resource.SchemaRequest{}, &resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
			}

			diags := resp.Schema.ValidateImplementation(ctx)
			if diags.HasError() {
				t.Errorf("schema implementation: %v", diags)
			}
		})
	}
}

func assertValidatorDescription(t *testing.T, name string, attr schema.StringAttribute, want string) {
	t.Helper()

	for _, v := range attr.Validators {
		if strings.Contains(v.Description(t.Context()), want) {
			return
		}
	}

	t.Errorf("%s has no validator referring to %s", name, want)
}

func assertWriteOnlyPair(t *testing.T, attrs map[string]schema.Attribute, name string) {
	t.Helper()

	wo, ok := attrs[name+writeOnlySuffix].(schema.StringAttribute)
	if !ok {
		t.Fatalf("%s_wo missing or not a string attribute", name)
	}

	if !wo.WriteOnly || !wo.Sensitive || !wo.Optional || wo.Computed || wo.Required {
		t.Errorf("%s_wo: WriteOnly=%t Sensitive=%t Optional=%t Computed=%t Required=%t",
			name, wo.WriteOnly, wo.Sensitive, wo.Optional, wo.Computed, wo.Required)
	}

	assertValidatorDescription(t, name+writeOnlySuffix, wo, name+writeOnlyVersionSuffix)

	version, ok := attrs[name+writeOnlyVersionSuffix].(schema.StringAttribute)
	if !ok {
		t.Fatalf("%s_wo_version missing or not a string attribute", name)
	}

	if version.WriteOnly || version.Sensitive || !version.Optional {
		t.Errorf("%s_wo_version: WriteOnly=%t Sensitive=%t Optional=%t",
			name, version.WriteOnly, version.Sensitive, version.Optional)
	}

	assertValidatorDescription(t, name+writeOnlyVersionSuffix, version, name+writeOnlySuffix)
}

func stringAttribute(t *testing.T, attrs map[string]schema.Attribute, name string) schema.StringAttribute {
	t.Helper()

	attr, ok := attrs[name].(schema.StringAttribute)
	if !ok {
		t.Fatalf("%s missing or not a string attribute", name)
	}

	return attr
}

func TestWriteOnlyComputedPlan(t *testing.T) {
	s := schema.Schema{Attributes: map[string]schema.Attribute{
		"api_key":            schema.StringAttribute{Optional: true, Computed: true},
		"api_key_wo_version": schema.StringAttribute{Optional: true},
	}}
	objectType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"api_key":            tftypes.String,
		"api_key_wo_version": tftypes.String,
	}}
	object := func(apiKey, version any) tftypes.Value {
		return tftypes.NewValue(objectType, map[string]tftypes.Value{
			"api_key":            tftypes.NewValue(tftypes.String, apiKey),
			"api_key_wo_version": tftypes.NewValue(tftypes.String, version),
		})
	}

	tests := map[string]struct {
		config tftypes.Value
		state  tftypes.Value
		want   types.String
	}{
		"stored value configured": {
			config: object("key", nil),
			state:  object("key", nil),
			want:   types.StringValue("prior"),
		},
		"neither configured": {
			config: object(nil, nil),
			state:  object("key", nil),
			want:   types.StringValue("prior"),
		},
		"create without write-only": {
			config: object(nil, nil),
			state:  tftypes.NewValue(objectType, nil),
			want:   types.StringValue("prior"),
		},
		"write-only in use": {
			config: object(nil, "1"),
			state:  object("key", nil),
			want:   types.StringNull(),
		},
		"write-only removed": {
			config: object(nil, nil),
			state:  object(nil, "1"),
			want:   types.StringUnknown(),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var configValue types.String

			config := tfsdk.Config{Schema: s, Raw: tc.config}
			if diags := config.GetAttribute(t.Context(), path.Root("api_key"), &configValue); diags.HasError() {
				t.Fatalf("read config: %v", diags)
			}

			req := planmodifier.StringRequest{
				Config:      config,
				ConfigValue: configValue,
				State:       tfsdk.State{Schema: s, Raw: tc.state},
				PlanValue:   types.StringValue("prior"),
			}
			resp := planmodifier.StringResponse{PlanValue: req.PlanValue}

			writeOnlyComputedPlan{versionName: "api_key_wo_version"}.PlanModifyString(t.Context(), req, &resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}

			if !resp.PlanValue.Equal(tc.want) {
				t.Errorf("plan value = %s, want %s", resp.PlanValue, tc.want)
			}
		})
	}
}
