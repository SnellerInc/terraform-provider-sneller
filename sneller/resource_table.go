package sneller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	client *Client
}

type tableResourceModel struct {
	ID           types.String                  `tfsdk:"id"`
	LastUpdated  types.String                  `tfsdk:"last_updated"`
	Region       types.String                  `tfsdk:"region"`
	Database     types.String                  `tfsdk:"database"`
	Table        types.String                  `tfsdk:"table"`
	Location     types.String                  `tfsdk:"location"`
	Inputs       []tableInputResourceModel     `tfsdk:"inputs"`
	Partitions   []tablePartitionResourceModel `tfsdk:"partitions"`
	BetaFeatures []string                      `tfsdk:"beta_features"`
	SkipBackfill types.Bool                    `tfsdk:"skip_backfill"`
}

type tableInputResourceModel struct {
	Pattern string `tfsdk:"pattern"`
	Format  string `tfsdk:"format"`
}

type tablePartitionResourceModel struct {
	Field string `tfsdk:"field"`
	Type  string `tfsdk:"type"`
	Value string `tfsdk:"value"`
}

func (r *tableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (r *tableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configure a Sneller table.",
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
				Description:   "Region where the table should be created. If not set, then the table is created in the tenant's home region.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
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
							Description: "Pattern definition to specify the source pattern (i.e. `s3://sneller-source-bucket/data/*.ndjson`).",
							Required:    true,
						},
						"format": schema.StringAttribute{
							Description:         "Format of the input data.",
							MarkdownDescription: "Format of the input data (" + formats + ").",
							Required:            true,
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
							Description:   "Type of the partition field.",
							Optional:      true,
							Computed:      true,
							PlanModifiers: []planmodifier.String{StringDefaultValue("string")},
						},
						"value": schema.StringAttribute{
							Description: "Template string that is used to produce the value for the partition field. If this is empty (or not set), the field name is used to determine the input URI part that will be used to determine the value.",
							Optional:    true,
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
	r.client = req.ProviderData.(*Client)
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
			fmt.Sprintf("Invalid ID %q", data.ID.ValueString()),
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

	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", tenantInfo.TenantID, region, database, table))
	data.Region = types.StringValue(region)
	data.Database = types.StringValue(database)
	data.Table = types.StringValue(table)
	data.Location = types.StringValue(fmt.Sprintf("%s/db/%s/%s/", tenantInfo.Regions[region].Bucket, database, table))
	if len(tableDescription.Input) > 0 {
		data.Inputs = make([]tableInputResourceModel, 0, len(tableDescription.Input))
		for _, input := range tableDescription.Input {
			data.Inputs = append(data.Inputs, tableInputResourceModel(input))
		}
	}
	if len(tableDescription.Partitions) > 0 {
		data.Partitions = make([]tablePartitionResourceModel, 0, len(tableDescription.Partitions))
		for _, partition := range tableDescription.Partitions {
			data.Partitions = append(data.Partitions, tablePartitionResourceModel(partition))
		}
	}
	data.BetaFeatures = tableDescription.BetaFeatures
	data.SkipBackfill = types.BoolValue(tableDescription.SkipBackfill)

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

	var inputs []TableInput
	if len(data.Inputs) > 0 {
		inputs = make([]TableInput, 0, len(data.Inputs))
		for _, input := range data.Inputs {
			inputs = append(inputs, TableInput(input))
		}
	}

	var partitions []TablePartition
	if len(data.Partitions) > 0 {
		partitions = make([]TablePartition, 0, len(data.Partitions))
		for _, partition := range data.Partitions {
			partitions = append(partitions, TablePartition(partition))
		}
	}

	database, table := data.Database.ValueString(), data.Table.ValueString()

	err = r.client.SetTable(ctx, region, database, TableDescription{
		Name:         table,
		Input:        inputs,
		Partitions:   partitions,
		BetaFeatures: data.BetaFeatures,
		SkipBackfill: data.SkipBackfill.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot create table",
			fmt.Sprintf("Unable to create table %s/%s in region %s: %v", database, table, region, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", tenantInfo.TenantID, region, database, table))
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
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
			fmt.Sprintf("Invalid ID %q", data.ID.ValueString()),
		)
		return
	}
	tenantID := parts[0]
	region := parts[1]
	database := parts[2]
	table := parts[3]

	var inputs []TableInput
	if len(data.Inputs) > 0 {
		inputs = make([]TableInput, 0, len(data.Inputs))
		for _, input := range data.Inputs {
			inputs = append(inputs, TableInput(input))
		}
	}

	var partitions []TablePartition
	if len(data.Partitions) > 0 {
		partitions = make([]TablePartition, 0, len(data.Partitions))
		for _, partition := range data.Partitions {
			partitions = append(partitions, TablePartition(partition))
		}
	}

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

	err = r.client.SetTable(ctx, region, database, TableDescription{
		Name:         table,
		Input:        inputs,
		Partitions:   partitions,
		BetaFeatures: data.BetaFeatures,
		SkipBackfill: data.SkipBackfill.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot update table",
			fmt.Sprintf("Unable to update table %s/%s in region %s: %v", database, table, region, err.Error()),
		)
		return
	}

	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
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
			fmt.Sprintf("Invalid ID %q", data.ID.ValueString()),
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
