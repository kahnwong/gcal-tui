package calendar

import (
	cliBase "github.com/kahnwong/cli-base"
	"github.com/kahnwong/gcal-tui/configs"
	"github.com/kahnwong/gcal-tui/internal/gcal"
	"google.golang.org/api/calendar/v3"
	"time"
)

type CalendarEvent struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
}

func ParseCalendars(events *calendar.Events) []CalendarEvent {
	var calendarEvents []CalendarEvent

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

	return calendarEvents
}

func FetchAllEvents() []CalendarEvent {
	var allEvents []CalendarEvent
	for _, c := range configs.AppConfig.Accounts {
		oathClientIDJson := gcal.ReadOauthClientID(cliBase.ExpandHome(c.Credentials))
		client := gcal.GetClient(c.Name, oathClientIDJson)

		for _, calendarId := range c.Calendars {
			events := gcal.GetEvents(calendarId, client)
			calendarEvents := ParseCalendars(events)

			allEvents = append(allEvents, calendarEvents...)
		}
	}

	return allEvents
}
