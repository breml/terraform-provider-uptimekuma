package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &TagDataSource{}

func NewTagDataSource() datasource.DataSource {
	return &TagDataSource{}
}

type TagDataSource struct {
	client *kuma.Client
}

type TagDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Color types.String `tfsdk:"color"`
}

func (d *TagDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (d *TagDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get tag information by ID or name",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Tag identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Tag name",
				Optional:            true,
				Computed:            true,
			},
			"color": schema.StringAttribute{
				MarkdownDescription: "Tag color (hex color code)",
				Computed:            true,
			},
		},
	}
}

func (d *TagDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *kuma.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *TagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TagDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is provided, use it directly
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		tag, err := d.client.GetTag(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("failed to read tag", err.Error())
			return
		}

		data.Name = types.StringValue(tag.Name)
		data.Color = types.StringValue(tag.Color)

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// If name is provided, search for it
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		tags, err := d.client.GetTags(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read tags", err.Error())
			return
		}

		var foundTag *struct {
			ID    int64
			Name  string
			Color string
		}

		for _, tag := range tags {
			if tag.Name == data.Name.ValueString() {
				if foundTag != nil {
					resp.Diagnostics.AddError(
						"Multiple tags found",
						fmt.Sprintf("Multiple tags with name '%s' found. Please use 'id' to specify the tag uniquely.", data.Name.ValueString()),
					)
					return
				}
				foundTag = &struct {
					ID    int64
					Name  string
					Color string
				}{
					ID:    tag.ID,
					Name:  tag.Name,
					Color: tag.Color,
				}
			}
		}

		if foundTag == nil {
			resp.Diagnostics.AddError(
				"Tag not found",
				fmt.Sprintf("No tag with name '%s' found.", data.Name.ValueString()),
			)
			return
		}

		data.ID = types.Int64Value(foundTag.ID)
		data.Color = types.StringValue(foundTag.Color)

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
