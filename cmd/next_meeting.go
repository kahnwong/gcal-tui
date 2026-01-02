package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kahnwong/gcal-tui/internal/calendar"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var nextMeetingCmd = &cobra.Command{
	Use:   "next-meeting",
	Short: "Show the next upcoming calendar event",
	Long:  `Display the next upcoming calendar event and the time remaining until it starts. Updates every minute.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(calendar.InitialNextMeetingModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal().Err(err).Msg("Error running next-meeting command")
		}
	},
}

func init() {
	rootCmd.AddCommand(nextMeetingCmd)
}
