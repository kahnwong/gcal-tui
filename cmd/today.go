package cmd

import (
	"github.com/kahnwong/gcal-tui/internal/calendar"
	"github.com/spf13/cobra"

	tea "charm.land/bubbletea/v2"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Calendar today view",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a today view with 1 column and 20 width
		model := calendar.NewModel(1, 20)
		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}
