package update

import "testing"

func TestParseVersionCleanTag(t *testing.T) {
	v, err := ParseVersion("v1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Fatalf("got %d.%d.%d, want 1.2.3", v.Major, v.Minor, v.Patch)
	}
}

func TestParseVersionNoPrefix(t *testing.T) {
	v, err := ParseVersion("0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Major != 0 || v.Minor != 1 || v.Patch != 0 {
		t.Fatalf("got %d.%d.%d, want 0.1.0", v.Major, v.Minor, v.Patch)
	}
}

func TestParseVersionInvalid(t *testing.T) {
	cases := []string{"dev", "v1.2.3-4-gabcdef", "", "v1.2", "abc", "v1.2.3-rc1"}
	for _, c := range cases {
		if _, err := ParseVersion(c); err == nil {
			t.Errorf("expected error for %q", c)
		}
	}
}

func TestNewerThan(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"v1.1.0", "v1.0.0", true},
		{"v2.0.0", "v1.9.9", true},
		{"v1.0.1", "v1.0.0", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v1.1.0", false},
		{"v0.0.1", "v0.0.2", false},
	}
	for _, c := range cases {
		a, _ := ParseVersion(c.a)
		b, _ := ParseVersion(c.b)
		if got := a.NewerThan(b); got != c.want {
			t.Errorf("%s.NewerThan(%s) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestVersionString(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	if got := v.String(); got != "v1.2.3" {
		t.Fatalf("got %q, want %q", got, "v1.2.3")
	}
}
