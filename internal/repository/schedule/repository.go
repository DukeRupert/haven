// internal/repository/schedule/repository.go
package schedule

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/DukeRupert/haven/internal/model/entity"
    "github.com/DukeRupert/haven/internal/model/params"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles schedule-related database operations
type Repository struct {
    pool *pgxpool.Pool
}

// New creates a new schedule repository
func New(pool *pgxpool.Pool) *Repository {
    return &Repository{
        pool: pool,
    }
}

// Common errors
var (
    ErrNotFound         = fmt.Errorf("schedule not found")
    ErrAlreadyExists    = fmt.Errorf("user already has a schedule")
    ErrDateNotFound     = fmt.Errorf("protected date not found")
)

func (r *Repository) Create(ctx context.Context, params params.CreateScheduleByCodeParams) (*entity.Schedule, error) {
    // First, get the user ID using a join between facilities and users
    var userID, facilityID int
    err := r.pool.QueryRow(ctx, `
        SELECT u.id, u.facility_id
        FROM users u
        JOIN facilities f ON u.facility_id = f.id
        WHERE f.code = $1 AND u.initials = $2
    `, params.FacilityCode, params.UserInitials).Scan(&userID, &facilityID)
    if err == pgx.ErrNoRows {
        return nil, fmt.Errorf("no user found with facility code %s and initials %s",
            params.FacilityCode, params.UserInitials)
    }
    if err != nil {
        return nil, fmt.Errorf("finding user: %w", err)
    }

    // Check for existing schedule
    exists, err := r.hasSchedule(ctx, userID)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, ErrAlreadyExists
    }

    // Create new schedule
    var schedule entity.Schedule
    now := time.Now()
    err = r.pool.QueryRow(ctx, `
        WITH inserted_schedule AS (
            INSERT INTO schedules (
                created_at, updated_at, user_id,
                first_weekday, second_weekday, start_date
            )
            VALUES ($1, $2, $3, $4, $5, $6)
            RETURNING *
        )
        SELECT 
            s.id, s.created_at, s.updated_at, s.user_id,
            u.facility_id,
            s.first_weekday, s.second_weekday, s.start_date
        FROM inserted_schedule s
        JOIN users u ON s.user_id = u.id
    `, now, now, userID, params.FirstWeekday, params.SecondWeekday, params.StartDate).Scan(
        &schedule.ID,
        &schedule.CreatedAt,
        &schedule.UpdatedAt,
        &schedule.UserID,
        &schedule.FacilityID,
        &schedule.FirstWeekday,
        &schedule.SecondWeekday,
        &schedule.StartDate,
    )
    if err != nil {
        return nil, fmt.Errorf("creating schedule: %w", err)
    }
    return &schedule, nil
}

func (r *Repository) Update(ctx context.Context, scheduleID int, params params.UpdateScheduleParams) (*entity.Schedule, error) {
    var schedule entity.Schedule
    err := r.pool.QueryRow(ctx, `
        UPDATE schedules s
        SET 
            updated_at = $1,
            first_weekday = $2,
            second_weekday = $3,
            start_date = $4
        FROM users u
        WHERE s.id = $5 AND s.user_id = u.id
        RETURNING 
            s.id, s.created_at, s.updated_at, s.user_id,
            u.facility_id,
            s.first_weekday, s.second_weekday, s.start_date
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
        &schedule.FacilityID,
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

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `
        DELETE FROM schedules
        WHERE id = $1
    `, id)
	if err != nil {
		return fmt.Errorf("error deleting schedule: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (*entity.Schedule, error) {
    var schedule entity.Schedule
    err := r.pool.QueryRow(ctx, `
        SELECT 
            s.id,
            s.created_at,
            s.updated_at,
            s.user_id,
            u.facility_id,
            s.first_weekday,
            s.second_weekday,
            s.start_date
        FROM schedules s
        JOIN users u ON s.user_id = u.id
        WHERE s.id = $1
    `, id).Scan(
        &schedule.ID,
        &schedule.CreatedAt,
        &schedule.UpdatedAt,
        &schedule.UserID,
        &schedule.FacilityID,
        &schedule.FirstWeekday,
        &schedule.SecondWeekday,
        &schedule.StartDate,
    )

    if err == pgx.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("getting schedule: %w", err)
    }

    return &schedule, nil
}

func (r *Repository) GetByUserID(ctx context.Context, userID int) (*entity.Schedule, error) {
    var schedule entity.Schedule
    err := r.pool.QueryRow(ctx, `
        SELECT 
            s.id,
            s.created_at,
            s.updated_at,
            s.user_id,
            u.facility_id,
            s.first_weekday,
            s.second_weekday,
            s.start_date
        FROM schedules s
        JOIN users u ON s.user_id = u.id
        WHERE s.user_id = $1
    `, userID).Scan(
        &schedule.ID,
        &schedule.CreatedAt,
        &schedule.UpdatedAt,
        &schedule.UserID,
        &schedule.FacilityID,
        &schedule.FirstWeekday,
        &schedule.SecondWeekday,
        &schedule.StartDate,
    )
 
    if err == pgx.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("getting schedule by user ID: %w", err)
    }
 
    return &schedule, nil
 }

func (r *Repository) GetProtectedDateByID(ctx context.Context, id int) (entity.ProtectedDate, error) {
	var date entity.ProtectedDate
	err := r.pool.QueryRow(ctx, `
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

func (r *Repository) GetProtectedDatesByFacilityCode(ctx context.Context, facilityCode string) ([]entity.ProtectedDate, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT 
            pd.id, pd.created_at, pd.updated_at,
            pd.schedule_id, pd.date, pd.available,
            pd.user_id, pd.facility_id
        FROM protected_dates pd
        JOIN facilities f ON pd.facility_id = f.id
        WHERE f.code = $1
        ORDER BY pd.date ASC, pd.user_id
    `, facilityCode)
    if err != nil {
        return nil, fmt.Errorf("getting protected dates: %w", err)
    }
    defer rows.Close()

    return r.scanProtectedDates(rows)
}

func (r *Repository) ToggleProtectedDateAvailability(ctx context.Context, dateID int) (entity.ProtectedDate, error) {
    var date entity.ProtectedDate
	err := r.pool.QueryRow(ctx, `
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

// Helper methods
func (r *Repository) hasSchedule(ctx context.Context, userID int) (bool, error) {
    var exists bool
    err := r.pool.QueryRow(ctx, `
        SELECT EXISTS(SELECT 1 FROM schedules WHERE user_id = $1)
    `, userID).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("checking schedule existence: %w", err)
    }
    return exists, nil
}

func (r *Repository) scanProtectedDates(rows pgx.Rows) ([]entity.ProtectedDate, error) {
	var dates []entity.ProtectedDate
	for rows.Next() {
		var date entity.ProtectedDate
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