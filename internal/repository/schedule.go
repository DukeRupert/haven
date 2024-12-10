package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DukeRupert/haven/types"

	"github.com/jackc/pgx/v5"
)

// GetScheduleByUserID retrieves a schedule by user ID
func (db *DB) GetScheduleByUserID(ctx context.Context, userID int) (*types.Schedule, error) {
	var schedule types.Schedule
	err := db.pool.QueryRow(ctx, `
        SELECT 
            id,
            created_at,
            updated_at,
            user_id,
            first_weekday,
            second_weekday,
            start_date
        FROM schedules
        WHERE user_id = $1
    `, userID).Scan(
		&schedule.ID,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
		&schedule.UserID,
		&schedule.FirstWeekday,
		&schedule.SecondWeekday,
		&schedule.StartDate,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting schedule by user ID: %w", err)
	}

	return &schedule, nil
}

// GetScheduleByID retrieves a schedule by its ID. This is a direct lookup on the schedules
// table and returns all schedule fields. Returns nil if no schedule is found.
func (db *DB) GetScheduleByID(ctx context.Context, id int) (*types.Schedule, error) {
	var schedule types.Schedule
	err := db.pool.QueryRow(ctx, `
        SELECT 
            id,
            created_at,
            updated_at,
            user_id,
            first_weekday,
            second_weekday,
            start_date
        FROM schedules
        WHERE id = $1
    `, id).Scan(
		&schedule.ID,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
		&schedule.UserID,
		&schedule.FirstWeekday,
		&schedule.SecondWeekday,
		&schedule.StartDate,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting schedule: %w", err)
	}

	return &schedule, nil
}

// UpdateSchedule updates an existing schedule record with the provided schedule data.
// It returns the updated schedule or an error if the update fails.
func (db *DB) UpdateSchedule(ctx context.Context, scheduleID int, params types.UpdateScheduleParams) (*types.Schedule, error) {
	var schedule types.Schedule
	err := db.QueryRow(ctx, `
        UPDATE schedules 
        SET 
            updated_at = $1,
            first_weekday = $2,
            second_weekday = $3,
            start_date = $4
        WHERE id = $5
        RETURNING id, created_at, updated_at, user_id, first_weekday, second_weekday, start_date
    `,
		time.Now(),
		params.FirstWeekday,
		params.SecondWeekday,
		params.StartDate,
		scheduleID,
	).Scan(
		&schedule.ID,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
		&schedule.UserID,
		&schedule.FirstWeekday,
		&schedule.SecondWeekday,
		&schedule.StartDate,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("no schedule found with ID %d", scheduleID)
	}
	if err != nil {
		return nil, fmt.Errorf("error updating schedule: %w", err)
	}

	return &schedule, nil
}

// DeleteSchedule deletes a schedule and its associated protected dates
func (db *DB) DeleteSchedule(ctx context.Context, id int) error {
	_, err := db.pool.Exec(ctx, `
        DELETE FROM schedules
        WHERE id = $1
    `, id)
	if err != nil {
		return fmt.Errorf("error deleting schedule: %w", err)
	}

	return nil
}

func (db *DB) CreateScheduleByCode(ctx context.Context, params types.CreateScheduleByCodeParams) (*types.Schedule, error) {
	// First, get the user ID using a join between facilities and users
	var userID int
	err := db.QueryRow(ctx, `
        SELECT u.id 
        FROM users u
        JOIN facilities f ON u.facility_id = f.id
        WHERE f.code = $1 AND u.initials = $2
    `, params.FacilityCode, params.UserInitials).Scan(&userID)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("no user found with facility code %s and initials %s",
			params.FacilityCode, params.UserInitials)
	}
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	// Check if user already has a schedule
	hasSchedule, err := db.doesUserHaveSchedule(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error checking existing schedule: %w", err)
	}
	if hasSchedule {
		return nil, ErrUserScheduleExists
	}

	var schedule types.Schedule
	now := time.Now()
	err = db.QueryRow(ctx, `
        INSERT INTO schedules (
            created_at,
            updated_at,
            user_id,
            first_weekday,
            second_weekday,
            start_date
        )
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at, user_id, first_weekday, second_weekday, start_date
    `, now, now, userID, params.FirstWeekday, params.SecondWeekday, params.StartDate).Scan(
		&schedule.ID,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
		&schedule.UserID,
		&schedule.FirstWeekday,
		&schedule.SecondWeekday,
		&schedule.StartDate,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating schedule: %w", err)
	}

	return &schedule, nil
}

