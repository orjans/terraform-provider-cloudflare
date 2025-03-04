package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func init() {
	resource.AddTestSweepers("cloudflare_record", &resource.Sweeper{
		Name: "cloudflare_record",
		F:    testSweepCloudflareRecord,
	})
}

func testSweepCloudflareRecord(r string) error {
	ctx := context.Background()
	client, clientErr := sharedClient()
	if clientErr != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed to create Cloudflare client: %s", clientErr))
	}

	// Clean up the account level rulesets
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	if zoneID == "" {
		return errors.New("CLOUDFLARE_ZONE_ID must be set")
	}

	records, err := client.DNSRecords(context.Background(), zoneID, cloudflare.DNSRecord{})
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed to fetch Cloudflare DNS records: %s", err))
	}

	if len(records) == 0 {
		log.Print("[DEBUG] No Cloudflare DNS records to sweep")
		return nil
	}

	for _, record := range records {
		tflog.Info(ctx, fmt.Sprintf("Deleting Cloudflare DNS record ID: %s", record.ID))
		//nolint:errcheck
		client.DeleteDNSRecord(context.Background(), zoneID, record.ID)
	}

	return nil
}

func TestAccCloudflareRecord_Basic(t *testing.T) {
	t.Parallel()
	var record cloudflare.DNSRecord
	testStartTime := time.Now().UTC()
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigBasic(zoneID, "tf-acctest-basic", rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					testAccCheckCloudflareRecordAttributes(&record),
					testAccCheckCloudflareRecordDates(resourceName, &record, testStartTime),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acctest-basic"),
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "value", "192.168.0.10"),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("tf-acctest-basic.%s", zoneName)),
					resource.TestMatchResourceAttr(resourceName, "zone_id", regexp.MustCompile("^[a-z0-9]{32}$")),
					resource.TestCheckResourceAttr(resourceName, "ttl", "3600"),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "metadata.auto_added", "false"),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_CaseInsensitive(t *testing.T) {
	t.Parallel()
	var record cloudflare.DNSRecord
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigBasic(zoneID, "tf-acctest-case-insensitive", rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					testAccCheckCloudflareRecordAttributes(&record),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acctest-case-insensitive"),
				),
			},
			{
				Config:   testAccCheckCloudflareRecordConfigBasic(zoneID, "tf-acctest-CASE-INSENSITIVE", rnd),
				PlanOnly: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					testAccCheckCloudflareRecordAttributes(&record),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acctest-case-insensitive"),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_Apex(t *testing.T) {
	t.Parallel()
	var record cloudflare.DNSRecord
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigApex(zoneID, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					testAccCheckCloudflareRecordAttributes(&record),
					resource.TestCheckResourceAttr(resourceName, "name", "@"),
					resource.TestCheckResourceAttr(resourceName, "zone_id", zoneID),
					resource.TestCheckResourceAttr(resourceName, "value", "192.168.0.10"),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_LOC(t *testing.T) {
	t.Parallel()
	var record cloudflare.DNSRecord
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigLOC(zoneID, "tf-acctest-loc", rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					resource.TestCheckResourceAttr(resourceName, "value", "37 46 46.000 N 122 23 35.000 W 0.00 100.00 0.00 0.00"),
					resource.TestCheckResourceAttr(resourceName, "proxiable", "false"),
					resource.TestCheckResourceAttr(resourceName, "data.0.lat_degrees", "37"),
					resource.TestCheckResourceAttr(resourceName, "data.0.lat_degrees", "37"),
					resource.TestCheckResourceAttr(resourceName, "data.0.lat_minutes", "46"),
					resource.TestCheckResourceAttr(resourceName, "data.0.lat_seconds", "46"),
					resource.TestCheckResourceAttr(resourceName, "data.0.lat_direction", "N"),
					resource.TestCheckResourceAttr(resourceName, "data.0.long_degrees", "122"),
					resource.TestCheckResourceAttr(resourceName, "data.0.long_minutes", "23"),
					resource.TestCheckResourceAttr(resourceName, "data.0.long_seconds", "35"),
					resource.TestCheckResourceAttr(resourceName, "data.0.long_direction", "W"),
					resource.TestCheckResourceAttr(resourceName, "data.0.altitude", "0"),
					resource.TestCheckResourceAttr(resourceName, "data.0.size", "100"),
					resource.TestCheckResourceAttr(resourceName, "data.0.precision_horz", "0"),
					resource.TestCheckResourceAttr(resourceName, "data.0.precision_vert", "0"),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_SRV(t *testing.T) {
	t.Parallel()
	var record cloudflare.DNSRecord
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigSRV(zoneID, "tf-acctest-srv", rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					resource.TestCheckResourceAttr(resourceName, "hostname", fmt.Sprintf("_xmpp-client._tcp.tf-acctest-srv.%s", domain)),
					resource.TestCheckResourceAttr(resourceName, "value", "0	5222	talk.l.google.com"),
					resource.TestCheckResourceAttr(resourceName, "proxiable", "false"),
					resource.TestCheckResourceAttr(resourceName, "data.0.priority", "5"),
					resource.TestCheckResourceAttr(resourceName, "data.0.weight", "0"),
					resource.TestCheckResourceAttr(resourceName, "data.0.port", "5222"),
					resource.TestCheckResourceAttr(resourceName, "data.0.target", "talk.l.google.com"),
					resource.TestCheckResourceAttr(resourceName, "data.0.service", "_xmpp-client"),
					resource.TestCheckResourceAttr(resourceName, "data.0.proto", "_tcp"),
					resource.TestCheckResourceAttr(resourceName, "data.0.name", "tf-acctest-srv"),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_CAA(t *testing.T) {
	t.Parallel()
	var record cloudflare.DNSRecord
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigCAA(rnd, zoneID, fmt.Sprintf("tf-acctest-caa.%s", domain), 600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					resource.TestCheckResourceAttr(resourceName, "data.0.flags", "0"),
					resource.TestCheckResourceAttr(resourceName, "data.0.tag", "issue"),
					resource.TestCheckResourceAttr(resourceName, "data.0.value", "letsencrypt.org"),
				),
			},
			{
				Config: testAccCheckCloudflareRecordConfigCAA(rnd, zoneID, fmt.Sprintf("tf-acctest-caa.%s", domain), 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					resource.TestCheckResourceAttr(resourceName, "data.0.flags", "0"),
					resource.TestCheckResourceAttr(resourceName, "data.0.tag", "issue"),
					resource.TestCheckResourceAttr(resourceName, "data.0.value", "letsencrypt.org"),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_Proxied(t *testing.T) {
	t.Parallel()
	var record cloudflare.DNSRecord
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigProxied(zoneID, domain, "tf-acctest-proxied", rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					resource.TestCheckResourceAttr(resourceName, "proxiable", "true"),
					resource.TestCheckResourceAttr(resourceName, "proxied", "true"),
					resource.TestCheckResourceAttr(resourceName, "type", "CNAME"),
					resource.TestCheckResourceAttr(resourceName, "value", domain),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_Updated(t *testing.T) {
	t.Parallel()
	var record cloudflare.DNSRecord
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	recordName := "tf-acctest-update"
	rnd := generateRandomResourceName()
	resourceName := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigBasic(zoneID, recordName, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					testAccCheckCloudflareRecordAttributes(&record),
				),
			},
			{
				Config: testAccCheckCloudflareRecordConfigNewValue(zoneID, recordName, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &record),
					testAccCheckCloudflareRecordAttributesUpdated(&record),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_typeForceNewRecord(t *testing.T) {
	t.Parallel()
	var afterCreate, afterUpdate cloudflare.DNSRecord
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	recordName := "tf-acctest-type-force-new"
	rnd := generateRandomResourceName()
	resourceName := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigBasic(zoneID, recordName, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &afterCreate),
				),
			},
			{
				Config: testAccCheckCloudflareRecordConfigChangeType(zoneID, recordName, zoneName, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &afterUpdate),
					testAccCheckCloudflareRecordRecreated(&afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_hostnameForceNewRecord(t *testing.T) {
	t.Parallel()
	var afterCreate, afterUpdate cloudflare.DNSRecord
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	recordName := "tf-acctest-hostname-force-new"
	rnd := generateRandomResourceName()
	resourceName := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigBasic(zoneID, recordName, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &afterCreate),
				),
			},
			{
				Config: testAccCheckCloudflareRecordConfigChangeHostname(zoneID, recordName, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(resourceName, &afterUpdate),
					testAccCheckCloudflareRecordRecreated(&afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_CreateAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var afterCreate, afterRecreate cloudflare.DNSRecord
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigBasic(zoneID, name, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(name, &afterCreate),
					testAccManuallyDeleteRecord(&afterCreate),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCheckCloudflareRecordConfigBasic(zoneID, name, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(name, &afterRecreate),
					testAccCheckCloudflareRecordRecreated(&afterCreate, &afterRecreate),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_TtlValidation(t *testing.T) {
	t.Parallel()
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	recordName := "tf-acctest-ttl-validation"
	rnd := generateRandomResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckCloudflareRecordConfigTtlValidation(zoneID, recordName, zoneName, rnd),
				ExpectError: regexp.MustCompile(fmt.Sprintf("error validating record %s: ttl must be set to 1 when `proxied` is true", recordName)),
			},
		},
	})
}

func TestAccCloudflareRecord_ExplicitProxiedFalse(t *testing.T) {
	t.Parallel()
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_record." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigExplicitProxied(zoneID, rnd, zoneName, "false", "300"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "proxied", "false"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "300"),
				),
			},
			{
				Config: testAccCheckCloudflareRecordConfigExplicitProxied(zoneID, rnd, zoneName, "true", "1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "proxied", "true"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "1"),
				),
			},
			{
				Config: testAccCheckCloudflareRecordConfigExplicitProxied(zoneID, rnd, zoneName, "false", "300"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "proxied", "false"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "300"),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_MXWithPriorityZero(t *testing.T) {
	t.Parallel()
	zoneName := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rnd := generateRandomResourceName()
	resourceName := "cloudflare_record." + rnd

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigMXWithPriorityZero(zoneID, rnd, zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "value", "mail.terraform.cfapi.net"),
				),
			},
		},
	})
}

func TestAccCloudflareRecord_TtlValidationUpdate(t *testing.T) {
	t.Parallel()
	domain := os.Getenv("CLOUDFLARE_DOMAIN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	recordName := "tf-acctest-ttl-validation"
	rnd := generateRandomResourceName()
	name := fmt.Sprintf("cloudflare_record.%s", rnd)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckCloudflareRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCloudflareRecordConfigProxied(zoneID, domain, recordName, rnd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudflareRecordExists(name, &cloudflare.DNSRecord{}),
				),
			},
			{
				Config:      testAccCheckCloudflareRecordConfigTtlValidation(zoneID, recordName, domain, rnd),
				ExpectError: regexp.MustCompile(fmt.Sprintf("error validating record %s: ttl must be set to 1 when `proxied` is true", recordName)),
			},
		},
	})
}

func testAccCheckCloudflareRecordRecreated(before, after *cloudflare.DNSRecord) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID == after.ID {
			return fmt.Errorf("expected change of Record Ids, but both were %v", before.ID)
		}
		return nil
	}
}

func testAccCheckCloudflareRecordDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudflare.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudflare_record" {
			continue
		}

		_, err := client.DNSRecord(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Record still exists")
		}
	}

	return nil
}

