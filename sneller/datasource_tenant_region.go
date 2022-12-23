package sneller

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTenantRegionDataSource() datasource.DataSource {
	return &tenantRegionDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &tenantRegionDataSource{}
	_ datasource.DataSourceWithConfigure = &tenantRegionDataSource{}
)

const defaultDbPrefix = "db/"

type tenantRegionDataSource struct {
	client *Client
}

type tenantRegionDataSourceModel struct {
	Region     types.String `tfsdk:"region"`
	Bucket     types.String `tfsdk:"bucket"`
	Prefix     types.String `tfsdk:"prefix"`
	RoleARN    types.String `tfsdk:"role_arn"`
	ExternalID types.String `tfsdk:"external_id"`
	SqsARN     types.String `tfsdk:"sqs_arn"`
}

func (r *tenantRegionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_region"
}

func (r *tenantRegionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides configuration of the tenant.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional: true,
			},
			"bucket": schema.StringAttribute{
				Computed: true,
			},
			"prefix": schema.StringAttribute{
				Computed: true,
			},
			"role_arn": schema.StringAttribute{
				Computed: true,
			},
			"external_id": schema.StringAttribute{
				Computed: true,
			},
			"sqs_arn": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *tenantRegionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *tenantRegionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data tenantRegionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := data.Region.ValueString()
	if region == "" {
		tenantInfo, err := d.client.Tenant(ctx, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Cannot get tenant info",
				fmt.Sprintf("Unable to get tenant info: %v", err.Error()),
			)
			return
		}
		region = tenantInfo.HomeRegion
	}

	tenantInfo, err := d.client.Tenant(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info",
			fmt.Sprintf("Unable to get tenant info in region %s: %v", region, err.Error()),
		)
		return
	}

	tenantRegionInfo := tenantInfo.Regions[region]

	sqsARN := tenantRegionInfo.SqsArn
	if tenantRegionInfo.SqsArn == "" {
		// workaround for older API versions that
		// don't expose the SQS queue ARN
		tenantRoleARN, err := arn.Parse(tenantInfo.TenantRoleArn)
		if err == nil {
			sqsARN = fmt.Sprintf("arn:aws:sqs:%s:%s:tenant-sdb-%s", region, tenantRoleARN.AccountID, tenantInfo.TenantID)
		}
	}

	data.Region = types.StringValue(region)
	data.Bucket = types.StringValue(strings.TrimPrefix(tenantRegionInfo.Bucket, "s3://"))
	data.Prefix = types.StringValue(defaultDbPrefix)
	data.RoleARN = types.StringValue(tenantRegionInfo.RegionRoleArn)
	data.ExternalID = types.StringValue(tenantRegionInfo.RegionExternalID)
	data.SqsARN = types.StringValue(sqsARN)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
