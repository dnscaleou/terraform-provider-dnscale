package provider

import (
	"testing"
)

// TestProvider_Instantiation tests that the provider can be instantiated
func TestProvider_Instantiation(t *testing.T) {
	t.Parallel()

	p := New("test")()
	if p == nil {
		t.Fatal("provider should not be nil")
	}
}
