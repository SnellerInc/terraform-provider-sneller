package sneller

import (
	"context"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("SNELLER_TOKEN", nil),
			},
			"default_region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SNELLER_REGION", nil),
			},
			"api_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "https://latest-api-production.__REGION__.sneller.io",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"sneller_tenant_region": resourceTenantRegion(),
			"sneller_table":         resourceTable(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"sneller_tenant":        dataSourceTenant(),
			"sneller_tenant_region": dataSourceTenantRegion(),
			"sneller_databases":     dataSourceDatabases(),
			"sneller_database":      dataSourceDatabase(),
			"sneller_table":         dataSourceTable(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	defaultRegion := d.Get("default_region").(string)
	apiEndPoint := d.Get("api_endpoint").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	apiURL, err := url.Parse(apiEndPoint)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	c := Client{
		Client:        http.DefaultClient,
		Token:         token,
		DefaultRegion: defaultRegion,
		apiURL:        apiURL,
	}

	if err = c.Ping(defaultRegion); err != nil {
		return nil, diag.FromErr(err)
	}

	return &c, diags
}
