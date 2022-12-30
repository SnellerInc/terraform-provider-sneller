package sneller

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewDatabasesDataSource() datasource.DataSource {
	return &databasesDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &databasesDataSource{}
	_ datasource.DataSourceWithConfigure = &databasesDataSource{}
)

type databasesDataSource struct {
	client *Client
}

type databasesDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Region    types.String `tfsdk:"region"`
	Location  types.String `tfsdk:"location"`
	Databases types.Set    `tfsdk:"databases"`
}

func (d *databasesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databases"
}

func (d *databasesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides all databases.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform identifier.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region from which to fetch the databases. When not set, then it default's to the tenant's home region.",
				Optional:    true,
			},
			"location": schema.StringAttribute{
				Description:         "S3 url where the databases are stored (i.e. `s3://sneller-cache-bucket/db/`).",
				MarkdownDescription: "S3 url where the databases are stored (i.e. `s3://sneller-cache-bucket/db/`).",
				Computed:            true,
			},
			"databases": schema.SetAttribute{
				Description: "Set of databases in the specified region.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *databasesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *databasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data databasesDataSourceModel
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
			fmt.Sprintf("Unable to get tenant info: %v", err.Error()),
		)
		return
	}
	tenantRegionInfo := tenantInfo.Regions[region]

	databases, err := d.client.Databases(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get databases",
			fmt.Sprintf("Unable to get databases in region %s: %v", region, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", tenantInfo.TenantID, region))
	data.Region = types.StringValue(region)
	data.Location = types.StringValue(fmt.Sprintf("%s/%s", tenantRegionInfo.Bucket, defaultDbPrefix))
	var diags diag.Diagnostics
	data.Databases, diags = types.SetValueFrom(ctx, types.StringType, databases)
	resp.Diagnostics.Append(diags...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
