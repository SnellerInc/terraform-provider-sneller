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
	Region    types.String `tfsdk:"region"`
	Databases types.Set    `tfsdk:"databases"`
}

func (r *databasesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databases"
}

func (r *databasesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides configuration of the tenant.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional: true,
			},
			"databases": schema.SetAttribute{
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

	databases, err := d.client.Databases(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get databases",
			fmt.Sprintf("Unable to get databases in region %s: %v", region, err.Error()),
		)
		return
	}

	data.Region = types.StringValue(region)
	var diags diag.Diagnostics
	data.Databases, diags = types.SetValueFrom(ctx, types.StringType, databases)
	resp.Diagnostics.Append(diags...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
