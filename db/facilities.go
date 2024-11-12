package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Facility represents a facility in the database
type Facility struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Name      string    `json:"name" db:"name"`
	Code      string    `json:"code" db:"code"`
}

// CreateFacilityParams holds the parameters needed to create a new facility
type CreateFacilityParams struct {
    Name string `json:"name" form:"name"`  // Add form tag
    Code string `json:"code" form:"code"`  // Add form tag
}

// UpdateFacilityParams holds the parameters needed to update a facility
type UpdateFacilityParams struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// ListFacilities retrieves all facilities from the database
func (db *DB) ListFacilities(ctx context.Context) ([]Facility, error) {
	// Option 2: Use a unique name for the prepared statement
	rows, err := db.pool.Query(ctx, `
	       SELECT id, created_at, name, code
	       FROM facilities
	       ORDER BY name ASC
	   `)
	if err != nil {
		return nil, fmt.Errorf("error listing facilities: %w", err)
	}
	defer rows.Close()

	var facs []Facility
	for rows.Next() {
		var facility Facility
		err := rows.Scan(
			&facility.ID,
			&facility.CreatedAt,
			&facility.Name,
			&facility.Code,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning facility row: %w", err)
		}
		facs = append(facs, facility)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating facility rows: %w", err)
	}

	return facs, nil
}

// IsFacilityCodeUnique checks if a facility code is unique, optionally excluding a specific facility ID
func (db *DB) IsFacilityCodeUnique(ctx context.Context, code string, excludeID *int) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM facilities WHERE code = $1"
	args := []interface{}{code}

	if excludeID != nil {
		query += " AND id != $2"
		args = append(args, *excludeID)
	}
	query += ")"

	var exists bool
	err := db.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking facility code uniqueness: %w", err)
	}
	return !exists, nil
}

// CreateFacility creates a new facility in the database
func (db *DB) CreateFacility(ctx context.Context, params CreateFacilityParams) (*Facility, error) {
	// Check for unique code first
	isUnique, err := db.IsFacilityCodeUnique(ctx, params.Code, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking facility code uniqueness: %w", err)
	}
	if !isUnique {
		return nil, ErrDuplicateFacilityCode
	}

	var facility Facility
	now := time.Now()

	err = db.QueryRow(ctx, `
		INSERT INTO facilities (created_at, name, code)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at, name, code
	`, now, now, params.Name, params.Code).Scan(
		&facility.ID,
		&facility.CreatedAt,
		&facility.Name,
		&facility.Code,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating facility: %w", err)
	}

	return &facility, nil
}

// UpdateFacility updates an existing facility in the database
func (db *DB) UpdateFacility(ctx context.Context, id int, params UpdateFacilityParams) (*Facility, error) {
	// Check for unique code first, excluding the current facility ID
	isUnique, err := db.IsFacilityCodeUnique(ctx, params.Code, &id)
	if err != nil {
		return nil, fmt.Errorf("error checking facility code uniqueness: %w", err)
	}
	if !isUnique {
		return nil, ErrDuplicateFacilityCode
	}

	var facility Facility
    now := time.Now()

    err = db.QueryRow(ctx, `
        UPDATE facilities
        SET updated_at = $1, name = $2, code = $3
        WHERE id = $4
        RETURNING id, created_at, updated_at, name, code
    `, now, params.Name, params.Code, id).Scan(
        &facility.ID,
        &facility.CreatedAt,
        &facility.UpdatedAt,
        &facility.Name,
        &facility.Code,
    )

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, ErrFacilityNotFound
        }
        return nil, fmt.Errorf("error updating facility: %w", err)
    }

    return &facility, nil
}

// Custom errors for better error handling
var (
	ErrDuplicateFacilityCode = fmt.Errorf("facility code already exists")
	ErrFacilityNotFound      = fmt.Errorf("facility not found")
)