package aggregates

import (
	"testing"
	"time"

	"easi/backend/internal/releases/domain/valueobjects"
)

func mustCreateVersion(t *testing.T, value string) valueobjects.Version {
	t.Helper()
	version, err := valueobjects.NewVersion(value)
	if err != nil {
		t.Fatalf("Failed to create version: %v", err)
	}
	return version
}

func TestNewRelease_CreatesReleaseWithCorrectValues(t *testing.T) {
	version := mustCreateVersion(t, "1.0.0")
	releaseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	notes := "## Features\n- New feature"

	release := NewRelease(version, releaseDate, notes)

	if !release.Version().Equals(version) {
		t.Errorf("Release.Version() = %v, want %v", release.Version(), version)
	}
	if !release.ReleaseDate().Equal(releaseDate) {
		t.Errorf("Release.ReleaseDate() = %v, want %v", release.ReleaseDate(), releaseDate)
	}
	if release.Notes() != notes {
		t.Errorf("Release.Notes() = %q, want %q", release.Notes(), notes)
	}
}

func TestNewRelease_SetsCreatedAtToNow(t *testing.T) {
	version := mustCreateVersion(t, "1.0.0")
	releaseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	beforeCreation := time.Now()

	release := NewRelease(version, releaseDate, "notes")

	afterCreation := time.Now()
	if release.CreatedAt().Before(beforeCreation) || release.CreatedAt().After(afterCreation) {
		t.Errorf("Release.CreatedAt() = %v, expected between %v and %v",
			release.CreatedAt(), beforeCreation, afterCreation)
	}
}

func TestReconstructRelease_RestoresAllFields(t *testing.T) {
	version := mustCreateVersion(t, "2.1.0")
	releaseDate := time.Date(2024, 3, 20, 10, 30, 0, 0, time.UTC)
	notes := "## Bug Fixes\n- Fixed bug"
	createdAt := time.Date(2024, 3, 15, 8, 0, 0, 0, time.UTC)

	release := ReconstructRelease(version, releaseDate, notes, createdAt)

	if !release.Version().Equals(version) {
		t.Errorf("Reconstructed Release.Version() = %v, want %v", release.Version(), version)
	}
	if !release.ReleaseDate().Equal(releaseDate) {
		t.Errorf("Reconstructed Release.ReleaseDate() = %v, want %v", release.ReleaseDate(), releaseDate)
	}
	if release.Notes() != notes {
		t.Errorf("Reconstructed Release.Notes() = %q, want %q", release.Notes(), notes)
	}
	if !release.CreatedAt().Equal(createdAt) {
		t.Errorf("Reconstructed Release.CreatedAt() = %v, want %v", release.CreatedAt(), createdAt)
	}
}

func TestRelease_UpdateNotes_ChangesNotes(t *testing.T) {
	version := mustCreateVersion(t, "1.0.0")
	releaseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	originalNotes := "Original notes"
	newNotes := "## Updated Features\n- Updated feature description"

	release := NewRelease(version, releaseDate, originalNotes)

	release.UpdateNotes(newNotes)

	if release.Notes() != newNotes {
		t.Errorf("After UpdateNotes, Release.Notes() = %q, want %q", release.Notes(), newNotes)
	}
}

func TestRelease_UpdateNotes_DoesNotAffectOtherFields(t *testing.T) {
	version := mustCreateVersion(t, "1.0.0")
	releaseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	originalNotes := "Original notes"

	release := NewRelease(version, releaseDate, originalNotes)
	originalCreatedAt := release.CreatedAt()

	release.UpdateNotes("New notes")

	if !release.Version().Equals(version) {
		t.Error("UpdateNotes should not change version")
	}
	if !release.ReleaseDate().Equal(releaseDate) {
		t.Error("UpdateNotes should not change release date")
	}
	if !release.CreatedAt().Equal(originalCreatedAt) {
		t.Error("UpdateNotes should not change created at")
	}
}

func TestRelease_UpdateNotes_AllowsEmptyNotes(t *testing.T) {
	version := mustCreateVersion(t, "1.0.0")
	releaseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	release := NewRelease(version, releaseDate, "Some notes")

	release.UpdateNotes("")

	if release.Notes() != "" {
		t.Errorf("After UpdateNotes(\"\"), Release.Notes() = %q, want empty string", release.Notes())
	}
}

func TestRelease_GettersMaintainImmutability(t *testing.T) {
	version := mustCreateVersion(t, "1.0.0")
	releaseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	notes := "Original notes"

	release := NewRelease(version, releaseDate, notes)

	v1 := release.Version()
	v2 := release.Version()

	if !v1.Equals(v2) {
		t.Error("Multiple calls to Version() should return equal values")
	}

	rd1 := release.ReleaseDate()
	rd2 := release.ReleaseDate()

	if !rd1.Equal(rd2) {
		t.Error("Multiple calls to ReleaseDate() should return equal values")
	}
}

func TestRelease_WithMarkdownNotes(t *testing.T) {
	version := mustCreateVersion(t, "2.0.0")
	releaseDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	markdownNotes := `## Major Features
- **New Dashboard**: Complete redesign of the main dashboard
- Added ` + "`dark mode`" + ` support

## Bug Fixes
- Fixed authentication timeout issue
- Resolved *memory leak* in data processing

## API Changes
- New endpoint for batch operations`

	release := NewRelease(version, releaseDate, markdownNotes)

	if release.Notes() != markdownNotes {
		t.Errorf("Release should preserve markdown notes exactly")
	}
}

func TestRelease_WithEmptyNotes(t *testing.T) {
	version := mustCreateVersion(t, "1.0.0")
	releaseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	release := NewRelease(version, releaseDate, "")

	if release.Notes() != "" {
		t.Errorf("Release.Notes() = %q, want empty string", release.Notes())
	}
}
