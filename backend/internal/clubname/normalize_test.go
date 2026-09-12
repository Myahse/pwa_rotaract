package clubname

import "testing"

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Rotaract IUGB Club", "iugb"},
		{"ROTARACT IUGB", "iugb"},
		{"UUGB", "iugb"},
		{"  rotaract-iugb  ", "iugb"},
	}
	for _, tc := range tests {
		if got := NormalizeKey(tc.in); got != tc.want {
			t.Fatalf("NormalizeKey(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