func testAccManuallyDeleteRecord(record *cloudflare.DNSRecord) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*cloudflare.API)
		err := client.DeleteDNSRecord(context.Background(), record.ZoneID, record.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckCloudflareRecordAttributes(record *cloudflare.DNSRecord) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if record.Content != "192.168.0.10" {
			return fmt.Errorf("bad content: %s", record.Content)
		}

		return nil
	}
}

func testAccCheckCloudflareRecordAttributesUpdated(record *cloudflare.DNSRecord) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if record.Content != "192.168.0.11" {
			return fmt.Errorf("bad content: %s", record.Content)
		}

		return nil
	}
}

func testAccCheckCloudflareRecordDates(n string, record *cloudflare.DNSRecord, testStartTime time.Time) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, _ := s.RootModule().Resources[n]

		for timeStampAttr, serverVal := range map[string]time.Time{"created_on": record.CreatedOn, "modified_on": record.ModifiedOn} {
			timeStamp, err := time.Parse(time.RFC3339Nano, rs.Primary.Attributes[timeStampAttr])
			if err != nil {
				return err
			}

			if timeStamp != serverVal {
				return fmt.Errorf("state value of %s: %s is different than server created value: %s", timeStampAttr, rs.Primary.Attributes[timeStampAttr], serverVal.Format(time.RFC3339Nano))
			}

			// check retrieved values are reasonable
			// note this could fail if local time is out of sync with server time
			if timeStamp.Before(testStartTime) {
				return fmt.Errorf("state value of %s: %s should be greater than test start time: %s", timeStampAttr, timeStamp.Format(time.RFC3339Nano), testStartTime.Format(time.RFC3339Nano))
			}
		}

		return nil
	}
}

