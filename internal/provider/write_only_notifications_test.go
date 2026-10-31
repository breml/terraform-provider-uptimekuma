package provider

import (
	"fmt"
	"strings"
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

	// importOmits are the attributes of attrs whose Read never refreshes the secret, see
	// writeOnlyTestCase.importOmits.
	importOmits []string
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
		importOmits:   tc.importOmits,
	}))
}

// singleSecret returns the secrets map for a notification with one secret attribute.
func singleSecret(attr string, value string) map[string]string {
	return map[string]string{attr: value}
}

// TestAccNotificationResourcesWriteOnly runs the shared write-only test for every notification with
// write-only secrets; each such notification adds an entry here. A notification whose secrets
// cannot share one configuration adds one entry per variant, keyed "<resource type>/<variant>".
// The full set of transitions and the validators are covered once, in
// TestAccNotificationSMTPResourcePasswordWO.
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
		"uptimekuma_notification_bearsms": {
			body: `
  username     = "bearsms-user"
  phone_number = "41790000000"`,
			attrs: []string{"hash_key"},
			serverSecrets: notificationSecrets(func(n *notification.BearSMS) map[string]string {
				return singleSecret("hash_key", n.HashKey)
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
		"uptimekuma_notification_halopsa": {
			body: `
  username = "halo-user"`,
			attrs:       []string{"webhook_url", "password"},
			importOmits: []string{"password"},
			serverSecrets: notificationSecrets(func(n *notification.HaloPSA) map[string]string {
				return map[string]string{
					"webhook_url": n.WebhookURL,
					"password":    n.Password,
				}
			}),
		},
		"uptimekuma_notification_heiioncall": {
			body: `
  trigger_id = "trigger-123"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.HeiiOnCall) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_homeassistant": {
			body: `
  home_assistant_url   = "https://homeassistant.example.com"
  notification_service = "notify.mobile_app"`,
			attrs: []string{"long_lived_access_token"},
			serverSecrets: notificationSecrets(func(n *notification.HomeAssistant) map[string]string {
				return singleSecret("long_lived_access_token", n.LongLivedAccessToken)
			}),
		},
		"uptimekuma_notification_jiraservicemanagement": {
			body: `
  cloud_id = "jsm-cloud-id"
  email    = "ops@example.com"`,
			attrs: []string{"api_token"},
			serverSecrets: notificationSecrets(func(n *notification.JiraServiceManagement) map[string]string {
				return singleSecret("api_token", n.APIToken)
			}),
		},
		"uptimekuma_notification_keep": {
			attrs: []string{"webhook_url", "api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Keep) map[string]string {
				return map[string]string{
					"webhook_url": n.WebhookURL,
					"api_key":     n.APIKey,
				}
			}),
		},
		"uptimekuma_notification_kook": {
			body: `
  guild_id = "382941547624206336"`,
			attrs: []string{"bot_token"},
			serverSecrets: notificationSecrets(func(n *notification.Kook) map[string]string {
				return singleSecret("bot_token", n.BotToken)
			}),
		},
		"uptimekuma_notification_line": {
			body: `
  user_id = "U0000000000"`,
			attrs: []string{"channel_access_token"},
			serverSecrets: notificationSecrets(func(n *notification.Line) map[string]string {
				return singleSecret("channel_access_token", n.ChannelAccessToken)
			}),
		},
		"uptimekuma_notification_matrix": {
			body: `
  homeserver_url   = "https://matrix.example.com"
  internal_room_id = "!abc123:example.com"`,
			attrs: []string{"access_token"},
			serverSecrets: notificationSecrets(func(n *notification.Matrix) map[string]string {
				return singleSecret("access_token", n.AccessToken)
			}),
		},
		"uptimekuma_notification_mattermost": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Mattermost) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_max": {
			body: `
  chat_id = "-123456789"`,
			attrs: []string{"bot_token"},
			serverSecrets: notificationSecrets(func(n *notification.Max) map[string]string {
				return singleSecret("bot_token", n.BotToken)
			}),
		},
		"uptimekuma_notification_milky": {
			body: `
  http_addr   = "http://localhost:3000"
  msg_type    = "group"
  receiver_id = "123456789"`,
			attrs: []string{"access_token"},
			serverSecrets: notificationSecrets(func(n *notification.Milky) map[string]string {
				return singleSecret("access_token", n.AccessToken)
			}),
		},
		"uptimekuma_notification_nextcloudtalk": {
			body: `
  host               = "https://nextcloud.example.com"
  conversation_token = "room-token"`,
			attrs: []string{"bot_secret"},
			serverSecrets: notificationSecrets(func(n *notification.NextcloudTalk) map[string]string {
				return singleSecret("bot_secret", n.BotSecret)
			}),
		},
		"uptimekuma_notification_nostr": {
			body: `
  recipients = "npub1qypt4l5elx7qjxapqvzc3gw7nj5zxwq5r5rzc5yqgj5j5j5j5j5j5j5j5j5j"
  relays     = "wss://relay.example.com"`,
			attrs: []string{"sender"},
			serverSecrets: notificationSecrets(func(n *notification.Nostr) map[string]string {
				return singleSecret("sender", n.Sender)
			}),
		},
		"uptimekuma_notification_notifery": {
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Notifery) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_notifyapp": {
			body: `
  device_id = "ABC12345"`,
			attrs: []string{"token"},
			serverSecrets: notificationSecrets(func(n *notification.NotifyApp) map[string]string {
				return singleSecret("token", n.Token)
			}),
		},
		"uptimekuma_notification_ntfy": {
			body: `
  topic                 = "uptime_kuma"
  authentication_method = "accessToken"`,
			attrs: []string{"access_token"},
			serverSecrets: notificationSecrets(func(n *notification.Ntfy) map[string]string {
				return singleSecret("access_token", n.AccessToken)
			}),
		},
		"uptimekuma_notification_ntfy/usernamePassword": {
			body: `
  topic                 = "uptime_kuma"
  authentication_method = "usernamePassword"
  username              = "ntfy-user"`,
			attrs: []string{"password"},
			serverSecrets: notificationSecrets(func(n *notification.Ntfy) map[string]string {
				return singleSecret("password", n.Password)
			}),
		},
		"uptimekuma_notification_octopush": {
			body: `
  version      = "2"
  login        = "octopush-user"
  phone_number = "+33600000000"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Octopush) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_octopush/v1": {
			body: `
  version         = "1"
  dm_login        = "octopush-user"
  dm_phone_number = "+33600000000"`,
			attrs: []string{"dm_api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Octopush) map[string]string {
				return singleSecret("dm_api_key", n.DMAPIKey)
			}),
		},
		"uptimekuma_notification_onebot": {
			body: `
  http_addr   = "http://localhost:5700"
  msg_type    = "group"
  receiver_id = "123456789"`,
			attrs:       []string{"access_token"},
			importOmits: []string{"access_token"},
			serverSecrets: notificationSecrets(func(n *notification.OneBot) map[string]string {
				return singleSecret("access_token", n.AccessToken)
			}),
		},
		"uptimekuma_notification_onechat": {
			body: `
  receiver_id = "user123"
  bot_id      = "bot-001"`,
			attrs: []string{"access_token"},
			serverSecrets: notificationSecrets(func(n *notification.OneChat) map[string]string {
				return singleSecret("access_token", n.AccessToken)
			}),
		},
		"uptimekuma_notification_onesender": {
			body: `
  url           = "https://onesender.example.com/api/v1/send"
  receiver      = "+6281234567890"
  type_receiver = "private"`,
			attrs: []string{"token"},
			serverSecrets: notificationSecrets(func(n *notification.OneSender) map[string]string {
				return singleSecret("token", n.Token)
			}),
		},
		"uptimekuma_notification_ooredoo": {
			body: `
  username  = "bulk-sms-user"
  to_number = "9601234567"`,
			attrs: []string{"access_key", "bearer_token"},
			serverSecrets: notificationSecrets(func(n *notification.Ooredoo) map[string]string {
				return map[string]string{"access_key": n.AccessKey, "bearer_token": n.BearerToken}
			}),
		},
		"uptimekuma_notification_openwa": {
			body: `
  api_url = "http://localhost:2785/"
  session = "default"
  chat_id = "00117612345678@c.us"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.OpenWa) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_opsgenie": {
			body: `
  region = "us"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Opsgenie) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_pagerduty": {
			attrs: []string{"integration_url", "integration_key"},
			serverSecrets: notificationSecrets(func(n *notification.PagerDuty) map[string]string {
				return map[string]string{"integration_url": n.IntegrationURL, "integration_key": n.IntegrationKey}
			}),
		},
		"uptimekuma_notification_pagertree": {
			attrs: []string{"integration_url"},
			serverSecrets: notificationSecrets(func(n *notification.PagerTree) map[string]string {
				return singleSecret("integration_url", n.IntegrationURL)
			}),
		},
		"uptimekuma_notification_pinglet": {
			body: `
  publish_url = "https://app.pinglet.co.uk/test-namespace/alerts"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Pinglet) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_plivo": {
			body: `
  auth_id     = "MA0123456789ABCDEFGH"
  from_number = "+15550001111"
  to_number   = "+15550002222"`,
			attrs: []string{"auth_token"},
			serverSecrets: notificationSecrets(func(n *notification.Plivo) map[string]string {
				return singleSecret("auth_token", n.AuthToken)
			}),
		},
		"uptimekuma_notification_promosms": {
			body: `
  login        = "promosms-user"
  phone_number = "+48501234567"`,
			attrs: []string{"password"},
			serverSecrets: notificationSecrets(func(n *notification.PromoSMS) map[string]string {
				return singleSecret("password", n.Password)
			}),
		},
		"uptimekuma_notification_pumble": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Pumble) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_pushbullet": {
			attrs: []string{"access_token"},
			serverSecrets: notificationSecrets(func(n *notification.Pushbullet) map[string]string {
				return singleSecret("access_token", n.AccessToken)
			}),
		},
		"uptimekuma_notification_pushdeer": {
			attrs: []string{"key"},
			serverSecrets: notificationSecrets(func(n *notification.PushDeer) map[string]string {
				return singleSecret("key", n.Key)
			}),
		},
		"uptimekuma_notification_pushover": {
			body: `
  user_key = "pushover-user-key"`,
			attrs: []string{"app_token"},
			serverSecrets: notificationSecrets(func(n *notification.Pushover) map[string]string {
				return singleSecret("app_token", n.AppToken)
			}),
		},
		"uptimekuma_notification_pushplus": {
			attrs: []string{"send_key"},
			serverSecrets: notificationSecrets(func(n *notification.PushPlus) map[string]string {
				return singleSecret("send_key", n.SendKey)
			}),
		},
		"uptimekuma_notification_pushy": {
			body: `
  token = "pushy-device-token"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Pushy) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_resend": {
			body: `
  from_email = "monitoring@example.com"
  to_email   = "alerts@example.com"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Resend) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_rocketchat": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.RocketChat) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_sendgrid": {
			body: `
  to_email   = "alerts@example.com"
  from_email = "monitoring@example.com"
  subject    = "Uptime Alert"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.SendGrid) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_serverchan": {
			attrs: []string{"send_key"},
			serverSecrets: notificationSecrets(func(n *notification.ServerChan) map[string]string {
				return singleSecret("send_key", n.SendKey)
			}),
		},
		"uptimekuma_notification_serwersms": {
			body: `
  username     = "serwersms-user"
  phone_number = "+48501234567"`,
			attrs: []string{"password"},
			serverSecrets: notificationSecrets(func(n *notification.SerwerSMS) map[string]string {
				return singleSecret("password", n.Password)
			}),
		},
		"uptimekuma_notification_sevenio": {
			body: `
  sender = "+491234567890"
  to     = "+491111111111"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.SevenIO) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_signalgrid": {
			body: `
  channel = "alerts"`,
			attrs: []string{"client_key"},
			serverSecrets: notificationSecrets(func(n *notification.Signalgrid) map[string]string {
				return singleSecret("client_key", n.ClientKey)
			}),
		},
		"uptimekuma_notification_signl4": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.SIGNL4) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_slack": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Slack) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_smsc": {
			body: `
  login     = "smsc-user"
  to_number = "77123456789"`,
			attrs: []string{"password"},
			serverSecrets: notificationSecrets(func(n *notification.SMSC) map[string]string {
				return singleSecret("password", n.Password)
			}),
		},
		"uptimekuma_notification_smseagle": {
			body: `
  url            = "https://smseagle.example.com"
  recipient_type = "smseagle-to"
  recipient_to   = "+1234567890"`,
			attrs: []string{"token"},
			serverSecrets: notificationSecrets(func(n *notification.SMSEagle) map[string]string {
				return singleSecret("token", n.Token)
			}),
		},
		"uptimekuma_notification_smsgateway": {
			body: `
  url        = "http://sms-gateway.example.com:8080"
  recipients = "+15551234567"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.SMSGateway) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_smsir": {
			body: `
  number   = "09123456789"
  template = "123456"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.SMSIR) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_smsmanager": {
			body: `
  numbers = "+41790000000"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.SMSManager) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_smspartner": {
			body: `
  phone_number = "+33612345678"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.SMSPartner) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_smsplanet": {
			body: `
  phone_numbers = "+48123456789"`,
			attrs: []string{"api_token"},
			serverSecrets: notificationSecrets(func(n *notification.SMSPlanet) map[string]string {
				return singleSecret("api_token", n.APIToken)
			}),
		},
		"uptimekuma_notification_smtp": {
			body: `
  host              = "smtp.example.com"
  from              = "uptime-kuma@example.com"
  to                = "admin@example.com"
  dkim_domain       = "example.com"
  dkim_key_selector = "default"`,
			attrs: []string{"dkim_private_key"},
			serverSecrets: notificationSecrets(func(n *notification.SMTP) map[string]string {
				return singleSecret("dkim_private_key", n.DkimPrivateKey)
			}),
		},
		"uptimekuma_notification_splunk": {
			body: `
  severity     = "critical"
  auto_resolve = "resolve"`,
			attrs: []string{"rest_url", "integration_key"},
			serverSecrets: notificationSecrets(func(n *notification.Splunk) map[string]string {
				return map[string]string{
					"rest_url":        n.RestURL,
					"integration_key": n.IntegrationKey,
				}
			}),
		},
		"uptimekuma_notification_spugpush": {
			attrs: []string{"template_key"},
			serverSecrets: notificationSecrets(func(n *notification.SpugPush) map[string]string {
				return singleSecret("template_key", n.TemplateKey)
			}),
		},
		"uptimekuma_notification_squadcast": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Squadcast) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_stackfield": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Stackfield) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_teams": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Teams) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_techuluspush": {
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.TechulusPush) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_telegram": {
			body: `
  chat_id = "123456789"`,
			attrs: []string{"bot_token"},
			serverSecrets: notificationSecrets(func(n *notification.Telegram) map[string]string {
				return singleSecret("bot_token", n.BotToken)
			}),
		},
		"uptimekuma_notification_telnyx": {
			body: `
  phone_number = "+15550001111"
  to_number    = "+15550002222"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.Telnyx) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_teltonika": {
			body: `
  url          = "https://192.168.1.1"
  username     = "admin"
  phone_number = "+1234567890"`,
			attrs: []string{"password"},
			serverSecrets: notificationSecrets(func(n *notification.Teltonika) map[string]string {
				return singleSecret("password", n.Password)
			}),
		},
		"uptimekuma_notification_threema": {
			body: `
  sender_identity = "*ABCDEFG"
  recipient       = "ECHOECHO"
  recipient_type  = "identity"`,
			attrs: []string{"secret"},
			serverSecrets: notificationSecrets(func(n *notification.Threema) map[string]string {
				return singleSecret("secret", n.Secret)
			}),
		},
		"uptimekuma_notification_turbosmtp": {
			body: `
  consumer_key = "turbosmtp-consumer-key"
  from_email   = "alerts@example.com"
  to_email     = "oncall@example.com"`,
			attrs: []string{"consumer_secret"},
			serverSecrets: notificationSecrets(func(n *notification.TurboSMTP) map[string]string {
				return singleSecret("consumer_secret", n.ConsumerSecret)
			}),
		},
		"uptimekuma_notification_twilio": {
			body: `
  account_sid = "AC00000000000000000000000000000000"
  to_number   = "+12025550123"
  from_number = "+12025550456"`,
			attrs: []string{"api_key", "auth_token"},
			serverSecrets: notificationSecrets(func(n *notification.Twilio) map[string]string {
				return map[string]string{"api_key": n.APIKey, "auth_token": n.AuthToken}
			}),
		},
		"uptimekuma_notification_vk": {
			body: `
  peer_id = "12345"`,
			attrs: []string{"access_token"},
			serverSecrets: notificationSecrets(func(n *notification.VK) map[string]string {
				return singleSecret("access_token", n.AccessToken)
			}),
		},
		"uptimekuma_notification_waha": {
			body: `
  api_url = "https://waha.example.com"
  session = "default"
  chat_id = "120363101234567890@g.us"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.WAHA) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_webhook": {
			body: `
  webhook_content_type = "json"`,
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.Webhook) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
		"uptimekuma_notification_wecom": {
			attrs: []string{"bot_key"},
			serverSecrets: notificationSecrets(func(n *notification.WeCom) map[string]string {
				return singleSecret("bot_key", n.BotKey)
			}),
		},
		"uptimekuma_notification_whapi": {
			body: `
  api_url   = "https://gate.whapi.cloud"
  recipient = "1234567890"`,
			attrs: []string{"auth_token"},
			serverSecrets: notificationSecrets(func(n *notification.Whapi) map[string]string {
				return singleSecret("auth_token", n.AuthToken)
			}),
		},
		"uptimekuma_notification_whatsapp360messenger": {
			body: `
  recipient = "+15551234567"`,
			attrs: []string{"auth_token"},
			serverSecrets: notificationSecrets(func(n *notification.Whatsapp360messenger) map[string]string {
				return singleSecret("auth_token", n.AuthToken)
			}),
		},
		"uptimekuma_notification_wpush": {
			body: `
  channel = "monitoring"`,
			attrs: []string{"api_key"},
			serverSecrets: notificationSecrets(func(n *notification.WPush) map[string]string {
				return singleSecret("api_key", n.APIKey)
			}),
		},
		"uptimekuma_notification_wxpusher": {
			attrs: []string{"spt"},
			serverSecrets: notificationSecrets(func(n *notification.WxPusher) map[string]string {
				return singleSecret("spt", n.SPT)
			}),
		},
		"uptimekuma_notification_yzj": {
			attrs: []string{"webhook_url", "token"},
			serverSecrets: notificationSecrets(func(n *notification.YZJ) map[string]string {
				return map[string]string{"webhook_url": n.WebHookURL, "token": n.Token}
			}),
		},
		"uptimekuma_notification_zohocliq": {
			attrs: []string{"webhook_url"},
			serverSecrets: notificationSecrets(func(n *notification.ZohoCliq) map[string]string {
				return singleSecret("webhook_url", n.WebhookURL)
			}),
		},
	}

	for key, tc := range tests {
		t.Run(key, func(t *testing.T) {
			resourceType, _, _ := strings.Cut(key, "/")
			testAccNotificationWriteOnly(t, resourceType, tc)
		})
	}
}
