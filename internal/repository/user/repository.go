// internal/repository/user/repository.go
package user

import (
	"context"
	"fmt"
	"time"

	"github.com/DukeRupert/haven/internal/model/dto"
	"github.com/DukeRupert/haven/internal/model/entity"
	"github.com/DukeRupert/haven/internal/model/params"
	"github.com/DukeRupert/haven/internal/repository/facility"
	"github.com/DukeRupert/haven/internal/repository/schedule"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Repository handles user-related database operations
type Repository struct {
	pool     *pgxpool.Pool
	facility *facility.Repository
	schedule *schedule.Repository
}

// New creates a new user repository
func New(pool *pgxpool.Pool, facility *facility.Repository, schedule *schedule.Repository) *Repository {
	return &Repository{
		pool:     pool,
		facility: facility,
		schedule: schedule,
	}
}

// Custom errors
var (
	ErrNotFound    = fmt.Errorf("user not found")
	ErrEmailExists = fmt.Errorf("email already exists")
)

func (r *Repository) GetByID(ctx context.Context, id int) (*entity.User, error) {
	var user entity.User
	err := r.pool.QueryRow(ctx, `
        SELECT 
            id, created_at, updated_at, first_name, last_name, 
            initials, email, facility_id, role
        FROM users
        WHERE id = $1
    `, id).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
		&user.FirstName, &user.LastName, &user.Initials,
		&user.Email, &user.FacilityID, &user.Role,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	return &user, nil
}

func (r *Repository) GetByFacilityCode(ctx context.Context, facilityCode string) ([]entity.User, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT 
            u.id, u.created_at, u.updated_at, u.first_name, u.last_name,
            u.initials, u.email, u.facility_id, u.role
        FROM users u
        JOIN facilities f ON u.facility_id = f.id
        WHERE f.code = $1
        ORDER BY u.last_name, u.first_name ASC
    `, facilityCode)
	if err != nil {
		return nil, fmt.Errorf("querying users by facility code: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		err := rows.Scan(
			&user.ID, &user.CreatedAt, &user.UpdatedAt,
			&user.FirstName, &user.LastName, &user.Initials,
			&user.Email, &user.FacilityID, &user.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning user row: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user rows: %w", err)
	}

	return users, nil
}

func (r *Repository) Create(ctx context.Context, params params.CreateUserParams) (*entity.User, error) {
	log.Debug().
		Str("first_name", params.FirstName).
		Str("last_name", params.LastName).
		Str("initials", params.Initials).
		Str("email", params.Email).
		Str("role", string(params.Role)).
		Int("facility_id", params.FacilityID).
		Msg("creating user")

	// Check email uniqueness
	isUnique, err := r.IsEmailUnique(ctx, params.Email, nil)
	if err != nil {
		return nil, err
	}
	if !isUnique {
		return nil, ErrEmailExists
	}

	var user entity.User
	now := time.Now()
	err = r.pool.QueryRow(ctx, `
        INSERT INTO users (
            created_at, updated_at, first_name, last_name, 
            initials, email, password, facility_id, role
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING 
            id, created_at, updated_at, first_name, last_name,
            initials, email, facility_id, role
    `,
		now, now, params.FirstName, params.LastName,
		params.Initials, params.Email, params.Password,
		params.FacilityID, params.Role,
	).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
		&user.FirstName, &user.LastName, &user.Initials,
		&user.Email, &user.FacilityID, &user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return &user, nil
}

func (r *Repository) Update(ctx context.Context, userID int, params params.UpdateUserParams) (*entity.User, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
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

	var user entity.User
	now := time.Now()

	err = r.pool.QueryRow(ctx, `
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
            role`,
		now,
		params.FirstName,
		params.LastName,
		params.Initials,
		params.Email,
		params.FacilityID,
		params.Role,
		userID,
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
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &user, nil
}

func (r *Repository) Delete(ctx context.Context, userID int) error {
	result, err := r.pool.Exec(ctx, `
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
func (r *Repository) VerifyCredentials(ctx context.Context, facilityID int, initials, email string) (*entity.User, error) {
	var user entity.User
	err := r.pool.QueryRow(ctx, `
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

func (r *Repository) SetPassword(ctx context.Context, userID int, hashedPassword string) error {
	_, err := r.pool.Exec(ctx, `
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

func (r *Repository) UpdatePassword(ctx context.Context, userID int, hashedPassword string) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE users 
        SET 
            password = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("error updating user password: %w", err)
	}
	return nil
}

// GetDetails retrieves full user details including facility and schedule
func (r *Repository) GetDetails(ctx context.Context, initials string, facility string) (*dto.UserDetails, error) {
	// Get user first
	user, err := r.GetByInitialsAndFacility(ctx, initials, facility)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	// Create result channels
	type facilityResult struct {
		facility *entity.Facility
		err      error
	}
	type scheduleResult struct {
		schedule *entity.Schedule
		err      error
	}

	facilityChan := make(chan facilityResult, 1)
	scheduleChan := make(chan scheduleResult, 1)

	// Fetch facility and schedule concurrently
	go func() {
		facility, err := r.facility.GetByID(ctx, user.FacilityID)
		facilityChan <- facilityResult{facility, err}
	}()

	go func() {
		schedule, err := r.schedule.GetByUserID(ctx, user.ID)
        if err != nil {
            // Return empty schedule with ID 0 for any error (including no rows)
            schedule = &entity.Schedule{
                ID:     0,
                UserID: user.ID,
            }
            err = nil
        }
        scheduleChan <- scheduleResult{schedule, err}
	}()

	// Collect results
	fResult := <-facilityChan
	if fResult.err != nil {
		return nil, fmt.Errorf("getting facility: %w", fResult.err)
	}

	sResult := <-scheduleChan
	if sResult.err != nil {
		return nil, fmt.Errorf("getting schedule: %w", sResult.err)
	}

	return &dto.UserDetails{
		User:     *user,
		Facility: *fResult.facility,
		Schedule: *sResult.schedule,
	}, nil
}

// GetByInitialsAndFacility finds a user by their initials and facility code
func (r *Repository) GetByInitialsAndFacility(ctx context.Context, initials string, facilityCode string) (*entity.User, error) {
    var user entity.User
    err := r.pool.QueryRow(ctx, `
        SELECT 
            u.id, u.created_at, u.updated_at, u.first_name, u.last_name, 
            u.initials, u.email, u.facility_id, u.role, u.registration_completed,
            u.password
        FROM users u
        JOIN facilities f ON u.facility_id = f.id
        WHERE u.initials = $1 AND f.code = $2
    `, initials, facilityCode).Scan(
        &user.ID, &user.CreatedAt, &user.UpdatedAt,
        &user.FirstName, &user.LastName, &user.Initials,
        &user.Email, &user.FacilityID, &user.Role,
        &user.RegistrationCompleted, &user.Password,
    )
    if err == pgx.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("finding user by initials and facility code: %w", err)
    }

    return &user, nil
}

// GetByEmail retrieves a user by their email address
func (r *Repository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.pool.QueryRow(ctx, `
        SELECT 
            id, created_at, updated_at, first_name, last_name,
            initials, email, password, facility_id, role
        FROM users
        WHERE email = $1
    `, email).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
		&user.FirstName, &user.LastName, &user.Initials,
		&user.Email, &user.Password, &user.FacilityID, &user.Role,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return &user, nil
}

// IsEmailUnique checks if an email is unique, excluding the specified user ID
func (r *Repository) IsEmailUnique(ctx context.Context, email string, excludeID *int) (bool, error) {
	query := `
        SELECT NOT EXISTS (
            SELECT 1 FROM users 
            WHERE email = $1
            AND ($2::int IS NULL OR id != $2)
        )`

	var isUnique bool
	err := r.pool.QueryRow(ctx, query, email, excludeID).Scan(&isUnique)
	if err != nil {
		return false, fmt.Errorf("checking email uniqueness: %w", err)
	}
	return isUnique, nil
}
