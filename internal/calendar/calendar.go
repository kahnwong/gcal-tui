// initial code mostly generated via gemini
package calendar

import (
	"fmt"
	"sync"
	"time"

	"github.com/kahnwong/gcal-tui/internal/utils"

	cliBase "github.com/kahnwong/cli-base"
	"github.com/kahnwong/gcal-tui/configs"
	"github.com/kahnwong/gcal-tui/internal/gcal"
	"google.golang.org/api/calendar/v3"
)

type CalendarEvent struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
	Color     string
}

func ParseCalendars(color string, events *calendar.Events) ([]CalendarEvent, error) {
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
				return nil, fmt.Errorf("error parsing start time for event '%s': %w", item.Summary, err)
			}
			event.StartTime = startTime

			endTime, err := time.Parse(time.RFC3339, item.End.DateTime)
			if err != nil {
				return nil, fmt.Errorf("error parsing end time for event '%s': %w", item.Summary, err)
			}
			event.EndTime = endTime
		} else if item.Start.Date != "" {
			//// All-day event
			//// For all-day events, the API returns "YYYY-MM-DD".
			//// We can interpret this as the start of the day in UTC, or the local timezone
			//// depending on requirements. For simplicity, we'll parse it as a date and
			//// set time to midnight. Note that Google Calendar's all-day events
			//// endDate is exclusive, so it might be the next day.
			//startDate, err := time.Parse("2006-01-02", item.Start.Date)
			//if err != nil {
			//	return nil, fmt.Errorf("error parsing all-day start date for event '%s': %w", item.Summary, err)
			//}
			//event.StartTime = startDate
			//
			//endDate, err := time.Parse("2006-01-02", item.End.Date)
			//if err != nil {
			//	return nil, fmt.Errorf("error parsing all-day end date for event '%s': %w", item.Summary, err)
			//}
			//// For all-day events, Google Calendar's end date is exclusive.
			//// To represent the end of the last day, subtract a nanosecond.
			//event.EndTime = endDate.Add(-time.Nanosecond)
		} else {
			return nil, fmt.Errorf("event '%s' has no start or end time/date", item.Summary)
		}

		// time adjustment
		_, startTimeOffsetSeconds := event.StartTime.Zone()

		event.StartTime = event.StartTime.Add(time.Second * time.Duration(startTimeOffsetSeconds))
		event.EndTime = event.EndTime.Add(time.Second * time.Duration(startTimeOffsetSeconds))

		calendarEvents = append(calendarEvents, event)
	}

	return calendarEvents, nil
}

func FetchAllEvents(weekStart time.Time) ([]CalendarEvent, error) {
	var allEvents []CalendarEvent

	resultsCh := make(chan []CalendarEvent, 100) // Buffer size can be tuned
	errorsCh := make(chan error, 100)
	var accountsWg sync.WaitGroup
	for _, c := range configs.AppConfig.Accounts {
		accountsWg.Add(1)
		go func(account configs.Account) {
			defer accountsWg.Done()
			expandedPath, err := cliBase.ExpandHome(account.Credentials)
			if err != nil {
				errorsCh <- fmt.Errorf("failed to expand home path for account '%s': %w", account.Name, err)
				return
			}
			oathClientIDJson, err := gcal.ReadOauthClientID(expandedPath)
			if err != nil {
				errorsCh <- fmt.Errorf("failed to read OAuth client ID for account '%s': %w", account.Name, err)
				return
			}
			client, err := gcal.GetClient(account.Name, oathClientIDJson)
			if err != nil {
				errorsCh <- fmt.Errorf("failed to get client for account '%s': %w", account.Name, err)
				return
			}

			var calendarsWg sync.WaitGroup
			for _, calendarInfo := range account.Calendars {
				calendarsWg.Add(1)
				go func(calInfo configs.Calendar) {
					defer calendarsWg.Done()
					events, err := gcal.GetEvents(weekStart, calInfo.Id, client)
					if err != nil {
						errorsCh <- fmt.Errorf("failed to get events for calendar '%s': %w", calInfo.Id, err)
						return
					}
					calendarEvents, err := ParseCalendars(calInfo.Color, events)
					if err != nil {
						errorsCh <- fmt.Errorf("failed to parse calendars for calendar '%s': %w", calInfo.Id, err)
						return
					}
					resultsCh <- calendarEvents
				}(calendarInfo)
			}
			calendarsWg.Wait()
		}(c)
	}

	go func() {
		accountsWg.Wait()
		close(resultsCh)
		close(errorsCh)
	}()

	// Collect results and errors
	var errors []error
	done := make(chan struct{})
	go func() {
		for events := range resultsCh {
			allEvents = append(allEvents, events...)
		}
		close(done)
	}()

	for err := range errorsCh {
		errors = append(errors, err)
	}

	<-done

	// Return first error if any occurred
	if len(errors) > 0 {
		return nil, errors[0]
	}

	// for making current time in calendar
	now := roundToNearestHalfHour(utils.GetNowLocalAdjusted())
	allEvents = append(allEvents, CalendarEvent{
		Title:     "CURRENT TIME",
		StartTime: now,
		EndTime:   now.Add(time.Minute * 30),
		Color:     "red",
	})

	return allEvents, nil
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
