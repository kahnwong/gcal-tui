package main

import (
	"github.com/kahnwong/gcal-tui/cmd"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Error executing command")
	}
}
