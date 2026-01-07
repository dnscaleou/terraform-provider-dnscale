package client

import "time"

// Zone represents a DNS zone
type Zone struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Region     string    `json:"region"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ZoneInput represents the input for creating/updating a zone
type ZoneInput struct {
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Region string `json:"region,omitempty"`
	Status string `json:"status,omitempty"`
}

// ZonesResponse represents the response from listing zones
type ZonesResponse struct {
	Zones      []Zone     `json:"zones"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination info
type Pagination struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Count  int `json:"count"`
}

// Record represents a DNS record
type Record struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Priority *int   `json:"priority,omitempty"`
	Disabled bool   `json:"disabled"`
}

// RecordInput represents the input for creating/updating a record
type RecordInput struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl,omitempty"`
	Priority *int   `json:"priority,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

// RecordsResponse represents the response from listing records
type RecordsResponse struct {
	Records    []Record   `json:"records"`
	Pagination Pagination `json:"pagination"`
}

// Cryptokey represents a DNSSEC cryptokey
type Cryptokey struct {
	ID        int      `json:"id"`
	KeyTag    int      `json:"keyTag"`
	Active    bool     `json:"active"`
	Published bool     `json:"published"`
	KeyType   string   `json:"keytype"`
	Algorithm string   `json:"algorithm"`
	Bits      int      `json:"bits"`
	DNSKey    string   `json:"dnskey"`
	DS        []string `json:"ds"`
}

// CryptokeyInput represents the input for creating a cryptokey
type CryptokeyInput struct {
	KeyType   string `json:"keytype"`
	Algorithm string `json:"algorithm,omitempty"`
	Bits      *int   `json:"bits,omitempty"`
	Active    *bool  `json:"active,omitempty"`
	Published *bool  `json:"published,omitempty"`
}

// CryptokeyUpdate represents the input for updating a cryptokey
type CryptokeyUpdate struct {
	Active    *bool `json:"active,omitempty"`
	Published *bool `json:"published,omitempty"`
}

// DNSSECStatus represents the DNSSEC status of a zone
type DNSSECStatus struct {
	Enabled   bool `json:"enabled"`
	KeysCount int  `json:"keys_count"`
	HasKSK    bool `json:"has_ksk"`
	HasZSK    bool `json:"has_zsk"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// API Response wrapper types - the API wraps all responses in a data envelope

// ZoneAPIResponse wraps a single zone response
type ZoneAPIResponse struct {
	Data struct {
		Zone Zone `json:"zone"`
	} `json:"data"`
	Status string `json:"status"`
}

// ZonesAPIResponse wraps a zones list response
type ZonesAPIResponse struct {
	Data struct {
		Zones      []Zone     `json:"zones"`
		Pagination Pagination `json:"pagination"`
	} `json:"data"`
	Status string `json:"status"`
}

// RecordAPIResponse wraps a single record response
type RecordAPIResponse struct {
	Data struct {
		Record Record `json:"record"`
	} `json:"data"`
	Status string `json:"status"`
}

// RecordsAPIResponse wraps a records list response
type RecordsAPIResponse struct {
	Data struct {
		Records    []Record   `json:"records"`
		Pagination Pagination `json:"pagination"`
	} `json:"data"`
	Status string `json:"status"`
}

// CryptokeyAPIResponse wraps a single cryptokey response
type CryptokeyAPIResponse struct {
	Data struct {
		Cryptokey Cryptokey `json:"cryptokey"`
	} `json:"data"`
	Status string `json:"status"`
}

// CryptokeysAPIResponse wraps a cryptokeys list response
type CryptokeysAPIResponse struct {
	Data struct {
		Cryptokeys []Cryptokey `json:"cryptokeys"`
	} `json:"data"`
	Status string `json:"status"`
}

// DNSSECStatusAPIResponse wraps a DNSSEC status response
type DNSSECStatusAPIResponse struct {
	Data   DNSSECStatus `json:"data"`
	Status string       `json:"status"`
}
