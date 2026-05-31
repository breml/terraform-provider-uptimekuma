package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	// Ensure MonitorKafkaProducerResource satisfies various resource interfaces.
	_ resource.Resource                = &MonitorKafkaProducerResource{}
	_ resource.ResourceWithImportState = &MonitorKafkaProducerResource{}
)

// NewMonitorKafkaProducerResource returns a new instance of the Kafka Producer monitor resource.
func NewMonitorKafkaProducerResource() resource.Resource {
	return &MonitorKafkaProducerResource{}
}

// MonitorKafkaProducerResource defines the resource implementation.
type MonitorKafkaProducerResource struct {
	client *kuma.Client
}

// MonitorKafkaProducerResourceModel describes the resource data model for Kafka Producer monitors.
type MonitorKafkaProducerResourceModel struct {
	MonitorBaseModel

	Brokers                types.List   `tfsdk:"brokers"`
	Topic                  types.String `tfsdk:"topic"`
	Message                types.String `tfsdk:"message"`
	SSL                    types.Bool   `tfsdk:"ssl"`
	AllowAutoTopicCreation types.Bool   `tfsdk:"allow_auto_topic_creation"`
	SASLOptions            types.String `tfsdk:"sasl_options"`
}

// Metadata returns the metadata for the resource.
func (*MonitorKafkaProducerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_kafka_producer"
}

// Schema returns the schema for the resource.
func (*MonitorKafkaProducerResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Kafka Producer monitor resource for testing connectivity to Apache Kafka clusters " +
			"and verifying message publishing capabilities.",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"brokers": schema.ListAttribute{
				MarkdownDescription: "List of Kafka broker addresses (e.g. `host:port`).",
				Required:            true,
				ElementType:         types.StringType,
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "Kafka topic to publish the test message to.",
				Required:            true,
			},
			"message": schema.StringAttribute{
				MarkdownDescription: "Test message to publish to the Kafka topic.",
				Required:            true,
			},
			"ssl": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable SSL/TLS for the Kafka connection.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"allow_auto_topic_creation": schema.BoolAttribute{
				MarkdownDescription: "Whether to allow automatic topic creation on the Kafka broker.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"sasl_options": schema.StringAttribute{
				MarkdownDescription: "SASL authentication options as a JSON-encoded object (e.g. " +
					"`{\"mechanism\":\"plain\",\"username\":\"u\",\"password\":\"p\"}`).",
				Optional:  true,
				Sensitive: true,
			},
		}),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorKafkaProducerResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Kafka Producer monitor resource.
