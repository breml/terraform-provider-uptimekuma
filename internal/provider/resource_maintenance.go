package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/maintenance"
)

var (
	// Ensure MaintenanceResource satisfies various resource interfaces.
	_ resource.Resource                = &MaintenanceResource{}
	_ resource.ResourceWithImportState = &MaintenanceResource{}
)

// NewMaintenanceResource returns a new instance of the Maintenance resource.
func NewMaintenanceResource() resource.Resource {
	return &MaintenanceResource{}
}

// MaintenanceResource defines the resource implementation for maintenance windows.
type MaintenanceResource struct {
	client *kuma.Client
}

// TimeOfDayModel describes the time of day data model.
type TimeOfDayModel struct {
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
	Seconds types.Int64 `tfsdk:"seconds"`
}

// TimeslotModel describes the timeslot data model with start and end date times for the maintenance window.
type TimeslotModel struct {
	StartDate types.String `tfsdk:"start_date"`
	EndDate   types.String `tfsdk:"end_date"`
}

// MaintenanceResourceModel describes the Maintenance resource data model.
type MaintenanceResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Title            types.String `tfsdk:"title"`
	Description      types.String `tfsdk:"description"`
	Strategy         types.String `tfsdk:"strategy"`
	Active           types.Bool   `tfsdk:"active"`
	StartDate        types.String `tfsdk:"start_date"`
	EndDate          types.String `tfsdk:"end_date"`
	IntervalDay      types.Int64  `tfsdk:"interval_day"`
	Weekdays         types.List   `tfsdk:"weekdays"`
	DaysOfMonth      types.List   `tfsdk:"days_of_month"`
	Cron             types.String `tfsdk:"cron"`
	DurationMinutes  types.Int64  `tfsdk:"duration_minutes"`
	StartTime        types.Object `tfsdk:"start_time"`
	EndTime          types.Object `tfsdk:"end_time"`
	Timezone         types.String `tfsdk:"timezone"`
	Status           types.String `tfsdk:"status"`
	TimezoneResolved types.String `tfsdk:"timezone_resolved"`
	TimezoneOffset   types.String `tfsdk:"timezone_offset"`
	Duration         types.Int64  `tfsdk:"duration"`
	TimeslotList     types.List   `tfsdk:"timeslot_list"`
}

// Metadata returns the metadata for the resource.
func (*MaintenanceResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenance"
}

// Schema returns the schema for the resource.
func (*MaintenanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Maintenance window resource",
		Attributes:          maintenanceSchemaAttributes(),
	}
}

// maintenanceSchemaAttributes builds the schema for all maintenance window resource attributes.
// Combines all attribute types including schedule, timezone, and status fields.
func maintenanceSchemaAttributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"id":                maintenanceIDAttribute(),
		"title":             maintenanceTitleAttribute(),
		"description":       maintenanceDescriptionAttribute(),
		"strategy":          maintenanceStrategyAttribute(),
		"active":            maintenanceActiveAttribute(),
		"start_date":        maintenanceStartDateAttribute(),
		"end_date":          maintenanceEndDateAttribute(),
		"interval_day":      maintenanceIntervalDayAttribute(),
		"weekdays":          maintenanceWeekdaysAttribute(),
		"days_of_month":     maintenanceDaysOfMonthAttribute(),
		"cron":              maintenanceCronAttribute(),
		"duration_minutes":  maintenanceDurationMinutesAttribute(),
		"start_time":        maintenanceStartTimeAttribute(),
		"end_time":          maintenanceEndTimeAttribute(),
		"timezone":          maintenanceTimezoneAttribute(),
		"status":            maintenanceStatusAttribute(),
		"timezone_resolved": maintenanceTimezoneResolvedAttribute(),
		"timezone_offset":   maintenanceTimezoneOffsetAttribute(),
		"duration":          maintenanceDurationAttribute(),
		"timeslot_list":     maintenanceTimeslotListAttribute(),
	}
	return attrs
}

