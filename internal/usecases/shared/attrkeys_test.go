package shared

import "testing"

// TC-UC-19: shared semantic attribute-key constants must hold the canonical
// OpenTelemetry-style values consumed by every use case.
func TestAttrKeyConstants(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"AttrKeyAppResult", AttrKeyAppResult, "app.result"},
		{"AttrKeyAppValidationError", AttrKeyAppValidationError, "app.validation_error"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("%s = %q, want %q", tc.name, tc.got, tc.want)
			}
		})
	}
}
