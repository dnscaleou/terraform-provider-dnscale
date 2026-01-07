package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dnscale": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates required environment variables are set before running acceptance tests
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("DNSCALE_API_KEY"); v == "" {
		t.Fatal("DNSCALE_API_KEY must be set for acceptance tests")
	}
}

// TestProvider_Instantiation tests that the provider can be instantiated
func TestProvider_Instantiation(t *testing.T) {
	t.Parallel()

	p := New("test")()
	if p == nil {
		t.Fatal("provider should not be nil")
	}
}