// maintenanceIDAttribute returns the schema for the maintenance ID.
// This is computed by the server and used to track the resource.
func maintenanceIDAttribute() schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: "Maintenance window ID",
		Computed:            true,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	}
}

// maintenanceTitleAttribute returns the schema for maintenance title/name.
// A required field that identifies the maintenance window.
func maintenanceTitleAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Name of the maintenance window",
		Required:            true,
	}
}

// maintenanceDescriptionAttribute returns the schema for maintenance description.
// Provides optional details about the maintenance window purpose or scope.
func maintenanceDescriptionAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Additional details about the maintenance",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(""),
	}
}

// maintenanceStrategyAttribute returns the schema for maintenance scheduling strategy.
// Determines how the maintenance window recurs or is scheduled.
func maintenanceStrategyAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Scheduling pattern: single, recurring-interval, recurring-weekday, recurring-day-of-month, cron, manual",
		Required:            true,
		Validators: []validator.String{
			stringvalidator.OneOf(
				"single",
				"recurring-interval",
				"recurring-weekday",
				"recurring-day-of-month",
				"cron",
				"manual",
			),
		},
	}
}

// maintenanceActiveAttribute returns the schema for the active flag.
// Controls whether the maintenance window is enforced.
func maintenanceActiveAttribute() schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: "Whether the maintenance window is active",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(true),
	}
}

func maintenanceStartDateAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Start date/time for single strategy (RFC3339 format)",
		Optional:            true,
	}
}

func maintenanceEndDateAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "End date/time for single strategy (RFC3339 format)",
		Optional:            true,
	}
}

func maintenanceIntervalDayAttribute() schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: "Interval in days for recurring-interval strategy",
		Optional:            true,
	}
}

// maintenanceWeekdaysAttribute returns the schema for weekdays selection (recurring-weekday strategy).
// Values: 1=Monday, 2=Tuesday, ..., 7=Sunday.
func maintenanceWeekdaysAttribute() schema.ListAttribute {
	return schema.ListAttribute{
		MarkdownDescription: "Days of week for recurring-weekday (1=Monday...7=Sunday)",
		Optional:            true,
		ElementType:         types.Int64Type,
	}
}

// maintenanceDaysOfMonthAttribute returns the schema for days of month selection (recurring-day-of-month strategy).
// Accepts numeric days (1-31) or special values like lastDay1-lastDay4 for last occurrence.
func maintenanceDaysOfMonthAttribute() schema.ListAttribute {
	return schema.ListAttribute{
		MarkdownDescription: "Days of month for recurring-day-of-month (1-31 or lastDay1-lastDay4)",
		Optional:            true,
		ElementType:         types.StringType,
	}
}

// maintenanceCronAttribute returns the schema for cron expression (cron strategy).
// Allows flexible scheduling using standard cron syntax.
func maintenanceCronAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Cron expression for cron strategy",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

// maintenanceDurationMinutesAttribute returns the schema for maintenance duration (cron strategy).
// Specifies how long the maintenance window lasts.
func maintenanceDurationMinutesAttribute() schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: "Duration in minutes for cron strategy",
		Optional:            true,
	}
}

func maintenanceStartTimeAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Start time for recurring strategies",
		Optional:            true,
		Attributes:          timeOfDaySchemaAttributes(),
	}
}

// maintenanceEndTimeAttribute returns the schema for the end time.
// Only used for recurring maintenance strategies.
func maintenanceEndTimeAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "End time for recurring strategies",
		Optional:            true,
		Attributes:          timeOfDaySchemaAttributes(),
	}
}

// maintenanceTimezoneAttribute returns the schema for timezone configuration.
// Defaults to UTC but can use server's timezone or specific IANA timezone.
func maintenanceTimezoneAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Timezone option: UTC, SAME_AS_SERVER, or IANA timezone (e.g., America/New_York)",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString("UTC"),
	}
}

func maintenanceStatusAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Current status: inactive, scheduled, under-maintenance, ended, unknown",
		Computed:            true,
	}
}

func maintenanceTimezoneResolvedAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Resolved IANA timezone",
		Computed:            true,
	}
}

func maintenanceTimezoneOffsetAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Timezone offset from UTC",
		Computed:            true,
	}
}

func maintenanceDurationAttribute() schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: "Duration in seconds (computed)",
		Computed:            true,
	}
}

func maintenanceTimeslotListAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Scheduled maintenance windows",
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"start_date": schema.StringAttribute{
					MarkdownDescription: "RFC3339 timestamp",
					Computed:            true,
				},
				"end_date": schema.StringAttribute{
					MarkdownDescription: "RFC3339 timestamp",
					Computed:            true,
				},
			},
		},
	}
}

func timeOfDaySchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"hours": schema.Int64Attribute{
			MarkdownDescription: "Hours (0-23)",
			Required:            true,
		},
		"minutes": schema.Int64Attribute{
			MarkdownDescription: "Minutes (0-59)",
			Required:            true,
		},
		"seconds": schema.Int64Attribute{
			MarkdownDescription: "Seconds (0-59)",
			Required:            true,
		},
	}
}

// Configure configures the resource with the API client.
func (r *MaintenanceResource) Configure(
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
func (r *MaintenanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Extract and validate configuration.
	var data MaintenanceResourceModel

	// Extract plan data.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	m := &maintenance.Maintenance{
		Title:       data.Title.ValueString(),
		Description: data.Description.ValueString(),
		Strategy:    data.Strategy.ValueString(),
		Active:      data.Active.ValueBool(),
	}

	err := r.populateMaintenanceFromModel(ctx, &data, m, &resp.Diagnostics)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to populate maintenance", err.Error())
		return
	}

	created, err := r.client.CreateMaintenance(ctx, m)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create maintenance", err.Error())
		return
	}

	data.ID = types.Int64Value(created.ID)
	r.populateModelFromMaintenance(ctx, created, &data, &resp.Diagnostics)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the resource.
func (r *MaintenanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MaintenanceResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.GetMaintenance(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read maintenance", err.Error())
		return
	}

	r.populateModelFromMaintenance(ctx, m, &data, &resp.Diagnostics)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MaintenanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MaintenanceResourceModel

	// Extract plan data.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	m := &maintenance.Maintenance{
		ID:          data.ID.ValueInt64(),
		Title:       data.Title.ValueString(),
		Description: data.Description.ValueString(),
		Strategy:    data.Strategy.ValueString(),
		Active:      data.Active.ValueBool(),
	}

	err := r.populateMaintenanceFromModel(ctx, &data, m, &resp.Diagnostics)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to populate maintenance", err.Error())
		return
	}

	err = r.client.UpdateMaintenance(ctx, m)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update maintenance", err.Error())
		return
	}

	updated, err := r.client.GetMaintenance(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to read updated maintenance", err.Error())
		return
	}

	r.populateModelFromMaintenance(ctx, updated, &data, &resp.Diagnostics)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *MaintenanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MaintenanceResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMaintenance(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete maintenance", err.Error())
		return
	}
}

// ValidateConfig validates the resource configuration.
func (*MaintenanceResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data MaintenanceResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	strategy := data.Strategy.ValueString()

	switch strategy {
	case "single":
		validateMaintenanceSingleStrategy(&data, &resp.Diagnostics)

	case "recurring-interval":
		validateMaintenanceRecurringIntervalStrategy(&data, &resp.Diagnostics)

	case "recurring-weekday":
		validateMaintenanceRecurringWeekdayStrategy(&data, &resp.Diagnostics)

	case "recurring-day-of-month":
		validateMaintenanceRecurringDayOfMonthStrategy(&data, &resp.Diagnostics)

	case "cron":
		validateMaintenanceCronStrategy(&data, &resp.Diagnostics)

	default:
		// manual strategy has no validation
	}
}

