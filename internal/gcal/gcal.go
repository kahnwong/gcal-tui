package gcal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kahnwong/gcal-tui/internal/utils"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var ctx = context.Background()

func ListCalendars(srv *calendar.Service) {
	calendarListCall := srv.CalendarList.List()
	calendarList, err := calendarListCall.Do()
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve calendar list")
	}

	if len(calendarList.Items) == 0 {
		fmt.Println("No calendars found.")
	} else {
		fmt.Println("Available Calendar IDs:")
		for _, item := range calendarList.Items {
			fmt.Printf("- %s (Summary: %s)\n", item.Id, item.Summary)
		}
	}
}

func GetEvents(dayAdjustment int, calendarId string, client *http.Client) *calendar.Events {
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve Calendar client")
	}

	//// show calendars list: run manually because I'm too lazy to expose it
	// ListCalendars(srv)

	// show events
	startTime, endTime := utils.GenerateStartAndEndOfWeekTime(dayAdjustment)
	events, err := srv.Events.List(calendarId).ShowDeleted(false).
		SingleEvents(true).
		TimeMin(startTime.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		MaxResults(50).OrderBy("startTime").Do()
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve next ten of the user's events")
	}

	return events
}
