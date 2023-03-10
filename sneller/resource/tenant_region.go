package resource

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-sneller/sneller/api"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTenantRegionResource() resource.Resource {
	return &tenantRegionResource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &tenantRegionResource{}
	_ resource.ResourceWithConfigure   = &tenantRegionResource{}
	_ resource.ResourceWithImportState = &tenantRegionResource{}
)

type tenantRegionResource struct {
	client *api.Client
}

type tenantRegionResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Region                types.String `tfsdk:"region"`
	Bucket                types.String `tfsdk:"bucket"`
	Prefix                types.String `tfsdk:"prefix"`
	RoleARN               types.String `tfsdk:"role_arn"`
	ExternalID            types.String `tfsdk:"external_id"`
	MaxScanBytes          types.Int64  `tfsdk:"max_scan_bytes"`
	EffectiveMaxScanBytes types.Int64  `tfsdk:"effective_max_scan_bytes"`
	SqsARN                types.String `tfsdk:"sqs_arn"`
}

func (r *tenantRegionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_region"
}

func (r *tenantRegionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configure a tenant's regional configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Terraform identifier.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"region": schema.StringAttribute{
				Description: "Region from which to fetch the tenant configuration. When not set, then it default's to the tenant's home region.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bucket": schema.StringAttribute{
				Description: "Sneller cache bucket name.",
				Required:    true,
			},
			"prefix": schema.StringAttribute{
				Description:   "Prefix of the files in the Sneller cache bucket (always 'db/').",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"role_arn": schema.StringAttribute{
				Description: "ARN of the role that is used to access the S3 data in this region's cache bucket. It is also used by the ingestion process to read the source data.",
				Required:    true,
			},
			"external_id": schema.StringAttribute{
				Description:   "External ID (typically the same as the tenant ID) that is passed when assuming the IAM role",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"max_scan_bytes": schema.Int64Attribute{
				Description: "Maximum number of bytes scanned per query",
				Optional:    true,
			},
			"effective_max_scan_bytes": schema.Int64Attribute{
				Description: "Effective maximum number of bytes scanned per query",
				Computed:    true,
			},
			"sqs_arn": schema.StringAttribute{
				Description:   "ARN of the SQS resource that is used to signal the ingestion process when new data arrives.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *tenantRegionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*api.Client)
}

func (r *tenantRegionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var data tenantRegionResourceModel
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

	tenantRegionInfo := tenantInfo.Regions[region]

	sqsARN := tenantRegionInfo.SqsArn
	if tenantRegionInfo.SqsArn == "" {
		// workaround for older API versions that
		// don't expose the SQS queue ARN
		tenantRoleARN, err := arn.Parse(tenantInfo.TenantRoleArn)
		if err == nil {
			sqsARN = fmt.Sprintf("arn:aws:sqs:%s:%s:tenant-sdb-%s", region, tenantRoleARN.AccountID, tenantInfo.TenantID)
		} else {
			sqsARN = "invalid"
		}
	}

	data.Region = types.StringValue(region)
	data.Bucket = types.StringValue(strings.TrimPrefix(tenantRegionInfo.Bucket, "s3://"))
	data.Prefix = types.StringValue(api.DefaultDbPrefix)
	data.RoleARN = types.StringValue(tenantRegionInfo.RegionRoleArn)
	data.ExternalID = types.StringValue(tenantRegionInfo.RegionExternalID)
	if tenantRegionInfo.MaxScanBytes != nil {
		data.MaxScanBytes = types.Int64Value(int64(*tenantRegionInfo.MaxScanBytes))
	}
	data.EffectiveMaxScanBytes = types.Int64Value(int64(tenantRegionInfo.EffectiveMaxScanBytes))
	data.SqsARN = types.StringValue(sqsARN)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *tenantRegionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data tenantRegionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := data.Region.ValueString()
	if region == "" {
		tenantInfo, err := r.client.Tenant(ctx, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Cannot get tenant info",
				fmt.Sprintf("Unable to get tenant info in region %s: %v", region, err.Error()),
			)
			return
		}
		region = tenantInfo.HomeRegion
	}

	bucket := data.Bucket.ValueString()
	roleARN := data.RoleARN.ValueString()
	err := r.client.SetBucket(ctx, region, "s3://"+bucket, roleARN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot set tenant region bucket",
			fmt.Sprintf("Unable to set tenant bucket in region %s: %v", region, err.Error()),
		)
		return
	}

	if !data.MaxScanBytes.IsNull() && !data.MaxScanBytes.IsUnknown() {
		_, err = r.client.SetMaxScanBytes(ctx, region, ptr(uint64(data.MaxScanBytes.ValueInt64())))
		if err != nil {
			resp.Diagnostics.AddError(
				"Cannot set max-scan-bytes value",
				fmt.Sprintf("Unable to set max-scan-bytes value in region %s: %v", region, err.Error()),
			)
			return
		}
	}

	// refresh tenant information
	tenantInfo, err := r.client.Tenant(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info (after set-bucket)",
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
		} else {
			sqsARN = "invalid"
		}
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", tenantInfo.TenantID, region))
	data.Region = types.StringValue(region)
	data.Prefix = types.StringValue(api.DefaultDbPrefix)
	data.ExternalID = types.StringValue(tenantRegionInfo.RegionExternalID)
	if tenantRegionInfo.MaxScanBytes != nil {
		data.MaxScanBytes = types.Int64Value(int64(*tenantRegionInfo.MaxScanBytes))
	}
	data.EffectiveMaxScanBytes = types.Int64Value(int64(tenantRegionInfo.EffectiveMaxScanBytes))
	data.SqsARN = types.StringValue(sqsARN)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *tenantRegionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data tenantRegionResourceModel

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

	tenantRegionInfo := tenantInfo.Regions[region]

	bucket := data.Bucket.ValueString()
	roleARN := data.RoleARN.ValueString()
	err = r.client.SetBucket(ctx, region, "s3://"+bucket, roleARN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot set tenant region configuration",
			fmt.Sprintf("Unable to set tenant configuration in region %s: %v", region, err.Error()),
		)
		return
	}

	var maxScanBytes *uint64
	if !data.MaxScanBytes.IsNull() && !data.MaxScanBytes.IsUnknown() {
		maxScanBytes = ptr(uint64(data.MaxScanBytes.ValueInt64()))
	}
	currentMaxScanBytes := tenantRegionInfo.MaxScanBytes

	if (maxScanBytes == nil && currentMaxScanBytes != nil) ||
		(maxScanBytes != nil && currentMaxScanBytes == nil) ||
		(maxScanBytes != nil && currentMaxScanBytes != nil && *maxScanBytes != *currentMaxScanBytes) {
		_, err = r.client.SetMaxScanBytes(ctx, region, maxScanBytes)
		if err != nil {
			resp.Diagnostics.AddError(
				"Cannot set max-scan-bytes value",
				fmt.Sprintf("Unable to set max-scan-bytes value in region %s: %v", region, err.Error()),
			)
			return
		}
	}

	// refresh tenant information
	tenantInfo, err = r.client.Tenant(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot get tenant info (after set-bucket)",
			fmt.Sprintf("Unable to get tenant info in region %s: %v", region, err.Error()),
		)
		return
	}
	tenantRegionInfo = tenantInfo.Regions[region]

	sqsARN := tenantRegionInfo.SqsArn
	if tenantRegionInfo.SqsArn == "" {
		// workaround for older API versions that
		// don't expose the SQS queue ARN
		tenantRoleARN, err := arn.Parse(tenantInfo.TenantRoleArn)
		if err == nil {
			sqsARN = fmt.Sprintf("arn:aws:sqs:%s:%s:tenant-sdb-%s", region, tenantRoleARN.AccountID, tenantInfo.TenantID)
		} else {
			sqsARN = "invalid"
		}
	}

	data.Prefix = types.StringValue(api.DefaultDbPrefix)
	data.ExternalID = types.StringValue(tenantRegionInfo.RegionExternalID)
	if tenantRegionInfo.MaxScanBytes != nil {
		data.MaxScanBytes = types.Int64Value(int64(*tenantRegionInfo.MaxScanBytes))
	}
	data.EffectiveMaxScanBytes = types.Int64Value(int64(tenantRegionInfo.EffectiveMaxScanBytes))
	data.SqsARN = types.StringValue(sqsARN)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *tenantRegionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data tenantRegionResourceModel

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

	err = r.client.ResetBucket(ctx, region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot delete tenant region configuration",
			fmt.Sprintf("Unable to reset tenant configuration in region %s: %v", region, err.Error()),
		)
		return
	}
}

func (r *tenantRegionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
