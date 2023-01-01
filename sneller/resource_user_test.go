package sneller

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceUser(t *testing.T) {
	initVars(t)

	email := fmt.Sprintf("john.doe+%s@example.com", acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum))
	resourceName := "sneller_user.test"
	baseConfig := providerConfig

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
					resource "sneller_user" "test" {
						email       = "` + email + `"
						is_admin    = true
						locale      = "nl-NL"
						given_name  = "John"
						family_name = "Doe"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "true"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "is_federated", "false"),
					resource.TestCheckResourceAttr(resourceName, "locale", "nl-NL"),
					resource.TestCheckResourceAttr(resourceName, "given_name", "John"),
					resource.TestCheckResourceAttr(resourceName, "family_name", "Doe"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Import testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"}, // TF only
			},
			// Update and Read testing
			{
				Config: baseConfig + `
				resource "sneller_user" "test" {
					email       = "` + email + `"
					is_admin    = false
					is_enabled  = false
					locale      = "de-DE"
					given_name  = "Max"
					family_name = "Mustermann"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "false"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "is_federated", "false"),
					resource.TestCheckResourceAttr(resourceName, "locale", "de-DE"),
					resource.TestCheckResourceAttr(resourceName, "given_name", "Max"),
					resource.TestCheckResourceAttr(resourceName, "family_name", "Mustermann"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Delete is automatically tested
		},
	})
}
