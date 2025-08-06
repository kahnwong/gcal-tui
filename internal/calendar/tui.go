// generated via gemini. refactored by Karn Wong <karn@karnwong.me>
package calendar

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/kahnwong/gcal-tui/internal/utils"
	"github.com/rivo/tview"
	"golang.org/x/term"
	"os"
	"time"
)

var currentOffset int = 0 // Track horizontal scroll position

func RenderTUI(events []CalendarEvent) {
	app := tview.NewApplication()

	weekDays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	var dayViews []*tview.TextView

	maxVisibleDays := getMaxVisibleDays()

	// Ensure the week starts on Monday for consistent display
	startOfWeek, _ := utils.GenerateStartAndEndOfWeekTime()

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

				for _, event := range events {
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

func getMaxVisibleDays() int {
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
	return maxVisibleDays
}
