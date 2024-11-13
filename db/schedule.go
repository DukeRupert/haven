package db

import (
	"time"
	"fmt"
	"context"
)

type Schedule struct {
	ID        int         `db:"id" json:"id"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`
	UserID    int         `db:"user_id" json:"user_id" validate:"required"`
	FirstDay  time.Weekday `db:"first_day" json:"first_day" validate:"required,min=0,max=6"`
	SecondDay time.Weekday `db:"second_day" json:"second_day" validate:"required,min=0,max=6"`
	StartDate time.Time   `db:"start_date" json:"start_date" validate:"required"`
}

type CreateScheduleParams struct {
    UserID    int
    FirstDay  time.Weekday
    SecondDay time.Weekday
    StartDate time.Time
}

type ProtectedDate struct {
	ID         int       `db:"id" json:"id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	ScheduleID int       `db:"schedule_id" json:"schedule_id" validate:"required"`
	Date       time.Time `db:"date" json:"date" validate:"required"`
	Available  bool      `db:"available" json:"available"`
}

// CreateSchedule creates a new schedule in the database
func (db *DB) CreateSchedule(ctx context.Context, params CreateScheduleParams) (*Schedule, error) {
    // First check if user already has a schedule
    var count int
    err := db.pool.QueryRow(ctx, `
        SELECT COUNT(*) 
        FROM schedules 
        WHERE user_id = $1
    `, params.UserID).Scan(&count)
    if err != nil {
        return nil, fmt.Errorf("error checking existing schedule: %w", err)
    }
    if count > 0 {
        return nil, fmt.Errorf("user already has a schedule: %d", params.UserID)
    }

    var schedule Schedule
    now := time.Now()
    err = db.pool.QueryRow(ctx, `
        INSERT INTO schedules (
            created_at,
            updated_at,
            user_id,
            first_day,
            second_day,
            start_date
        )
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING
            id,
            created_at,
            updated_at,
            user_id,
            first_day,
            second_day,
            start_date
    `,
        now,            // $1
        now,            // $2
        params.UserID,  // $3
        params.FirstDay,    // $4
        params.SecondDay,   // $5
        params.StartDate,   // $6
    ).Scan(
        &schedule.ID,
        &schedule.CreatedAt,
        &schedule.UpdatedAt,
        &schedule.UserID,
        &schedule.FirstDay,
        &schedule.SecondDay,
        &schedule.StartDate,
    )
    if err != nil {
        return nil, fmt.Errorf("error creating schedule: %w", err)
    }

    return &schedule, nil
}

// GetSchedule retrieves a schedule by ID
func (db *DB) GetSchedule(ctx context.Context, id int) (*Schedule, error) {
    var schedule Schedule
    err := db.pool.QueryRow(ctx, `
        SELECT 
            id,
            created_at,
            updated_at,
            user_id,
            first_day,
            second_day,
            start_date
        FROM schedules
        WHERE id = $1
    `, id).Scan(
        &schedule.ID,
        &schedule.CreatedAt,
        &schedule.UpdatedAt,
        &schedule.UserID,
        &schedule.FirstDay,
        &schedule.SecondDay,
        &schedule.StartDate,
    )
    if err != nil {
        return nil, fmt.Errorf("error getting schedule: %w", err)
    }

    return &schedule, nil
}

// GetScheduleByUserID retrieves a schedule by user ID
func (db *DB) GetScheduleByUserID(ctx context.Context, userID int) (*Schedule, error) {
    var schedule Schedule
    err := db.pool.QueryRow(ctx, `
        SELECT 
            id,
            created_at,
            updated_at,
            user_id,
            first_day,
            second_day,
            start_date
        FROM schedules
        WHERE user_id = $1
    `, userID).Scan(
        &schedule.ID,
        &schedule.CreatedAt,
        &schedule.UpdatedAt,
        &schedule.UserID,
        &schedule.FirstDay,
        &schedule.SecondDay,
        &schedule.StartDate,
    )
    if err != nil {
        return nil, fmt.Errorf("error getting schedule by user ID: %w", err)
    }

    return &schedule, nil
}

// GetScheduleByUserInitials retrieves a schedule by user initials and facility ID
func (db *DB) GetScheduleByUserInitials(ctx context.Context, initials string, facilityID int) (*Schedule, error) {
    var schedule Schedule
    err := db.pool.QueryRow(ctx, `
        SELECT 
            s.id,
            s.created_at,
            s.updated_at,
            s.user_id,
            s.first_day,
            s.second_day,
            s.start_date
        FROM schedules s
        JOIN users u ON s.user_id = u.id
        WHERE u.initials = $1 
        AND u.facility_id = $2
    `, initials, facilityID).Scan(
        &schedule.ID,
        &schedule.CreatedAt,
        &schedule.UpdatedAt,
        &schedule.UserID,
        &schedule.FirstDay,
        &schedule.SecondDay,
        &schedule.StartDate,
    )
    if err != nil {
        return nil, fmt.Errorf("error getting schedule by user initials %s: %w", initials, err)
    }

    return &schedule, nil
}

type UpdateScheduleParams struct {
    ID        int
    FirstDay  time.Weekday
    SecondDay time.Weekday
    StartDate time.Time
}

// UpdateSchedule updates an existing schedule
func (db *DB) UpdateSchedule(ctx context.Context, params UpdateScheduleParams) (*Schedule, error) {
    var schedule Schedule
    now := time.Now()
    err := db.pool.QueryRow(ctx, `
        UPDATE schedules 
        SET 
            updated_at = $1,
            first_day = $2,
            second_day = $3,
            start_date = $4
        WHERE id = $5
        RETURNING
            id,
            created_at,
            updated_at,
            user_id,
            first_day,
            second_day,
            start_date
    `,
        now,            // $1
        params.FirstDay,    // $2
        params.SecondDay,   // $3
        params.StartDate,   // $4
        params.ID,         // $5
    ).Scan(
        &schedule.ID,
        &schedule.CreatedAt,
        &schedule.UpdatedAt,
        &schedule.UserID,
        &schedule.FirstDay,
        &schedule.SecondDay,
        &schedule.StartDate,
    )
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

// ListSchedules retrieves all schedules for a facility
func (db *DB) ListSchedules(ctx context.Context, facilityID int) ([]Schedule, error) {
    rows, err := db.pool.Query(ctx, `
        SELECT 
            s.id,
            s.created_at,
            s.updated_at,
            s.user_id,
            s.first_day,
            s.second_day,
            s.start_date
        FROM schedules s
        JOIN users u ON s.user_id = u.id
        WHERE u.facility_id = $1
        ORDER BY s.created_at DESC
    `, facilityID)
    if err != nil {
        return nil, fmt.Errorf("error listing schedules: %w", err)
    }
    defer rows.Close()

    var schedules []Schedule
    for rows.Next() {
        var schedule Schedule
        err := rows.Scan(
            &schedule.ID,
            &schedule.CreatedAt,
            &schedule.UpdatedAt,
            &schedule.UserID,
            &schedule.FirstDay,
            &schedule.SecondDay,
            &schedule.StartDate,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning schedule row: %w", err)
        }
        schedules = append(schedules, schedule)
    }

    return schedules, nil
}