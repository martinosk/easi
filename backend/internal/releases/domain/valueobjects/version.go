package valueobjects

import (
	"fmt"
	"regexp"
)

var semverRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

type Version struct {
	value string
}

func NewVersion(value string) (Version, error) {
	if value == "" {
		return Version{}, fmt.Errorf("version cannot be empty")
	}
	if !semverRegex.MatchString(value) {
		return Version{}, fmt.Errorf("version must be in semver format (e.g., 1.0.0)")
	}
	return Version{value: value}, nil
}

func (v Version) Value() string {
	return v.value
}

func (v Version) String() string {
	return v.value
}

func (v Version) Equals(other Version) bool {
	return v.value == other.value
}
