package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

// httpWriteOnlyAttrs returns the secrets every monitor built on MonitorHTTPBaseModel offers a
// write-only variant for.
func httpWriteOnlyAttrs() []string {
	return []string{"basic_auth_pass", "tls_key", "oauth_client_secret"}
}

func httpDetailsSecrets(d *monitor.HTTPDetails) map[string]string {
	return map[string]string{
		"basic_auth_pass":     d.BasicAuthPass,
		"tls_key":             d.TLSKey,
		"oauth_client_secret": d.OAuthClientSecret,
	}
}

// testAccMonitorWriteOnly runs the shared write-only test for a monitor whose resource block
// body, without the write-only lines, is body; %[1]q in body is the monitor name.
func testAccMonitorWriteOnly(
	t *testing.T,
	resourceType string,
	body string,
	attrs []string,
	serverSecrets serverSecretsFunc,
) {
	t.Helper()

	name := acctest.RandomWithPrefix("WO" + resourceType)

	resource.ParallelTest(t, testAccWriteOnlyTestCase(t, writeOnlyTestCase{
		addr:  resourceType + ".test",
		attrs: attrs,
		config: func(secret string, version string) string {
			return providerConfig() + fmt.Sprintf(`
resource %q "test" {
%s
%s}
`, resourceType, fmt.Sprintf(body, name), writeOnlyAttrLines(secret, version, attrs...))
		},
		serverSecrets: serverSecrets,
	}))
}

func TestAccMonitorHTTPResourceWriteOnly(t *testing.T) {
	testAccMonitorWriteOnly(t, "uptimekuma_monitor_http", `
  name        = %[1]q
  url         = "https://example.com"
  auth_method = "basic"`,
		httpWriteOnlyAttrs(),
		monitorSecrets(func(m *monitor.HTTP) map[string]string { return httpDetailsSecrets(&m.HTTPDetails) }),
	)
}

func TestAccMonitorHTTPKeywordResourceWriteOnly(t *testing.T) {
	testAccMonitorWriteOnly(t, "uptimekuma_monitor_http_keyword", `
  name        = %[1]q
  url         = "https://example.com"
  keyword     = "Example"
  auth_method = "basic"`,
		httpWriteOnlyAttrs(),
		monitorSecrets(func(m *monitor.HTTPKeyword) map[string]string { return httpDetailsSecrets(&m.HTTPDetails) }),
	)
}

func TestAccMonitorHTTPJSONQueryResourceWriteOnly(t *testing.T) {
	testAccMonitorWriteOnly(t, "uptimekuma_monitor_http_json_query", `
  name           = %[1]q
  url            = "https://example.com"
  json_path      = "$.status"
  expected_value = "ok"
  auth_method    = "basic"`,
		httpWriteOnlyAttrs(),
		monitorSecrets(func(m *monitor.HTTPJSONQuery) map[string]string { return httpDetailsSecrets(&m.HTTPDetails) }),
	)
}

func TestAccMonitorWebsocketUpgradeResourceWriteOnly(t *testing.T) {
	testAccMonitorWriteOnly(t, "uptimekuma_monitor_websocket_upgrade", `
  name        = %[1]q
  url         = "wss://example.com"
  auth_method = "basic"`,
		httpWriteOnlyAttrs(),
		monitorSecrets(func(m *monitor.WebsocketUpgrade) map[string]string {
			return httpDetailsSecrets(&m.HTTPDetails)
		}),
	)
}

func TestAccMonitorGlobalpingResourceWriteOnly(t *testing.T) {
	testAccMonitorWriteOnly(t, "uptimekuma_monitor_globalping", `
  name        = %[1]q
  subtype     = "http"
  url         = "https://example.com"
  auth_method = "basic"`,
		httpWriteOnlyAttrs(),
		monitorSecrets(func(m *monitor.Globalping) map[string]string { return httpDetailsSecrets(&m.HTTPDetails) }),
	)
}

// connectionStringSecrets reads database_connection_string for a database monitor.
func connectionStringSecrets(s string) map[string]string {
	return map[string]string{"database_connection_string": s}
}

func TestAccMonitorDatabaseResourcesWriteOnly(t *testing.T) {
	tests := map[string]serverSecretsFunc{
		"uptimekuma_monitor_mysql": monitorSecrets(func(m *monitor.MySQL) map[string]string {
			return connectionStringSecrets(m.DatabaseConnectionString)
		}),
		"uptimekuma_monitor_postgres": monitorSecrets(func(m *monitor.Postgres) map[string]string {
			return connectionStringSecrets(m.DatabaseConnectionString)
		}),
		"uptimekuma_monitor_mongodb": monitorSecrets(func(m *monitor.MongoDB) map[string]string {
			return connectionStringSecrets(m.DatabaseConnectionString)
		}),
		"uptimekuma_monitor_redis": monitorSecrets(func(m *monitor.Redis) map[string]string {
			return connectionStringSecrets(m.ConnectionString)
		}),
		"uptimekuma_monitor_sqlserver": monitorSecrets(func(m *monitor.SQLServer) map[string]string {
			return connectionStringSecrets(m.DatabaseConnectionString)
		}),
	}

	for resourceType, serverSecrets := range tests {
		t.Run(resourceType, func(t *testing.T) {
			testAccMonitorWriteOnly(t, resourceType, `
  name = %[1]q`,
				[]string{"database_connection_string"},
				serverSecrets,
			)
		})
	}
}

