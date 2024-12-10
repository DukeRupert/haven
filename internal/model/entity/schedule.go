package entity

import "time"

type Schedule struct {
	ID            int          `db:"id" json:"id"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
	UserID        int          `db:"user_id" json:"user_id" validate:"required"`
	FirstWeekday  time.Weekday `db:"first_weekday" json:"first_weekday" validate:"required,min=0,max=6"`
	SecondWeekday time.Weekday `db:"second_weekday" json:"second_weekday" validate:"required,min=0,max=6"`
	StartDate     time.Time    `db:"start_date" json:"start_date" validate:"required"`
}

func (s Schedule) IsZero() bool {
	return s.ID == 0 &&
		s.CreatedAt.IsZero() &&
		s.UpdatedAt.IsZero() &&
		s.UserID == 0 &&
		s.StartDate.IsZero()
}
