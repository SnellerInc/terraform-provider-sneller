package sneller

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the Sneller client is properly configured.
	// It is also possible to use the SNELLER_TOKEN and SNELLER_REGION
	// environment variables instead, such as updating the Makefile and
	// running the testing through that tool.
	providerConfig = `provider "sneller" {}` + "\n"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"sneller": providerserver.NewProtocol6WithError(New()),
}