func TestAccMonitorOracleDBResourceWriteOnly(t *testing.T) {
	testAccMonitorWriteOnly(t, "uptimekuma_monitor_oracledb", `
  name            = %[1]q
  basic_auth_user = "monitor"`,
		[]string{"database_connection_string", "basic_auth_pass"},
		monitorSecrets(func(m *monitor.OracleDB) map[string]string {
			return map[string]string{
				"database_connection_string": m.DatabaseConnectionString,
				"basic_auth_pass":            m.Password,
			}
		}),
	)
}

func TestAccMonitorMQTTResourceWriteOnly(t *testing.T) {
	testAccMonitorWriteOnly(t, "uptimekuma_monitor_mqtt", `
  name            = %[1]q
  hostname        = "mqtt.example.com"
  port            = 1883
  mqtt_topic      = "status"
  mqtt_username   = "monitor"
  mqtt_check_type = "keyword"`,
		[]string{"mqtt_password"},
		monitorSecrets(func(m *monitor.MQTT) map[string]string {
			return map[string]string{"mqtt_password": stringPtrValue(m.MQTTPassword)}
		}),
	)
}

func TestAccMonitorRabbitMQResourceWriteOnly(t *testing.T) {
	testAccMonitorWriteOnly(t, "uptimekuma_monitor_rabbitmq", `
  name     = %[1]q
  nodes    = jsonencode(["http://rabbitmq.example.com:15672/"])
  username = "monitor"`,
		[]string{"password"},
		monitorSecrets(func(m *monitor.RabbitMQ) map[string]string {
			return map[string]string{"password": stringPtrValue(m.Password)}
		}),
	)
}

func TestAccMonitorRadiusResourceWriteOnly(t *testing.T) {
	testAccMonitorWriteOnly(t, "uptimekuma_monitor_radius", `
  name            = %[1]q
  hostname        = "radius.example.com"
  radius_username = "monitor"`,
		[]string{"radius_password", "radius_secret"},
		monitorSecrets(func(m *monitor.Radius) map[string]string {
			return map[string]string{"radius_password": m.Password, "radius_secret": m.Secret}
		}),
	)
}

func TestAccMonitorSFTPResourceWriteOnly(t *testing.T) {
	t.Run("private key", func(t *testing.T) {
		testAccMonitorWriteOnly(t, "uptimekuma_monitor_sftp", `
  name            = %[1]q
  hostname        = "sftp.example.com"
  ssh_username    = "backup"
  ssh_auth_method = "privateKey"`,
			[]string{"ssh_private_key", "ssh_passphrase"},
			monitorSecrets(func(m *monitor.SFTP) map[string]string {
				return map[string]string{
					"ssh_private_key": stringPtrValue(m.SSHPrivateKey),
					"ssh_passphrase":  stringPtrValue(m.SSHPassphrase),
				}
			}),
		)
	})

	t.Run("password", func(t *testing.T) {
		testAccMonitorWriteOnly(t, "uptimekuma_monitor_sftp", `
  name         = %[1]q
  hostname     = "sftp.example.com"
  ssh_username = "backup"`,
			[]string{"ssh_password"},
			monitorSecrets(func(m *monitor.SFTP) map[string]string {
				return map[string]string{"ssh_password": stringPtrValue(m.SSHPassword)}
			}),
		)
	})
}

func TestAccMonitorSNMPResourceWriteOnly(t *testing.T) {
	testAccMonitorWriteOnly(t, "uptimekuma_monitor_snmp", `
  name         = %[1]q
  hostname     = "192.0.2.1"
  snmp_version = "2c"
  snmp_oid     = ".1.3.6.1.2.1.1.5.0"`,
		[]string{"snmp_community"},
		monitorSecrets(func(m *monitor.SNMP) map[string]string {
			return map[string]string{"snmp_community": m.SNMPCommunity}
		}),
	)
}

// TestAccMonitorKafkaProducerResourceWriteOnly cannot use testAccMonitorWriteOnly, because
// sasl_options_wo takes JSON; the secret is the password inside it.
func TestAccMonitorKafkaProducerResourceWriteOnly(t *testing.T) {
	name := acctest.RandomWithPrefix("WOKafkaProducer")

	resource.ParallelTest(t, testAccWriteOnlyTestCase(t, writeOnlyTestCase{
		addr:  "uptimekuma_monitor_kafka_producer.test",
		attrs: []string{"sasl_options"},
		// Read never refreshes sasl_options, so an import leaves it unset.
		importOmits: []string{"sasl_options"},
		config: func(secret string, version string) string {
			return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_kafka_producer" "test" {
  name    = %[1]q
  brokers = ["kafka.example.com:9092"]
  topic   = "monitor-topic"
  message = "ping"

  sasl_options_wo         = jsonencode({ mechanism = "plain", username = "u", password = %[2]q })
  sasl_options_wo_version = %[3]q
}
`, name, writeOnlyTestSecret("sasl_options", secret), version)
		},
		serverSecrets: func(ctx context.Context, kumaClient *kuma.Client, id int64) (map[string]string, error) {
			return monitorSecrets(func(m *monitor.KafkaProducer) map[string]string {
				var password string

				if m.SASLOptions != nil {
					if p, ok := (*m.SASLOptions)["password"].(string); ok {
						password = p
					}
				}

				return map[string]string{"sasl_options": password}
			})(ctx, kumaClient, id)
		},
	}))
}
