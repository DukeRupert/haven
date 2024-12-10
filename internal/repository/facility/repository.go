// internal/repository/facility/repository.go
package facility

import (
	"context"
	"fmt"
	"time"

	"github.com/DukeRupert/haven/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles facility-related database operations
type Repository struct {
	pool *pgxpool.Pool
}

// New creates a new facility repository
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// Custom errors
var (
	ErrDuplicateCode = fmt.Errorf("facility code already exists")
	ErrNotFound      = fmt.Errorf("facility not found")
)

func (r *Repository) List(ctx context.Context) ([]model.Facility, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, created_at, name, code
        FROM facilities
        ORDER BY name ASC
    `)
	if err != nil {
		return nil, fmt.Errorf("error listing facilities: %w", err)
	}
	defer rows.Close()

	var facilities []model.Facility
	for rows.Next() {
		var f model.Facility
		err := rows.Scan(
			&f.ID,
			&f.CreatedAt,
			&f.Name,
			&f.Code,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning facility row: %w", err)
		}
		facilities = append(facilities, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating facility rows: %w", err)
	}

	return facilities, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (*model.Facility, error) {
	var f model.Facility
	err := r.pool.QueryRow(ctx, `
        SELECT id, created_at, updated_at, name, code
        FROM facilities
        WHERE id = $1
    `, id).Scan(
		&f.ID,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.Name,
		&f.Code,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("facility not found with id %d", id)
		}
		return nil, fmt.Errorf("error getting facility by id: %w", err)
	}
	return &f, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*model.Facility, error) {
	var f model.Facility
	err := r.pool.QueryRow(ctx, `
        SELECT id, created_at, updated_at, name, code
        FROM facilities
        WHERE code = $1
    `, code).Scan(
		&f.ID,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.Name,
		&f.Code,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting facility by code: %w", err)
	}
	return &f, nil
}

func (r *Repository) Create(ctx context.Context, params model.CreateFacilityParams) (*model.Facility, error) {
	// Check for unique code first
	isUnique, err := r.IsCodeUnique(ctx, params.Code, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking facility code uniqueness: %w", err)
	}
	if !isUnique {
		return nil, ErrDuplicateCode
	}

	var f model.Facility
	now := time.Now()
	err = r.pool.QueryRow(ctx, `
		INSERT INTO facilities (created_at, updated_at, name, code)  -- Added updated_at
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at, name, code
	`, now, now, params.Name, params.Code).Scan(
		&f.ID,
		&f.CreatedAt,
		&f.UpdatedAt, // Add this field to match RETURNING clause
		&f.Name,
		&f.Code,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating facility: %w", err)
	}
	return &f, nil
}

func (r *Repository) UpdateFacility(ctx context.Context, id int, params model.UpdateFacilityParams) (*model.Facility, error) {
	// Check for unique code first, excluding the current facility ID
	isUnique, err := r.IsCodeUnique(ctx, params.Code, &id)
	if err != nil {
		return nil, fmt.Errorf("error checking facility code uniqueness: %w", err)
	}
	if !isUnique {
		return nil, ErrDuplicateCode
	}

	var f model.Facility
	now := time.Now()

	err = r.pool.QueryRow(ctx, `
        UPDATE facilities
        SET updated_at = $1, name = $2, code = $3
        WHERE id = $4
        RETURNING id, created_at, updated_at, name, code
    `, now, params.Name, params.Code, id).Scan(
		&f.ID,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.Name,
		&f.Code,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error updating facility: %w", err)
	}

	return &f, nil
}

func (r *Repository) IsCodeUnique(ctx context.Context, code string, excludeID *int) (bool, error) {
	// Build query with optional ID exclusion
	// $2::int is used to explicitly cast the nullable ID parameter
	// to ensure proper type handling in PostgreSQL.
	query := `
        SELECT NOT EXISTS (
            SELECT 1 FROM facilities 
            WHERE code = $1
            AND ($2::int IS NULL OR id != $2)
        )`

	var isUnique bool
	err := r.pool.QueryRow(ctx, query, code, excludeID).Scan(&isUnique)
	if err != nil {
		return false, fmt.Errorf("checking code uniqueness: %w", err)
	}

	return isUnique, nil
}
