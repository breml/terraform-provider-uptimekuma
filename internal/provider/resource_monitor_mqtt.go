package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	// Ensure MonitorMQTTResource satisfies various resource interfaces.
	_ resource.Resource                = &MonitorMQTTResource{}
	_ resource.ResourceWithImportState = &MonitorMQTTResource{}
)

// NewMonitorMQTTResource returns a new instance of the MQTT monitor resource.
func NewMonitorMQTTResource() resource.Resource {
	return &MonitorMQTTResource{}
}

// MonitorMQTTResource defines the resource implementation.
type MonitorMQTTResource struct {
	client *kuma.Client
}

// MonitorMQTTResourceModel describes the resource data model for MQTT monitors.
type MonitorMQTTResourceModel struct {
	MonitorBaseModel
	MonitorMQTTBaseModel
}

// Metadata returns the metadata for the resource.
func (*MonitorMQTTResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_mqtt"
}

// Schema returns the schema for the resource.
func (*MonitorMQTTResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Define resource schema attributes and validation.
	resp.Schema = schema.Schema{
		MarkdownDescription: "MQTT monitor resource",
		Attributes:          withMonitorBaseAttributes(withMQTTMonitorBaseAttributes(map[string]schema.Attribute{})),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorMQTTResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	r.client = client
}

// Create creates a new resource.
func (r *MonitorMQTTResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorMQTTResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mqttMonitor := buildMQTTMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, mqttMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create MQTT monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildMQTTMonitor(ctx context.Context, data *MonitorMQTTResourceModel, diags *diag.Diagnostics) monitor.MQTT {
	mqttMonitor := monitor.MQTT{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		MQTTDetails: monitor.MQTTDetails{
			Hostname:      data.Hostname.ValueString(),
			MQTTTopic:     data.MQTTTopic.ValueString(),
			MQTTCheckType: monitor.MQTTCheckType(data.MQTTCheckType.ValueString()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		mqttMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		mqttMonitor.Parent = &parent
	}

	if !data.Port.IsNull() {
		port := data.Port.ValueInt64()
		mqttMonitor.Port = &port
	}

	if !data.MQTTUsername.IsNull() {
		username := data.MQTTUsername.ValueString()
		mqttMonitor.MQTTUsername = &username
	}

	if !data.MQTTPassword.IsNull() {
		password := data.MQTTPassword.ValueString()
		mqttMonitor.MQTTPassword = &password
	}

	if !data.MQTTWebsocketPath.IsNull() {
		websocketPath := data.MQTTWebsocketPath.ValueString()
		mqttMonitor.MQTTWebsocketPath = &websocketPath
	}

	if !data.MQTTSuccessMessage.IsNull() {
		msg := data.MQTTSuccessMessage.ValueString()
		mqttMonitor.MQTTSuccessMessage = &msg
	}

	if !data.JSONPath.IsNull() {
		jp := data.JSONPath.ValueString()
		mqttMonitor.JSONPath = &jp
	}

	if !data.ExpectedValue.IsNull() {
		ev := data.ExpectedValue.ValueString()
		mqttMonitor.ExpectedValue = &ev
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			mqttMonitor.NotificationIDs = notificationIDs
		}
	}

	return mqttMonitor
}

// populateMQTTMonitorBaseFieldsForMQTT populates the base MQTT monitor fields from the API response.
// Extracts MQTT-specific fields from the API response into the model.
// Handles base monitor fields and all MQTT configuration options.
func populateMQTTMonitorBaseFieldsForMQTT(mqttMonitor *monitor.MQTT, m *MonitorMQTTResourceModel) {
	m.Name = types.StringValue(mqttMonitor.Name)
	if mqttMonitor.Description != nil {
		m.Description = types.StringValue(*mqttMonitor.Description)
	} else {
		m.Description = types.StringNull()
	}

	m.Interval = types.Int64Value(mqttMonitor.Interval)
	m.RetryInterval = types.Int64Value(mqttMonitor.RetryInterval)
	m.ResendInterval = types.Int64Value(mqttMonitor.ResendInterval)
	m.MaxRetries = types.Int64Value(mqttMonitor.MaxRetries)
	m.UpsideDown = types.BoolValue(mqttMonitor.UpsideDown)
	m.Active = types.BoolValue(mqttMonitor.IsActive)
	m.Hostname = types.StringValue(mqttMonitor.Hostname)
	m.MQTTTopic = types.StringValue(mqttMonitor.MQTTTopic)
	m.MQTTCheckType = types.StringValue(string(mqttMonitor.MQTTCheckType))
	m.MQTTSuccessMessage = stringOrNull(stringPtrValue(mqttMonitor.MQTTSuccessMessage))
	m.JSONPath = stringOrNull(stringPtrValue(mqttMonitor.JSONPath))
	m.ExpectedValue = stringOrNull(stringPtrValue(mqttMonitor.ExpectedValue))
	m.MQTTUsername = stringOrNull(stringPtrValue(mqttMonitor.MQTTUsername))
	m.MQTTPassword = stringOrNull(stringPtrValue(mqttMonitor.MQTTPassword))
	m.MQTTWebsocketPath = stringOrNull(stringPtrValue(mqttMonitor.MQTTWebsocketPath))
}

// stringPtrValue returns the value of a string pointer, or empty string if nil.
func stringPtrValue(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// populateOptionalFieldsForMQTT populates optional fields for MQTT monitor.
// Handles parent group, port, and notification IDs.
// Converts null API values to Terraform null types appropriately.
func populateOptionalFieldsForMQTT(
	ctx context.Context,
	mqttMonitor *monitor.MQTT,
	m *MonitorMQTTResourceModel,
	diags *diag.Diagnostics,
) {
	// Set parent monitor group if present.
	if mqttMonitor.Parent != nil {
		m.Parent = types.Int64Value(*mqttMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	// Set port if configured.
	if mqttMonitor.Port != nil {
		m.Port = types.Int64Value(*mqttMonitor.Port)
	} else {
		m.Port = types.Int64Null()
	}

	// Convert notification IDs list if present.
	if len(mqttMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, mqttMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// Read reads the current state of the resource.
func (r *MonitorMQTTResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorMQTTResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mqttMonitor monitor.MQTT
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &mqttMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read MQTT monitor", err.Error())
		return
	}

	populateMQTTMonitorBaseFieldsForMQTT(&mqttMonitor, &data)
	populateOptionalFieldsForMQTT(ctx, &mqttMonitor, &data, &resp.Diagnostics)

	data.Tags = handleMonitorTagsRead(ctx, mqttMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MonitorMQTTResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorMQTTResourceModel
	var state MonitorMQTTResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mqttMonitor := buildMQTTMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	mqttMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, mqttMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update MQTT monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *MonitorMQTTResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorMQTTResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API.
	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete MQTT monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorMQTTResource) ImportState(
	// Import monitor by ID.
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
