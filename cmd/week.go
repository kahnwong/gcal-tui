package cmd

import (
	"github.com/kahnwong/gcal-tui/internal/calendar"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

var weekCmd = &cobra.Command{
	Use:   "week",
	Short: "Calendar week view",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a week view with 7 columns and 20 width, starting on Monday
		model := calendar.NewModel(7, 20)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(weekCmd)
}
