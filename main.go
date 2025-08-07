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

	events := calendar.FetchAllEvents()
	calendar.RenderTUI(events)
}