func validateMaintenanceSingleStrategy(data *MaintenanceResourceModel, diags *diag.Diagnostics) {
	if data.StartDate.IsNull() || data.EndDate.IsNull() {
		diags.AddAttributeError(
			path.Root("strategy"),
			"Invalid Configuration",
			"start_date and end_date are required for single strategy",
		)
	}
}

func validateMaintenanceRecurringIntervalStrategy(data *MaintenanceResourceModel, diags *diag.Diagnostics) {
	if data.IntervalDay.IsNull() {
		diags.AddAttributeError(
			path.Root("strategy"),
			"Invalid Configuration",
			"interval_day is required for recurring-interval strategy",
		)
	}

	if data.StartTime.IsNull() || data.EndTime.IsNull() {
		diags.AddAttributeError(
			path.Root("strategy"),
			"Invalid Configuration",
			"start_time and end_time are required for recurring-interval strategy",
		)
	}
}

func validateMaintenanceRecurringWeekdayStrategy(data *MaintenanceResourceModel, diags *diag.Diagnostics) {
	if data.Weekdays.IsNull() {
		diags.AddAttributeError(
			path.Root("strategy"),
			"Invalid Configuration",
			"weekdays is required for recurring-weekday strategy",
		)
	}

	if data.StartTime.IsNull() || data.EndTime.IsNull() {
		diags.AddAttributeError(
			path.Root("strategy"),
			"Invalid Configuration",
			"start_time and end_time are required for recurring-weekday strategy",
		)
	}
}

func validateMaintenanceRecurringDayOfMonthStrategy(data *MaintenanceResourceModel, diags *diag.Diagnostics) {
	if data.DaysOfMonth.IsNull() {
		diags.AddAttributeError(
			path.Root("strategy"),
			"Invalid Configuration",
			"days_of_month is required for recurring-day-of-month strategy",
		)
	}

	if data.StartTime.IsNull() || data.EndTime.IsNull() {
		diags.AddAttributeError(
			path.Root("strategy"),
			"Invalid Configuration",
			"start_time and end_time are required for recurring-day-of-month strategy",
		)
	}
}

func validateMaintenanceCronStrategy(data *MaintenanceResourceModel, diags *diag.Diagnostics) {
	if data.Cron.IsNull() {
		diags.AddAttributeError(
			path.Root("strategy"),
			"Invalid Configuration",
			"cron is required for cron strategy",
		)
	}

	if data.DurationMinutes.IsNull() {
		diags.AddAttributeError(
			path.Root("strategy"),
			"Invalid Configuration",
			"duration_minutes is required for cron strategy",
		)
	}
}

