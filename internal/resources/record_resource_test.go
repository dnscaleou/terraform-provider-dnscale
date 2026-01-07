package resources

import "testing"

// TestNormalizeRecordContent tests the normalizeRecordContent function
func TestNormalizeRecordContent(t *testing.T) {
	tests := []struct {
		name       string
		recordType string
		content    string
		priority   *int
		want       string
	}{
		// TXT record tests
		{
			name:       "TXT with double quotes",
			recordType: "TXT",
			content:    "\"v=spf1 -all\"",
			priority:   nil,
			want:       "v=spf1 -all",
		},
		{
			name:       "TXT with single quotes",
			recordType: "TXT",
			content:    "'v=spf1 -all'",
			priority:   nil,
			want:       "v=spf1 -all",
		},
		{
			name:       "TXT without quotes",
			recordType: "TXT",
			content:    "v=spf1 -all",
			priority:   nil,
			want:       "v=spf1 -all",
		},
		{
			name:       "TXT empty string",
			recordType: "TXT",
			content:    "",
			priority:   nil,
			want:       "",
		},
		{
			name:       "TXT single char",
			recordType: "TXT",
			content:    "a",
			priority:   nil,
			want:       "a",
		},
		{
			name:       "TXT only quotes",
			recordType: "TXT",
			content:    "\"\"",
			priority:   nil,
			want:       "",
		},
		{
			name:       "TXT DKIM key with quotes",
			recordType: "TXT",
			content:    "\"v=DKIM1; k=rsa; p=MIGfMA0...\"",
			priority:   nil,
			want:       "v=DKIM1; k=rsa; p=MIGfMA0...",
		},

		// MX record tests
		{
			name:       "MX with priority prefix",
			recordType: "MX",
			content:    "10 mail.example.com.",
			priority:   intPtr(10),
			want:       "mail.example.com.",
		},
		{
			name:       "MX with zero priority",
			recordType: "MX",
			content:    "0 mail.example.com.",
			priority:   intPtr(0),
			want:       "mail.example.com.",
		},
		{
			name:       "MX with high priority",
			recordType: "MX",
			content:    "99 backup-mx.example.com.",
			priority:   intPtr(99),
			want:       "backup-mx.example.com.",
		},
		{
			name:       "MX without priority prefix",
			recordType: "MX",
			content:    "mail.example.com.",
			priority:   nil,
			want:       "mail.example.com.",
		},
		{
			name:       "MX with non-numeric prefix",
			recordType: "MX",
			content:    "abc mail.example.com.",
			priority:   nil,
			want:       "abc mail.example.com.",
		},
		{
			name:       "MX empty content",
			recordType: "MX",
			content:    "",
			priority:   nil,
			want:       "",
		},

		// A record tests (should not be modified)
		{
			name:       "A record",
			recordType: "A",
			content:    "192.0.2.1",
			priority:   nil,
			want:       "192.0.2.1",
		},

		// AAAA record tests (should not be modified)
		{
			name:       "AAAA record",
			recordType: "AAAA",
			content:    "2001:db8::1",
			priority:   nil,
			want:       "2001:db8::1",
		},

		// CNAME record tests (should not be modified)
		{
			name:       "CNAME record",
			recordType: "CNAME",
			content:    "example.com.",
			priority:   nil,
			want:       "example.com.",
		},

		// SRV record tests (should not be modified - weight/port/target are part of content)
		{
			name:       "SRV record",
			recordType: "SRV",
			content:    "10 20 443 server.example.com.",
			priority:   intPtr(10),
			want:       "10 20 443 server.example.com.",
		},

		// NS record tests (should not be modified)
		{
			name:       "NS record",
			recordType: "NS",
			content:    "ns1.example.com.",
			priority:   nil,
			want:       "ns1.example.com.",
		},

		// CAA record tests (should not be modified)
		{
			name:       "CAA record",
			recordType: "CAA",
			content:    "0 issue \"letsencrypt.org\"",
			priority:   nil,
			want:       "0 issue \"letsencrypt.org\"",
		},

		// PTR record tests (should not be modified)
		{
			name:       "PTR record",
			recordType: "PTR",
			content:    "host.example.com.",
			priority:   nil,
			want:       "host.example.com.",
		},

		// ALIAS record tests (should not be modified)
		{
			name:       "ALIAS record",
			recordType: "ALIAS",
			content:    "example.com.",
			priority:   nil,
			want:       "example.com.",
		},

		// SVCB record tests (should not be modified)
		{
			name:       "SVCB record",
			recordType: "SVCB",
			content:    "1 api.example.com.",
			priority:   nil,
			want:       "1 api.example.com.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRecordContent(tt.recordType, tt.content, tt.priority)
			if got != tt.want {
				t.Errorf("normalizeRecordContent(%q, %q) = %q, want %q",
					tt.recordType, tt.content, got, tt.want)
			}
		})
	}
}

// intPtr is a helper to create *int values for tests
func intPtr(i int) *int {
	return &i
}

// TestNormalizeRecordContent_EdgeCases tests edge cases for record normalization
func TestNormalizeRecordContent_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		recordType string
		content    string
		want       string
	}{
		{
			name:       "TXT with nested quotes",
			recordType: "TXT",
			content:    "\"hello \"world\"\"",
			want:       "hello \"world\"",
		},
		{
			name:       "MX with multiple spaces",
			recordType: "MX",
			content:    "10  mail.example.com.",
			want:       " mail.example.com.",
		},
		{
			name:       "TXT with only opening quote",
			recordType: "TXT",
			content:    "\"hello",
			want:       "\"hello",
		},
		{
			name:       "TXT with mismatched quotes",
			recordType: "TXT",
			content:    "\"hello'",
			want:       "\"hello'",
		},
		{
			name:       "MX with space only",
			recordType: "MX",
			content:    " ",
			want:       " ",
		},
		{
			name:       "Unknown record type",
			recordType: "UNKNOWN",
			content:    "some value",
			want:       "some value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRecordContent(tt.recordType, tt.content, nil)
			if got != tt.want {
				t.Errorf("normalizeRecordContent(%q, %q) = %q, want %q",
					tt.recordType, tt.content, got, tt.want)
			}
		})
	}
}
