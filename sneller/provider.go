package sneller

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func New() provider.Provider {
	return &snellerProvider{}
}

var _ provider.Provider = &snellerProvider{}

type snellerProvider struct{}

func (p *snellerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sneller"
}

func (p *snellerProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for interacting with Sneller.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "Sneller token to authenticate to the Sneller API. It defaults to the " + envSnellerToken + " " +
					"environment variable.",
				Sensitive: true,
				Optional:  true,
			},
			"default_region": schema.StringAttribute{
				Description: "Default AWS region to use. It defaults to the " + envSnellerRegion + " " +
					"environment variable. If this variable isn't set, then it default to us-east-1",
				Optional: true,
			},
			"api_endpoint": schema.StringAttribute{
				Description: "Endpoint of the Sneller API (intended for internal use).",
				Optional:    true,
			},
		},
	}
}

type snellerProviderModel struct {
	Token         types.String `tfsdk:"token"`
	DefaultRegion types.String `tfsdk:"default_region"`
	Endpoint      types.String `tfsdk:"api_endpoint"`
}

const (
	envSnellerApiEndpoint = "SNELLER_API_ENDPOINT"
	envSnellerToken       = "SNELLER_TOKEN"
	envSnellerRegion      = "SNELLER_REGION"
	defaultApiEndPoint    = "https://latest-api-production.__REGION__.sneller.io"
	defaultSnellerRegion  = "us-east-1"
)

func (p *snellerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data snellerProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Check configuration data, which should take precedence over
	token := os.Getenv(envSnellerToken)
	if data.Token.ValueString() != "" {
		token = data.Token.ValueString()
	}
	if token == "" {
		resp.Diagnostics.AddError(
			"Missing token",
			"While configuring the provider, the token was not found in "+
				"the "+envSnellerToken+" environment variable or provider "+
				"configuration block token attribute.",
		)
	}

	defaultRegion := os.Getenv(envSnellerRegion)
	if data.DefaultRegion.ValueString() != "" {
		defaultRegion = data.DefaultRegion.ValueString()
	}
	if defaultRegion == "" {
		defaultRegion = defaultSnellerRegion
	}

	apiEndPoint := os.Getenv(envSnellerApiEndpoint)
	if data.Endpoint.ValueString() != "" {
		apiEndPoint = data.Endpoint.ValueString()
	}
	if apiEndPoint == "" {
		apiEndPoint = defaultApiEndPoint
	}
	apiURL, err := url.Parse(apiEndPoint)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid API url",
			fmt.Sprintf("The Sneller API url %q is invalid", apiEndPoint),
		)
		return
	}

	c := Client{
		Token:         token,
		DefaultRegion: defaultRegion,
		apiURL:        apiURL,
	}

	if err = c.Ping(ctx, defaultRegion); err != nil {
		resp.Diagnostics.AddError(
			"Cannot access Sneller API",
			fmt.Sprintf("The Sneller API cannot be contacted: %v", err.Error()),
		)
	}

	resp.DataSourceData = &c
	resp.ResourceData = &c
}

func (p *snellerProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTenantRegionResource,
		NewTableResource,
	}
}

func (p *snellerProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTenantDataSource,
		NewTenantRegionDataSource,
		NewDatabasesDataSource,
		NewDatabaseDataSource,
		NewTableDataSource,
	}
}
