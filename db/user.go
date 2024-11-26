package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DukeRupert/haven/types"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

func (db *DB) GetUsersByFacilityCode(ctx context.Context, facilityCode string) ([]types.User, error) {
	rows, err := db.pool.Query(ctx, `
        SELECT u.id, u.created_at, u.updated_at, u.first_name, u.last_name, 
               u.initials, u.email, u.facility_id, u.role
        FROM users u
        JOIN facilities f ON u.facility_id = f.id
        WHERE f.code = $1
        ORDER BY u.last_name, u.first_name ASC
    `, facilityCode)
	if err != nil {
		return nil, fmt.Errorf("error getting users by facility code: %w", err)
	}
	defer rows.Close()

	var users []types.User
	for rows.Next() {
		var user types.User
		err := rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.FirstName,
			&user.LastName,
			&user.Initials,
			&user.Email,
			&user.FacilityID,
			&user.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

func (db *DB) GetUserByID(ctx context.Context, id int) (*types.User, error) {
	var user types.User
	err := db.pool.QueryRow(ctx, `
        SELECT 
            id, 
            created_at, 
            updated_at, 
            first_name, 
            last_name, 
            initials, 
            email, 
            facility_id, 
            role
        FROM users
        WHERE id = $1
    `, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.FirstName,
		&user.LastName,
		&user.Initials,
		&user.Email,
		&user.FacilityID,
		&user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting user by id: %w", err)
	}
	return &user, nil
}

func (db *DB) GetUserByInitialsAndFacility(ctx context.Context, initials string, facilityID int) (*types.User, error) {
	var user types.User
	err := db.pool.QueryRow(ctx, `
        SELECT 
            id, 
            created_at, 
            updated_at, 
            first_name, 
            last_name, 
            initials, 
            email, 
            facility_id, 
            role
        FROM users
        WHERE initials = $1 
        AND facility_id = $2
    `, initials, facilityID).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.FirstName,
		&user.LastName,
		&user.Initials,
		&user.Email,
		&user.FacilityID,
		&user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting user by initials and facility: %w", err)
	}

	return &user, nil
}

func (db *DB) GetUserDetails(ctx context.Context, initials string, facilityID int) (*types.UserDetails, error) {
	// Get user first (need this to get other data)
	user, err := db.GetUserByInitialsAndFacility(ctx, initials, facilityID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Create channels for results
	facilityChan := make(chan struct {
		facility *types.Facility
		err      error
	})
	scheduleChan := make(chan struct {
		schedule *types.Schedule
		err      error
	})

	// Get facility and schedule concurrently
	go func() {
		facility, err := db.GetFacilityByID(ctx, user.FacilityID)
		facilityChan <- struct {
			facility *types.Facility
			err      error
		}{facility, err}
	}()

	go func() {
		schedule, err := db.GetScheduleByUserID(ctx, user.ID)
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			err = nil // Convert "no rows" to nil error
		}
		scheduleChan <- struct {
			schedule *types.Schedule
			err      error
		}{schedule, err}
	}()

	// Wait for both results
	facilityResult := <-facilityChan
	if facilityResult.err != nil {
		return nil, fmt.Errorf("error getting facility: %w", facilityResult.err)
	}

	scheduleResult := <-scheduleChan
	if scheduleResult.err != nil {
		return nil, fmt.Errorf("error getting schedule: %w", scheduleResult.err)
	}

	// Create the schedule - either empty or from result
	var schedule types.Schedule
	if scheduleResult.schedule == nil {
		schedule = types.Schedule{
			UserID: user.ID, // Set only the UserID for empty schedule
		}
	} else {
		schedule = *scheduleResult.schedule
	}

	return &types.UserDetails{
		User:     *user,
		Facility: *facilityResult.facility,
		Schedule: schedule, // Not a pointer here
	}, nil
}

func (db *DB) CreateUser(ctx context.Context, params types.CreateUserParams) (*types.User, error) {
	// Log params received by database method
	log.Debug().
		Str("first_name", params.FirstName).
		Str("last_name", params.LastName).
		Str("initials", params.Initials).
		Str("email", params.Email).
		Str("role", string(params.Role)).
		Int("facility_id", params.FacilityID).
		Msg("received create user params")

		// First check if email is unique
	var count int
	err := db.pool.QueryRow(ctx, `
        SELECT COUNT(*) 
        FROM users 
        WHERE email = $1
    `, params.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("error checking email uniqueness: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("email already exists: %s", params.Email)
	}

	var user types.User
	now := time.Now()
	err = db.pool.QueryRow(ctx, `
        INSERT INTO users (
            created_at, 
            updated_at, 
            first_name, 
            last_name, 
            initials, 
            email, 
            password,
            facility_id, 
            role
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING 
            id, 
            created_at, 
            updated_at, 
            first_name, 
            last_name, 
            initials, 
            email, 
            facility_id, 
            role
    `,
		now,               // $1
		now,               // $2
		params.FirstName,  // $3
		params.LastName,   // $4
		params.Initials,   // $5
		params.Email,      // $6
		params.Password,   // $7 (should be pre-hashed)
		params.FacilityID, // $8
		params.Role,       // $9
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.FirstName,
		&user.LastName,
		&user.Initials,
		&user.Email,
		&user.FacilityID,
		&user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &user, nil
}

func (db *DB) UpdateUser(ctx context.Context, userID int, params types.CreateUserParams) (*types.User, error) {
	// First check if email is unique (excluding current user)
	var count int
	err := db.pool.QueryRow(ctx, `
        SELECT COUNT(*) 
        FROM users 
        WHERE email = $1 AND id != $2
    `, params.Email, userID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("error checking email uniqueness: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("email already exists: %s", params.Email)
	}

	var user types.User
	now := time.Now()

	// If password is empty, keep existing password
	var query string
	var args []interface{}
	if params.Password != "" {
		query = `
            UPDATE users 
            SET updated_at = $1,
                first_name = $2,
                last_name = $3,
                initials = $4,
                email = $5,
                password = $6,
                facility_id = $7,
                role = $8
            WHERE id = $9
            RETURNING 
                id, 
                created_at, 
                updated_at, 
                first_name, 
                last_name, 
                initials, 
                email, 
                facility_id, 
                role`
		args = []interface{}{
			now,               // $1
			params.FirstName,  // $2
			params.LastName,   // $3
			params.Initials,   // $4
			params.Email,      // $5
			params.Password,   // $6 (should be pre-hashed)
			params.FacilityID, // $7
			params.Role,       // $8
			userID,            // $9
		}
	} else {
		query = `
            UPDATE users 
            SET updated_at = $1,
                first_name = $2,
                last_name = $3,
                initials = $4,
                email = $5,
                facility_id = $6,
                role = $7
            WHERE id = $8
            RETURNING 
                id, 
                created_at, 
                updated_at, 
                first_name, 
                last_name, 
                initials, 
                email, 
                facility_id, 
                role`
		args = []interface{}{
			now,               // $1
			params.FirstName,  // $2
			params.LastName,   // $3
			params.Initials,   // $4
			params.Email,      // $5
			params.FacilityID, // $6
			params.Role,       // $7
			userID,            // $8
		}
	}

	err = db.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.FirstName,
		&user.LastName,
		&user.Initials,
		&user.Email,
		&user.FacilityID,
		&user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &user, nil
}

func (db *DB) DeleteUser(ctx context.Context, userID int) error {
	result, err := db.pool.Exec(ctx, `
        DELETE FROM users 
        WHERE id = $1
    `, userID)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	// Check if any row was actually deleted
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found with ID %d", userID)
	}

	return nil
}

// VerifyUserCredentials checks if a user exists with matching facility_id, initials, and email
func (db *DB) VerifyUserCredentials(ctx context.Context, facilityID int, initials, email string) (*types.User, error) {
	var user types.User
	err := db.pool.QueryRow(ctx, `
        SELECT 
            id, 
            created_at, 
            updated_at, 
            first_name, 
            last_name, 
            initials, 
            email,
            facility_id,
            role
        FROM users
        WHERE facility_id = $1 
        AND UPPER(initials) = UPPER($2)
        AND LOWER(email) = LOWER($3)
    `, facilityID, initials, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.FirstName,
		&user.LastName,
		&user.Initials,
		&user.Email,
		&user.FacilityID,
		&user.Role,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error verifying user credentials: %w", err)
	}
	return &user, nil
}

func (db *DB) SetUserPassword(ctx context.Context, userID int, hashedPassword string) error {
	_, err := db.pool.Exec(ctx, `
        UPDATE users 
        SET 
            password = $1,
            registration_completed = true,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("error setting user password: %w", err)
	}

	return nil
}
