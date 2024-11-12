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
	FirstName  string   `json:"first_name" validate:"required"`
	LastName   string   `json:"last_name" validate:"required"`
	Initials   string   `json:"initials" validate:"required,max=10"`
	Email      string   `json:"email" validate:"required,email"`
	Password   string   `json:"password" validate:"required,min=8"`
	FacilityID int      `json:"facility_id" validate:"required,min=1"`
	Role       UserRole `json:"role" validate:"required,oneof=super admin user"`
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
