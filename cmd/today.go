package cmd

import (
	"github.com/kahnwong/gcal-tui/internal/calendar"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Calendar today view",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a today view with 1 column and 20 width
		model := calendar.NewModel(1, 20)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}
