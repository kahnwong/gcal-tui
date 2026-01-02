package cmd

import (
	"os"

	"github.com/kahnwong/gcal-tui/internal/calendar"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

var rootCmd = &cobra.Command{
	Use:   "gcal-tui",
	Short: "A terminal-based Google Calendar viewer",
	Long:  `A beautiful terminal user interface for viewing your Google Calendar events in a weekly format.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(calendar.InitialModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal().Err(err).Msg("Error running root command")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Error executing root command")
	}
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}
