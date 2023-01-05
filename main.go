package main

import (
	"context"
	"flag"
	"log"
	"terraform-provider-sneller/sneller/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name sneller

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address:         "registry.terraform.io/SnellerInc/sneller",
		Debug:           debug,
		ProtocolVersion: 6,
	})
	if err != nil {
		log.Fatal(err)
	}
}
