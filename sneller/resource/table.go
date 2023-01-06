package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"terraform-provider-sneller/sneller/api"
	"terraform-provider-sneller/sneller/model"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTableResource() resource.Resource {
	return &tableResource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &tableResource{}
	_ resource.ResourceWithConfigure   = &tableResource{}
	_ resource.ResourceWithImportState = &tableResource{}
)

type tableResource struct {
	client *api.Client
}

type tableResourceModel struct {
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

func (r *tableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (r *tableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	skipRecords := schema.Int64Attribute{
		Description:         "skip the first N records (useful when headers are used).",
		MarkdownDescription: "skip the first *N* records (useful when headers are used).",
		Optional:            true,
	}
	missingValues := schema.ListAttribute{
		Description: "list of values that represent a missing value.",
		Optional:    true,
		ElementType: types.StringType,
	}
	fields := schema.ListNestedAttribute{
		Description: "specify hints for each field.",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "Field-name (use dots to make it a subfield)",
					Required:    true,
				},
				"type": schema.StringAttribute{
					Description: "Type of field (or ignore)",
					Optional:    true,
				},
				"default": schema.StringAttribute{
					Description: "Default value if the column is an empty string",
					Optional:    true,
				},
				"format": schema.StringAttribute{
					Description: "Ingestion format (i.e. different data formats)",
					Optional:    true,
				},
				"allow_empty": schema.BoolAttribute{
					Description:   "Allow empty values (only valid for strings) to be ingested. If flag is set to false, then the field won't be written for the record instead.",
					Optional:      true,
					Computed:      true,
					PlanModifiers: []planmodifier.Bool{BoolDefaultValue(false)},
				},
				"no_index": schema.BoolAttribute{
					Description:   "ADon't use sparse-indexing for this value (only valid for date-time type).",
					Optional:      true,
					Computed:      true,
					PlanModifiers: []planmodifier.Bool{BoolDefaultValue(false)},
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
		Description: "Configure a Sneller table.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Terraform identifier.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"region": schema.StringAttribute{
				Description: "Region where the table should be created. If not set, then the table is created in the tenant's home region.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"database": schema.StringAttribute{
				Description:   "Database name.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"table": schema.StringAttribute{
				Description:   "Table name.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"location": schema.StringAttribute{
				Description:         "S3 url of the database location (i.e. `s3://sneller-cache-bucket/db/test-db/test-table/`).",
				MarkdownDescription: "S3 url of the database location (i.e. `s3://sneller-cache-bucket/db/test-db/test-table/`).",
				Computed:            true,
			},
			"inputs": schema.ListNestedAttribute{
				Description: "The input definition specifies where the source data is located and it format.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"pattern": schema.StringAttribute{
							Description:         "Pattern definition to specify the source pattern (i.e. 's3://sneller-source-bucket/data/*.ndjson').",
							MarkdownDescription: "Pattern definition to specify the source pattern (i.e. `s3://sneller-source-bucket/data/*.ndjson`).",
							Required:            true,
						},
						"format": schema.StringAttribute{
							Description:         fmt.Sprintf("Format of the input data ('%s').", strings.Join(api.Formats, "', '")),
							MarkdownDescription: fmt.Sprintf("Format of the input data (`%s`).", strings.Join(api.Formats, "`, `")),
							Required:            true,
							Validators:          []validator.String{stringvalidator.OneOf(api.Formats...)},
						},
						"json_hints": schema.ListNestedAttribute{
							Description: "Ingestion hints for JSON input.",
							Optional:    true,
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
						"csv_hints": schema.MapNestedAttribute{
							Description: "Ingestion hints for CSV input.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"separator": schema.StringAttribute{
										Description:         "specify a custom separator (defaults to ',').",
										MarkdownDescription: "specify a custom separator (defaults to `,`).",
										Optional:            true,
										Validators:          []validator.String{stringvalidator.LengthBetween(1, 1)},
									},
									"skip_records":   skipRecords,
									"missing_values": missingValues,
									"fields":         fields,
								},
							},
							// TODO: Only allows for CSV format
						},
						"tsv_hints": schema.MapNestedAttribute{
							Description: "Ingestion hints for TSV input.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"skip_records":   skipRecords,
									"missing_values": missingValues,
									"fields":         fields,
								},
							},
							// TODO: Only allows for TSV format
						},
					},
				},
			},
			"partitions": schema.ListNestedAttribute{
				Description: "Synthetic field that is generated from parts of an input URI and used to partition table data..",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Description: "Name of the partition field. If this field conflicts with a field in the input data, the partition field will override it.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "Type of the partition field.",
							Optional:    true,
						},
						"value": schema.StringAttribute{
							Description: "Template string that is used to produce the value for the partition field. If this is empty (or not set), the field name is used to determine the input URI part that will be used to determine the value.",
							Optional:    true,
						},
					},
				},
			},
			"retention_policy": schema.SingleNestedAttribute{
				Description: "Synthetic field that is generated from parts of an input URI and used to partition table data.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"field": schema.StringAttribute{
						Description: "Path expression for the field used to determine the age of a record for the purpose of the data retention policy. Currently only timestamp fields are supported.",
						Required:    true,
					},
					"valid_for": schema.StringAttribute{
						Description:         "ValidFor is the validity window relative to now. This is a string with a format like '<n>y<n>m<n>d' where '<n>' is a number and any component can be omitted.",
						MarkdownDescription: "ValidFor is the validity window relative to now. This is a string with a format like `<n>y<n>m<n>d` where `<n>` is a number and any component can be omitted.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile("([0-9]+y)?([0-9]+m)?([0-9]+d)?"), "value must be '<n>y<n>m<n>d'  where '<n>' is a number and any component can be omitted."),
							stringvalidator.LengthAtLeast(2),
						},
					},
				},
			},
			"beta_features": schema.ListAttribute{
				Description: "List of feature flags that can be used to turn on features for beta-testing.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"skip_backfill": schema.BoolAttribute{
				Description:   "Skip scanning the source bucket(s) for matching objects when the first objects are inserted into the table.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{BoolDefaultValue(false)},
			},
		},
	}
}

