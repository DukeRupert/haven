package db

import (
	"context"
	"fmt"
	"time"
)

// UserRole is a custom type for the user_role enum
type UserRole string

const (
	UserRoleSuper UserRole = "super"
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

type User struct {
	ID         int       `db:"id" json:"id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	FirstName  string    `db:"first_name" json:"first_name"`
	LastName   string    `db:"last_name" json:"last_name"`
	Initials   string    `db:"initials" json:"initials"`
	Email      string    `db:"email" json:"email"`
	Password   string    `db:"password" json:"-"` // Hashed password
	FacilityID int       `db:"facility_id" json:"facility_id"`
	Role       UserRole  `db:"role" json:"role" validate:"required,oneof=super admin user"`
}

// CreateUserParams represents the parameters needed to create a new user
type CreateUserParams struct {
    FirstName  string   `json:"first_name" form:"first_name" validate:"required"`
    LastName   string   `json:"last_name" form:"last_name" validate:"required"`
    Initials   string   `json:"initials" form:"initials" validate:"required,max=10"`
    Email      string   `json:"email" form:"email" validate:"required,email"`
    Password   string   `json:"password" form:"password" validate:"required,min=8"`
    FacilityID int      `json:"facility_id" form:"facility_id" validate:"required,min=1"`
    Role       UserRole `json:"role" form:"role" validate:"required,oneof=super admin user"`
}

func (db *DB) GetUsersByFacilityCode(ctx context.Context, facilityCode string) ([]User, error) {
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

	var users []User
	for rows.Next() {
		var user User
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

func (db *DB) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
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

    var user User
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
        now,                // $1
        now,                // $2
        params.FirstName,   // $3
        params.LastName,    // $4
        params.Initials,    // $5
        params.Email,       // $6
        params.Password,    // $7 (should be pre-hashed)
        params.FacilityID,  // $8
        params.Role,        // $9
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

func (db *DB) UpdateUser(ctx context.Context, userID int, params CreateUserParams) (*User, error) {
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

    var user User
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
            now,                // $1
            params.FirstName,   // $2
            params.LastName,    // $3
            params.Initials,    // $4
            params.Email,       // $5
            params.Password,    // $6 (should be pre-hashed)
            params.FacilityID,  // $7
            params.Role,        // $8
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
            now,                // $1
            params.FirstName,   // $2
            params.LastName,    // $3
            params.Initials,    // $4
            params.Email,       // $5
            params.FacilityID,  // $6
            params.Role,        // $7
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