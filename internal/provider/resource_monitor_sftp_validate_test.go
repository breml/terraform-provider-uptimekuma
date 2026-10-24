package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestValidateSFTPCredentials covers the credential validation for every auth method, including
// the omitted ssh_auth_method, which is null during validation and means the schema default
// (password). A credential left unset in a row is null, the zero value of types.String.
func TestValidateSFTPCredentials(t *testing.T) {
	tests := []struct {
		name              string
		authMethod        types.String
		password          types.String
		privateKey        types.String
		passphrase        types.String
		passwordWO        types.String
		privateKeyWO      types.String
		passphraseWO      types.String
		expectedErrors    []string
		expectedSummaries []string
	}{
		{
			name:       "auth method omitted, no key material",
			authMethod: types.StringNull(),
			privateKey: types.StringNull(),
			passphrase: types.StringNull(),
		},
		{
			name:              "auth method omitted, passphrase set",
			authMethod:        types.StringNull(),
			privateKey:        types.StringNull(),
			passphrase:        types.StringValue("phrase"),
			expectedErrors:    []string{"ssh_passphrase"},
			expectedSummaries: []string{"Invalid Attribute Combination"},
		},
		{
			name:       "auth method unknown, passphrase set",
			authMethod: types.StringUnknown(),
			privateKey: types.StringNull(),
			passphrase: types.StringValue("phrase"),
		},
		{
			name:       "auth method unknown, no private key",
			authMethod: types.StringUnknown(),
			privateKey: types.StringNull(),
			passphrase: types.StringNull(),
		},
		{
			name:       "password, no key material",
			authMethod: types.StringValue("password"),
			privateKey: types.StringNull(),
			passphrase: types.StringNull(),
		},
		{
			name:              "password, passphrase set",
			authMethod:        types.StringValue("password"),
			privateKey:        types.StringNull(),
			passphrase:        types.StringValue("phrase"),
			expectedErrors:    []string{"ssh_passphrase"},
			expectedSummaries: []string{"Invalid Attribute Combination"},
		},
		{
			name:       "password, passphrase only known after apply",
			authMethod: types.StringValue("password"),
			privateKey: types.StringNull(),
			passphrase: types.StringUnknown(),
		},
		{
			name:       "private key, key set",
			authMethod: types.StringValue("privateKey"),
			privateKey: types.StringValue("key"),
			passphrase: types.StringNull(),
		},
		{
			name:       "private key, key and passphrase set",
			authMethod: types.StringValue("privateKey"),
			privateKey: types.StringValue("key"),
			passphrase: types.StringValue("phrase"),
		},
		{
			name:       "private key, key only known after apply",
			authMethod: types.StringValue("privateKey"),
			privateKey: types.StringUnknown(),
			passphrase: types.StringNull(),
		},
		{
			name:              "private key, key missing",
			authMethod:        types.StringValue("privateKey"),
			privateKey:        types.StringNull(),
			passphrase:        types.StringNull(),
			expectedErrors:    []string{"ssh_private_key"},
			expectedSummaries: []string{"Missing Attribute Configuration"},
		},
		{
			name:       "auth method omitted, password set",
			authMethod: types.StringNull(),
			password:   types.StringValue("s3cret"),
			privateKey: types.StringNull(),
			passphrase: types.StringNull(),
		},
		{
			name:              "auth method omitted, private key set",
			authMethod:        types.StringNull(),
			privateKey:        types.StringValue("key"),
			passphrase:        types.StringNull(),
			expectedErrors:    []string{"ssh_private_key"},
			expectedSummaries: []string{"Invalid Attribute Combination"},
		},
		{
			name:              "password, private key and passphrase set",
			authMethod:        types.StringValue("password"),
			password:          types.StringValue("s3cret"),
			privateKey:        types.StringValue("key"),
			passphrase:        types.StringValue("phrase"),
			expectedErrors:    []string{"ssh_private_key", "ssh_passphrase"},
			expectedSummaries: []string{"Invalid Attribute Combination", "Invalid Attribute Combination"},
		},
		{
			name:       "password, private key only known after apply",
			authMethod: types.StringValue("password"),
			privateKey: types.StringUnknown(),
			passphrase: types.StringNull(),
		},
		{
			name:              "private key, password set",
			authMethod:        types.StringValue("privateKey"),
			password:          types.StringValue("s3cret"),
			privateKey:        types.StringValue("key"),
			passphrase:        types.StringNull(),
			expectedErrors:    []string{"ssh_password"},
			expectedSummaries: []string{"Invalid Attribute Combination"},
		},
		{
			name:       "private key, password only known after apply",
			authMethod: types.StringValue("privateKey"),
			password:   types.StringUnknown(),
			privateKey: types.StringValue("key"),
			passphrase: types.StringNull(),
		},
		{
			name:              "private key, key missing and password set",
			authMethod:        types.StringValue("privateKey"),
			password:          types.StringValue("s3cret"),
			privateKey:        types.StringNull(),
			passphrase:        types.StringNull(),
			expectedErrors:    []string{"ssh_private_key", "ssh_password"},
			expectedSummaries: []string{"Missing Attribute Configuration", "Invalid Attribute Combination"},
		},
		{
			name:         "private key, write-only key set",
			authMethod:   types.StringValue("privateKey"),
			privateKeyWO: types.StringValue("key"),
		},
		{
			name:              "private key, write-only password set",
			authMethod:        types.StringValue("privateKey"),
			privateKeyWO:      types.StringValue("key"),
			passwordWO:        types.StringValue("s3cret"),
			expectedErrors:    []string{"ssh_password_wo"},
			expectedSummaries: []string{"Invalid Attribute Combination"},
		},
		{
			name:              "password, write-only key material set",
			authMethod:        types.StringValue("password"),
			passwordWO:        types.StringValue("s3cret"),
			privateKeyWO:      types.StringValue("key"),
			passphraseWO:      types.StringValue("phrase"),
			expectedErrors:    []string{"ssh_private_key_wo", "ssh_passphrase_wo"},
			expectedSummaries: []string{"Invalid Attribute Combination", "Invalid Attribute Combination"},
		},
		{
			name:       "unexpected auth method is left to the schema validator",
			authMethod: types.StringValue("publickey"),
			privateKey: types.StringNull(),
			passphrase: types.StringValue("phrase"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MonitorSFTPResourceModel{
				SSHAuthMethod: tt.authMethod,
				SSHPassword:   tt.password,
				SSHPrivateKey: tt.privateKey,
				SSHPassphrase: tt.passphrase,

				SSHPasswordWO:   tt.passwordWO,
				SSHPrivateKeyWO: tt.privateKeyWO,
				SSHPassphraseWO: tt.passphraseWO,
			}

			resp := &resource.ValidateConfigResponse{}

			validateSFTPCredentials(&config, resp)

			gotPaths := make([]string, 0, len(resp.Diagnostics))
			gotSummaries := make([]string, 0, len(resp.Diagnostics))

			for _, d := range resp.Diagnostics {
				withPath, ok := d.(diag.DiagnosticWithPath)
				if !ok {
					t.Fatalf("diagnostic %q has no attribute path", d.Summary())
				}

				gotPaths = append(gotPaths, withPath.Path().String())
				gotSummaries = append(gotSummaries, d.Summary())
			}

			if len(gotPaths) != len(tt.expectedErrors) {
				t.Fatalf("got errors for %v, want errors for %v", gotPaths, tt.expectedErrors)
			}

			for i, want := range tt.expectedErrors {
				if gotPaths[i] != want {
					t.Errorf("error %d is for %q, want %q", i, gotPaths[i], want)
				}
			}

			if len(gotSummaries) != len(tt.expectedSummaries) {
				t.Fatalf("got summaries %v, want %v", gotSummaries, tt.expectedSummaries)
			}

			for i, want := range tt.expectedSummaries {
				if gotSummaries[i] != want {
					t.Errorf("error %d has summary %q, want %q", i, gotSummaries[i], want)
				}
			}
		})
	}
}
