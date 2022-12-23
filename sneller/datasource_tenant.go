package sneller

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTenantDataSource() datasource.DataSource {
	return &tenantDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &tenantDataSource{}
	_ datasource.DataSourceWithConfigure = &tenantDataSource{}
)

type tenantDataSource struct {
	client *Client
}

type tenantDataSourceModel struct {
	TenantID      types.String `tfsdk:"tenant_id"`
	State         types.String `tfsdk:"state"`
	Name          types.String `tfsdk:"name"`
	HomeRegion    types.String `tfsdk:"home_region"`
	Email         types.String `tfsdk:"email"`
	TenantRoleARN types.String `tfsdk:"tenant_role_arn"`
	Mfa           types.String `tfsdk:"mfa"`
	CreatedAt     types.String `tfsdk:"created_at"`
	ActivatedAt   types.String `tfsdk:"activated_at"`
	DeactivatedAt types.String `tfsdk:"deactivated_at"`
}

func (r *tenantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *tenantDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides configuration of the tenant.",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				Optional: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"home_region": schema.StringAttribute{
				Computed: true,
			},
			"email": schema.StringAttribute{
				Computed: true,
			},
			"tenant_role_arn": schema.StringAttribute{
				Computed: true,
			},
			"mfa": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"activated_at": schema.StringAttribute{
				Computed: true,
			},
			"deactivated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *tenantDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *tenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data tenantDataSourceModel
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

	data.TenantID = types.StringValue(tenantInfo.TenantID)
	data.State = types.StringValue(tenantInfo.TenantState)
	data.Name = types.StringValue(tenantInfo.TenantName)
	data.Email = types.StringValue(tenantInfo.Email)
	data.HomeRegion = types.StringValue(tenantInfo.HomeRegion)
	data.TenantRoleARN = types.StringValue(tenantInfo.TenantRoleArn)
	data.Mfa = types.StringValue(string(tenantInfo.Mfa))
	data.CreatedAt = types.StringValue(tenantInfo.CreatedAt.Format(time.RFC3339))
	if tenantInfo.ActivatedAt != nil {
		data.ActivatedAt = types.StringValue(tenantInfo.ActivatedAt.Format(time.RFC3339))
	}
	if tenantInfo.DeactivatedAt != nil {
		data.DeactivatedAt = types.StringValue(tenantInfo.DeactivatedAt.Format(time.RFC3339))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
