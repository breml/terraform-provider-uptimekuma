package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MonitorMQTTBaseModel describes the base data model for MQTT monitors.
// While currently MQTT has only one monitor type, this model is kept for organizational
// consistency with other monitor types and to allow for potential future MQTT variants.
type MonitorMQTTBaseModel struct {
	Hostname           types.String `tfsdk:"hostname"`             // MQTT broker hostname or IP.
	Port               types.Int64  `tfsdk:"port"`                 // MQTT broker port.
	MQTTTopic          types.String `tfsdk:"mqtt_topic"`           // Topic to subscribe to.
	MQTTUsername       types.String `tfsdk:"mqtt_username"`        // Optional username for MQTT authentication.
	MQTTPassword       types.String `tfsdk:"mqtt_password"`        // Optional password for MQTT authentication.
	MQTTWebsocketPath  types.String `tfsdk:"mqtt_websocket_path"`  // Optional WebSocket path for WebSocket connections.
	MQTTCheckType      types.String `tfsdk:"mqtt_check_type"`      // Check type: keyword or json-query.
	MQTTSuccessMessage types.String `tfsdk:"mqtt_success_message"` // Expected message for keyword check.
	JSONPath           types.String `tfsdk:"json_path"`            // JSON path for json-query check.
	ExpectedValue      types.String `tfsdk:"expected_value"`       // Expected value for json-query check.
}

// withMQTTMonitorBaseAttributes adds MQTT-specific schema attributes to the provided attribute map.
// This includes hostname, port, topic, authentication, and check configuration options.
func withMQTTMonitorBaseAttributes(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	attrs["hostname"] = mqttHostnameAttribute()
	attrs["port"] = mqttPortAttribute()
	attrs["mqtt_topic"] = mqttTopicAttribute()
	attrs["mqtt_username"] = mqttUsernameAttribute()
	attrs["mqtt_password"] = mqttPasswordAttribute()
	attrs["mqtt_websocket_path"] = mqttWebsocketPathAttribute()
	attrs["mqtt_check_type"] = mqttCheckTypeAttribute()
	attrs["mqtt_success_message"] = mqttSuccessMessageAttribute()
	attrs["json_path"] = mqttJSONPathAttribute()
	attrs["expected_value"] = mqttExpectedValueAttribute()
	return attrs
}

func mqttHostnameAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "MQTT broker hostname or IP address",
		Required:            true,
	}
}

func mqttPortAttribute() schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: "MQTT broker port",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(1883),
	}
}

func mqttTopicAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Topic to subscribe to",
		Required:            true,
	}
}

func mqttUsernameAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "MQTT username for authentication",
		Optional:            true,
	}
}

func mqttPasswordAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "MQTT password for authentication",
		Optional:            true,
		Sensitive:           true,
	}
}

func mqttWebsocketPathAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "WebSocket path for WebSocket connections",
		Optional:            true,
	}
}

func mqttCheckTypeAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Check type: keyword or json-query",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("keyword"),
		Validators: []validator.String{
			stringvalidator.OneOf("keyword", "json-query"),
		},
	}
}

func mqttSuccessMessageAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Expected message for keyword check",
		Optional:            true,
	}
}

func mqttJSONPathAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "JSON path for json-query check",
		Optional:            true,
	}
}

func mqttExpectedValueAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Expected value for json-query check",
		Optional:            true,
	}
}