func testAccCheckCloudflareRecordExists(n string, record *cloudflare.DNSRecord) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*cloudflare.API)
		foundRecord, err := client.DNSRecord(context.Background(), rs.Primary.Attributes["zone_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundRecord.ID != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*record = foundRecord

		return nil
	}
}

func testAccCheckCloudflareRecordConfigBasic(zoneID, name, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[3]s" {
	zone_id = "%[1]s"
	name = "%[2]s"
	value = "192.168.0.10"
	type = "A"
	ttl = 3600
}`, zoneID, name, rnd)
}

func testAccCheckCloudflareRecordConfigApex(zoneID, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[2]s" {
	zone_id = "%[1]s"
	name = "@"
	value = "192.168.0.10"
	type = "A"
	ttl = 3600
}`, zoneID, rnd)
}

func testAccCheckCloudflareRecordConfigLOC(zoneID, name, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[3]s" {
  zone_id = "%[1]s"
  name = "%[2]s"
  data {
    lat_degrees    = "37"
    lat_minutes    = "46"
    lat_seconds    = 46.000
    lat_direction  = "N"
    long_degrees   = "122"
    long_minutes   = "23"
    long_seconds   = 35.000
    long_direction = "W"
    altitude       = 0.00
    size           = 100.00
    precision_horz = 0.00
    precision_vert = 0.00
  }
  type = "LOC"
  ttl = 3600
}`, zoneID, name, rnd)
}

func testAccCheckCloudflareRecordConfigSRV(zoneID, name, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[3]s" {
  zone_id = "%[1]s"
  name = "%[2]s"
  data {
    priority = 5
    weight = 0
    port = 5222
    target = "talk.l.google.com"
    service = "_xmpp-client"
    proto = "_tcp"
    name = "%[2]s"
  }
  type = "SRV"
  ttl = 3600
}`, zoneID, name, rnd)
}

