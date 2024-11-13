package utils

import (
	"time"
)

// WeekdayString returns a readable string of the weekday
func WeekdayString(d time.Weekday) string {
    return d.String() // Built-in method returns "Sunday", "Monday", etc.
}

// WeekdayShort returns a 3-letter abbreviation
func WeekdayShort(d time.Weekday) string {
    return d.String()[:3] // Returns "Sun", "Mon", etc.
}

// WeekdayFriendly returns a more user-friendly format
func WeekdayFriendly(d time.Weekday) string {
    switch d {
    case time.Sunday:
        return "Every Sunday"
    case time.Monday:
        return "Every Monday"
    case time.Tuesday:
        return "Every Tuesday"
    case time.Wednesday:
        return "Every Wednesday"
    case time.Thursday:
        return "Every Thursday"
    case time.Friday:
        return "Every Friday"
    case time.Saturday:
        return "Every Saturday"
    default:
        return "Unknown Day"
    }
}