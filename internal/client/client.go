package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "https://api.dnscale.eu"
	DefaultTimeout = 30 * time.Second
)

// Client is the DNScale API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new DNScale API client
func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// doRequest performs an HTTP request and returns the response body
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("API error: %s (code: %s)", errResp.Error.Message, errResp.Error.Code)
		}
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Zone Operations

// CreateZone creates a new DNS zone
func (c *Client) CreateZone(ctx context.Context, input ZoneInput) (*Zone, error) {
	respBody, err := c.doRequest(ctx, http.MethodPost, "/v1/zones", input)
	if err != nil {
		return nil, err
	}

	var response ZoneAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal zone response: %w", err)
	}

	return &response.Data.Zone, nil
}

// GetZone retrieves a zone by ID
func (c *Client) GetZone(ctx context.Context, id string) (*Zone, error) {
	respBody, err := c.doRequest(ctx, http.MethodGet, "/v1/zones/"+id, nil)
	if err != nil {
		return nil, err
	}

	var response ZoneAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal zone response: %w", err)
	}

	return &response.Data.Zone, nil
}

// GetZoneByName retrieves a zone by name
func (c *Client) GetZoneByName(ctx context.Context, name string) (*Zone, error) {
	zones, err := c.ListZones(ctx)
	if err != nil {
		return nil, err
	}

	for _, zone := range zones {
		if zone.Name == name {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf("zone not found: %s", name)
}

// UpdateZone updates an existing zone
func (c *Client) UpdateZone(ctx context.Context, id string, input ZoneInput) (*Zone, error) {
	respBody, err := c.doRequest(ctx, http.MethodPut, "/v1/zones/"+id, input)
	if err != nil {
		return nil, err
	}

	var response ZoneAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal zone response: %w", err)
	}

	return &response.Data.Zone, nil
}

// DeleteZone deletes a zone
func (c *Client) DeleteZone(ctx context.Context, id string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, "/v1/zones/"+id, nil)
	return err
}

// ListZones lists all zones
func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	respBody, err := c.doRequest(ctx, http.MethodGet, "/v1/zones", nil)
	if err != nil {
		return nil, err
	}

	var response ZonesAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal zones response: %w", err)
	}

	if response.Data.Zones == nil {
		return []Zone{}, nil
	}
	return response.Data.Zones, nil
}

// Record Operations

// CreateRecord creates a new DNS record
func (c *Client) CreateRecord(ctx context.Context, zoneID string, input RecordInput) (*Record, error) {
	path := fmt.Sprintf("/v1/zones/%s/records", zoneID)
	respBody, err := c.doRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, err
	}

	var response RecordAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record response: %w", err)
	}

	return &response.Data.Record, nil
}

// GetRecord retrieves a record by ID
func (c *Client) GetRecord(ctx context.Context, zoneID, recordID string) (*Record, error) {
	path := fmt.Sprintf("/v1/zones/%s/records/%s", zoneID, recordID)
	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response RecordAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record response: %w", err)
	}

	return &response.Data.Record, nil
}

// UpdateRecord updates an existing record
func (c *Client) UpdateRecord(ctx context.Context, zoneID, recordID string, input RecordInput) (*Record, error) {
	path := fmt.Sprintf("/v1/zones/%s/records/%s", zoneID, recordID)
	respBody, err := c.doRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, err
	}

	var response RecordAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record response: %w", err)
	}

	return &response.Data.Record, nil
}

// DeleteRecord deletes a record
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	path := fmt.Sprintf("/v1/zones/%s/records/%s", zoneID, recordID)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

// ListRecords lists all records in a zone
func (c *Client) ListRecords(ctx context.Context, zoneID string) ([]Record, error) {
	var allRecords []Record
	offset := 0
	limit := 100

	for {
		path := fmt.Sprintf("/v1/zones/%s/records?offset=%d&limit=%d", zoneID, offset, limit)
		respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		var response RecordsAPIResponse
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal records response: %w", err)
		}

		if response.Data.Records != nil {
			allRecords = append(allRecords, response.Data.Records...)
		}

		if response.Data.Records == nil || len(response.Data.Records) < limit || offset+len(response.Data.Records) >= response.Data.Pagination.Total {
			break
		}
		offset += limit
	}

	return allRecords, nil
}

// DNSSEC Operations

// CreateCryptokey creates a new DNSSEC cryptokey
func (c *Client) CreateCryptokey(ctx context.Context, zoneID string, input CryptokeyInput) (*Cryptokey, error) {
	path := fmt.Sprintf("/v1/zones/%s/dnssec/cryptokeys", zoneID)
	respBody, err := c.doRequest(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, err
	}

	var response CryptokeyAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cryptokey response: %w", err)
	}

	return &response.Data.Cryptokey, nil
}

// GetCryptokey retrieves a cryptokey by ID
func (c *Client) GetCryptokey(ctx context.Context, zoneID string, keyID int) (*Cryptokey, error) {
	path := fmt.Sprintf("/v1/zones/%s/dnssec/cryptokeys/%d", zoneID, keyID)
	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response CryptokeyAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cryptokey response: %w", err)
	}

	return &response.Data.Cryptokey, nil
}

// UpdateCryptokey updates a cryptokey
func (c *Client) UpdateCryptokey(ctx context.Context, zoneID string, keyID int, input CryptokeyUpdate) (*Cryptokey, error) {
	path := fmt.Sprintf("/v1/zones/%s/dnssec/cryptokeys/%d", zoneID, keyID)
	respBody, err := c.doRequest(ctx, http.MethodPut, path, input)
	if err != nil {
		return nil, err
	}

	var response CryptokeyAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cryptokey response: %w", err)
	}

	return &response.Data.Cryptokey, nil
}

// DeleteCryptokey deletes a cryptokey
func (c *Client) DeleteCryptokey(ctx context.Context, zoneID string, keyID int) error {
	path := fmt.Sprintf("/v1/zones/%s/dnssec/%d", zoneID, keyID)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

// ListCryptokeys lists all cryptokeys in a zone
func (c *Client) ListCryptokeys(ctx context.Context, zoneID string) ([]Cryptokey, error) {
	path := fmt.Sprintf("/v1/zones/%s/dnssec/cryptokeys", zoneID)
	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response CryptokeysAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cryptokeys response: %w", err)
	}

	if response.Data.Cryptokeys == nil {
		return []Cryptokey{}, nil
	}
	return response.Data.Cryptokeys, nil
}

// GetDNSSECStatus gets the DNSSEC status of a zone
func (c *Client) GetDNSSECStatus(ctx context.Context, zoneID string) (*DNSSECStatus, error) {
	path := fmt.Sprintf("/v1/zones/%s/dnssec/status", zoneID)
	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response DNSSECStatusAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DNSSEC status response: %w", err)
	}

	return &response.Data, nil
}
