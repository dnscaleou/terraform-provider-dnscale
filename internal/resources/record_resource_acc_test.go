package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRecordResource_A tests A record creation
func TestAccRecordResource_A(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRecordResourceConfig_A(rName, "192.0.2.1", 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_record.test", "name", rName+"."),
					resource.TestCheckResourceAttr("dnscale_record.test", "type", "A"),
					resource.TestCheckResourceAttr("dnscale_record.test", "content", "192.0.2.1"),
					resource.TestCheckResourceAttr("dnscale_record.test", "ttl", "300"),
					resource.TestCheckResourceAttrSet("dnscale_record.test", "id"),
				),
			},
			// Update testing
			{
				Config: testAccRecordResourceConfig_A(rName, "192.0.2.2", 600),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_record.test", "content", "192.0.2.2"),
					resource.TestCheckResourceAttr("dnscale_record.test", "ttl", "600"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dnscale_record.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccRecordResource_AAAA tests AAAA record creation
func TestAccRecordResource_AAAA(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_AAAA(rName, "2001:db8::1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_record.test", "type", "AAAA"),
					resource.TestCheckResourceAttr("dnscale_record.test", "content", "2001:db8::1"),
				),
			},
		},
	})
}

// TestAccRecordResource_CNAME tests CNAME record creation
func TestAccRecordResource_CNAME(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_CNAME(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_record.test", "name", "www."+rName+"."),
					resource.TestCheckResourceAttr("dnscale_record.test", "type", "CNAME"),
					resource.TestCheckResourceAttr("dnscale_record.test", "content", rName+"."),
				),
			},
		},
	})
}

// TestAccRecordResource_MX tests MX record creation with priority
func TestAccRecordResource_MX(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_MX(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_record.test", "type", "MX"),
					resource.TestCheckResourceAttr("dnscale_record.test", "priority", "10"),
				),
			},
			// Update priority
			{
				Config: testAccRecordResourceConfig_MX(rName, 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_record.test", "priority", "20"),
				),
			},
		},
	})
}

// TestAccRecordResource_TXT tests TXT record creation
func TestAccRecordResource_TXT(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_TXT(rName, "v=spf1 -all"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_record.test", "type", "TXT"),
					resource.TestCheckResourceAttr("dnscale_record.test", "content", "v=spf1 -all"),
				),
			},
		},
	})
}

// TestAccRecordResource_NS tests NS record creation
func TestAccRecordResource_NS(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_NS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_record.test", "type", "NS"),
					resource.TestCheckResourceAttr("dnscale_record.test", "content", "ns1.example.com."),
				),
			},
		},
	})
}

// TestAccRecordResource_CAA tests CAA record creation
func TestAccRecordResource_CAA(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_CAA(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_record.test", "type", "CAA"),
				),
			},
		},
	})
}

// Config helper functions

func testAccRecordResourceConfig_A(zoneName, content string, ttl int) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

resource "dnscale_record" "test" {
  zone_id = dnscale_zone.test.id
  name    = "%[1]s."
  type    = "A"
  content = %[2]q
  ttl     = %[3]d
}
`, zoneName, content, ttl)
}

func testAccRecordResourceConfig_AAAA(zoneName, content string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

resource "dnscale_record" "test" {
  zone_id = dnscale_zone.test.id
  name    = "ipv6.%[1]s."
  type    = "AAAA"
  content = %[2]q
  ttl     = 300
}
`, zoneName, content)
}

func testAccRecordResourceConfig_CNAME(zoneName string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

resource "dnscale_record" "test" {
  zone_id = dnscale_zone.test.id
  name    = "www.%[1]s."
  type    = "CNAME"
  content = "%[1]s."
  ttl     = 3600
}
`, zoneName)
}

func testAccRecordResourceConfig_MX(zoneName string, priority int) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

resource "dnscale_record" "test" {
  zone_id  = dnscale_zone.test.id
  name     = "%[1]s."
  type     = "MX"
  content  = "mail.%[1]s."
  ttl      = 3600
  priority = %[2]d
}
`, zoneName, priority)
}

func testAccRecordResourceConfig_TXT(zoneName, content string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

resource "dnscale_record" "test" {
  zone_id = dnscale_zone.test.id
  name    = "%[1]s."
  type    = "TXT"
  content = %[2]q
  ttl     = 3600
}
`, zoneName, content)
}

func testAccRecordResourceConfig_NS(zoneName string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

resource "dnscale_record" "test" {
  zone_id = dnscale_zone.test.id
  name    = "%[1]s."
  type    = "NS"
  content = "ns1.example.com."
  ttl     = 3600
}
`, zoneName)
}

func testAccRecordResourceConfig_CAA(zoneName string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

resource "dnscale_zone" "test" {
  name   = %[1]q
  region = "EU"
  type   = "master"
}

resource "dnscale_record" "test" {
  zone_id = dnscale_zone.test.id
  name    = "%[1]s."
  type    = "CAA"
  content = "0 issue \"letsencrypt.org\""
  ttl     = 3600
}
`, zoneName)
}

// TestAccRecordResource_ConcurrentCreation tests concurrent record creation
// This validates the fix for the race condition issue with zone-level locking
func TestAccRecordResource_ConcurrentCreation(t *testing.T) {
	testAccPreCheck(t)
	rName := fmt.Sprintf("tf-acc-test-%s.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_Concurrent(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dnscale_record.a", "type", "A"),
					resource.TestCheckResourceAttr("dnscale_record.aaaa", "type", "AAAA"),
					resource.TestCheckResourceAttr("dnscale_record.cname", "type", "CNAME"),
					resource.TestCheckResourceAttr("dnscale_record.txt", "type", "TXT"),
					resource.TestCheckResourceAttr("dnscale_record.mx", "type", "MX"),
				),
			},
		},
	})
}

func testAccRecordResourceConfig_Concurrent(zoneName string) string {
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

resource "dnscale_record" "aaaa" {
  zone_id = dnscale_zone.test.id
  name    = "ipv6.%[1]s."
  type    = "AAAA"
  content = "2001:db8::1"
  ttl     = 300
}

resource "dnscale_record" "cname" {
  zone_id = dnscale_zone.test.id
  name    = "www.%[1]s."
  type    = "CNAME"
  content = "%[1]s."
  ttl     = 3600
}

resource "dnscale_record" "txt" {
  zone_id = dnscale_zone.test.id
  name    = "%[1]s."
  type    = "TXT"
  content = "v=spf1 -all"
  ttl     = 3600
}

resource "dnscale_record" "mx" {
  zone_id  = dnscale_zone.test.id
  name     = "%[1]s."
  type     = "MX"
  content  = "mail.%[1]s."
  ttl      = 3600
  priority = 10
}
`, zoneName)
}
