package datasource_test

import (
	"fmt"
	"terraform-provider-sneller/sneller/acctest"
	"terraform-provider-sneller/sneller/api"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTenant(t *testing.T) {
	resourceName := "data.sneller_tenant.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: acctest.ProviderConfig + `data "sneller_tenant" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", acctest.SnellerTenantID),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", acctest.SnellerTenantID),
					resource.TestCheckResourceAttr(resourceName, "state", "active"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "home_region", api.DefaultSnellerRegion),
					resource.TestCheckResourceAttrSet(resourceName, "email"),
					resource.TestCheckResourceAttr(resourceName, "tenant_role_arn", fmt.Sprintf("arn:aws:iam::%s:role/tenant-%s", acctest.SnellerAccountID, acctest.SnellerTenantID)),
					resource.TestCheckResourceAttr(resourceName, "mfa", "off"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "activated_at"),
					resource.TestCheckResourceAttr(resourceName, "deactivated_at", ""),
				),
			},
		},
	})
}
