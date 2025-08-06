package main

import (
	"github.com/kahnwong/gcal-tui/internal/calendar"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	events := calendar.FetchAllEvents()
	calendar.RenderTUI(events)
}
