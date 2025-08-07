// generated via gemini. refactored by Karn Wong <karn@karnwong.me>
package calendar

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gdamore/tcell/v2"
	"github.com/kahnwong/gcal-tui/internal/utils"
	"github.com/rivo/tview"
	"golang.org/x/term"
)

var (
	currentOffset  int = 0 // Track horizontal scroll position
	maxVisibleDays int
	dayViews       []*tview.TextView
	flex           *tview.Flex
	timeScale      *tview.TextView
	mainFlex       *tview.Flex
	statusBarText  = "[yellow]Keys: [white]Ctrl+F[yellow]=Down, [white]Ctrl+B[yellow]=Up, [white]h[yellow]=Left, [white]l[yellow]=Right, [white]q[yellow]=Exit"
)

func RenderTUI(dayAdjustment int, events []CalendarEvent) {
	app := tview.NewApplication()

	startOfWeek, _ := utils.GenerateStartAndEndOfWeekTime(dayAdjustment) // Ensure the week starts on Monday for consistent display
	maxVisibleDays = getMaxVisibleDays()

	weekDays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	weekDaysWithDate := make([]string, len(weekDays))
	for i, day := range weekDays {
		currentDayDate := startOfWeek.AddDate(0, 0, i)
		dateString := currentDayDate.Format("Jan 2")
		weekDaysWithDate[i] = fmt.Sprintf(" %s - %s ", day, dateString)
	}

	// add marker for current time
	events = append(events, CalendarEvent{
		Title:     "",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(30 * time.Minute),
		Color:     "red",
	})

	const eventPadding = 23 // Event display padding width

	// Create all day views (but don't add to flex yet)
	for i, dayName := range weekDaysWithDate {
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
				var eventTitle string
				var eventColor string

				for _, event := range events {
					// Check if this slot is the exact start of an event
					if slotTime.Equal(event.StartTime) {
						isEventStart = true

						eventTitle = tview.Escape(event.Title)
						if len(eventTitle) >= 21 {
							eventTitle = tview.Escape(event.Title[:21]) // prevent new line overflow when rendering
						}
						eventColor = event.Color
						break
					}
					// Check if this slot is within an event (but not its start)
					if slotTime.After(event.StartTime) && slotTime.Before(event.EndTime) {
						isEventContinuing = true
						eventColor = event.Color
						break
					}
				}

				if isEventStart {
					// fill with title
					formatString := fmt.Sprintf("[black:%s]%%-%ds[-:-]\n", eventColor, eventPadding)
					_, err := fmt.Fprintf(dayTextView, formatString, eventTitle)
					if err != nil {
						log.Error().Err(err).Msg("failed to write event start fill")
					}
				} else if isEventContinuing {
					// fill without title
					_, err := fmt.Fprintf(dayTextView, "[black:%s]%*s[-:-]\n", eventColor, eventPadding, "")
					if err != nil {
						log.Error().Err(err).Msg("failed to write event end fill")
					}
				} else {
					// Empty slot
					_, err := fmt.Fprintf(dayTextView, "       \n")
					if err != nil {
						log.Error().Err(err).Msg("failed to write empty slot")
					}
				}
			}
		}
	}

	// Initial layout build
	rebuildLayout()

	// Set up context for proper cleanup of goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.SetInputCapture(setKeys(cancel, app))
	periodicResize(ctx, app)
	setStatusBar()

	if err := app.SetRoot(mainFlex, true).Run(); err != nil {
		panic(err)
	}
}

func getMaxVisibleDays() int {
	// Calculate initial number of visible days based on terminal width
	maxVisibleDays = 7 // Default to show all days
	if termWidth, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		// Each day column needs ~25 chars, time scale needs 8 chars
		availableWidth := termWidth - 8
		calculatedDays := availableWidth / 25
		if calculatedDays > 0 && calculatedDays < 7 {
			maxVisibleDays = calculatedDays
		}
	}
	return maxVisibleDays
}

func setKeys(cancel context.CancelFunc, app *tview.Application) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
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
			case 'q':
				// Exit the application
				cancel() // Cancel the resize detection goroutine
				app.Stop()
				return nil
			}
		}
		return event
	}
}
func rebuildLayout() {
	// Function to rebuild the flex layout based on current offset
	flex = tview.NewFlex()

	// Add time scale column
	timeScale = tview.NewTextView().SetDynamicColors(true)
	timeScale.SetBorder(true).SetTitle("Time")
	for i := 0; i < 24; i++ {
		// Display every hour, leave space for minutes
		_, err := fmt.Fprintf(timeScale, "%02d:00\n\n", i)
		if err != nil {
			log.Error().Err(err).Msg("failed to write hour")
		}
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
		statusText.SetText(statusBarText)
		statusText.SetTextAlign(tview.AlignCenter)
		mainFlex.AddItem(statusText, 1, 1, false)
	}
}

func handleResize() {
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
}

func periodicResize(ctx context.Context, app *tview.Application) {
	// Set up periodic resize checking with context for proper cleanup
	go func() { // don't remove `go func`
		var lastWidth int

		// Get initial terminal size
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			lastWidth = w
		}

		// Check for size changes every 250ms for more responsive resizing
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
					// Only trigger resize handling for horizontal changes
					// since vertical changes are handled automatically by tview
					if w != lastWidth {
						lastWidth = w
						app.QueueUpdateDraw(func() {
							handleResize()
						})
					}
				}
			}
		}
	}()
}

func setStatusBar() {
	// Create main layout with status bar at the bottom
	mainFlex = tview.NewFlex().SetDirection(tview.FlexRow)
	mainFlex.AddItem(flex, 0, 1, true)

	statusText := tview.NewTextView().SetDynamicColors(true)
	statusText.SetText(statusBarText)
	statusText.SetTextAlign(tview.AlignCenter)
	mainFlex.AddItem(statusText, 1, 1, false)
}
