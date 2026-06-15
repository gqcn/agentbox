// This file verifies provider projection helpers that do not require database
// state.

package catalog

import "testing"

// TestMaskSecret verifies provider API keys are not exposed verbatim.
func TestMaskSecret(t *testing.T) {
	cases := map[string]string{
		"":                 "",
		"short":            "****",
		"sk-1234567890":    "sk-1****7890",
		"  sk-abcdefghi  ": "sk-a****fghi",
	}
	for input, expected := range cases {
		if actual := maskSecret(input); actual != expected {
			t.Fatalf("maskSecret(%q)=%q, want %q", input, actual, expected)
		}
	}
}
