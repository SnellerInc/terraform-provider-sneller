package sneller

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewDatabaseDataSource() datasource.DataSource {
	return &databaseDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &databaseDataSource{}
	_ datasource.DataSourceWithConfigure = &databaseDataSource{}
)

type databaseDataSource struct {
	client *Client
}

type databaseDataSourceModel struct {
	Region   types.String `tfsdk:"region"`
	Database types.String `tfsdk:"database"`
	Location types.String `tfsdk:"location"`
	Tables   types.Set    `tfsdk:"tables"`
}

func (r *databaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *databaseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides configuration of the tenant.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional: true,
			},
			"database": schema.StringAttribute{
				Required: true,
			},
			"location": schema.StringAttribute{
				Computed: true,
			},
			"tables": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *databaseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *databaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data databaseDataSourceModel
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

	database := data.Database.ValueString()

	tableInfos, err := d.client.Database(ctx, region, database)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get databases",
			fmt.Sprintf("Unable to get databases in region %s: %v", region, err.Error()),
		)
		return
	}

	data.Region = types.StringValue(region)
	data.Location = types.StringValue(fmt.Sprintf("%s/db/%s/", tenantInfo.Regions[region].Bucket, database))

	tables := make([]string, 0, len(tableInfos))
	for _, ti := range tableInfos {
		tables = append(tables, ti.Name)
	}

	var diags diag.Diagnostics
	data.Tables, diags = types.SetValueFrom(ctx, types.StringType, tables)
	resp.Diagnostics.Append(diags...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
