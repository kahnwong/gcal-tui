package calendar

import (
	"sync"
	"time"

	cliBase "github.com/kahnwong/cli-base"
	"github.com/kahnwong/gcal-tui/configs"
	"github.com/kahnwong/gcal-tui/internal/gcal"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/calendar/v3"
)

type CalendarEvent struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
	Color     string // see <https://github.com/rivo/tview/blob/a4a78f1e05cbdedca1e2a5f2a1120fea713674db/ansi.go#L109-L124>
}

func ParseCalendars(color string, events *calendar.Events) []CalendarEvent {
	var calendarEvents []CalendarEvent

	for _, item := range events.Items {
		event := CalendarEvent{
			Title: item.Summary,
			Color: color,
		}

		// Handle all-day events vs. timed events
		if item.Start.DateTime != "" {
			// Timed event
			startTime, err := time.Parse(time.RFC3339, item.Start.DateTime)
			if err != nil {
				log.Error().Err(err).Msgf("error parsing start time for event: %s", item.Summary)
			}
			event.StartTime = startTime

			endTime, err := time.Parse(time.RFC3339, item.End.DateTime)
			if err != nil {
				log.Error().Err(err).Msgf("error parsing end time for event: %s", item.Summary)
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
				log.Error().Err(err).Msgf("error parsing all-day start date for event: %s", item.Summary)
			}
			event.StartTime = startDate

			endDate, err := time.Parse("2006-01-02", item.End.Date)
			if err != nil {
				log.Error().Err(err).Msgf("error parsing all-day end date for event: %s", item.Summary)
			}
			// For all-day events, Google Calendar's end date is exclusive.
			// To represent the end of the last day, subtract a nanosecond.
			event.EndTime = endDate.Add(-time.Nanosecond)
		} else {
			log.Error().Msgf("event '%s' has no start or end time/date", item.Summary)
		}

		calendarEvents = append(calendarEvents, event)
	}

	return calendarEvents
}

func FetchAllEvents(dayAdjustment int) []CalendarEvent {
	var allEvents []CalendarEvent

	resultsCh := make(chan []CalendarEvent, 100) // Buffer size can be tuned

	var accountsWg sync.WaitGroup

	for _, c := range configs.AppConfig.Accounts {
		accountsWg.Add(1)
		go func(account configs.Account) {
			defer accountsWg.Done()
			oathClientIDJson := gcal.ReadOauthClientID(cliBase.ExpandHome(account.Credentials))
			client := gcal.GetClient(account.Name, oathClientIDJson)

			var calendarsWg sync.WaitGroup
			for _, calendarInfo := range account.Calendars {
				calendarsWg.Add(1)
				go func(calInfo configs.Calendar) {
					defer calendarsWg.Done()
					events := gcal.GetEvents(dayAdjustment, calInfo.Id, client)
					calendarEvents := ParseCalendars(calInfo.Color, events)
					resultsCh <- calendarEvents
				}(calendarInfo)
			}
			calendarsWg.Wait()
		}(c)
	}

	go func() {
		accountsWg.Wait()
		close(resultsCh)
	}()

	for events := range resultsCh {
		allEvents = append(allEvents, events...)
	}

	return allEvents
}
