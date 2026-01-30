package gcal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var ctx = context.Background()

func ListCalendars(srv *calendar.Service) error {
	calendarListCall := srv.CalendarList.List()
	calendarList, err := calendarListCall.Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve calendar list: %w", err)
	}

	if len(calendarList.Items) == 0 {
		fmt.Println("No calendars found.")
	} else {
		fmt.Println("Available Calendar IDs:")
		for _, item := range calendarList.Items {
			fmt.Printf("- %s (Summary: %s)\n", item.Id, item.Summary)
		}
	}
	return nil
}

func GetEvents(weekStart time.Time, calendarId string, client *http.Client) (*calendar.Events, error) {
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %w", err)
	}

	//// show calendars list: run manually because I'm too lazy to expose it
	// ListCalendars(srv)

	// show events
	events, err := srv.Events.List(calendarId).ShowDeleted(false).
		SingleEvents(true).
		TimeMin(weekStart.Format(time.RFC3339)).
		MaxResults(15).OrderBy("startTime").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve next ten of the user's events: %w", err)
	}

	return events, nil
}
