package main

import (
	"os"

	"github.com/kahnwong/gcal-tui/internal/calendar"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	var events []calendar.CalendarEvent
	var dayAdjustment int
	if len(os.Args) == 1 {
		dayAdjustment = 0
	} else if len(os.Args) == 2 {
		if os.Args[1] == "next" {
			dayAdjustment = 7
		}
	}
	events = calendar.FetchAllEvents(dayAdjustment)

	calendar.RenderTUI(dayAdjustment, events)
}
