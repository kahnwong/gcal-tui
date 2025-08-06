package main

import (
	"fmt"
	"github.com/rivo/tview"
	"time"

	_ "github.com/kahnwong/gcal-tui/internal/logger"
)

type Event struct {
	Title     string
	StartTime time.Time
	EndTime   time.Time
}

func main() {
	//oathClientIDJson := gcal.ReadOauthClientIDJSON()
	//client := gcal.GetClient(oathClientIDJson)
	//
	//ctx := context.Background()
	//srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	//if err != nil {
	//	log.Fatal().Err(err).Msg("Unable to retrieve Calendar client")
	//}

	//// show calendar lists: run manually because I'm too lazy to expose it
	//gcal.ListCalendars(srv)
	//
	//// show events
	//t := time.Now().Format(time.RFC3339)
	//events, err := srv.Events.List("primary").ShowDeleted(false).
	//	SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	//if err != nil {
	//	log.Fatal().Err(err).Msg("Unable to retrieve next ten of the user's events")
	//}
	//fmt.Println("Upcoming events:")
	//if len(events.Items) == 0 {
	//	fmt.Println("No upcoming events found.")
	//} else {
	//	for _, item := range events.Items {
	//		date := item.Start.DateTime
	//		if date == "" {
	//			date = item.Start.Date
	//		}
	//		fmt.Printf("%v (%v)\n", item.Summary, date)
	//	}
	//}

	//

	app := tview.NewApplication()
	textView := tview.NewTextView()
	textView.SetDynamicColors(true)
	textView.SetRegions(true)
	textView.SetWrap(false)

	// Define a slice of events for a single day (e.g., today).
	today := time.Now().Truncate(24 * time.Hour) // Get the start of today
	events := []Event{
		{
			Title:     "Daily Standup",
			StartTime: today.Add(9 * time.Hour),
			EndTime:   today.Add(9*time.Hour + 30*time.Minute),
		},
		{
			Title:     "Project Meeting",
			StartTime: today.Add(11 * time.Hour),
			EndTime:   today.Add(12 * time.Hour),
		},
		{
			Title:     "Lunch Break",
			StartTime: today.Add(12*time.Hour + 30*time.Minute),
			EndTime:   today.Add(13*time.Hour + 30*time.Minute),
		},
		{
			Title:     "Coding Session",
			StartTime: today.Add(14 * time.Hour),
			EndTime:   today.Add(17 * time.Hour),
		},
	}

	// Constants for display
	const totalHours = 24
	const rowsPerHour = 4 // 4 rows per hour (15-minute intervals)
	const totalRows = totalHours * rowsPerHour
	const colWidth = 40 // Width of the calendar view

	// Generate the calendar view
	calendarContent := generateCalendarView(events, today, totalRows, colWidth, rowsPerHour)
	textView.SetText(calendarContent)

	if err := app.SetRoot(textView, true).Run(); err != nil {
		panic(err)
	}
}

// generateCalendarView creates the string representation of the day's calendar.
func generateCalendarView(events []Event, day time.Time, totalRows, colWidth, rowsPerHour int) string {
	//var builder tview.ANSIWriter // Use ANSIWriter for color codes if not using SetDynamicColors(true)
	// Or simply use strings.Builder and raw ANSI codes if not relying on tview's dynamic colors for regions
	// For simplicity with tview.TextView and its dynamic colors, we'll build the string directly.

	// Initialize a 2D array to represent the screen cells for the day
	// Each cell will hold a character to be printed.
	screen := make([][]rune, totalRows)
	for i := range screen {
		screen[i] = make([]rune, colWidth)
		for j := range screen[i] {
			if j == 0 || j == colWidth-1 || j == 7 { // Vertical line for time and event separation
				screen[i][j] = '|'
			} else {
				screen[i][j] = ' '
			}
		}
	}

	// Populate time labels
	for h := 0; h < 24; h++ {
		timeStr := fmt.Sprintf("%02d:00", h)
		row := h * rowsPerHour
		for i, r := range timeStr {
			if row < totalRows && i < colWidth {
				screen[row][i] = r
			}
		}
	}

	// Fill events
	for _, event := range events {
		// Calculate start and end rows for the event
		startOfDay := day
		eventStartMinutes := float64(event.StartTime.Sub(startOfDay).Minutes())
		eventEndMinutes := float64(event.EndTime.Sub(startOfDay).Minutes())

		startRow := int(eventStartMinutes / (60.0 / float64(rowsPerHour)))
		endRow := int(eventEndMinutes / (60.0 / float64(rowsPerHour)))

		// Ensure rows are within bounds
		if startRow < 0 {
			startRow = 0
		}
		if endRow > totalRows {
			endRow = totalRows
		}
		if startRow >= totalRows || endRow <= 0 || startRow >= endRow {
			continue // Event is outside the displayable day or invalid
		}

		// Fill the rectangle for the event
		for r := startRow; r < endRow; r++ {
			if r < totalRows {
				// Fill with a background color and print title if at the start
				fillChar := '█' // Unicode block character for solid fill
				// You could also use ANSI escape codes directly for background color
				// For simplicity with tview's SetDynamicColors, we'll use a specific char and then let tview color it.
				// A more advanced approach would involve tcell primitives or custom drawing.

				// Print the title on the first row of the event, centered or at a fixed position
				eventTitle := fmt.Sprintf(" %s ", event.Title) // Add spaces for padding
				if r == startRow {
					// Clear the space for the title first
					for c := 8; c < colWidth-1; c++ {
						screen[r][c] = fillChar
					}
					// Place title
					titleStartCol := 8 + (colWidth-8-len(eventTitle))/2 // Center in the event column
					if titleStartCol < 8 {
						titleStartCol = 8
					}
					for i, char := range eventTitle {
						if titleStartCol+i < colWidth-1 {
							screen[r][titleStartCol+i] = char
						}
					}
				} else {
					for c := 8; c < colWidth-1; c++ {
						screen[r][c] = fillChar
					}
				}
			}
		}
	}

	// Build the final string
	var output string
	for r := 0; r < totalRows; r++ {
		output += string(screen[r]) + "\n"
	}
	return output
}
