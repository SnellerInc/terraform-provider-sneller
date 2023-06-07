package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"terraform-provider-sneller/sneller/api"
	"terraform-provider-sneller/sneller/model"

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
	client *api.Client
}

type tableDataSourceModel struct {
	ID              types.String                `tfsdk:"id" json:"ignore"`
	Region          types.String                `tfsdk:"region" json:"ignore"`
	Database        types.String                `tfsdk:"database" json:"ignore"`
	Location        types.String                `tfsdk:"location" json:"ignore"`
	Table           *string                     `tfsdk:"table" json:"name"`
	Inputs          []model.TableInputModel     `tfsdk:"inputs" json:"input"`
	Partitions      []model.TablePartitionModel `tfsdk:"partitions" json:"partitions,omitempty"`
	RetentionPolicy *model.TableRetentionModel  `tfsdk:"retention_policy" json:"retention_policy,omitempty"`
	BetaFeatures    []string                    `tfsdk:"beta_features" json:"beta_features,omitempty"`
	SkipBackfill    *bool                       `tfsdk:"skip_backfill" json:"skip_backfill,omitempty"`
}

func (d *tableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (d *tableDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	skipRecords := schema.Int64Attribute{
		Description:         "skip the first N records (useful when headers are used).",
		MarkdownDescription: "skip the first *N* records (useful when headers are used).",
		Computed:            true,
	}
	missingValues := schema.ListAttribute{
		Description: "list of values that represent a missing value.",
		Computed:    true,
		ElementType: types.StringType,
	}
	fields := schema.ListNestedAttribute{
		Description: "specify hints for each field.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "Field-name (use dots to make it a subfield)",
					Computed:    true,
				},
				"type": schema.StringAttribute{
					Description: "Type of field (or ignore)",
					Computed:    true,
				},
				"default": schema.StringAttribute{
					Description: "Default value if the column is an empty string",
					Computed:    true,
				},
				"format": schema.StringAttribute{
					Description: "Ingestion format (i.e. different data formats)",
					Computed:    true,
				},
				"allow_empty": schema.BoolAttribute{
					Description: "Allow empty values (only valid for strings) to be ingested. If flag is set to false, then the field won't be written for the record instead.",
					Optional:    true,
					Computed:    true,
				},
				"no_index": schema.BoolAttribute{
					Description: "ADon't use sparse-indexing for this value (only valid for date-time type).",
					Optional:    true,
					Computed:    true,
				},
				"true_values": schema.ListAttribute{
					Description: "Optional list of values that represent TRUE (only valid for bool type).",
					Optional:    true,
					ElementType: types.StringType,
				},
				"false_values": schema.ListAttribute{
					Description: "Optional list of values that represent FALSE (only valid for bool type).",
					Optional:    true,
					ElementType: types.StringType,
				},
				"missing_values": schema.ListAttribute{
					Description: "Optional list of values that represents a missing value.",
					Optional:    true,
					ElementType: types.StringType,
				},
			},
		},
	}

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
						"json_hints": schema.ListNestedAttribute{
							Description: "Ingestion hints for JSON input.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"field": schema.StringAttribute{
										Description: "Field name.",
										Required:    true,
									},
									"hints": schema.ListAttribute{
										Description: "Hints.",
										Required:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"csv_hints": schema.SingleNestedAttribute{
							Description: "Ingestion hints for CSV input.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"separator": schema.StringAttribute{
									Description:         "specify a custom separator (defaults to ',').",
									MarkdownDescription: "specify a custom separator (defaults to `,`).",
									Computed:            true,
								},
								"skip_records":   skipRecords,
								"missing_values": missingValues,
								"fields":         fields,
							},
						},
						"tsv_hints": schema.SingleNestedAttribute{
							Description: "Ingestion hints for TSV input.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"skip_records":   skipRecords,
								"missing_values": missingValues,
								"fields":         fields,
							},
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
			"retention_policy": schema.SingleNestedAttribute{
				Description: "Synthetic field that is generated from parts of an input URI and used to partition table data.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"field": schema.StringAttribute{
						Description: "Path expression for the field used to determine the age of a record for the purpose of the data retention policy. Currently only timestamp fields are supported.",
						Computed:    true,
					},
					"valid_for": schema.StringAttribute{
						Description:         "ValidFor is the validity window relative to now. This is a string with a format like '<n>y<n>m<n>d' where '<n>' is a number and any component can be omitted.",
						MarkdownDescription: "ValidFor is the validity window relative to now. This is a string with a format like `<n>y<n>m<n>d` where `<n>` is a number and any component can be omitted.",
						Computed:            true,
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
	d.client = req.ProviderData.(*api.Client)
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

	database, table := data.Database.ValueString(), *data.Table
	tableDescription, err := d.client.Table(ctx, region, database, table)
	if err != nil {
		if err == api.ErrNotFound {
			resp.Diagnostics.AddError(
				"Table not found",
				fmt.Sprintf("Table %q not found in database %q (region %s)", table, database, region),
			)
		} else {
			resp.Diagnostics.AddError(
				"Cannot get table configuration",
				fmt.Sprintf("Unable to get tenant table configuration of table %s:%s (region %s): %v", database, table, region, err.Error()),
			)
		}
		return
	}

	err = json.Unmarshal(tableDescription, &data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot decode table configuration",
			fmt.Sprintf("Unable to decode tenant table configuration of table %s:%s in region %s: %v", database, table, region, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", tenantInfo.TenantID, region, database, table))
	data.Region = types.StringValue(region)
	data.Location = types.StringValue(fmt.Sprintf("%s/db/%s/%s/", tenantInfo.Regions[region].Bucket, database, table))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
