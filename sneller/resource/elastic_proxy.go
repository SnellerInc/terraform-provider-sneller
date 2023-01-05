package resource

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-sneller/sneller/api"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewElasticProxyResource() resource.Resource {
	return &elasticProxyResource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &elasticProxyResource{}
	_ resource.ResourceWithConfigure   = &elasticProxyResource{}
	_ resource.ResourceWithImportState = &elasticProxyResource{}
)

type elasticProxyResource struct {
	client *api.Client
}

type elasticProxyResourceModel struct {
	ID          types.String                              `tfsdk:"id"`
	LastUpdated types.String                              `tfsdk:"last_updated"`
	Region      types.String                              `tfsdk:"region"`
	Location    types.String                              `tfsdk:"location"`
	LogPath     types.String                              `tfsdk:"log_path"`
	LogFlags    *elasticProxyLogFlagsResourceModel        `tfsdk:"log_flags"`
	Index       map[string]elasticProxyIndexResourceModel `tfsdk:"index"`
}

type elasticProxyLogFlagsResourceModel struct {
	LogRequest         types.Bool `tfsdk:"log_request"`
	LogQueryParameters types.Bool `tfsdk:"log_query_parameters"`
	LogSQL             types.Bool `tfsdk:"log_sql"`
	LogSnellerResult   types.Bool `tfsdk:"log_sneller_result"`
	LogPreprocessed    types.Bool `tfsdk:"log_preprocessed"`
	LogResult          types.Bool `tfsdk:"log_result"`
}

type elasticProxyIndexResourceModel struct {
	Database        types.String                                    `tfsdk:"database"`
	Table           types.String                                    `tfsdk:"table"`
	IgnoreTotalHits types.Bool                                      `tfsdk:"ignore_total_hits"`
	TypeMapping     map[string]elasticProxyTypeMappingResourceModel `tfsdk:"type_mapping"`
}

type elasticProxyTypeMappingResourceModel struct {
	Type   types.String            `tfsdk:"type"`
	Fields map[string]types.String `tfsdk:"fields"`
}

func (r *elasticProxyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elastic_proxy"
}

func (r *elasticProxyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configure the Elastic proxy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Terraform identifier.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description:   "Region for which to configure the Elastic Proxy.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Location of the Elastic proxy configuration file (i.e. `s3://sneller-cache-bucket/db/elastic-proxy.json`).",
				Description:         "Location of the Elastic proxy configuration file (i.e. 's3://sneller-cache-bucket/db/elastic-proxy.json').",
				Computed:            true,
			},
			"log_path": schema.StringAttribute{
				MarkdownDescription: "Location where Elastic Proxy logging is stored (i.e. `s3://logging-bucket/elastic-proxy/`).",
				Description:         "Location where Elastic Proxy logging is stored (i.e. 's3://logging-bucket/elastic-proxy/').",
				Optional:            true,
			},
			"log_flags": schema.SingleNestedAttribute{
				Description: "Logging flags",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"log_request": schema.BoolAttribute{
						Description: "Log all requests.",
						Optional:    true,
					},
					"log_query_parameters": schema.BoolAttribute{
						Description: "Log query parameters.",
						Optional:    true,
					},
					"log_sql": schema.BoolAttribute{
						Description: "Log generated SQL query.",
						Optional:    true,
					},
					"log_sneller_result": schema.BoolAttribute{
						Description: "Log Sneller query result (may be verbose and contain sensitive data).",
						Optional:    true,
					},
					"log_preprocessed": schema.BoolAttribute{
						Description: "Log preprocessed Sneller results (may be verbose and contain sensitive data).",
						Optional:    true,
					},
					"log_result": schema.BoolAttribute{
						Description: "Log result (may be verbose and contain sensitive data).",
						Optional:    true,
					},
				},
			},
			"index": schema.MapNestedAttribute{
				Description: "Configures an Elastic index that maps to a Sneller table.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"database": schema.StringAttribute{
							Description: "Sneller database.",
							Required:    true,
						},
						"table": schema.StringAttribute{
							Description: "Sneller table.",
							Required:    true,
						},
						"ignore_total_hits": schema.BoolAttribute{
							Description:   "Ignore 'total_hits' in Elastic response (more efficient).",
							Optional:      true,
							Computed:      true,
							PlanModifiers: []planmodifier.Bool{BoolDefaultValue(false)},
						},
						"type_mapping": schema.MapNestedAttribute{
							Description: "Custom type mappings.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type.",
										Required:    true,
									},
									"fields": schema.MapAttribute{
										Description: "Field mappings.",
										Optional:    true,
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

func (r *elasticProxyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*api.Client)
}