// GetProtectedDate retrieves a single protected date by ID
func (db *DB) GetProtectedDate(ctx context.Context, id int) (types.ProtectedDate, error) {
	var date types.ProtectedDate
	err := db.pool.QueryRow(ctx, `
        SELECT 
            id,
            created_at,
            updated_at,
            schedule_id,
            date,
            available,
            user_id,
            facility_id
        FROM protected_dates
        WHERE id = $1
    `, id).Scan(
		&date.ID,
		&date.CreatedAt,
		&date.UpdatedAt,
		&date.ScheduleID,
		&date.Date,
		&date.Available,
		&date.UserID,
		&date.FacilityID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return date, fmt.Errorf("protected date %d not found: %w", id, err)
		}
		return date, fmt.Errorf("error getting protected date %d: %w", id, err)
	}

	return date, nil
}

// GetProtectedDatesByFacilityCode retrieves all protected dates for all users at a facility
func (db *DB) GetProtectedDatesByFacilityCode(ctx context.Context, facilityCode string) ([]types.ProtectedDate, error) {
	rows, err := db.pool.Query(ctx, `
        SELECT 
            pd.id,
            pd.created_at,
            pd.updated_at,
            pd.schedule_id,
            pd.date,
            pd.available,
            pd.user_id,
            pd.facility_id
        FROM protected_dates pd
        JOIN facilities f ON pd.facility_id = f.id
        WHERE f.code = $1
        ORDER BY pd.date ASC, pd.user_id
    `, facilityCode)
	if err != nil {
		return nil, fmt.Errorf("error getting protected dates for facility code %s: %w", facilityCode, err)
	}
	defer rows.Close()

	return scanProtectedDates(rows)
}

// Helper function to scan rows into ProtectedDate slices
func scanProtectedDates(rows pgx.Rows) ([]types.ProtectedDate, error) {
	var dates []types.ProtectedDate
	for rows.Next() {
		var date types.ProtectedDate
		err := rows.Scan(
			&date.ID,
			&date.CreatedAt,
			&date.UpdatedAt,
			&date.ScheduleID,
			&date.Date,
			&date.Available,
			&date.UserID,
			&date.FacilityID,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning protected date row: %w", err)
		}
		dates = append(dates, date)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating protected date rows: %w", err)
	}

	return dates, nil
}

// ToggleProtectedDateAvailability toggles the available status of a protected date and returns the updated record
func (db *DB) ToggleProtectedDateAvailability(ctx context.Context, dateID int) (types.ProtectedDate, error) {
	var date types.ProtectedDate
	err := db.pool.QueryRow(ctx, `
        UPDATE protected_dates 
        SET 
            available = NOT available,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
        RETURNING 
            id,
            created_at,
            updated_at,
            schedule_id,
            date,
            available,
            user_id,
            facility_id
    `, dateID).Scan(
		&date.ID,
		&date.CreatedAt,
		&date.UpdatedAt,
		&date.ScheduleID,
		&date.Date,
		&date.Available,
		&date.UserID,
		&date.FacilityID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return date, fmt.Errorf("protected date %d not found: %w", dateID, err)
		}
		return date, fmt.Errorf("error toggling protected date %d: %w", dateID, err)
	}

	return date, nil
}

// Helper method to check if user exists
func (db *DB) doesUserExist(ctx context.Context, userID int) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
	`, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking user existence: %w", err)
	}
	return exists, nil
}

// Helper method to check if user already has a schedule
func (db *DB) doesUserHaveSchedule(ctx context.Context, userID int) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM schedules WHERE user_id = $1)
	`, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking existing schedule: %w", err)
	}
	return exists, nil
}

// Error definitions (add these to your errors.go file)
var (
	ErrUserNotFound       = fmt.Errorf("user not found")
	ErrUserScheduleExists = fmt.Errorf("user already has a schedule")
)