// ImportState imports an existing resource by ID.
func (*MaintenanceResource) ImportState(
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

func (r *MaintenanceResource) populateMaintenanceFromModel(
	ctx context.Context,
	data *MaintenanceResourceModel,
	m *maintenance.Maintenance,
	diags *diag.Diagnostics,
) error {
	strategy := data.Strategy.ValueString()

	switch strategy {
	case "single":
		return r.populateMaintenanceFromModelSingle(data, m)

	case "recurring-interval":
		return r.populateMaintenanceFromModelRecurringInterval(ctx, data, m, diags)

	case "recurring-weekday":
		return r.populateMaintenanceFromModelRecurringWeekday(ctx, data, m, diags)

	case "recurring-day-of-month":
		return r.populateMaintenanceFromModelRecurringDayOfMonth(ctx, data, m, diags)

	case "cron":
		return r.populateMaintenanceFromModelCron(data, m)

	default:
		m.DateRange = []*time.Time{nil, nil}
	}

	return nil
}

func (*MaintenanceResource) populateMaintenanceFromModelSingle(
	data *MaintenanceResourceModel,
	m *maintenance.Maintenance,
) error {
	if !data.StartDate.IsNull() && !data.EndDate.IsNull() {
		startDate, err := time.Parse(time.RFC3339, data.StartDate.ValueString())
		if err != nil {
			return fmt.Errorf("invalid start_date: %w", err)
		}

		endDate, err := time.Parse(time.RFC3339, data.EndDate.ValueString())
		if err != nil {
			return fmt.Errorf("invalid end_date: %w", err)
		}

		m.DateRange = []*time.Time{&startDate, &endDate}
	}

	if !data.Timezone.IsNull() {
		m.TimezoneOption = data.Timezone.ValueString()
	}

	return nil
}

func (r *MaintenanceResource) populateMaintenanceFromModelRecurringInterval(
	ctx context.Context,
	data *MaintenanceResourceModel,
	m *maintenance.Maintenance,
	diags *diag.Diagnostics,
) error {
	if !data.IntervalDay.IsNull() {
		m.IntervalDay = int(data.IntervalDay.ValueInt64())
	}

	m.DateRange = []*time.Time{nil, nil}
	err := r.populateTimeRange(ctx, data, m, diags)
	if err != nil {
		return err
	}

	if !data.Timezone.IsNull() {
		m.TimezoneOption = data.Timezone.ValueString()
	}

	return nil
}

func (r *MaintenanceResource) populateMaintenanceFromModelRecurringWeekday(
	ctx context.Context,
	data *MaintenanceResourceModel,
	m *maintenance.Maintenance,
	diags *diag.Diagnostics,
) error {
	if !data.Weekdays.IsNull() {
		var weekdays []int64
		diags.Append(data.Weekdays.ElementsAs(ctx, &weekdays, false)...)
		if diags.HasError() {
			return errors.New("invalid weekdays")
		}

		m.Weekdays = make([]int, len(weekdays))
		for i, w := range weekdays {
			m.Weekdays[i] = int(w)
		}
	}

	m.DateRange = []*time.Time{nil, nil}
	err := r.populateTimeRange(ctx, data, m, diags)
	if err != nil {
		return err
	}

	if !data.Timezone.IsNull() {
		m.TimezoneOption = data.Timezone.ValueString()
	}

	return nil
}

func (r *MaintenanceResource) populateMaintenanceFromModelRecurringDayOfMonth(
	ctx context.Context,
	data *MaintenanceResourceModel,
	m *maintenance.Maintenance,
	diags *diag.Diagnostics,
) error {
	if !data.DaysOfMonth.IsNull() {
		var daysOfMonth []string
		diags.Append(data.DaysOfMonth.ElementsAs(ctx, &daysOfMonth, false)...)
		if diags.HasError() {
			return errors.New("invalid days_of_month")
		}

		m.DaysOfMonth = make([]any, len(daysOfMonth))
		for i, d := range daysOfMonth {
			m.DaysOfMonth[i] = d
		}
	}

	m.DateRange = []*time.Time{nil, nil}
	err := r.populateTimeRange(ctx, data, m, diags)
	if err != nil {
		return err
	}

	if !data.Timezone.IsNull() {
		m.TimezoneOption = data.Timezone.ValueString()
	}

	return nil
}

func (*MaintenanceResource) populateMaintenanceFromModelCron(
	data *MaintenanceResourceModel,
	m *maintenance.Maintenance,
) error {
	if !data.Cron.IsNull() {
		m.Cron = data.Cron.ValueString()
	}

	if !data.DurationMinutes.IsNull() {
		m.DurationMinutes = int(data.DurationMinutes.ValueInt64())
	}

	m.DateRange = []*time.Time{nil, nil}
	if !data.Timezone.IsNull() {
		m.TimezoneOption = data.Timezone.ValueString()
	}

	return nil
}

func (*MaintenanceResource) populateTimeRange(
	ctx context.Context,
	data *MaintenanceResourceModel,
	m *maintenance.Maintenance,
	diags *diag.Diagnostics,
) error {
	if !data.StartTime.IsNull() && !data.EndTime.IsNull() {
		var startTime TimeOfDayModel
		diags.Append(data.StartTime.As(ctx, &startTime, basetypes.ObjectAsOptions{})...)
		// Check for configuration errors.
		if diags.HasError() {
			return errors.New("invalid start_time")
		}

		var endTime TimeOfDayModel
		diags.Append(data.EndTime.As(ctx, &endTime, basetypes.ObjectAsOptions{})...)
		// Check for configuration errors.
		if diags.HasError() {
			return errors.New("invalid end_time")
		}

		m.TimeRange = []maintenance.TimeOfDay{
			{
				Hours:   int(startTime.Hours.ValueInt64()),
				Minutes: int(startTime.Minutes.ValueInt64()),
				Seconds: int(startTime.Seconds.ValueInt64()),
			},
			{
				Hours:   int(endTime.Hours.ValueInt64()),
				Minutes: int(endTime.Minutes.ValueInt64()),
				Seconds: int(endTime.Seconds.ValueInt64()),
			},
		}
	}

	return nil
}

func (r *MaintenanceResource) populateModelFromMaintenance(
	ctx context.Context,
	m *maintenance.Maintenance,
	data *MaintenanceResourceModel,
	diags *diag.Diagnostics,
) {
	data.Title = types.StringValue(m.Title)
	data.Description = types.StringValue(m.Description)
	data.Strategy = types.StringValue(m.Strategy)
	data.Active = types.BoolValue(m.Active)

	if m.Status != "" {
		data.Status = types.StringValue(m.Status)
	}

	if m.Timezone != "" {
		data.TimezoneResolved = types.StringValue(m.Timezone)
	}

	if m.TimezoneOption != "" {
		data.Timezone = types.StringValue(m.TimezoneOption)
	}

	if m.TimezoneOffset != "" {
		data.TimezoneOffset = types.StringValue(m.TimezoneOffset)
	}

	if m.Duration > 0 {
		data.Duration = types.Int64Value(int64(m.Duration))
	} else {
		data.Duration = types.Int64Null()
	}

	if m.Cron != "" {
		data.Cron = types.StringValue(m.Cron)
	} else {
		data.Cron = types.StringNull()
	}

	switch m.Strategy {
	case "single":
		r.populateModelFromMaintenanceSingle(m, data)

	case "recurring-interval":
		r.populateModelFromMaintenanceRecurringInterval(ctx, m, data, diags)

	case "recurring-weekday":
		r.populateModelFromMaintenanceRecurringWeekday(ctx, m, data, diags)

	case "recurring-day-of-month":
		r.populateModelFromMaintenanceRecurringDayOfMonth(ctx, m, data, diags)

	case "cron":
		r.populateModelFromMaintenanceCron(m, data)

	default:
		// manual strategy has no special handling
	}

	r.populateMaintenanceTimeslotList(ctx, m, data, diags)
}

func (*MaintenanceResource) populateModelFromMaintenanceSingle(
	m *maintenance.Maintenance,
	data *MaintenanceResourceModel,
) {
	if len(m.DateRange) == 2 && m.DateRange[0] != nil && m.DateRange[1] != nil {
		data.StartDate = types.StringValue(m.DateRange[0].Format(time.RFC3339))
		data.EndDate = types.StringValue(m.DateRange[1].Format(time.RFC3339))
	}
}

func (r *MaintenanceResource) populateModelFromMaintenanceRecurringInterval(
	_ context.Context,
	m *maintenance.Maintenance,
	data *MaintenanceResourceModel,
	diags *diag.Diagnostics,
) {
	if m.IntervalDay > 0 {
		data.IntervalDay = types.Int64Value(int64(m.IntervalDay))
	}

	r.populateModelTimeRange(m, data, diags)
}

func (r *MaintenanceResource) populateModelFromMaintenanceRecurringWeekday(
	ctx context.Context,
	m *maintenance.Maintenance,
	data *MaintenanceResourceModel,
	diags *diag.Diagnostics,
) {
	if len(m.Weekdays) > 0 {
		weekdays := make([]int64, len(m.Weekdays))
		for i, w := range m.Weekdays {
			weekdays[i] = int64(w)
		}

		listValue, d := types.ListValueFrom(ctx, types.Int64Type, weekdays)
		diags.Append(d...)
		data.Weekdays = listValue
	}

	r.populateModelTimeRange(m, data, diags)
}

func (r *MaintenanceResource) populateModelFromMaintenanceRecurringDayOfMonth(
	ctx context.Context,
	m *maintenance.Maintenance,
	data *MaintenanceResourceModel,
	diags *diag.Diagnostics,
) {
	if len(m.DaysOfMonth) > 0 {
		daysOfMonth := make([]string, len(m.DaysOfMonth))
		for i, d := range m.DaysOfMonth {
			daysOfMonth[i] = fmt.Sprintf("%v", d)
		}

		listValue, d := types.ListValueFrom(ctx, types.StringType, daysOfMonth)
		diags.Append(d...)
		data.DaysOfMonth = listValue
	}

	r.populateModelTimeRange(m, data, diags)
}

func (*MaintenanceResource) populateModelFromMaintenanceCron(
	m *maintenance.Maintenance,
	data *MaintenanceResourceModel,
) {
	if m.DurationMinutes > 0 {
		data.DurationMinutes = types.Int64Value(int64(m.DurationMinutes))
	}
}

func (*MaintenanceResource) populateMaintenanceTimeslotList(
	_ context.Context,
	m *maintenance.Maintenance,
	data *MaintenanceResourceModel,
	diags *diag.Diagnostics,
) {
	timeslotAttrTypes := map[string]attr.Type{
		"start_date": types.StringType,
		"end_date":   types.StringType,
	}

	if len(m.TimeslotList) > 0 {
		timeslots := make([]TimeslotModel, len(m.TimeslotList))
		for i, ts := range m.TimeslotList {
			timeslots[i] = TimeslotModel{
				StartDate: types.StringValue(ts.StartDate.Format(time.RFC3339)),
				EndDate:   types.StringValue(ts.EndDate.Format(time.RFC3339)),
			}
		}

		timeslotList := make([]attr.Value, len(timeslots))
		for i, ts := range timeslots {
			objValue, d := types.ObjectValue(timeslotAttrTypes, map[string]attr.Value{
				"start_date": ts.StartDate,
				"end_date":   ts.EndDate,
			})
			diags.Append(d...)
			timeslotList[i] = objValue
		}

		listValue, d := types.ListValue(types.ObjectType{AttrTypes: timeslotAttrTypes}, timeslotList)
		diags.Append(d...)
		data.TimeslotList = listValue
	} else {
		listValue, d := types.ListValue(types.ObjectType{AttrTypes: timeslotAttrTypes}, []attr.Value{})
		diags.Append(d...)
		data.TimeslotList = listValue
	}
}

func (*MaintenanceResource) populateModelTimeRange(
	m *maintenance.Maintenance,
	data *MaintenanceResourceModel,
	diags *diag.Diagnostics,
) {
	if len(m.TimeRange) == 2 {
		timeOfDayAttrTypes := map[string]attr.Type{
			"hours":   types.Int64Type,
			"minutes": types.Int64Type,
			"seconds": types.Int64Type,
		}

		startTimeObj, d := types.ObjectValue(timeOfDayAttrTypes, map[string]attr.Value{
			"hours":   types.Int64Value(int64(m.TimeRange[0].Hours)),
			"minutes": types.Int64Value(int64(m.TimeRange[0].Minutes)),
			"seconds": types.Int64Value(int64(m.TimeRange[0].Seconds)),
		})
		diags.Append(d...)
		data.StartTime = startTimeObj

		endTimeObj, d := types.ObjectValue(timeOfDayAttrTypes, map[string]attr.Value{
			"hours":   types.Int64Value(int64(m.TimeRange[1].Hours)),
			"minutes": types.Int64Value(int64(m.TimeRange[1].Minutes)),
			"seconds": types.Int64Value(int64(m.TimeRange[1].Seconds)),
		})
		diags.Append(d...)
		data.EndTime = endTimeObj
	}
}