func (r *elasticProxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var data elasticProxyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Cannot parse ID",
			fmt.Sprintf("Invalid ID %q", data.ID.ValueString()),
		)
		return
	}
	tenantID := parts[0]
	region := parts[1]

	tenantInfo, err := r.client.Tenant(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info in region %s: %v", region, err.Error()),
		)
		return
	}
	if tenantInfo.TenantID != tenantID {
		resp.Diagnostics.AddError(
			"Invalid tenant",
			fmt.Sprintf("Expected tenant %s, but got %s", tenantID, tenantInfo.TenantID),
		)
		return
	}

	tenantInfo, err = r.client.Tenant(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info in region %s: %v", region, err.Error()),
		)
		return
	}

	elasticProxyPath := fmt.Sprintf("%s/%selastic-proxy.json", tenantInfo.Regions[region].Bucket, api.DefaultDbPrefix)
	config, err := r.client.ElasticProxyConfig(ctx, region)
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
		data.LogFlags = &elasticProxyLogFlagsResourceModel{
			LogRequest:         types.BoolValue(config.LogFlags.LogRequest),
			LogQueryParameters: types.BoolValue(config.LogFlags.LogQueryParameters),
			LogSQL:             types.BoolValue(config.LogFlags.LogSQL),
			LogSnellerResult:   types.BoolValue(config.LogFlags.LogSnellerResult),
			LogPreprocessed:    types.BoolValue(config.LogFlags.LogPreprocessed),
			LogResult:          types.BoolValue(config.LogFlags.LogResult),
		}
	}

	if len(config.Mapping) > 0 {
		data.Index = make(map[string]elasticProxyIndexResourceModel, len(config.Mapping))
		for index, config := range config.Mapping {
			mapping := elasticProxyIndexResourceModel{
				Database:        types.StringValue(config.Database),
				Table:           types.StringValue(config.Table),
				IgnoreTotalHits: types.BoolValue(config.IgnoreTotalHits),
			}
			if len(config.TypeMapping) > 0 {
				mapping.TypeMapping = make(map[string]elasticProxyTypeMappingResourceModel, len(config.TypeMapping))
				for tm, config := range config.TypeMapping {
					typeMapping := elasticProxyTypeMappingResourceModel{
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

func (r *elasticProxyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data elasticProxyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := data.Region.ValueString()
	tenantInfo, err := r.client.Tenant(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info in region %s: %v", region, err.Error()),
		)
		return
	}

	if region == "" {
		region = tenantInfo.HomeRegion
	}

	config := elasticProxyConfigFromData(data)

	err = r.client.SetElasticProxyConfig(ctx, region, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot create elastic proxy configuration",
			fmt.Sprintf("Unable to create elastic proxy configuration in region %s: %v", region, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", tenantInfo.TenantID, region))
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	data.Location = types.StringValue(fmt.Sprintf("%s/%selastic-proxy.json", tenantInfo.Regions[region].Bucket, api.DefaultDbPrefix))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *elasticProxyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data elasticProxyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Cannot parse ID",
			fmt.Sprintf("Invalid ID %q", data.ID.ValueString()),
		)
		return
	}
	tenantID := parts[0]
	region := parts[1]

	tenantInfo, err := r.client.Tenant(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info in region %s: %v", region, err.Error()),
		)
		return
	}
	if tenantInfo.TenantID != tenantID {
		resp.Diagnostics.AddError(
			"Invalid tenant",
			fmt.Sprintf("Expected tenant %s, but got %s", tenantID, tenantInfo.TenantID),
		)
		return
	}

	config := elasticProxyConfigFromData(data)

	err = r.client.SetElasticProxyConfig(ctx, region, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot update elastic proxy configuration",
			fmt.Sprintf("Unable to update elastic proxy configuration in region %s: %v", region, err.Error()),
		)
		return
	}

	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	data.Location = types.StringValue(fmt.Sprintf("%s/%selastic-proxy.json", tenantInfo.Regions[region].Bucket, api.DefaultDbPrefix))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *elasticProxyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data elasticProxyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Cannot parse ID",
			fmt.Sprintf("Invalid ID %q", data.ID.ValueString()),
		)
		return
	}
	tenantID := parts[0]
	region := parts[1]

	tenantInfo, err := r.client.Tenant(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info in region %s: %v", region, err.Error()),
		)
		return
	}
	if tenantInfo.TenantID != tenantID {
		resp.Diagnostics.AddError(
			"Invalid tenant",
			fmt.Sprintf("Expected tenant %s, but got %s", tenantID, tenantInfo.TenantID),
		)
		return
	}

	err = r.client.DeleteElasticProxyConfig(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot delete elastic proxy configuration",
			fmt.Sprintf("Unable to delete elastic proxy configuration in region %s: %v", region, err.Error()),
		)
		return
	}
}

func (r *elasticProxyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func elasticProxyConfigFromData(data elasticProxyResourceModel) api.ElasticProxyConfig {
	elasticProxyConfig := api.ElasticProxyConfig{
		LogPath: data.LogPath.ValueString(),
		Mapping: make(map[string]api.ElasticProxyMappingConfig, len(data.Index)),
	}
	if data.LogFlags != nil {
		elasticProxyConfig.LogFlags = &api.ElasticProxyLogFlagsConfig{
			LogRequest:         data.LogFlags.LogRequest.ValueBool(),
			LogQueryParameters: data.LogFlags.LogQueryParameters.ValueBool(),
			LogSQL:             data.LogFlags.LogSQL.ValueBool(),
			LogSnellerResult:   data.LogFlags.LogSnellerResult.ValueBool(),
			LogPreprocessed:    data.LogFlags.LogPreprocessed.ValueBool(),
			LogResult:          data.LogFlags.LogResult.ValueBool(),
		}
	}
	if data.Index != nil {
		for index, mapping := range data.Index {
			mappingConfig := api.ElasticProxyMappingConfig{
				Database:        mapping.Database.ValueString(),
				Table:           mapping.Table.ValueString(),
				IgnoreTotalHits: mapping.IgnoreTotalHits.ValueBool(),
				TypeMapping:     make(map[string]api.ElasticProxyTypeMapping, len(mapping.TypeMapping)),
			}
			if mapping.TypeMapping != nil {
				for typ, config := range mapping.TypeMapping {
					typeMappingConfig := api.ElasticProxyTypeMapping{
						Type:   config.Type.ValueString(),
						Fields: make(map[string]string, len(config.Fields)),
					}
					if config.Fields != nil {
						for f, typ := range config.Fields {
							typeMappingConfig.Fields[f] = typ.ValueString()
						}
					}
					mappingConfig.TypeMapping[typ] = typeMappingConfig
				}
			}
			elasticProxyConfig.Mapping[index] = mappingConfig
		}
	}
	return elasticProxyConfig
}
