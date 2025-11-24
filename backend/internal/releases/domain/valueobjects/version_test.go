package valueobjects

import (
	"testing"
)

func TestNewVersion_ValidVersions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
	}{
		{
			name:  "standard semver version",
			input: "1.0.0",
			want:  "1.0.0",
		},
		{
			name:  "version with double digits",
			input: "10.20.30",
			want:  "10.20.30",
		},
		{
			name:  "version with zeros",
			input: "0.0.0",
			want:  "0.0.0",
		},
		{
			name:  "version with large numbers",
			input: "100.200.300",
			want:  "100.200.300",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := NewVersion(tt.input)

			if err != nil {
				t.Errorf("NewVersion(%q) unexpected error: %v", tt.input, err)
				return
			}
			if version.Value() != tt.want {
				t.Errorf("NewVersion(%q).Value() = %q, want %q", tt.input, version.Value(), tt.want)
			}
		})
	}
}

func TestNewVersion_InvalidVersions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: "version cannot be empty",
		},
		{
			name:    "missing patch version",
			input:   "1.0",
			wantErr: "version must be in semver format",
		},
		{
			name:    "missing minor version",
			input:   "1",
			wantErr: "version must be in semver format",
		},
		{
			name:    "version with v prefix",
			input:   "v1.0.0",
			wantErr: "version must be in semver format",
		},
		{
			name:    "version with prerelease",
			input:   "1.0.0-alpha",
			wantErr: "version must be in semver format",
		},
		{
			name:    "version with build metadata",
			input:   "1.0.0+build",
			wantErr: "version must be in semver format",
		},
		{
			name:    "version with letters",
			input:   "a.b.c",
			wantErr: "version must be in semver format",
		},
		{
			name:    "version with extra parts",
			input:   "1.0.0.0",
			wantErr: "version must be in semver format",
		},
		{
			name:    "version with negative numbers",
			input:   "-1.0.0",
			wantErr: "version must be in semver format",
		},
		{
			name:    "version with spaces",
			input:   "1. 0.0",
			wantErr: "version must be in semver format",
		},
		{
			name:    "only dots",
			input:   "..",
			wantErr: "version must be in semver format",
		},
		{
			name:    "random string",
			input:   "not-a-version",
			wantErr: "version must be in semver format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewVersion(tt.input)

			if err == nil {
				t.Errorf("NewVersion(%q) expected error, got nil", tt.input)
				return
			}
			if err.Error() != tt.wantErr && !contains(err.Error(), tt.wantErr) {
				t.Errorf("NewVersion(%q) error = %q, want to contain %q", tt.input, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	version, err := NewVersion("1.2.3")
	if err != nil {
		t.Fatalf("NewVersion failed: %v", err)
	}

	if version.String() != "1.2.3" {
		t.Errorf("Version.String() = %q, want %q", version.String(), "1.2.3")
	}
}

func TestVersion_Equals(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		want     bool
	}{
		{
			name:     "equal versions",
			version1: "1.0.0",
			version2: "1.0.0",
			want:     true,
		},
		{
			name:     "different major",
			version1: "1.0.0",
			version2: "2.0.0",
			want:     false,
		},
		{
			name:     "different minor",
			version1: "1.0.0",
			version2: "1.1.0",
			want:     false,
		},
		{
			name:     "different patch",
			version1: "1.0.0",
			version2: "1.0.1",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := NewVersion(tt.version1)
			if err != nil {
				t.Fatalf("NewVersion(%q) failed: %v", tt.version1, err)
			}

			v2, err := NewVersion(tt.version2)
			if err != nil {
				t.Fatalf("NewVersion(%q) failed: %v", tt.version2, err)
			}

			if v1.Equals(v2) != tt.want {
				t.Errorf("Version(%q).Equals(Version(%q)) = %v, want %v",
					tt.version1, tt.version2, v1.Equals(v2), tt.want)
			}
		})
	}
}

func TestVersion_ImmutabilityOnEquality(t *testing.T) {
	v1, _ := NewVersion("1.0.0")
	v2, _ := NewVersion("1.0.0")

	if !v1.Equals(v2) {
		t.Error("Two versions with the same value should be equal")
	}

	if v1.Value() != v2.Value() {
		t.Error("Two versions with the same value should have the same Value()")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
