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
	ID           types.String                    `tfsdk:"id"`
	Region       types.String                    `tfsdk:"region"`
	Database     types.String                    `tfsdk:"database"`
	Table        types.String                    `tfsdk:"table"`
	Location     types.String                    `tfsdk:"location"`
	Inputs       []tableInputDataSourceModel     `tfsdk:"inputs"`
	Partitions   []tablePartitionDataSourceModel `tfsdk:"partitions"`
	BetaFeatures []string                        `tfsdk:"beta_features"`
	SkipBackfill types.Bool                      `tfsdk:"skip_backfill"`
}

type tableInputDataSourceModel struct {
	Pattern string `tfsdk:"pattern"`
	Format  string `tfsdk:"format"`
}

type tablePartitionDataSourceModel struct {
	Field string `tfsdk:"field"`
	Type  string `tfsdk:"type"`
	Value string `tfsdk:"value"`
}

func (d *tableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (d *tableDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides table configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform identifier.",
				Computed:    true,
			},
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
				Description:         "S3 url where the table is stored (i.e. `s3://sneller-cache-bucket/db/test-db/test-table/`).",
				MarkdownDescription: "S3 url where the table is stored (i.e. `s3://sneller-cache-bucket/db/test-db/test-table/`).",
				Computed:            true,
			},
			"inputs": schema.ListNestedAttribute{
				Description: "The input definition specifies where the source data is located and it format.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
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
			"partitions": schema.ListNestedAttribute{
				Description: "Synthetic field that is generated from parts of an input URI and used to partition table data..",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Description: "Name of the partition field. If this field conflicts with a field in the input data, the partition field will override it.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Type of the partition field.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Template string that is used to produce the value for the partition field.",
							Computed:    true,
						},
					},
				},
			},
			"beta_features": schema.ListAttribute{
				Description: "List of feature flags that can be used to turn on features for beta-testing.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"skip_backfill": schema.BoolAttribute{
				Description: "Skip scanning the source bucket(s) for matching objects when the first objects are inserted into the table.",
				Computed:    true,
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

	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", tenantInfo.TenantID, region, database, table))
	data.Region = types.StringValue(region)
	data.Location = types.StringValue(fmt.Sprintf("%s/db/%s/%s/", tenantInfo.Regions[region].Bucket, database, table))
	if len(tableDescription.Input) > 0 {
		data.Inputs = make([]tableInputDataSourceModel, 0, len(tableDescription.Input))
		for _, input := range tableDescription.Input {
			data.Inputs = append(data.Inputs, tableInputDataSourceModel(input))
		}
	}
	if len(tableDescription.Partitions) > 0 {
		data.Partitions = make([]tablePartitionDataSourceModel, 0, len(tableDescription.Partitions))
		for _, partition := range tableDescription.Partitions {
			data.Partitions = append(data.Partitions, tablePartitionDataSourceModel(partition))
		}
	}
	data.BetaFeatures = tableDescription.BetaFeatures
	data.SkipBackfill = types.BoolValue(tableDescription.SkipBackfill)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
