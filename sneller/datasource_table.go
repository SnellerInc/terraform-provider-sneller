package sneller

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTableDataSource() datasource.DataSource {
	return &tableDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &tableDataSource{}
	_ datasource.DataSourceWithConfigure = &tableDataSource{}
)

type tableDataSource struct {
	client *Client
}

type tableDataSourceModel struct {
	Region   types.String                `tfsdk:"region"`
	Database types.String                `tfsdk:"database"`
	Table    types.String                `tfsdk:"table"`
	Location types.String                `tfsdk:"location"`
	Input    []tableInputDataSourceModel `tfsdk:"input"`
}

type tableInputDataSourceModel struct {
	Pattern string `tfsdk:"pattern"`
	Format  string `tfsdk:"format"`
}

func (r *tableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (r *tableDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides configuration for a table.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "Region where the table is located. If not set, then the tenant's home region is assumed.",
				Optional:    true,
			},
			"database": schema.StringAttribute{
				Description: "Database name.",
				Required:    true,
			},
			"table": schema.StringAttribute{
				Description: "Table name.",
				Required:    true,
			},
			"location": schema.StringAttribute{
				Description: "S3 url of the database location (i.e. `s3://sneller-cache-bucket/db/test-db/test-table/`).",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"input": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"pattern": schema.StringAttribute{
							Description: "Pattern definition to specify the source pattern (i.e. `s3://sneller-source-bucket/data/*.ndjson`).",
							Computed:    true,
						},
						"format": schema.StringAttribute{
							Description: "Format of the input data (`json`, `json.gz`, `json.zst`, `cloudtrail.json.gz`, `csv`, `csv.gz`, `csv.zst`, `tsv`, `tsv.gz`, `tsv.zst`).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *tableDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *tableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data tableDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := data.Region.ValueString()
	if region == "" {
		region = d.client.DefaultRegion
	}

	tenantInfo, err := d.client.Tenant(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info in region %s: %v", region, err.Error()),
		)
		return
	}

	database, table := data.Database.ValueString(), data.Table.ValueString()
	tableDescription, err := d.client.Table(ctx, region, database, table)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get table configuration",
			fmt.Sprintf("Unable to get tenant table configuration of table %s:%s in region %s: %v", database, table, region, err.Error()),
		)
		return
	}

	data.Region = types.StringValue(region)
	data.Location = types.StringValue(fmt.Sprintf("%s/db/%s/%s/", tenantInfo.Regions[region].Bucket, database, table))
	data.Input = make([]tableInputDataSourceModel, 0, len(tableDescription.Input))
	for _, input := range tableDescription.Input {
		data.Input = append(data.Input, tableInputDataSourceModel(input))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
