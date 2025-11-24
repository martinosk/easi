package repository

import (
	"context"
	"database/sql"
	"time"

	"easi/backend/internal/releases/domain"
	"easi/backend/internal/releases/domain/aggregates"
	"easi/backend/internal/releases/domain/valueobjects"
	sharedDomain "easi/backend/internal/shared/domain"
)

type PostgresReleaseRepository struct {
	db *sql.DB
}

func NewPostgresReleaseRepository(db *sql.DB) domain.ReleaseRepository {
	return &PostgresReleaseRepository{db: db}
}

func (r *PostgresReleaseRepository) Save(ctx context.Context, release *aggregates.Release) error {
	query := `
		INSERT INTO releases (version, release_date, notes, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (version) DO UPDATE SET
			release_date = EXCLUDED.release_date,
			notes = EXCLUDED.notes
	`
	_, err := r.db.ExecContext(ctx, query,
		release.Version().Value(),
		release.ReleaseDate(),
		release.Notes(),
		release.CreatedAt(),
	)
	return err
}

func (r *PostgresReleaseRepository) FindByVersion(ctx context.Context, version valueobjects.Version) (*aggregates.Release, error) {
	query := `SELECT version, release_date, notes, created_at FROM releases WHERE version = $1`
	row := r.db.QueryRowContext(ctx, query, version.Value())
	return r.scanRelease(row)
}

func (r *PostgresReleaseRepository) FindLatest(ctx context.Context) (*aggregates.Release, error) {
	query := `SELECT version, release_date, notes, created_at FROM releases ORDER BY release_date DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query)
	return r.scanRelease(row)
}

func (r *PostgresReleaseRepository) FindAll(ctx context.Context) ([]*aggregates.Release, error) {
	query := `SELECT version, release_date, notes, created_at FROM releases ORDER BY release_date DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var releases []*aggregates.Release
	for rows.Next() {
		release, err := scanReleaseFromScanner(rows)
		if err != nil {
			return nil, err
		}
		releases = append(releases, release)
	}
	return releases, rows.Err()
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (r *PostgresReleaseRepository) scanRelease(row *sql.Row) (*aggregates.Release, error) {
	release, err := scanReleaseFromScanner(row)
	if err == sql.ErrNoRows {
		return nil, sharedDomain.ErrNotFound
	}
	return release, err
}

func scanReleaseFromScanner(s scanner) (*aggregates.Release, error) {
	var versionStr string
	var releaseDate time.Time
	var notes string
	var createdAt time.Time

	if err := s.Scan(&versionStr, &releaseDate, &notes, &createdAt); err != nil {
		return nil, err
	}

	version, err := valueobjects.NewVersion(versionStr)
	if err != nil {
		return nil, err
	}

	return aggregates.ReconstructRelease(version, releaseDate, notes, createdAt), nil
}
