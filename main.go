package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kahnwong/gcal-tui/internal/gcal"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func main() {
	oathClientIDJson := gcal.ReadOauthClientIDJSON()
	client := gcal.GetClient(oathClientIDJson)

	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	// show calendar lists
	// 3. Call the CalendarList.List() method
	calendarListCall := srv.CalendarList.List()
	// You can add optional parameters, e.g., to show hidden calendars:
	// calendarListCall.ShowHidden(true)

	calendarList, err := calendarListCall.Do()
	if err != nil {
		log.Fatalf("Unable to retrieve calendar list: %v", err)
	}

	// 4. Iterate and print the calendar IDs
	if len(calendarList.Items) == 0 {
		fmt.Println("No calendars found.")
	} else {
		fmt.Println("Available Calendar IDs:")
		for _, item := range calendarList.Items {
			fmt.Printf("- %s (Summary: %s)\n", item.Id, item.Summary)
		}
	}

	// show events

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}
}
