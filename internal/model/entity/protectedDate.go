// internal/model/entity/protectedDate.go
package entity

import "time"

type ProtectedDate struct {
	ID         int       `db:"id" json:"id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	ScheduleID int       `db:"schedule_id" json:"schedule_id" validate:"required"`
	Date       time.Time `db:"date" json:"date" validate:"required"`
	Available  bool      `db:"available" json:"available"`
	UserID     int       `db:"user_id" json:"user_id" validate:"required"`
	FacilityID int       `db:"facility_id" json:"facility_id" validate:"required"`
}