func (r *MonitorKafkaProducerResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorKafkaProducerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kafkaMonitor := buildKafkaProducerMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &kafkaMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Kafka Producer monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = handleMonitorActiveStateCreate(ctx, r.client, id, data.Active)
	if err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError("failed to apply monitor active state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// buildKafkaProducerMonitor builds a Kafka Producer monitor API object from the resource model.
func buildKafkaProducerMonitor(
	ctx context.Context,
	data *MonitorKafkaProducerResourceModel,
	diags *diag.Diagnostics,
) monitor.KafkaProducer {
	kafkaMonitor := monitor.KafkaProducer{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		KafkaProducerDetails: monitor.KafkaProducerDetails{
			Topic:                  data.Topic.ValueString(),
			Message:                data.Message.ValueString(),
			SSL:                    data.SSL.ValueBool(),
			AllowAutoTopicCreation: data.AllowAutoTopicCreation.ValueBool(),
		},
	}

	if !data.Brokers.IsNull() && !data.Brokers.IsUnknown() {
		var brokers []string
		diags.Append(data.Brokers.ElementsAs(ctx, &brokers, false)...)
		if diags.HasError() {
			return kafkaMonitor
		}

		kafkaMonitor.Brokers = brokers
	}

	if !data.SASLOptions.IsNull() && !data.SASLOptions.IsUnknown() {
		raw := data.SASLOptions.ValueString()

		var saslOptions map[string]any
		err := json.Unmarshal([]byte(raw), &saslOptions)
		if err != nil {
			diags.AddError("failed to parse sasl_options", fmt.Sprintf("must be valid JSON: %s", err.Error()))
			return kafkaMonitor
		}

		kafkaMonitor.SASLOptions = &saslOptions
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		kafkaMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		kafkaMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			kafkaMonitor.NotificationIDs = notificationIDs
		}
	}

	return kafkaMonitor
}

// populateKafkaProducerMonitorBaseFields populates the resource model with data from the Kafka Producer monitor API
// response.
func populateKafkaProducerMonitorBaseFields(
	ctx context.Context,
	kafkaMonitor *monitor.KafkaProducer,
	m *MonitorKafkaProducerResourceModel,
	diags *diag.Diagnostics,
) {
	m.Name = types.StringValue(kafkaMonitor.Name)
	if kafkaMonitor.Description != nil {
		m.Description = types.StringValue(*kafkaMonitor.Description)
	} else {
		m.Description = types.StringNull()
	}

	m.Interval = types.Int64Value(kafkaMonitor.Interval)
	m.RetryInterval = types.Int64Value(kafkaMonitor.RetryInterval)
	m.ResendInterval = types.Int64Value(kafkaMonitor.ResendInterval)
	m.MaxRetries = types.Int64Value(kafkaMonitor.MaxRetries)
	m.UpsideDown = types.BoolValue(kafkaMonitor.UpsideDown)
	m.Active = types.BoolValue(kafkaMonitor.IsActive)

	m.Topic = types.StringValue(kafkaMonitor.Topic)
	m.SSL = types.BoolValue(kafkaMonitor.SSL)
	m.AllowAutoTopicCreation = types.BoolValue(kafkaMonitor.AllowAutoTopicCreation)

	// Uptime Kuma may not return the test message in the API response.
	// Preserve the existing state value to avoid perpetual diffs.
	if kafkaMonitor.Message != "" {
		m.Message = types.StringValue(kafkaMonitor.Message)
	}

	if len(kafkaMonitor.Brokers) > 0 {
		brokers, d := types.ListValueFrom(ctx, types.StringType, kafkaMonitor.Brokers)
		diags.Append(d...)
		m.Brokers = brokers
	} else {
		m.Brokers = types.ListNull(types.StringType)
	}
}

// populateOptionalFieldsForKafkaProducer populates optional parent and notification fields from the Kafka Producer
// monitor API response.
func populateOptionalFieldsForKafkaProducer(
	ctx context.Context,
	kafkaMonitor *monitor.KafkaProducer,
	m *MonitorKafkaProducerResourceModel,
	diags *diag.Diagnostics,
) {
	if kafkaMonitor.Parent != nil {
		m.Parent = types.Int64Value(*kafkaMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	if len(kafkaMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, kafkaMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// Read reads the current state of the Kafka Producer monitor resource.
func (r *MonitorKafkaProducerResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorKafkaProducerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var kafkaMonitor monitor.KafkaProducer
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &kafkaMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read Kafka Producer monitor", err.Error())
		return
	}

	if actual := kafkaMonitor.Base.Type(); actual != "" && actual != kafkaMonitor.Type() {
		tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
			"id":            data.ID.ValueInt64(),
			"expected_type": kafkaMonitor.Type(),
			"actual_type":   actual,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	populateKafkaProducerMonitorBaseFields(ctx, &kafkaMonitor, &data, &resp.Diagnostics)
	populateOptionalFieldsForKafkaProducer(ctx, &kafkaMonitor, &data, &resp.Diagnostics)

	data.Tags = handleMonitorTagsRead(ctx, kafkaMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Kafka Producer monitor resource.
func (r *MonitorKafkaProducerResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorKafkaProducerResourceModel
	var state MonitorKafkaProducerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kafkaMonitor := buildKafkaProducerMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	kafkaMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &kafkaMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update Kafka Producer monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	handleMonitorActiveStateUpdate(ctx, r.client, data.ID.ValueInt64(), state.Active, data.Active, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Kafka Producer monitor resource.
func (r *MonitorKafkaProducerResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorKafkaProducerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Kafka Producer monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorKafkaProducerResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
