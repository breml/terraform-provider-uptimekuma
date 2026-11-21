package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/maintenance"
)

var _ datasource.DataSource = &MaintenanceDataSource{}

// NewMaintenanceDataSource returns a new instance of the maintenance data source.
func NewMaintenanceDataSource() datasource.DataSource {
	return &MaintenanceDataSource{}
}

// MaintenanceDataSource manages maintenance data source operations.
type MaintenanceDataSource struct {
	client *kuma.Client
}

// MaintenanceDataSourceModel describes the data model for maintenance data source.
type MaintenanceDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Title types.String `tfsdk:"title"`
}

// Metadata returns the metadata for the data source.
func (*MaintenanceDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenance"
}

// Schema returns the schema for the data source.
func (*MaintenanceDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get maintenance window information by ID or name",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Maintenance identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Maintenance name (title)",
				Optional:            true,
				Computed:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Maintenance title",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MaintenanceDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MaintenanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MaintenanceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Attempt to read by ID if provided.
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		d.readByID(ctx, &data, resp)
		return
	}

	// Attempt to read by name if ID not provided.
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		d.readByName(ctx, &data, resp)
		return
	}

	// Error if neither ID nor name provided.
	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}

func (d *MaintenanceDataSource) readByID(
	ctx context.Context,
	data *MaintenanceDataSourceModel,
	resp *datasource.ReadResponse,
) {
	window, err := d.client.GetMaintenance(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read maintenance", err.Error())

		return
	}

	// Populate name and title from API response.
	data.Name = types.StringValue(window.Title)
	data.Title = types.StringValue(window.Title)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *MaintenanceDataSource) readByName(
	ctx context.Context,
	data *MaintenanceDataSourceModel,
	resp *datasource.ReadResponse,
) {
	window, ok := findMaintenanceByTitle(ctx, d.client, data.Name.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}

	// Populate ID and title from matched result.
	data.ID = types.Int64Value(window.ID)
	data.Title = types.StringValue(window.Title)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// maintenanceListResyncer is the part of *kuma.Client findMaintenanceByTitle
// needs, a seam so the lookup can be exercised without a socket.io server.
type maintenanceListResyncer interface {
	resyncer
	maintenanceLister
}

// findMaintenanceByTitle returns the one maintenance window carrying title. It
// reports false, with an error in diags, on a failed lookup, on a miss and on an
// ambiguous title.
//
// The list serves from the state cache, so a resync is forced once before a
// miss is believed, see findWithResync. A list that left out windows it could
// not decode (kuma.ErrMaintenanceDecode) is still searched: a match among the
// decoded windows comes with a warning, because it is unique only among those,
// and a miss names the undecodable windows, because one of them may be the
// window looked for.
func findMaintenanceByTitle(
	ctx context.Context,
	client maintenanceListResyncer,
	title string,
	diags *diag.Diagnostics,
) (maintenance.Maintenance, bool) {
	// decodeErr is the one of the last look, which is the one the outcome rests on.
	var decodeErr error

	matches, found, err := findWithResync(
		ctx, client,
		func(ctx context.Context) ([]maintenance.Maintenance, bool, error) {
			matched, listDecodeErr, matchErr := matchMaintenancesByTitle(ctx, client, title)
			decodeErr = listDecodeErr

			return matched, len(matched) > 0, matchErr
		},
	)
	if err != nil {
		diags.AddError("failed to read maintenances", err.Error())

		return maintenance.Maintenance{}, false
	}

	if !found {
		detail := fmt.Sprintf("No maintenance window with title '%s' found.", title)
		if decodeErr != nil {
			detail += fmt.Sprintf(
				" Some maintenance windows could not be decoded and could be the one looked for: %v. "+
					"Correct their dates in Uptime Kuma; looking the window up by id does not help.",
				decodeErr,
			)
		}

		reportMiss(diags, client, maintenanceListEvent, "Maintenance not found", detail)

		return maintenance.Maintenance{}, false
	}

	// Error if multiple matches found.
	if len(matches) > 1 {
		diags.AddError(
			"Multiple maintenances found",
			fmt.Sprintf(
				"Multiple maintenance windows with title '%s' found. Please use 'id' to specify the maintenance uniquely.",
				title,
			),
		)

		return maintenance.Maintenance{}, false
	}

	if decodeErr != nil {
		diags.AddWarning(
			"Some maintenance windows could not be decoded",
			fmt.Sprintf(
				"The maintenance window with title '%s' is unique only among the windows that could be "+
					"decoded: %v. Correct their dates in Uptime Kuma.",
				title, decodeErr,
			),
		)
	}

	return matches[0], true
}

// matchMaintenancesByTitle returns the cached maintenance windows carrying
// title. It reports every match rather than the first, so that the caller can
// tell an ambiguous title from a missing one.
//
// An error wrapping kuma.ErrMaintenanceDecode comes with the windows that did
// decode, which are searched all the same; it is returned as decodeErr, for the
// caller to mention. Any other failure to read the list is returned as err
// rather than as an empty result, because findWithResync must not mistake it
// for a miss and resync.
func matchMaintenancesByTitle(
	ctx context.Context,
	client maintenanceLister,
	title string,
) (matched []maintenance.Maintenance, decodeErr error, err error) {
	maintenances, err := client.GetMaintenances(ctx)
	if err != nil {
		if !errors.Is(err, kuma.ErrMaintenanceDecode) {
			return nil, nil, fmt.Errorf("get maintenances: %w", err)
		}

		decodeErr = err
	}

	for i := range maintenances {
		if maintenances[i].Title == title {
			matched = append(matched, maintenances[i])
		}
	}

	return matched, decodeErr, nil
}
