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
	// Use a fixed date for consistent example output, e.g., Monday, August 4, 2025
	today := time.Date(2025, time.August, 4, 0, 0, 0, 0, now.Location())

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
			StartTime: today.Add(24 * time.Hour).Add(14 * time.Hour), // Tomorrow 2 PM (Tuesday)
			EndTime:   today.Add(24 * time.Hour).Add(16 * time.Hour),
		},
		{
			Title:     "Client Call",
			StartTime: today.Add(48 * time.Hour).Add(10 * time.Hour).Add(30 * time.Minute), // Day after tomorrow 10:30 AM (Wednesday)
			EndTime:   today.Add(48 * time.Hour).Add(11 * time.Hour).Add(30 * time.Minute),
		},
		{
			Title:     "GoLang Workshop",
			StartTime: today.Add(72 * time.Hour).Add(9 * time.Hour), // Thursday 9 AM
			EndTime:   today.Add(72 * time.Hour).Add(12 * time.Hour),
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

				for _, event := range events {
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

				eventText := fmt.Sprintf("[white:blue]%s: %02d:%02d %-7s[-:-]\n", dayName[:3], hour, minute, eventTitle)
				if isEventStart {
					// Display title only at the start time
					fmt.Fprintf(dayTextView, eventText)
				} else if isEventContinuing {
					// Fill the slot without the title
					fmt.Fprintf(dayTextView, "[white:blue]   %02d:%02d       [-:-]\n", hour, minute)
				} else {
					// Empty slot
					fmt.Fprintf(dayTextView, "       \n")
				}
			}
		}
		flex.AddItem(dayTextView, 0, 1, false) // Each day takes equal flexible width
	}

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
