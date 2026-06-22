package datasources_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/dnscale/terraform-provider-dnscale/internal/client"
)

const defaultLargeRecordCount = 10001

// TestAccRecordsDataSource_moreThan10K verifies the Terraform data source can
// read a live zone with more than 10,000 records. This is intentionally opt-in:
// it creates many DNS records and can take a long time against the real API.
func TestAccRecordsDataSource_moreThan10K(t *testing.T) {
	testAccPreCheck(t)
	testAccLargeRecordsPreCheck(t)

	recordCount := testAccEnvInt(t, "DNSCALE_LARGE_RECORD_COUNT", defaultLargeRecordCount, defaultLargeRecordCount)
	parallelism := testAccEnvInt(t, "DNSCALE_LARGE_RECORD_PARALLELISM", 8, 1)
	if parallelism > 32 {
		parallelism = 32
	}

	timeout := time.Duration(max(30, recordCount/250)) * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	apiClient := client.NewClient(os.Getenv("DNSCALE_API_KEY"), os.Getenv("DNSCALE_API_URL"))
	zoneName := fmt.Sprintf(
		"tf-acc-large-%s.com",
		strings.ToLower(acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)),
	)

	zone, err := apiClient.CreateZone(ctx, client.ZoneInput{
		Name:   zoneName,
		Region: "EU",
		Type:   "master",
	})
	if err != nil {
		t.Fatalf("failed to create large-record test zone: %v", err)
	}
	t.Cleanup(func() {
		cleanupLargeRecordZone(t, apiClient, zone.ID, zone.Name, parallelism)
	})

	seedLargeRecords(ctx, t, apiClient, zone.ID, zone.Name, recordCount, parallelism)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLargeRecordsDataSourceConfig(zone.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordsAtLeast("data.dnscale_records.large", recordCount),
				),
			},
		},
	})
}

func testAccLargeRecordsPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("DNSCALE_RUN_LARGE_RECORD_TESTS") != "1" {
		t.Skip("set DNSCALE_RUN_LARGE_RECORD_TESTS=1 to run the live >10K record acceptance test")
	}
}

func testAccEnvInt(t *testing.T, name string, defaultValue, minValue int) int {
	t.Helper()
	raw := os.Getenv(name)
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("%s must be an integer, got %q", name, raw)
	}
	if value < minValue {
		t.Fatalf("%s must be at least %d, got %d", name, minValue, value)
	}
	return value
}

func seedLargeRecords(
	ctx context.Context,
	t *testing.T,
	apiClient *client.Client,
	zoneID string,
	zoneName string,
	count int,
	parallelism int,
) {
	t.Helper()

	start := time.Now()
	zoneName = strings.TrimSuffix(zoneName, ".")
	err := testAccRunConcurrent(ctx, count, parallelism, func(ctx context.Context, i int) error {
		recordInput := client.RecordInput{
			Name:    fmt.Sprintf("r%05d.%s.", i, zoneName),
			Type:    "A",
			Content: fmt.Sprintf("192.0.2.%d", (i%250)+1),
			TTL:     300,
		}
		return testAccWithRetry(ctx, func() error {
			_, err := apiClient.CreateRecord(ctx, zoneID, recordInput)
			return err
		})
	})
	if err != nil {
		t.Fatalf("failed to seed %d large-record test records: %v", count, err)
	}
	t.Logf("seeded %d records in %s", count, time.Since(start).Round(time.Second))
}

func cleanupLargeRecordZone(
	t *testing.T,
	apiClient *client.Client,
	zoneID string,
	zoneName string,
	parallelism int,
) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	deleteZone := func() error {
		return testAccWithRetry(ctx, func() error {
			return apiClient.DeleteZone(ctx, zoneID)
		})
	}

	if err := deleteZone(); err == nil || client.IsNotFound(err) {
		return
	} else {
		t.Logf("initial zone delete failed, deleting generated records before retry: %v", err)
	}

	records, err := apiClient.ListRecords(ctx, zoneID)
	if err != nil {
		t.Logf("failed to list generated records for cleanup: %v", err)
		return
	}

	zoneName = strings.TrimSuffix(zoneName, ".")
	generatedSuffix := "." + zoneName + "."
	generatedRecords := make([]client.Record, 0, len(records))
	for _, record := range records {
		if record.Type == "A" &&
			strings.HasPrefix(record.Name, "r") &&
			strings.HasSuffix(record.Name, generatedSuffix) {
			generatedRecords = append(generatedRecords, record)
		}
	}

	err = testAccRunConcurrent(ctx, len(generatedRecords), parallelism, func(ctx context.Context, i int) error {
		record := generatedRecords[i]
		return testAccWithRetry(ctx, func() error {
			err := apiClient.DeleteRecord(ctx, zoneID, record.ID)
			if client.IsNotFound(err) {
				return nil
			}
			return err
		})
	})
	if err != nil {
		t.Logf("failed to delete all generated records during cleanup: %v", err)
	}

	if err := deleteZone(); err != nil && !client.IsNotFound(err) {
		t.Logf("failed to delete large-record test zone %s after cleanup: %v", zoneID, err)
	}
}

func testAccRunConcurrent(
	ctx context.Context,
	count int,
	parallelism int,
	fn func(context.Context, int) error,
) error {
	if count == 0 {
		return nil
	}
	if parallelism < 1 {
		parallelism = 1
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan int)
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error
	setErr := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}

	for worker := 0; worker < parallelism; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case i, ok := <-jobs:
					if !ok {
						return
					}
					if err := fn(ctx, i); err != nil {
						setErr(err)
						return
					}
				}
			}
		}()
	}

sendLoop:
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			break sendLoop
		case jobs <- i:
		}
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return firstErr
	}
	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func testAccWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < 6; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if !testAccIsRetriableAPIError(err) {
				return err
			}
		} else {
			return nil
		}

		timer := time.NewTimer(time.Duration(attempt+1) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return lastErr
}

func testAccIsRetriableAPIError(err error) bool {
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500
}

func testAccCheckRecordsAtLeast(resourceName string, minCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}
		rawCount, ok := rs.Primary.Attributes["records.#"]
		if !ok {
			return fmt.Errorf("resource %q has no records.# attribute", resourceName)
		}
		count, err := strconv.Atoi(rawCount)
		if err != nil {
			return fmt.Errorf("resource %q records.# is not an integer: %q", resourceName, rawCount)
		}
		if count < minCount {
			return fmt.Errorf("resource %q has %d records, want at least %d", resourceName, count, minCount)
		}
		return nil
	}
}

func testAccLargeRecordsDataSourceConfig(zoneID string) string {
	return fmt.Sprintf(`
provider "dnscale" {}

data "dnscale_records" "large" {
  zone_id = %[1]q
}
`, zoneID)
}