func (r *tableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*api.Client)
}

func (r *tableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var data tableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), "/")
	if len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Cannot parse ID",
			fmt.Sprintf("Invalid ID %q", data.ID),
		)
		return
	}
	tenantID := parts[0]
	region := parts[1]
	database := parts[2]
	table := parts[3]

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

	tableDescription, err := r.client.Table(ctx, region, database, table)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get table configuration",
			fmt.Sprintf("Unable to get tenant table configuration of table %s:%s in region %s: %v", database, table, region, err.Error()),
		)
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
	data.Database = types.StringValue(database)
	data.Table = &table
	data.Location = types.StringValue(fmt.Sprintf("%s/db/%s/%s/", tenantInfo.Regions[region].Bucket, database, table))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *tableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data tableResourceModel

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

	database, table := data.Database.ValueString(), *data.Table
	if err = r.writeTable(ctx, data, region, database, table, resp.Diagnostics); err != nil {
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", tenantInfo.TenantID, region, database, table))
	data.Region = types.StringValue(region)
	data.Location = types.StringValue(fmt.Sprintf("%s/db/%s/%s/", tenantInfo.Regions[region].Bucket, database, table))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *tableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data tableResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), "/")
	if len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Cannot parse ID",
			fmt.Sprintf("Invalid ID %q", data.ID),
		)
		return
	}
	tenantID := parts[0]
	region := parts[1]
	database := parts[2]
	table := parts[3]

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

	if err = r.writeTable(ctx, data, region, database, table, resp.Diagnostics); err != nil {
		return
	}

	data.Location = types.StringValue(fmt.Sprintf("%s/db/%s/%s/", tenantInfo.Regions[region].Bucket, database, table))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *tableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data tableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.ID.ValueString(), "/")
	if len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Cannot parse ID",
			fmt.Sprintf("Invalid ID %q", data.ID),
		)
		return
	}
	tenantID := parts[0]
	region := parts[1]
	database := parts[2]
	table := parts[3]

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

	err = r.client.DeleteTable(ctx, region, database, table, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot delete table",
			fmt.Sprintf("Unable to delete table for %s:%s in region %s: %v", database, table, region, err.Error()),
		)
		return
	}
}

func (r *tableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *tableResource) writeTable(ctx context.Context, data tableResourceModel, region, database, table string, diags diag.Diagnostics) error {
	copy := data
	if copy.SkipBackfill != nil && !*copy.SkipBackfill {
		// this prevents writing the `false` value
		copy.SkipBackfill = nil
	}
	tableBytes, err := json.Marshal(&copy)
	if err != nil {
		diags.AddError(
			"Cannot encode table configuration",
			fmt.Sprintf("Unable to encode table configuration in region %s: %v", region, err.Error()),
		)
		return err
	}

	err = r.client.SetTable(ctx, region, database, table, tableBytes)
	if err != nil {
		diags.AddError(
			"Cannot create table",
			fmt.Sprintf("Unable to create table %s/%s in region %s: %v", database, table, region, err.Error()),
		)
		return err
	}

	return nil
}
