package main

import (
	"context"
	"log"

	"github.com/cepidalim-epam/terraform-provider-codemie/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version = "dev"

func main() {
	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/cepidalim-epam/codemie",
	})
	if err != nil {
		log.Fatal(err)
	}
}
