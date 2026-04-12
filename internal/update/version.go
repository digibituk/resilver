package update

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a parsed semantic version.
type Version struct {
	Major, Minor, Patch int
}

// ParseVersion parses a version string like "v1.2.3" or "1.2.3".
// Returns an error for non-clean semver strings (e.g. "dev", "v1.2.3-4-gabcdef").
func ParseVersion(s string) (Version, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version: %q", s)
	}
	var v Version
	var err error
	v.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid version: %q", s)
	}
	v.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid version: %q", s)
	}
	v.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid version: %q", s)
	}
	return v, nil
}

// NewerThan returns true if v is strictly newer than other.
func (v Version) NewerThan(other Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

// String returns the version as "v1.2.3".
func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}
