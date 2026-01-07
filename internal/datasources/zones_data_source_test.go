package datasources_test

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

// TestAccZonesDataSource_basic tests reading all zones
func TestAccZonesDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccZonesDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// At least one zone should exist (the one we created)
					resource.TestCheckResourceAttrSet("data.dnscale_zones.all", "zones.#"),
				),
			},
		},
	})
}

func testAccZonesDataSourceConfig(zoneName string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

data "dnscale_zones" "all" {
  depends_on = [dnscale_zone.test]
}
`, zoneName)
}

// TestAccZoneDataSource_basic tests reading a specific zone by ID
func TestAccZoneDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dnscale_zone.test", "name", rName),
					resource.TestCheckResourceAttr("data.dnscale_zone.test", "region", "EU"),
					resource.TestCheckResourceAttr("data.dnscale_zone.test", "type", "master"),
				),
			},
		},
	})
}

func testAccZoneDataSourceConfig(zoneName string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

data "dnscale_zone" "test" {
  id = dnscale_zone.test.id
}
`, zoneName)
}
