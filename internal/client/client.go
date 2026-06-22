package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// APIError captures structured DNScale API errors.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}

	if e.Message != "" && e.Code != "" {
		return fmt.Sprintf("API error: %s (code: %s)", e.Message, e.Code)
	}
	if e.Message != "" {
		return fmt.Sprintf("API error: %s (status: %d)", e.Message, e.StatusCode)
	}
	if strings.TrimSpace(e.Body) != "" {
		return fmt.Sprintf("API error: status %d, body: %s", e.StatusCode, strings.TrimSpace(e.Body))
	}
	return fmt.Sprintf("API error: status %d", e.StatusCode)
}

// IsNotFound returns true when an error represents a missing remote resource.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	if apiErr.StatusCode == http.StatusNotFound {
		return true
	}

	code := strings.ToUpper(apiErr.Code)
	return strings.Contains(code, "NOT_FOUND")
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

func (c *Client) requestURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(c.BaseURL, "/") + path
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

	req, err := http.NewRequestWithContext(ctx, method, c.requestURL(path), reqBody)
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
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Code:       errResp.Error.Code,
				Message:    errResp.Error.Message,
				Body:       string(respBody),
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	return respBody, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	respBody, err := c.doRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	if out == nil || len(strings.TrimSpace(string(respBody))) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

// Zone Operations

// CreateZone creates a new DNS zone
func (c *Client) CreateZone(ctx context.Context, input ZoneInput) (*Zone, error) {
	var response ZoneAPIResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/zones", input, &response); err != nil {
		return nil, err
	}

	return &response.Data.Zone, nil
}

// GetZone retrieves a zone by ID
func (c *Client) GetZone(ctx context.Context, id string) (*Zone, error) {
	var response ZoneAPIResponse
	if err := c.doJSON(ctx, http.MethodGet, "/v1/zones/"+id, nil, &response); err != nil {
		return nil, err
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
	var response ZoneAPIResponse
	if err := c.doJSON(ctx, http.MethodPut, "/v1/zones/"+id, input, &response); err != nil {
		return nil, err
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
	var allZones []Zone
	offset := 0
	limit := 100

	for {
		path := fmt.Sprintf("/v1/zones?offset=%d&limit=%d", offset, limit)
		var response ZonesAPIResponse
		if err := c.doJSON(ctx, http.MethodGet, path, nil, &response); err != nil {
			return nil, err
		}

		if response.Data.Zones != nil {
			allZones = append(allZones, response.Data.Zones...)
		}

		pageCount := len(response.Data.Zones)
		if pageCount == 0 {
			break
		}
		if response.Data.Pagination.HasMore != nil {
			if !*response.Data.Pagination.HasMore {
				break
			}
			offset += pageCount
			continue
		}
		if response.Data.Pagination.Total > 0 && offset+pageCount >= response.Data.Pagination.Total {
			break
		}
		if pageCount < limit {
			break
		}
		offset += pageCount
	}

	if allZones == nil {
		return []Zone{}, nil
	}
	return allZones, nil
}

// Record Operations

// CreateRecord creates a new DNS record
func (c *Client) CreateRecord(ctx context.Context, zoneID string, input RecordInput) (*Record, error) {
	path := fmt.Sprintf("/v1/zones/%s/records", zoneID)
	var response RecordAPIResponse
	if err := c.doJSON(ctx, http.MethodPost, path, input, &response); err != nil {
		return nil, err
	}

	return &response.Data.Record, nil
}

// GetRecord retrieves a record by ID
func (c *Client) GetRecord(ctx context.Context, zoneID, recordID string) (*Record, error) {
	path := fmt.Sprintf("/v1/zones/%s/records/%s", zoneID, recordID)
	var response RecordAPIResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}

	return &response.Data.Record, nil
}

// UpdateRecord updates an existing record
func (c *Client) UpdateRecord(ctx context.Context, zoneID, recordID string, input RecordInput) (*Record, error) {
	path := fmt.Sprintf("/v1/zones/%s/records/%s", zoneID, recordID)
	var response RecordAPIResponse
	if err := c.doJSON(ctx, http.MethodPut, path, input, &response); err != nil {
		return nil, err
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
		var response RecordsAPIResponse
		if err := c.doJSON(ctx, http.MethodGet, path, nil, &response); err != nil {
			return nil, err
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
	var response CryptokeyAPIResponse
	if err := c.doJSON(ctx, http.MethodPost, path, input, &response); err != nil {
		return nil, err
	}

	return &response.Data.Cryptokey, nil
}

// GetCryptokey retrieves a cryptokey by ID
func (c *Client) GetCryptokey(ctx context.Context, zoneID string, keyID int) (*Cryptokey, error) {
	path := fmt.Sprintf("/v1/zones/%s/dnssec/cryptokeys/%d", zoneID, keyID)
	var response CryptokeyAPIResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
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

	if len(strings.TrimSpace(string(respBody))) > 0 {
		var response CryptokeyAPIResponse
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cryptokey response: %w", err)
		}
		return &response.Data.Cryptokey, nil
	}

	return c.GetCryptokey(ctx, zoneID, keyID)
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
	var response CryptokeysAPIResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}

	if response.Data.Cryptokeys == nil {
		return []Cryptokey{}, nil
	}
	return response.Data.Cryptokeys, nil
}

// GetDNSSECStatus gets the DNSSEC status of a zone
func (c *Client) GetDNSSECStatus(ctx context.Context, zoneID string) (*DNSSECStatus, error) {
	path := fmt.Sprintf("/v1/zones/%s/dnssec/status", zoneID)
	var response DNSSECStatusAPIResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}

	status := &DNSSECStatus{
		Enabled: response.Data.EnabledValue(),
	}

	keys, err := c.ListCryptokeys(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNSSEC keys while building DNSSEC status: %w", err)
	}

	status.KeysCount = len(keys)
	for _, key := range keys {
		switch strings.ToUpper(key.KeyType) {
		case "KSK":
			status.HasKSK = true
		case "ZSK":
			status.HasZSK = true
		case "CSK":
			status.HasKSK = true
			status.HasZSK = true
		}
	}

	return status, nil
}
