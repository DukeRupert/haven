package db

import "sort"

// GenerateProtectedDates creates ProtectedDate entries for a schedule for 1 year from start date
func GenerateProtectedDates(schedule Schedule) []ProtectedDate {
	protectedDates := []ProtectedDate{}

	// Create end date (1 year from start)
	endDate := schedule.StartDate.AddDate(1, 0, 0)

	// Initialize counters for both weekdays
	firstCount := 0
	secondCount := 0

	// Start from the schedule's start date
	currentDate := schedule.StartDate

	// Continue until we reach one year from start date
	for currentDate.Before(endDate) {
		currentWeekday := currentDate.Weekday()

		// Check if current date matches either weekday
		if currentWeekday == schedule.FirstWeekday {
			firstCount++
			// Mark every third occurrence as protected
			if firstCount%3 == 0 {
				protectedDates = append(protectedDates, ProtectedDate{
					ScheduleID: schedule.ID,
					Date:       currentDate,
					Available:  false, // marked as protected/unavailable
				})
			}
		}

		if currentWeekday == schedule.SecondWeekday {
			secondCount++
			// Mark every third occurrence as protected
			if secondCount%3 == 0 {
				protectedDates = append(protectedDates, ProtectedDate{
					ScheduleID: schedule.ID,
					Date:       currentDate,
					Available:  false, // marked as protected/unavailable
				})
			}
		}

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Sort the protected dates by date
	sort.Slice(protectedDates, func(i, j int) bool {
		return protectedDates[i].Date.Before(protectedDates[j].Date)
	})

	return protectedDates
}
