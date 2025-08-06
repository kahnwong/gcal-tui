package utils

import (
	"time"
)

func GenerateStartAndEndOfWeekTime() (time.Time, time.Time) { // generated via gemini
	now := time.Now()

	// current monday
	daysSinceMonday := (now.Weekday() - time.Monday + 7) % 7
	currentMonday := now.AddDate(0, 0, -int(daysSinceMonday))
	currentMonday = time.Date(currentMonday.Year(), currentMonday.Month(), currentMonday.Day(), 0, 0, 0, 0, currentMonday.Location())

	// next monday
	daysToNextMonday := (time.Monday - now.Weekday() + 7) % 7
	if daysToNextMonday == 0 { // If today is Monday
		daysToNextMonday = 7
	}

	upcomingMonday := now.Add(time.Duration(daysToNextMonday*24) * time.Hour)
	upcomingMonday = time.Date(upcomingMonday.Year(), upcomingMonday.Month(), upcomingMonday.Day(), 0, 0, 0, 0, upcomingMonday.Location())

	return currentMonday, upcomingMonday
}