func testAccCheckCloudflareRecordConfigCAA(resourceName, zoneID, name string, ttl int) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[1]s" {
  zone_id = "%[2]s"
  name = "%[3]s"
  data {
    flags = "0"
    tag   = "issue"
    value = "letsencrypt.org"
  }
  type = "CAA"
  ttl = %[4]d
}`, resourceName, zoneID, name, ttl)
}

func testAccCheckCloudflareRecordConfigProxied(zoneID, domain, name, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[4]s" {
	zone_id = "%[1]s"
	name = "%[3]s"
	value = "%[2]s"
	type = "CNAME"
	proxied = true
}`, zoneID, domain, name, rnd)
}

func testAccCheckCloudflareRecordConfigNewValue(zoneID, name, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[3]s" {
	zone_id = "%[1]s"
	name = "%[2]s"
	value = "192.168.0.11"
	type = "A"
	ttl = 3600
}`, zoneID, name, rnd)
}

func testAccCheckCloudflareRecordConfigChangeType(zoneID, name, zoneName, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[4]s" {
	zone_id = "%[1]s"
	name = "%[2]s"
	value = "%[3]s"
	type = "CNAME"
	ttl = 3600
}`, zoneID, name, zoneName, rnd)
}

func testAccCheckCloudflareRecordConfigChangeHostname(zoneID, name, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[3]s" {
	zone_id = "%[1]s"
	name = "%[2]s-changed"
	value = "192.168.0.10"
	type = "A"
	ttl = 3600
}`, zoneID, name, rnd)
}

func testAccCheckCloudflareRecordConfigTtlValidation(zoneID, name, zoneName, rnd string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[4]s" {
	zone_id = "%[1]s"
	name = "%[2]s"
	value = "%[3]s"
	type = "CNAME"
	proxied = true
	ttl = 3600
}`, zoneID, name, zoneName, rnd)
}

func testAccCheckCloudflareRecordConfigExplicitProxied(zoneID, name, zoneName, proxied, ttl string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[2]s" {
	zone_id = "%[1]s"
	name = "%[2]s"
	value = "%[3]s"
	type = "CNAME"
	proxied = %[4]s
	ttl = %[5]s
}`, zoneID, name, zoneName, proxied, ttl)
}

func testAccCheckCloudflareRecordConfigMXWithPriorityZero(zoneID, name, zoneName string) string {
	return fmt.Sprintf(`
resource "cloudflare_record" "%[2]s" {
	zone_id = "%[1]s"
	name = "%[2]s"
	value = "mail.terraform.cfapi.net"
	type = "MX"
	priority = 0
	proxied = false
	ttl = 300
}`, zoneID, name, zoneName)
}
