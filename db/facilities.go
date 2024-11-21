package db

import (
	"context"
	"fmt"
	"time"

	"github.com/DukeRupert/haven/types"

	"github.com/jackc/pgx/v5"
)

// ListFacilities retrieves all facilities from the database
func (db *DB) ListFacilities(ctx context.Context) ([]types.Facility, error) {
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

	var facs []types.Facility
	for rows.Next() {
		var facility types.Facility
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

// GetFacilityByID retrieves a facility from the database by its ID
func (db *DB) GetFacilityByID(ctx context.Context, id int) (*types.Facility, error) {
	var facility types.Facility
	err := db.QueryRow(ctx, `
        SELECT id, created_at, updated_at, name, code
        FROM facilities
        WHERE id = $1
    `, id).Scan(
		&facility.ID,
		&facility.CreatedAt,
		&facility.UpdatedAt,
		&facility.Name,
		&facility.Code,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("facility not found with id %d", id)
		}
		return nil, fmt.Errorf("error getting facility by id: %w", err)
	}
	return &facility, nil
}

// GetFacilityByCode retrieves a facility from the database by its code
func (db *DB) GetFacilityByCode(ctx context.Context, code string) (*types.Facility, error) {
	var facility types.Facility
	err := db.QueryRow(ctx, `
        SELECT id, created_at, updated_at, name, code
        FROM facilities
        WHERE code = $1
    `, code).Scan(
		&facility.ID,
		&facility.CreatedAt,
		&facility.UpdatedAt,
		&facility.Name,
		&facility.Code,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting facility by code: %w", err)
	}
	return &facility, nil
}

// CreateFacility creates a new facility in the database
func (db *DB) CreateFacility(ctx context.Context, params types.CreateFacilityParams) (*types.Facility, error) {
	// Check for unique code first
	isUnique, err := db.IsFacilityCodeUnique(ctx, params.Code, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking facility code uniqueness: %w", err)
	}
	if !isUnique {
		return nil, ErrDuplicateFacilityCode
	}

	var facility types.Facility
	now := time.Now()
	err = db.QueryRow(ctx, `
		INSERT INTO facilities (created_at, updated_at, name, code)  -- Added updated_at
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at, name, code
	`, now, now, params.Name, params.Code).Scan(
		&facility.ID,
		&facility.CreatedAt,
		&facility.UpdatedAt, // Add this field to match RETURNING clause
		&facility.Name,
		&facility.Code,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating facility: %w", err)
	}
	return &facility, nil
}

// UpdateFacility updates an existing facility in the database
func (db *DB) UpdateFacility(ctx context.Context, id int, params types.UpdateFacilityParams) (*types.Facility, error) {
	// Check for unique code first, excluding the current facility ID
	isUnique, err := db.IsFacilityCodeUnique(ctx, params.Code, &id)
	if err != nil {
		return nil, fmt.Errorf("error checking facility code uniqueness: %w", err)
	}
	if !isUnique {
		return nil, ErrDuplicateFacilityCode
	}

	var facility types.Facility
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
