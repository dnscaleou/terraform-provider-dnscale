package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/dnscale/terraform-provider-dnscale/internal/provider"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dnscale": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck validates required environment variables are set before running acceptance tests
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("DNSCALE_API_KEY"); v == "" {
		t.Skip("DNSCALE_API_KEY must be set for acceptance tests")
	}
}

// TestAccZoneResource_basic tests basic zone creation
func TestAccZoneResource_basic(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccZoneResourceConfig(rName, "EU"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_zone.test", "name", rName),
					resource.TestCheckResourceAttr("dnscale_zone.test", "region", "EU"),
					resource.TestCheckResourceAttr("dnscale_zone.test", "type", "master"),
					resource.TestCheckResourceAttrSet("dnscale_zone.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dnscale_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccZoneResource_update tests zone update (region change requires replacement)
func TestAccZoneResource_update(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with EU region
			{
				Config: testAccZoneResourceConfig(rName, "EU"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_zone.test", "region", "EU"),
				),
			},
			// Update to US region (should force replacement)
			{
				Config: testAccZoneResourceConfig(rName, "US"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_zone.test", "region", "US"),
				),
			},
		},
	})
}

func testAccZoneResourceConfig(name, region string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = %[2]q
  type   = "master"
}
`, name, region)
}
