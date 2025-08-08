package calendar

import (
	"sync"
	"time"

	"github.com/kahnwong/gcal-tui/internal/utils"

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
	Color     string
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

		// time adjustment
		_, startTimeOffsetSeconds := event.StartTime.Zone()

		event.StartTime = event.StartTime.Add(time.Second * time.Duration(startTimeOffsetSeconds))
		event.EndTime = event.EndTime.Add(time.Second * time.Duration(startTimeOffsetSeconds))

		calendarEvents = append(calendarEvents, event)
	}

	return calendarEvents
}

func FetchAllEvents(weekStart time.Time) []CalendarEvent {
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
					events := gcal.GetEvents(weekStart, calInfo.Id, client)
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

	// for making current time in calendar
	now := roundToNearestHalfHour(utils.GetNowLocalAdjusted())
	allEvents = append(allEvents, CalendarEvent{
		Title:     "CURRENT TIME",
		StartTime: now,
		EndTime:   now.Add(time.Minute * 30),
		Color:     "red",
	})

	return allEvents
}

func roundToNearestHalfHour(t time.Time) time.Time {
	minute := t.Minute()
	second := t.Second()
	nanosecond := t.Nanosecond()

	// Calculate total minutes past the hour, including seconds and nanoseconds as fractions of a minute
	totalMinutes := float64(minute) + float64(second)/60.0 + float64(nanosecond)/(60.0*1e9)

	var roundedTime time.Time

	if totalMinutes >= 45 {
		// Round up to the next hour
		roundedTime = t.Truncate(time.Hour).Add(time.Hour)
	} else if totalMinutes >= 15 {
		// Round to :30
		roundedTime = t.Truncate(time.Hour).Add(30 * time.Minute)
	} else {
		// Round down to :00
		roundedTime = t.Truncate(time.Hour)
	}

	return roundedTime
}
