package main

import (
	"log/slog"
	"os"

	"github.com/kahnwong/gcal-tui/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("Error executing command", "error", err)
		os.Exit(1)
	}
}
