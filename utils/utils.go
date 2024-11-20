package utils

import (
	"time"

	"github.com/DukeRupert/haven/db"
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

// Helper functions (in a separate .go file)
func getDaysInMonth(date time.Time) []time.Time {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	// Calculate padding days for previous month
	startPadding := int(firstDay.Weekday())
	if startPadding == 0 {
		startPadding = 7
	}
	startPadding--

	// Calculate padding days for next month
	endPadding := 42 - (startPadding + lastDay.Day()) // 42 = 6 weeks * 7 days

	var days []time.Time

	// Add previous month padding days
	for i := startPadding - 1; i >= 0; i-- {
		days = append(days, firstDay.AddDate(0, 0, -i-1))
	}

	// Add current month days
	for i := 0; i < lastDay.Day(); i++ {
		days = append(days, firstDay.AddDate(0, 0, i))
	}

	// Add next month padding days
	for i := 0; i < endPadding; i++ {
		days = append(days, lastDay.AddDate(0, 0, i+1))
	}

	return days
}

func canToggleDate(protectedDate *db.ProtectedDate, userRole db.UserRole, currentUserID int) bool {
	if protectedDate == nil {
		return false
	}
	return userRole == db.UserRoleSuper ||
		userRole == db.UserRoleAdmin ||
		protectedDate.UserID == currentUserID
}

func findProtectedDate(date time.Time, protectedDates []db.ProtectedDate) *db.ProtectedDate {
	for _, pd := range protectedDates {
		if pd.Date.Year() == date.Year() &&
			pd.Date.Month() == date.Month() &&
			pd.Date.Day() == date.Day() {
			return &pd
		}
	}
	return nil
}

