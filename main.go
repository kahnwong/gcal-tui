package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kahnwong/gcal-tui/internal/gcal"
	"github.com/kahnwong/gcal-tui/internal/utils"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type CalendarEvent struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
}

func main() {
	// init log
	log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// login
	oathClientIDJson := gcal.ReadOauthClientIDJSON()
	client := gcal.GetClient(oathClientIDJson)

	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve Calendar client")
	}

	//// show calendar lists: run manually because I'm too lazy to expose it
	//gcal.ListCalendars(srv)
	//
	// show events
	var calendarEvents []CalendarEvent
	//t := time.Now().Format(time.RFC3339)
	currentMonday, upcomingMonday := utils.GenerateStartAndEndOfWeekTime()
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(currentMonday.Format(time.RFC3339)).TimeMax(upcomingMonday.Format(time.RFC3339)).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve next ten of the user's events")
	}
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

	//
	app := tview.NewApplication()

	// Define some sample events
	now := time.Now()
	// Use a fixed date for consistent example output, e.g., Monday, August 4, 2025
	today := time.Date(2025, time.August, 4, 0, 0, 0, 0, now.Location())

	// Generate columns for each day of the week
	weekDays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	var dayViews []*tview.TextView
	var currentOffset int = 0 // Track horizontal scroll position

	// Calculate initial number of visible days based on terminal width
	var maxVisibleDays int = 7 // Default to show all days
	if termWidth, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		// Each day column needs ~25 chars, time scale needs 8 chars
		availableWidth := termWidth - 8
		calculatedDays := availableWidth / 25
		if calculatedDays > 0 && calculatedDays < 7 {
			maxVisibleDays = calculatedDays
		}
	}

	// Ensure the week starts on Monday for consistent display
	// 'today' is Monday, August 4, 2025, so startOfWeek will be today
	startOfWeek := today

	// Create all day views (but don't add to flex yet)
	for i, dayName := range weekDays {
		dayTextView := tview.NewTextView().SetDynamicColors(true)
		dayTextView.SetBorder(true).SetTitle(dayName)
		dayViews = append(dayViews, dayTextView)

		currentDay := startOfWeek.Add(time.Duration(i) * 24 * time.Hour)

		// Populate day view with time slots
		for hour := 0; hour < 24; hour++ {
			for minute := 0; minute < 60; minute += 30 { // 30-minute intervals
				slotTime := currentDay.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)

				isEventStart := false
				isEventContinuing := false
				eventTitle := ""

				for _, event := range calendarEvents {
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

				if isEventStart {
					// Calculate the length of the eventTitle
					paddingValue := 23 //- len(eventTitle)

					// Construct the format string dynamically
					formatString := fmt.Sprintf("[white:blue]%%-%ds[-:-]\n", paddingValue)

					// Use the dynamically created format string with Fprintf
					fmt.Fprintf(dayTextView, formatString, eventTitle)
					//fmt.Fprintf(dayTextView, "[white:blue]%-7s[-:-]\n", eventTitle)
				} else if isEventContinuing {
					// Fill the slot without the title
					fmt.Fprintf(dayTextView, "[white:blue]                       [-:-]\n")
				} else {
					// Empty slot
					fmt.Fprintf(dayTextView, "       \n")
				}
			}
		}
	}

	// Function to rebuild the flex layout based on current offset
	var flex *tview.Flex
	var timeScale *tview.TextView
	var mainFlex *tview.Flex

	rebuildLayout := func() {
		flex = tview.NewFlex()

		// Add time scale column
		timeScale = tview.NewTextView().SetDynamicColors(true)
		timeScale.SetBorder(true).SetTitle("Time")
		for i := 0; i < 24; i++ {
			fmt.Fprintf(timeScale, "%02d:00\n\n", i) // Display every hour, leave space for minutes
		}
		flex.AddItem(timeScale, 8, 1, false) // Fixed width for time scale

		// Add visible day columns based on current offset
		for i := 0; i < maxVisibleDays && (currentOffset+i) < len(dayViews); i++ {
			dayIndex := currentOffset + i
			flex.AddItem(dayViews[dayIndex], 25, 1, false)
		}

		// Update the main layout if it exists
		if mainFlex != nil {
			mainFlex.Clear()
			mainFlex.AddItem(flex, 0, 1, true)

			// Add status bar
			statusText := tview.NewTextView().SetDynamicColors(true)
			statusText.SetText("[yellow]Keys: [white]Ctrl+F[yellow]=Scroll Down, [white]Ctrl+B[yellow]=Scroll Up, [white]h[yellow]=Scroll Left, [white]l[yellow]=Scroll Right, [white]Esc[yellow]=Exit")
			statusText.SetTextAlign(tview.AlignCenter)
			mainFlex.AddItem(statusText, 1, 1, false)
		}
	}

	// Initial layout build
	rebuildLayout()

	// Add input handler for scrolling
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlF:
			// Scroll down all currently visible views
			for i := 0; i < maxVisibleDays && (currentOffset+i) < len(dayViews); i++ {
				dayIndex := currentOffset + i
				row, _ := dayViews[dayIndex].GetScrollOffset()
				dayViews[dayIndex].ScrollTo(row+5, 0) // Scroll down by 5 lines
			}
			// Also scroll the time scale
			if timeScale != nil {
				timeRow, _ := timeScale.GetScrollOffset()
				timeScale.ScrollTo(timeRow+5, 0)
			}
			return nil
		case tcell.KeyCtrlB:
			// Scroll up all currently visible views (bonus feature)
			for i := 0; i < maxVisibleDays && (currentOffset+i) < len(dayViews); i++ {
				dayIndex := currentOffset + i
				row, _ := dayViews[dayIndex].GetScrollOffset()
				if row >= 5 {
					dayViews[dayIndex].ScrollTo(row-5, 0) // Scroll up by 5 lines
				} else {
					dayViews[dayIndex].ScrollTo(0, 0) // Scroll to top
				}
			}
			// Also scroll the time scale up
			if timeScale != nil {
				timeRow, _ := timeScale.GetScrollOffset()
				if timeRow >= 5 {
					timeScale.ScrollTo(timeRow-5, 0)
				} else {
					timeScale.ScrollTo(0, 0)
				}
			}
			return nil
		case tcell.KeyRune:
			// Handle character keys for horizontal scrolling
			switch event.Rune() {
			case 'h':
				// Scroll left (show previous day columns)
				if currentOffset > 0 {
					currentOffset--
					rebuildLayout()
				}
				return nil
			case 'l':
				// Scroll right (show next day columns)
				if currentOffset+maxVisibleDays < len(dayViews) {
					currentOffset++
					rebuildLayout()
				}
				return nil
			case 'r':
				// Refresh and recalculate terminal size
				if termWidth, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
					availableWidth := termWidth - 8
					calculatedDays := availableWidth / 25
					if calculatedDays > 0 && calculatedDays <= 7 {
						maxVisibleDays = calculatedDays
						// Adjust offset if needed
						if currentOffset+maxVisibleDays > len(dayViews) {
							currentOffset = len(dayViews) - maxVisibleDays
							if currentOffset < 0 {
								currentOffset = 0
							}
						}
						rebuildLayout()
					}
				}
				return nil

			case 'q':
				// Exit the application
				app.Stop()
				return nil
			}
		}
		return event
	})

	// Create main layout with status bar at the bottom
	mainFlex = tview.NewFlex().SetDirection(tview.FlexRow)
	mainFlex.AddItem(flex, 0, 1, true)

	statusText := tview.NewTextView().SetDynamicColors(true)
	statusText.SetText("[yellow]Keys: [white]Ctrl+F[yellow]=Down, [white]Ctrl+B[yellow]=Up, [white]h[yellow]=Left, [white]l[yellow]=Right, [white]+/-[yellow]=Resize, [white]r[yellow]=Refresh Size, [white]Esc[yellow]=Exit")
	statusText.SetTextAlign(tview.AlignCenter)
	mainFlex.AddItem(statusText, 1, 1, false)

	if err := app.SetRoot(mainFlex, true).Run(); err != nil {
		panic(err)
	}
}
