package main

import (
	"context"
	"fmt"
	"github.com/kahnwong/gcal-tui/internal/gcal"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"time"

	_ "github.com/kahnwong/gcal-tui/internal/logger"
)

type CalendarEvent struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
}

func main() {
	oathClientIDJson := gcal.ReadOauthClientIDJSON()
	client := gcal.GetClient(oathClientIDJson)

	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve Calendar client")
	}

	//// show calendar lists: run manually because I'm too lazy to expose it
	//gcal.ListCalendars(srv)
	//
	// show events
	var calendarEvents []CalendarEvent
	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve next ten of the user's events")
	}
	for _, item := range events.Items {
		event := CalendarEvent{
			Title: item.Summary,
		}

		// Handle all-day events vs. timed events
		if item.Start.DateTime != "" {
			// Timed event
			startTime, err := time.Parse(time.RFC3339, item.Start.DateTime)
			if err != nil {
				//fmt.Errorf("error parsing start time for event '%s': %w", item.Summary, err)
			}
			event.StartTime = startTime

			endTime, err := time.Parse(time.RFC3339, item.End.DateTime)
			if err != nil {
				//fmt.Errorf("error parsing end time for event '%s': %w", item.Summary, err)
			}
			event.EndTime = endTime
		} else if item.Start.Date != "" {
			// All-day event
			// For all-day events, the API returns "YYYY-MM-DD".
			// We can interpret this as the start of the day in UTC, or the local timezone
			// depending on requirements. For simplicity, we'll parse it as a date and
			// set time to midnight. Note that Google Calendar's all-day events
			// endDate is exclusive, so it might be the next day.
			startDate, err := time.Parse("2006-01-02", item.Start.Date)
			if err != nil {
				//fmt.Errorf("error parsing all-day start date for event '%s': %w", item.Summary, err)
			}
			event.StartTime = startDate

			endDate, err := time.Parse("2006-01-02", item.End.Date)
			if err != nil {
				//fmt.Errorf("error parsing all-day end date for event '%s': %w", item.Summary, err)
			}
			// For all-day events, Google Calendar's end date is exclusive.
			// To represent the end of the last day, subtract a nanosecond.
			event.EndTime = endDate.Add(-time.Nanosecond)
		} else {
			//fmt.Errorf("event '%s' has no start or end time/date", item.Summary)
		}

		calendarEvents = append(calendarEvents, event)
	}

	//
	app := tview.NewApplication()

	// Define some sample events
	now := time.Now()
	// Use a fixed date for consistent example output, e.g., Monday, August 4, 2025
	today := time.Date(2025, time.August, 4, 0, 0, 0, 0, now.Location())

	//events := []CalendarEvent{
	//	{
	//		Title:     "Team Meeting",
	//		StartTime: today.Add(9 * time.Hour),
	//		EndTime:   today.Add(10 * time.Hour),
	//	},
	//	{
	//		Title:     "Lunch Break",
	//		StartTime: today.Add(12 * time.Hour).Add(30 * time.Minute),
	//		EndTime:   today.Add(13 * time.Hour),
	//	},
	//	{
	//		Title:     "Project Review",
	//		StartTime: today.Add(24 * time.Hour).Add(14 * time.Hour), // Tomorrow 2 PM (Tuesday)
	//		EndTime:   today.Add(24 * time.Hour).Add(16 * time.Hour),
	//	},
	//	{
	//		Title:     "Client Call",
	//		StartTime: today.Add(48 * time.Hour).Add(10 * time.Hour).Add(30 * time.Minute), // Day after tomorrow 10:30 AM (Wednesday)
	//		EndTime:   today.Add(48 * time.Hour).Add(11 * time.Hour).Add(30 * time.Minute),
	//	},
	//	{
	//		Title:     "GoLang Workshop",
	//		StartTime: today.Add(72 * time.Hour).Add(9 * time.Hour), // Thursday 9 AM
	//		EndTime:   today.Add(72 * time.Hour).Add(12 * time.Hour),
	//	},
	//}

	// Create the main flex layout for the week view
	flex := tview.NewFlex()

	// Add time scale column
	timeScale := tview.NewTextView().SetDynamicColors(true)
	timeScale.SetBorder(true).SetTitle("Time")
	for i := 0; i < 24; i++ {
		fmt.Fprintf(timeScale, "%02d:00\n\n", i) // Display every hour, leave space for minutes
	}
	flex.AddItem(timeScale, 8, 1, false) // Fixed width for time scale

	// Generate columns for each day of the week
	weekDays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

	// Ensure the week starts on Monday for consistent display
	// 'today' is Monday, August 4, 2025, so startOfWeek will be today
	startOfWeek := today

	for i, dayName := range weekDays {
		dayTextView := tview.NewTextView().SetDynamicColors(true)
		dayTextView.SetBorder(true).SetTitle(dayName)

		currentDay := startOfWeek.Add(time.Duration(i) * 24 * time.Hour)

		// Populate day view with time slots
		for hour := 0; hour < 24; hour++ {
			for minute := 0; minute < 60; minute += 30 { // 30-minute intervals
				slotTime := currentDay.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)

				isEventStart := false
				isEventContinuing := false
				eventTitle := ""

				for _, event := range calendarEvents {
					// Check if this slot is the exact start of an event
					if slotTime.Equal(event.StartTime) {
						isEventStart = true
						eventTitle = event.Title
						break
					}
					// Check if this slot is within an event (but not its start)
					if slotTime.After(event.StartTime) && slotTime.Before(event.EndTime) {
						isEventContinuing = true
						break
					}
				}

				if isEventStart {
					// Calculate the length of the eventTitle
					paddingValue := 23 //- len(eventTitle)

					// Construct the format string dynamically
					formatString := fmt.Sprintf("[white:blue]%%-%ds[-:-]\n", paddingValue)

					// Use the dynamically created format string with Fprintf
					fmt.Fprintf(dayTextView, formatString, eventTitle)
					//fmt.Fprintf(dayTextView, "[white:blue]%-7s[-:-]\n", eventTitle)
				} else if isEventContinuing {
					// Fill the slot without the title
					fmt.Fprintf(dayTextView, "[white:blue]                       [-:-]\n")
				} else {
					// Empty slot
					fmt.Fprintf(dayTextView, "       \n")
				}
			}
		}
		flex.AddItem(dayTextView, 25, 1, false) // Each day takes equal flexible width
	}

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
