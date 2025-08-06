package main

import (
	"fmt"
	"github.com/rivo/tview"
	"time"

	_ "github.com/kahnwong/gcal-tui/internal/logger"
)

type CalendarEvent struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
}

func main() {
	//oathClientIDJson := gcal.ReadOauthClientIDJSON()
	//client := gcal.GetClient(oathClientIDJson)
	//
	//ctx := context.Background()
	//srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	//if err != nil {
	//	log.Fatal().Err(err).Msg("Unable to retrieve Calendar client")
	//}

	//// show calendar lists: run manually because I'm too lazy to expose it
	//gcal.ListCalendars(srv)
	//
	//// show events
	//t := time.Now().Format(time.RFC3339)
	//events, err := srv.Events.List("primary").ShowDeleted(false).
	//	SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	//if err != nil {
	//	log.Fatal().Err(err).Msg("Unable to retrieve next ten of the user's events")
	//}
	//fmt.Println("Upcoming events:")
	//if len(events.Items) == 0 {
	//	fmt.Println("No upcoming events found.")
	//} else {
	//	for _, item := range events.Items {
	//		date := item.Start.DateTime
	//		if date == "" {
	//			date = item.Start.Date
	//		}
	//		fmt.Printf("%v (%v)\n", item.Summary, date)
	//	}
	//}

	//
	app := tview.NewApplication()

	// Define some sample events
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	events := []CalendarEvent{
		{
			Title:     "Team Meeting",
			StartTime: today.Add(9 * time.Hour),
			EndTime:   today.Add(10 * time.Hour),
		},
		{
			Title:     "Lunch Break",
			StartTime: today.Add(12 * time.Hour).Add(30 * time.Minute),
			EndTime:   today.Add(13 * time.Hour),
		},
		{
			Title:     "Project Review",
			StartTime: today.Add(24 * time.Hour).Add(14 * time.Hour), // Tomorrow 2 PM
			EndTime:   today.Add(24 * time.Hour).Add(16 * time.Hour),
		},
		{
			Title:     "Client Call",
			StartTime: today.Add(48 * time.Hour).Add(10 * time.Hour).Add(30 * time.Minute), // Day after tomorrow 10:30 AM
			EndTime:   today.Add(48 * time.Hour).Add(11 * time.Hour).Add(30 * time.Minute),
		},
	}

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

	// Adjust start of week to Monday
	startOfWeek := today.Add(time.Duration(time.Monday-today.Weekday()) * 24 * time.Hour)
	if today.Weekday() == time.Sunday { // Handle Sunday case for startOfWeek
		startOfWeek = today.Add(-6 * 24 * time.Hour)
	}

	for i, dayName := range weekDays {
		dayTextView := tview.NewTextView().SetDynamicColors(true)
		dayTextView.SetBorder(true).SetTitle(dayName)

		currentDay := startOfWeek.Add(time.Duration(i) * 24 * time.Hour)

		// Populate day view with time slots
		for hour := 0; hour < 24; hour++ {
			for minute := 0; minute < 60; minute += 30 { // 30-minute intervals
				slotTime := currentDay.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
				isEventSlot := false
				eventTitle := ""

				for _, event := range events {
					// Check if the current slot falls within an event's time
					if (slotTime.Equal(event.StartTime) || slotTime.After(event.StartTime)) && slotTime.Before(event.EndTime) {
						isEventSlot = true
						eventTitle = event.Title
						break // Found an event for this slot
					}
				}

				if isEventSlot {
					// Use a background color or specific characters to represent the event
					fmt.Fprintf(dayTextView, "[white:blue]%s: %02d:%02d %-7s[-:-]\n", dayName[:3], hour, minute, eventTitle) // Fill with blue background
				} else {
					fmt.Fprintf(dayTextView, "   %02d:%02d       \n", hour, minute) // Empty slot
				}
			}
		}
		flex.AddItem(dayTextView, 0, 1, false) // Each day takes equal flexible width
	}

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
