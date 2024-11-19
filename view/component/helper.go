package component

import (
	"strings"
	"time"

	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/types"
)

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

func getDayClasses(props types.CalendarDayProps) string {
	classes := []string{"py-1.5", "hover:bg-gray-100", "focus:z-10"}

	// Add background color
	if props.Date.Month() == props.CurrentMonth.Month() {
		classes = append(classes, "bg-white")
	} else {
		classes = append(classes, "bg-gray-50")
	}

	// Add text color based on protected status
	if props.ProtectedDate != nil {
		if props.ProtectedDate.Available {
			classes = append(classes, "text-green-600")
		} else {
			classes = append(classes, "text-red-600")
		}
	} else if props.Date.Month() == props.CurrentMonth.Month() {
		classes = append(classes, "text-gray-900")
	} else {
		classes = append(classes, "text-gray-400")
	}

	// Add corner rounding
	if props.Date.Day() == 1 && props.Date.Month() == props.CurrentMonth.Month() {
		classes = append(classes, "rounded-tl-lg")
	}
	// Add other corner cases...

	return strings.Join(classes, " ")
}

func getTimeClasses(props types.CalendarDayProps) string {
	classes := []string{"mx-auto", "flex", "size-7", "items-center", "justify-center", "rounded-full"}

	if props.ProtectedDate != nil {
		if !props.ProtectedDate.Available {
			classes = append(classes, "bg-red-100")
		} else {
			classes = append(classes, "bg-green-100")
		}
	}

	return strings.Join(classes, " ")
}

func canToggleDate(protectedDate *db.ProtectedDate, userRole db.UserRole, currentUserID int) bool {
	if protectedDate == nil {
		return false
	}
	return userRole == db.UserRoleSuper ||
		userRole == db.UserRoleAdmin ||
		protectedDate.ScheduleID == currentUserID
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
