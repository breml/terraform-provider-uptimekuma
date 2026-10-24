package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/breml/go-uptime-kuma-client/notification"
)

// notificationWriteOnlyTest describes one notification resource for
// testAccNotificationWriteOnly.
type notificationWriteOnlyTest struct {
	// body is the resource block body without the name and the write-only lines.
	body string

	// attrs are the secret attributes with a write-only variant.
	attrs []string

	// serverSecrets reads the secrets back from Uptime Kuma.
	serverSecrets serverSecretsFunc
}

// testAccNotificationWriteOnly runs the shared write-only test for the notification resource
// resourceType.
func testAccNotificationWriteOnly(t *testing.T, resourceType string, tc notificationWriteOnlyTest) {
	t.Helper()

	name := acctest.RandomWithPrefix("WO" + resourceType)

	resource.ParallelTest(t, testAccWriteOnlyTestCase(t, writeOnlyTestCase{
		addr:  resourceType + ".test",
		attrs: tc.attrs,
		config: func(secret string, version string) string {
			return providerConfig() + fmt.Sprintf(`
resource %q "test" {
  name = %q
%s
%s}
`, resourceType, name, tc.body, writeOnlyAttrLines(secret, version, tc.attrs...))
		},
		serverSecrets: tc.serverSecrets,
	}))
}

// singleSecret returns the secrets map for a notification with one secret attribute.
func singleSecret(attr string, value string) map[string]string {
	return map[string]string{attr: value}
}

