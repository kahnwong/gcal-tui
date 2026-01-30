package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kahnwong/gcal-tui/internal/calendar"
	"github.com/spf13/cobra"
)

var nextMeetingCmd = &cobra.Command{
	Use:   "next-meeting",
	Short: "Show the next upcoming calendar event",
	Long:  `Display the next upcoming calendar event and the time remaining until it starts. Updates every minute.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(calendar.InitialNextMeetingModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nextMeetingCmd)
}
