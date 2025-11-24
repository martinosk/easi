package aggregates

import (
	"time"

	"easi/backend/internal/releases/domain/valueobjects"
)

type Release struct {
	version     valueobjects.Version
	releaseDate time.Time
	notes       string
	createdAt   time.Time
}

func NewRelease(version valueobjects.Version, releaseDate time.Time, notes string) *Release {
	return &Release{
		version:     version,
		releaseDate: releaseDate,
		notes:       notes,
		createdAt:   time.Now(),
	}
}

func ReconstructRelease(version valueobjects.Version, releaseDate time.Time, notes string, createdAt time.Time) *Release {
	return &Release{
		version:     version,
		releaseDate: releaseDate,
		notes:       notes,
		createdAt:   createdAt,
	}
}

func (r *Release) Version() valueobjects.Version {
	return r.version
}

func (r *Release) ReleaseDate() time.Time {
	return r.releaseDate
}

func (r *Release) Notes() string {
	return r.notes
}

func (r *Release) CreatedAt() time.Time {
	return r.createdAt
}

func (r *Release) UpdateNotes(notes string) {
	r.notes = notes
}
