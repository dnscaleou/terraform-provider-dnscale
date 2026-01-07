package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a client configured to use a test server
func newTestClient(server *httptest.Server) *Client {
	return &Client{
		BaseURL:    server.URL,
		APIKey:     "test-api-key",
		HTTPClient: server.Client(),
	}
}

// TestNewClient tests client creation with default and custom base URLs
func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		baseURL     string
		wantBaseURL string
	}{
		{
			name:        "with default base URL",
			apiKey:      "test-key",
			baseURL:     "",
			wantBaseURL: DefaultBaseURL,
		},
		{
			name:        "with custom base URL",
			apiKey:      "test-key",
			baseURL:     "https://custom.api.example.com",
			wantBaseURL: "https://custom.api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.apiKey, tt.baseURL)
			if client.BaseURL != tt.wantBaseURL {
				t.Errorf("NewClient() BaseURL = %v, want %v", client.BaseURL, tt.wantBaseURL)
			}
			if client.APIKey != tt.apiKey {
				t.Errorf("NewClient() APIKey = %v, want %v", client.APIKey, tt.apiKey)
			}
			if client.HTTPClient == nil {
				t.Error("NewClient() HTTPClient is nil")
			}
		})
	}
}

// TestCreateZone tests zone creation
func TestCreateZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/zones" {
			t.Errorf("expected /v1/zones, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Bearer test-api-key, got %s", r.Header.Get("Authorization"))
		}

		// Return mock response
		response := ZoneAPIResponse{
			Data: struct {
				Zone Zone `json:"zone"`
			}{
				Zone: Zone{
					ID:     "test-zone-id",
					Name:   "example.com",
					Type:   "master",
					Region: "EU",
					Status: "active",
				},
			},
			Status: "success",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server)
	zone, err := client.CreateZone(context.Background(), ZoneInput{
		Name:   "example.com",
		Type:   "master",
		Region: "EU",
	})

	if err != nil {
		t.Fatalf("CreateZone() error = %v", err)
	}
	if zone.ID != "test-zone-id" {
		t.Errorf("CreateZone() zone.ID = %v, want test-zone-id", zone.ID)
	}
	if zone.Name != "example.com" {
		t.Errorf("CreateZone() zone.Name = %v, want example.com", zone.Name)
	}
}

// TestGetZone tests zone retrieval
func TestGetZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/zones/test-zone-id" {
			t.Errorf("expected /v1/zones/test-zone-id, got %s", r.URL.Path)
		}

		response := ZoneAPIResponse{
			Data: struct {
				Zone Zone `json:"zone"`
			}{
				Zone: Zone{
					ID:     "test-zone-id",
					Name:   "example.com",
					Type:   "master",
					Region: "EU",
					Status: "active",
				},
			},
			Status: "success",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server)
	zone, err := client.GetZone(context.Background(), "test-zone-id")

	if err != nil {
		t.Fatalf("GetZone() error = %v", err)
	}
	if zone.ID != "test-zone-id" {
		t.Errorf("GetZone() zone.ID = %v, want test-zone-id", zone.ID)
	}
}

// TestDeleteZone tests zone deletion
func TestDeleteZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/zones/test-zone-id" {
			t.Errorf("expected /v1/zones/test-zone-id, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteZone(context.Background(), "test-zone-id")

	if err != nil {
		t.Fatalf("DeleteZone() error = %v", err)
	}
}

// TestListZones tests listing all zones
func TestListZones(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/zones" {
			t.Errorf("expected /v1/zones, got %s", r.URL.Path)
		}

		response := ZonesAPIResponse{
			Data: struct {
				Zones      []Zone     `json:"zones"`
				Pagination Pagination `json:"pagination"`
			}{
				Zones: []Zone{
					{ID: "zone-1", Name: "example1.com"},
					{ID: "zone-2", Name: "example2.com"},
				},
				Pagination: Pagination{
					Total:  2,
					Offset: 0,
					Limit:  100,
					Count:  2,
				},
			},
			Status: "success",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server)
	zones, err := client.ListZones(context.Background())

	if err != nil {
		t.Fatalf("ListZones() error = %v", err)
	}
	if len(zones) != 2 {
		t.Errorf("ListZones() returned %d zones, want 2", len(zones))
	}
}

// TestCreateRecord tests record creation
func TestCreateRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/zones/zone-id/records" {
			t.Errorf("expected /v1/zones/zone-id/records, got %s", r.URL.Path)
		}

		response := RecordAPIResponse{
			Data: struct {
				Record Record `json:"record"`
			}{
				Record: Record{
					ID:      "record-id",
					Name:    "www.example.com.",
					Type:    "A",
					Content: "192.0.2.1",
					TTL:     3600,
				},
			},
			Status: "success",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server)
	record, err := client.CreateRecord(context.Background(), "zone-id", RecordInput{
		Name:    "www.example.com.",
		Type:    "A",
		Content: "192.0.2.1",
		TTL:     3600,
	})

	if err != nil {
		t.Fatalf("CreateRecord() error = %v", err)
	}
	if record.ID != "record-id" {
		t.Errorf("CreateRecord() record.ID = %v, want record-id", record.ID)
	}
	if record.Content != "192.0.2.1" {
		t.Errorf("CreateRecord() record.Content = %v, want 192.0.2.1", record.Content)
	}
}

