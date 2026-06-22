package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDNSSECKeyTypeStateValue(t *testing.T) {
	tests := []struct {
		name       string
		configured types.String
		apiKeyType string
		want       string
	}{
		{
			name:       "preserves configured KSK when API returns CSK",
			configured: types.StringValue("KSK"),
			apiKeyType: "csk",
			want:       "KSK",
		},
		{
			name:       "preserves configured CSK",
			configured: types.StringValue("csk"),
			apiKeyType: "csk",
			want:       "CSK",
		},
		{
			name:       "uses API value for import",
			configured: types.StringNull(),
			apiKeyType: "csk",
			want:       "CSK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dnssecKeyTypeStateValue(tt.configured, tt.apiKeyType)
			if got.ValueString() != tt.want {
				t.Fatalf("dnssecKeyTypeStateValue() = %q, want %q", got.ValueString(), tt.want)
			}
		})
	}
}
