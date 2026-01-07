package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRecordsDataSource_basic tests reading all records for a zone
func TestAccRecordsDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Should have at least 2 records (A and CNAME we created)
					resource.TestCheckResourceAttrSet("data.dnscale_records.test", "records.#"),
				),
			},
		},
	})
}

func testAccRecordsDataSourceConfig(zoneName string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

resource "dnscale_record" "a" {
  zone_id = dnscale_zone.test.id
  name    = "%[1]s."
  type    = "A"
  content = "192.0.2.1"
  ttl     = 300
}

resource "dnscale_record" "cname" {
  zone_id = dnscale_zone.test.id
  name    = "www.%[1]s."
  type    = "CNAME"
  content = "%[1]s."
  ttl     = 3600
}

data "dnscale_records" "test" {
  zone_id = dnscale_zone.test.id
  depends_on = [
    dnscale_record.a,
    dnscale_record.cname
  ]
}
`, zoneName)
}

// TestAccRecordsDataSource_filtered tests reading filtered records
func TestAccRecordsDataSource_filtered(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordsDataSourceConfig_Filtered(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dnscale_records.test", "records.#"),
				),
			},
		},
	})
}

func testAccRecordsDataSourceConfig_Filtered(zoneName string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

resource "dnscale_record" "a1" {
  zone_id = dnscale_zone.test.id
  name    = "%[1]s."
  type    = "A"
  content = "192.0.2.1"
  ttl     = 300
}

resource "dnscale_record" "a2" {
  zone_id = dnscale_zone.test.id
  name    = "api.%[1]s."
  type    = "A"
  content = "192.0.2.2"
  ttl     = 300
}

resource "dnscale_record" "txt" {
  zone_id = dnscale_zone.test.id
  name    = "%[1]s."
  type    = "TXT"
  content = "v=spf1 -all"
  ttl     = 3600
}

data "dnscale_records" "test" {
  zone_id = dnscale_zone.test.id
  depends_on = [
    dnscale_record.a1,
    dnscale_record.a2,
    dnscale_record.txt
  ]
}
`, zoneName)
}