// TestAccNotificationResourcesWriteOnly runs the shared write-only test for every notification with
// write-only secrets; each such notification adds an entry here. The full set of transitions and
// the validators are covered once, in TestAccNotificationSMTPResourcePasswordWO.
func TestAccNotificationResourcesWriteOnly(t *testing.T) {
	tests := map[string]notificationWriteOnlyTest{
		"uptimekuma_notification_46elks": {
			body: `
  username    = "elks-user"
  from_number = "UptimeKuma"
  to_number   = "+46700000000"`,
			attrs: []string{"auth_token"},
			serverSecrets: notificationSecrets(func(n *notification.FortySixElks) map[string]string {
				return singleSecret("auth_token", n.AuthToken)
			}),
		},
		"uptimekuma_notification_alerta": {
			body: `
  api_endpoint = "https://alerta.example.com/api"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Alerta) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_alertnow": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.AlertNow) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_aliyunsms": {
			body: `
  access_key_id = "LTAI0000"
  phone_number  = "13800000000"
  sign_name     = "UptimeKuma"
  template_code = "SMS_000000"`,
			attrs: []string{"secret_access_key"},
			serverSecrets: notificationSecrets(func(n *notification.AliyunSMS) map[string]string {
				return singleSecret("secret_access_key", n.SecretAccessKey)
			}),
		},
		"uptimekuma_notification_apprise": {
			attrs: []string{"apprise_url"},
			serverSecrets: notificationSecrets(func(n *notification.Apprise) map[string]string {
				return singleSecret("apprise_url", n.AppriseURL)
			}),
		},
		"uptimekuma_notification_bale": {
			body: `
  chat_id = "123456"`,
			attrs: []string{"bot_token"},
			serverSecrets: notificationSecrets(func(n *notification.Bale) map[string]string {
				return singleSecret("bot_token", n.BotToken)
			}),
		},
		"uptimekuma_notification_bark": {
			attrs: []string{"endpoint"},
			serverSecrets: notificationSecrets(func(n *notification.Bark) map[string]string {
				return singleSecret("endpoint", n.Endpoint)
			}),
		},
		"uptimekuma_notification_bitrix24": {
			body: `
  notification_user_id = "1"`,
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Bitrix24) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_brevo": {
			body: `
  to_email   = "ops@example.com"
  from_email = "uptime@example.com"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Brevo) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_callmebot": {
			attrs: []string{"endpoint"},
			serverSecrets: notificationSecrets(func(n *notification.CallMeBot) map[string]string {
				return singleSecret("endpoint", n.Endpoint)
			}),
		},
		"uptimekuma_notification_cellsynt": {
			body: `
  login       = "cellsynt-user"
  destination = "0046700000000"
  originator  = "UptimeKuma"`,
			attrs: []string{"password"},
			serverSecrets: notificationSecrets(func(n *notification.Cellsynt) map[string]string {
				return singleSecret("password", n.Password)
			}),
		},
		"uptimekuma_notification_clicksendsms": {
			body: `
  login       = "clicksend-user"
  to_number   = "+61400000000"
  sender_name = "UptimeKuma"`,
			attrs: []string{"password"},
			serverSecrets: notificationSecrets(func(n *notification.ClickSendSMS) map[string]string {
				return singleSecret("password", n.Password)
			}),
		},
		"uptimekuma_notification_clickup": {
			body: `
  workspace_id = "9012345678"
  channel_id   = "4-90123456-8"`,
			attrs: []string{"token"},
			serverSecrets: notificationSecrets(func(n *notification.ClickUp) map[string]string {
				return singleSecret("token", n.Token)
			}),
		},
		"uptimekuma_notification_dingding": {
			attrs: []string{"webhook_url", "secret_key"},
			serverSecrets: notificationSecrets(func(n *notification.DingDing) map[string]string {
				return map[string]string{"webhook_url": n.WebHookURL, "secret_key": n.SecretKey}
			}),
		},
		"uptimekuma_notification_discord": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Discord) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_evolution": {
			body: `
  api_url       = "https://evolution.example.com"
  instance_name = "uptime"
  recipient     = "5511999999999"`,
			attrs: []string{"auth_token"},
			serverSecrets: notificationSecrets(func(n *notification.Evolution) map[string]string {
				return singleSecret("auth_token", n.AuthToken)
			}),
		},
		"uptimekuma_notification_feishu": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Feishu) map[string]string {
				return singleSecret("webhook_url", n.WebHookURL)
			}),
		},
		"uptimekuma_notification_flashduty": {
			attrs: []string{"integration_key"},
			serverSecrets: notificationSecrets(func(n *notification.FlashDuty) map[string]string {
				return singleSecret("integration_key", n.IntegrationKey)
			}),
		},
		"uptimekuma_notification_flowtriq": {
			attrs: []string{"webhook_url", "api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Flowtriq) map[string]string {
				return map[string]string{"webhook_url": n.WebhookURL, "api_key": stringPtrValue(n.APIKey)}
			}),
		},
		"uptimekuma_notification_fluxer": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Fluxer) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_freemobile": {
			body: `
  user = "12345678"`,
			attrs: []string{"pass"},
			serverSecrets: notificationSecrets(func(n *notification.FreeMobile) map[string]string {
				return singleSecret("pass", n.Pass)
			}),
		},
		"uptimekuma_notification_goalert": {
			body: `
  base_url = "https://goalert.example.com"`,
			attrs: []string{"token"},
			serverSecrets: notificationSecrets(func(n *notification.GoAlert) map[string]string {
				return singleSecret("token", n.Token)
			}),
		},
		"uptimekuma_notification_googlechat": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.GoogleChat) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_googlesheets": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.GoogleSheets) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_gotify": {
			body: `
  server_url = "https://gotify.example.com"`,
			attrs: []string{"application_token"},
			serverSecrets: notificationSecrets(func(n *notification.Gotify) map[string]string {
				return singleSecret("application_token", n.ApplicationToken)
			}),
		},
		"uptimekuma_notification_grafanaoncall": {
			attrs: []string{"grafana_oncall_url"},
			serverSecrets: notificationSecrets(func(n *notification.GrafanaOncall) map[string]string {
				return singleSecret("grafana_oncall_url", n.GrafanaOncallURL)
			}),
		},
		"uptimekuma_notification_gtxmessaging": {
			body: `
  from = "UptimeKuma"
  to   = "+41790000000"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.GTXMessaging) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
	}

	for resourceType, tc := range tests {
		t.Run(resourceType, func(t *testing.T) {
			testAccNotificationWriteOnly(t, resourceType, tc)
		})
	}
}
