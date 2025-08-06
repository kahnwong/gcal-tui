package gcal

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/calendar/v3"

	_ "github.com/kahnwong/gcal-tui/internal/logger"
)

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
