package sneller

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewElasticProxyDataSource() datasource.DataSource {
	return &elasticProxyDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &elasticProxyDataSource{}
	_ datasource.DataSourceWithConfigure = &elasticProxyDataSource{}
)

type elasticProxyDataSource struct {
	client *Client
}

type elasticProxyDataSourceModel struct {
	ID       types.String                                `tfsdk:"id"`
	Region   types.String                                `tfsdk:"region"`
	Location types.String                                `tfsdk:"location"`
	LogPath  types.String                                `tfsdk:"log_path"`
	LogFlags *elasticProxyLogFlagsDataSourceModel        `tfsdk:"log_flags"`
	Index    map[string]elasticProxyIndexDataSourceModel `tfsdk:"index"`
}

type elasticProxyLogFlagsDataSourceModel struct {
	LogRequest         types.Bool `tfsdk:"log_request"`
	LogQueryParameters types.Bool `tfsdk:"log_query_parameters"`
	LogSQL             types.Bool `tfsdk:"log_sql"`
	LogSnellerResult   types.Bool `tfsdk:"log_sneller_result"`
	LogPreprocessed    types.Bool `tfsdk:"log_preprocessed"`
	LogResult          types.Bool `tfsdk:"log_result"`
}

type elasticProxyIndexDataSourceModel struct {
	Database        types.String                                      `tfsdk:"database"`
	Table           types.String                                      `tfsdk:"table"`
	IgnoreTotalHits types.Bool                                        `tfsdk:"ignore_total_hits"`
	TypeMapping     map[string]elasticProxyTypeMappingDataSourceModel `tfsdk:"type_mapping"`
}

type elasticProxyTypeMappingDataSourceModel struct {
	Type   types.String            `tfsdk:"type"`
	Fields map[string]types.String `tfsdk:"fields"`
}

func (d *elasticProxyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_proxy"
}

func (d *elasticProxyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Elastic proxy configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform identifier.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region for which to obtain the Elastic Proxy configuration.",
				Required:    true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Location of the Elastic proxy configuration file (i.e. `s3://sneller-cache-bucket/db/elastic-proxy.json`).",
				Description:         "Location of the Elastic proxy configuration file (i.e. 's3://sneller-cache-bucket/db/elastic-proxy.json').",
				Computed:            true,
			},
			"log_path": schema.StringAttribute{
				MarkdownDescription: "Location where Elastic Proxy logging is stored (i.e. `s3://logging-bucket/elastic-proxy/`).",
				Description:         "Location where Elastic Proxy logging is stored (i.e. 's3://logging-bucket/elastic-proxy/').",
				Computed:            true,
			},
			"log_flags": schema.SingleNestedAttribute{
				Description: "Logging flags",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"log_request": schema.BoolAttribute{
						Description: "Log all requests.",
						Computed:    true,
					},
					"log_query_parameters": schema.BoolAttribute{
						Description: "Log query parameters.",
						Computed:    true,
					},
					"log_sql": schema.BoolAttribute{
						Description: "Log generated SQL query.",
						Computed:    true,
					},
					"log_sneller_result": schema.BoolAttribute{
						Description: "Log Sneller query result (may be verbose and contain sensitive data).",
						Computed:    true,
					},
					"log_preprocessed": schema.BoolAttribute{
						Description: "Log preprocessed Sneller results (may be verbose and contain sensitive data).",
						Computed:    true,
					},
					"log_result": schema.BoolAttribute{
						Description: "Log result (may be verbose and contain sensitive data).",
						Computed:    true,
					},
				},
			},
			"index": schema.MapNestedAttribute{
				Description: "Configures an Elastic index that maps to a Sneller table.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"database": schema.StringAttribute{
							Description: "Sneller database.",
							Computed:    true,
						},
						"table": schema.StringAttribute{
							Description: "Sneller table.",
							Computed:    true,
						},
						"ignore_total_hits": schema.BoolAttribute{
							Description: "Ignore 'total_hits' in Elastic response (more efficient).",
							Computed:    true,
						},
						"type_mapping": schema.MapNestedAttribute{
							Description: "Custom type mappings.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type.",
										Computed:    true,
									},
									"fields": schema.MapAttribute{
										Description: "Field mappings.",
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *elasticProxyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *elasticProxyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data elasticProxyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantInfo, err := d.client.Tenant(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info: %v", err.Error()),
		)
		return
	}

	region := data.Region.ValueString()
	if region == "" {
		region = tenantInfo.HomeRegion
	}

	tenantInfo, err = d.client.Tenant(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info in region %s: %v", region, err.Error()),
		)
		return
	}

	elasticProxyPath := fmt.Sprintf("%s/%selastic-proxy.json", tenantInfo.Regions[region].Bucket, defaultDbPrefix)
	config, err := d.client.ElasticProxyConfig(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get elastic-proxy configuration",
			fmt.Sprintf("Unable to get elastic-proxy configuration in region %s: %v", region, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", tenantInfo.TenantID, region))
	data.Region = types.StringValue(region)
	data.Location = types.StringValue(elasticProxyPath)
	data.LogPath = types.StringValue(config.LogPath)
	if config.LogFlags != nil {
		data.LogFlags = &elasticProxyLogFlagsDataSourceModel{
			LogRequest:         types.BoolValue(config.LogFlags.LogRequest),
			LogQueryParameters: types.BoolValue(config.LogFlags.LogQueryParameters),
			LogSQL:             types.BoolValue(config.LogFlags.LogSQL),
			LogSnellerResult:   types.BoolValue(config.LogFlags.LogSnellerResult),
			LogPreprocessed:    types.BoolValue(config.LogFlags.LogPreprocessed),
			LogResult:          types.BoolValue(config.LogFlags.LogResult),
		}
	}

	if len(config.Mapping) > 0 {
		data.Index = make(map[string]elasticProxyIndexDataSourceModel, len(config.Mapping))
		for index, config := range config.Mapping {
			mapping := elasticProxyIndexDataSourceModel{
				Database:        types.StringValue(config.Database),
				Table:           types.StringValue(config.Table),
				IgnoreTotalHits: types.BoolValue(config.IgnoreTotalHits),
			}
			if len(config.TypeMapping) > 0 {
				mapping.TypeMapping = make(map[string]elasticProxyTypeMappingDataSourceModel, len(config.TypeMapping))
				for tm, config := range config.TypeMapping {
					typeMapping := elasticProxyTypeMappingDataSourceModel{
						Type: types.StringValue(config.Type),
					}
					if len(config.Fields) > 0 {
						typeMapping.Fields = make(map[string]basetypes.StringValue, len(config.Fields))

						for f, config := range config.Fields {
							typeMapping.Fields[f] = types.StringValue(config)
						}
					}
					mapping.TypeMapping[tm] = typeMapping
				}
			}
			data.Index[index] = mapping
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
