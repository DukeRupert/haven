package db

import (
	"context"
	"fmt"
)

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