// TestGetRecord tests record retrieval
func TestGetRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/zones/zone-id/records/record-id" {
			t.Errorf("expected /v1/zones/zone-id/records/record-id, got %s", r.URL.Path)
		}

		response := RecordAPIResponse{
			Data: struct {
				Record Record `json:"record"`
			}{
				Record: Record{
					ID:      "record-id",
					Name:    "www.example.com.",
					Type:    "A",
					Content: "192.0.2.1",
					TTL:     3600,
				},
			},
			Status: "success",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server)
	record, err := client.GetRecord(context.Background(), "zone-id", "record-id")

	if err != nil {
		t.Fatalf("GetRecord() error = %v", err)
	}
	if record.ID != "record-id" {
		t.Errorf("GetRecord() record.ID = %v, want record-id", record.ID)
	}
}

// TestDeleteRecord tests record deletion
func TestDeleteRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/zones/zone-id/records/record-id" {
			t.Errorf("expected /v1/zones/zone-id/records/record-id, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteRecord(context.Background(), "zone-id", "record-id")

	if err != nil {
		t.Fatalf("DeleteRecord() error = %v", err)
	}
}

// TestListRecords tests listing records with pagination
func TestListRecords(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		// Simulate pagination - return 2 records total
		response := RecordsAPIResponse{
			Data: struct {
				Records    []Record   `json:"records"`
				Pagination Pagination `json:"pagination"`
			}{
				Records: []Record{
					{ID: "record-1", Name: "www.example.com.", Type: "A"},
					{ID: "record-2", Name: "mail.example.com.", Type: "MX"},
				},
				Pagination: Pagination{
					Total:  2,
					Offset: 0,
					Limit:  100,
					Count:  2,
				},
			},
			Status: "success",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server)
	records, err := client.ListRecords(context.Background(), "zone-id")

	if err != nil {
		t.Fatalf("ListRecords() error = %v", err)
	}
	if len(records) != 2 {
		t.Errorf("ListRecords() returned %d records, want 2", len(records))
	}
	if callCount != 1 {
		t.Errorf("ListRecords() made %d API calls, want 1", callCount)
	}
}

// TestAPIError tests error handling for API errors
func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		response := ErrorResponse{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				Code:    "INVALID_INPUT",
				Message: "Zone name is invalid",
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateZone(context.Background(), ZoneInput{Name: ""})

	if err == nil {
		t.Fatal("CreateZone() expected error, got nil")
	}
	if err.Error() != "API error: Zone name is invalid (code: INVALID_INPUT)" {
		t.Errorf("CreateZone() error = %v, want specific error message", err)
	}
}

// TestGetDNSSECStatus tests DNSSEC status retrieval
func TestGetDNSSECStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/zones/zone-id/dnssec/status" {
			t.Errorf("expected /v1/zones/zone-id/dnssec/status, got %s", r.URL.Path)
		}

		response := DNSSECStatusAPIResponse{
			Data: DNSSECStatus{
				Enabled:   true,
				KeysCount: 2,
				HasKSK:    true,
				HasZSK:    true,
			},
			Status: "success",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server)
	status, err := client.GetDNSSECStatus(context.Background(), "zone-id")

	if err != nil {
		t.Fatalf("GetDNSSECStatus() error = %v", err)
	}
	if !status.Enabled {
		t.Error("GetDNSSECStatus() Enabled = false, want true")
	}
	if status.KeysCount != 2 {
		t.Errorf("GetDNSSECStatus() KeysCount = %d, want 2", status.KeysCount)
	}
}

// TestCreateCryptokey tests cryptokey creation
func TestCreateCryptokey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/zones/zone-id/dnssec/cryptokeys" {
			t.Errorf("expected /v1/zones/zone-id/dnssec/cryptokeys, got %s", r.URL.Path)
		}

		response := CryptokeyAPIResponse{
			Data: struct {
				Cryptokey Cryptokey `json:"cryptokey"`
			}{
				Cryptokey: Cryptokey{
					ID:        1,
					KeyType:   "KSK",
					Algorithm: "ECDSAP256SHA256",
					Active:    true,
					Published: true,
					KeyTag:    12345,
				},
			},
			Status: "success",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server)
	active := true
	published := true
	key, err := client.CreateCryptokey(context.Background(), "zone-id", CryptokeyInput{
		KeyType:   "KSK",
		Algorithm: "ECDSAP256SHA256",
		Active:    &active,
		Published: &published,
	})

	if err != nil {
		t.Fatalf("CreateCryptokey() error = %v", err)
	}
	if key.ID != 1 {
		t.Errorf("CreateCryptokey() key.ID = %d, want 1", key.ID)
	}
	if key.KeyType != "KSK" {
		t.Errorf("CreateCryptokey() key.KeyType = %v, want KSK", key.KeyType)
	}
}

// TestContextCancellation tests that requests respect context cancellation
func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should never be reached if context is cancelled
		t.Error("request should have been cancelled")
	}))
	defer server.Close()

	client := newTestClient(server)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetZone(ctx, "zone-id")
	if err == nil {
		t.Error("GetZone() expected error due to cancelled context, got nil")
	}
}
