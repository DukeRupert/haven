// internal/model/params/schedule.go
package params

import "time"

type CreateScheduleRequest struct {
	FirstWeekday  int    `form:"first_weekday" validate:"required,min=0,max=6"`
	SecondWeekday int    `form:"second_weekday" validate:"required,min=0,max=6"`
	StartDate     string `form:"start_date" validate:"required"`
}

type CreateScheduleByCodeParams struct {
	FacilityCode  string       `form:"facility_code"`
	UserInitials  string       `form:"user_initials"`
	FirstWeekday  time.Weekday `form:"first_weekday"`
	SecondWeekday time.Weekday `form:"second_weekday"`
	StartDate     time.Time    `form:"start_date"`
}

type UpdateScheduleParams struct {
	FirstWeekday  time.Weekday `form:"first_weekday" validate:"required,min=0,max=6"`
	SecondWeekday time.Weekday `form:"second_weekday" validate:"required,min=0,max=6"`
	StartDate     time.Time    `form:"start_date" validate:"required"`
}
