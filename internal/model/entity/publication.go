package entity

import "time"

type SchedulePublication struct {
   ID              int       `json:"id"`
   CreatedAt       time.Time `json:"created_at"`
   UpdatedAt       time.Time `json:"updated_at"`
   FacilityID      int       `json:"facility_id"`
   PublishedThrough time.Time `json:"published_through"`
}