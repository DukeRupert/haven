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

// Helper function to group protected dates by date
func groupProtectedDates(dates []db.ProtectedDate) map[string][]db.ProtectedDate {
    groups := make(map[string][]db.ProtectedDate)
    for _, date := range dates {
        key := date.Date.Format("2006-01-02")
        groups[key] = append(groups[key], date)
    }
    return groups
}

// Update helper function to return all protected dates for a given day
func findProtectedDates(date time.Time, dates []db.ProtectedDate) []db.ProtectedDate {
    var dayDates []db.ProtectedDate
    for _, pd := range dates {
        if pd.Date.Year() == date.Year() && 
           pd.Date.Month() == date.Month() && 
           pd.Date.Day() == date.Day() {
            dayDates = append(dayDates, pd)
        }
    }
    return dayDates
}

// Update the day classes to account for multiple protected dates
func getDayClasses(props types.CalendarDayProps) string {
    classes := []string{}
    
    // Add position-based classes
    if isFirstWeek(props.Date) && props.Date.Weekday() == time.Monday {
        classes = append(classes, "rounded-tl-lg")
    }
    if isFirstWeek(props.Date) && props.Date.Weekday() == time.Sunday {
        classes = append(classes, "rounded-tr-lg")
    }
    if isLastWeek(props.Date) && props.Date.Weekday() == time.Monday {
        classes = append(classes, "rounded-bl-lg")
    }
    if isLastWeek(props.Date) && props.Date.Weekday() == time.Sunday {
        classes = append(classes, "rounded-br-lg")
    }

    // Add month-based classes
    if props.Date.Month() == props.CurrentMonth.Month() {
        classes = append(classes, "bg-white")
    } else {
        classes = append(classes, "bg-gray-50 text-gray-400")
    }

    return strings.Join(classes, " ")
}

// isFirstWeek checks if the given date falls in the first week of the calendar grid
func isFirstWeek(date time.Time) bool {
    // Get the first day of the month
    firstOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
    
    // Find the Monday that starts the calendar grid
    var firstGridDay time.Time
    if firstOfMonth.Weekday() != time.Monday {
        daysToSubtract := int(firstOfMonth.Weekday() - time.Monday)
        if daysToSubtract < 0 {
            daysToSubtract += 7
        }
        firstGridDay = firstOfMonth.AddDate(0, 0, -daysToSubtract)
    } else {
        firstGridDay = firstOfMonth
    }
    
    // Check if the date is in the first week by comparing with firstGridDay
    return date.After(firstGridDay.Add(-24*time.Hour)) && 
           date.Before(firstGridDay.AddDate(0, 0, 7))
}

// isLastWeek checks if the given date falls in the last week of the calendar grid
func isLastWeek(date time.Time) bool {
    // Get the last day of the month
    firstOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
    lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
    
    // Find the Sunday that ends the calendar grid
    var lastGridDay time.Time
    if lastOfMonth.Weekday() != time.Sunday {
        daysToAdd := int(time.Sunday - lastOfMonth.Weekday())
        if daysToAdd <= 0 {
            daysToAdd += 7
        }
        lastGridDay = lastOfMonth.AddDate(0, 0, daysToAdd)
    } else {
        lastGridDay = lastOfMonth
    }
    
    // Check if the date is in the last week by comparing with lastGridDay
    return date.After(lastGridDay.AddDate(0, 0, -7).Add(-24*time.Hour)) && 
           date.Before(lastGridDay.AddDate(0, 0, 1))
}