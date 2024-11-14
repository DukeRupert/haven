package db

import (
	"time"
	"fmt"
	"context"
)

type Schedule struct {
    ID            int         `db:"id" json:"id"`
    CreatedAt     time.Time   `db:"created_at" json:"created_at"`
    UpdatedAt     time.Time   `db:"updated_at" json:"updated_at"`
    UserID        int         `db:"user_id" json:"user_id" validate:"required"`
    FirstWeekday  time.Weekday `db:"first_weekday" json:"first_weekday" validate:"required,min=0,max=6"`
    SecondWeekday time.Weekday `db:"second_weekday" json:"second_weekday" validate:"required,min=0,max=6"`
    StartDate     time.Time   `db:"start_date" json:"start_date" validate:"required"`
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

func (s Schedule) IsZero() bool {
    return s.ID == 0 && 
           s.CreatedAt.IsZero() && 
           s.UpdatedAt.IsZero() && 
           s.UserID == 0 && 
           s.StartDate.IsZero()
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

// NewSchedule creates a new Schedule with default values
func NewSchedule(userID int, firstWeekday, secondWeekday time.Weekday, startDate time.Time) *Schedule {
    now := time.Now()
    return &Schedule{
        CreatedAt:     now,
        UpdatedAt:     now,
        UserID:        userID,
        FirstWeekday:  firstWeekday,
        SecondWeekday: secondWeekday,
        StartDate:     startDate,
    }
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

// GetScheduleByUserInitials retrieves a schedule by user initials and facility ID
func (db *DB) GetScheduleByUserInitials(ctx context.Context, initials string, facilityID int) (*Schedule, error) {
    var schedule Schedule
    err := db.pool.QueryRow(ctx, `
        SELECT 
            s.id,
            s.created_at,
            s.updated_at,
            s.user_id,
            s.first_weekday,
            s.second_weekday,
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
        &schedule.FirstWeekday,
        &schedule.SecondWeekday,
        &schedule.StartDate,
    )
    if err != nil {
        return nil, fmt.Errorf("error getting schedule by user initials %s: %w", initials, err)
    }

    return &schedule, nil
}

type UpdateScheduleParams struct {
    ID            int
    FirstWeekday  time.Weekday
    SecondWeekday time.Weekday
    StartDate     time.Time
}

// UpdateSchedule updates an existing schedule
func (db *DB) UpdateSchedule(ctx context.Context, params UpdateScheduleParams) (*Schedule, error) {
    var schedule Schedule
    now := time.Now()
    err := db.pool.QueryRow(ctx, `
        UPDATE schedules 
        SET 
            updated_at = $1,
            first_weekday = $2,
            second_weekday = $3,
            start_date = $4
        WHERE id = $5
        RETURNING
            id,
            created_at,
            updated_at,
            user_id,
            first_weekday,
            second_weekday,
            start_date
    `,
        now,                // $1
        params.FirstWeekday,    // $2
        params.SecondWeekday,   // $3
        params.StartDate,       // $4
        params.ID,              // $5
    ).Scan(
        &schedule.ID,
        &schedule.CreatedAt,
        &schedule.UpdatedAt,
        &schedule.UserID,
        &schedule.FirstWeekday,
        &schedule.SecondWeekday,
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
            s.first_weekday,
            s.second_weekday,
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
            &schedule.FirstWeekday,
            &schedule.SecondWeekday,
            &schedule.StartDate,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning schedule row: %w", err)
        }
        schedules = append(schedules, schedule)
    }

    return schedules, nil
}

func (db *DB) CreateSchedule(ctx context.Context, params CreateScheduleParams) (*Schedule, error) {
	// Check if user exists first
	exists, err := db.doesUserExist(ctx, params.UserID)
	if err != nil {
		return nil, fmt.Errorf("error checking user existence: %w", err)
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	// Check if user already has a schedule
	hasSchedule, err := db.doesUserHaveSchedule(ctx, params.UserID)
	if err != nil {
		return nil, fmt.Errorf("error checking existing schedule: %w", err)
	}
	if hasSchedule {
		return nil, ErrUserScheduleExists
	}

	var schedule Schedule
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
	`, now, now, params.UserID, params.FirstDay, params.SecondDay, params.StartDate).Scan(
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
	ErrUserNotFound      = fmt.Errorf("user not found")
	ErrUserScheduleExists = fmt.Errorf("user already has a schedule")
)