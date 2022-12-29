package sneller

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTenant(t *testing.T) {
	initVars(t)

	resourceName := "data.sneller_tenant.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "sneller_tenant" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", snellerTenantID),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", snellerTenantID),
					resource.TestCheckResourceAttr(resourceName, "state", "active"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "home_region", defaultSnellerRegion),
					resource.TestCheckResourceAttrSet(resourceName, "email"),
					resource.TestCheckResourceAttr(resourceName, "tenant_role_arn", fmt.Sprintf("arn:aws:iam::%s:role/tenant-%s", snellerAccountID, snellerTenantID)),
					resource.TestCheckResourceAttr(resourceName, "mfa", "off"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "activated_at"),
					resource.TestCheckResourceAttr(resourceName, "deactivated_at", ""),
				),
			},
		},
	})
}
