package version

import "testing"

func TestIsVersionOutdated(t *testing.T) {
	cases := []struct {
		target   string
		latest   string
		outdated bool
	}{
		{"1.0.0", "2.0.0", true},
		{"1.0.0", "1.1.0", true},
		{"1.0.0", "1.0.1", true},
		{"1.8.1", "1.9.0", true},

		{"1.0.0", "1.0.0", false},
		{"2.0.0", "1.0.0", false},
		{"12.0.0", "2.0.0", false},
		{"undefined", "1.9.1", false},
	}

	for _, tc := range cases {
		if got := isVersionOutdated(tc.target, tc.latest); got != tc.outdated {
			t.Errorf("isVersionOutdated(%s, %s) = %v, want %v", tc.target, tc.latest, got, tc.outdated)
		}
	}
}
