package netbox

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"log"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	// newProvider is an example function that returns a provider.Provider
	"netbox": func() (tfprotov6.ProviderServer, error) {
		upgradedSdkServer, err := tf5to6server.UpgradeServer(
			context.Background(),
			Provider().GRPCProvider,
		)

		if err != nil {
			log.Fatal(err)
		}

		providers := []func() tfprotov6.ProviderServer{
			providerserver.NewProtocol6(New()),
			func() tfprotov6.ProviderServer {
				return upgradedSdkServer
			},
		}

		muxServer, err := tf6muxserver.NewMuxServer(context.Background(), providers...)
		if err != nil {
			log.Fatal(err)
		}
		return muxServer, err
	},
}
