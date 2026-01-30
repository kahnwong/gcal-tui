package calendar

import (
	"fmt"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kahnwong/gcal-tui/internal/utils"
)

// GetNextMeeting fetches all events and returns the next upcoming event
func GetNextMeeting() (*CalendarEvent, error) {
	now := utils.GetNowLocalAdjusted()

	// Fetch events starting from now for the next week
	weekStart := now.Truncate(24 * time.Hour)
	allEvents, err := FetchAllEvents(weekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	// Filter out past events and the "CURRENT TIME" marker
	var upcomingEvents []CalendarEvent
	for _, event := range allEvents {
		if event.StartTime.After(now) && event.Title != "CURRENT TIME" {
			upcomingEvents = append(upcomingEvents, event)
		}
	}

	if len(upcomingEvents) == 0 {
		return nil, fmt.Errorf("no upcoming events found")
	}

	// Sort events by start time
	sort.Slice(upcomingEvents, func(i, j int) bool {
		return upcomingEvents[i].StartTime.Before(upcomingEvents[j].StartTime)
	})

	return &upcomingEvents[0], nil
}

// FormatTimeUntil returns a human-readable string showing time remaining until the event
func FormatTimeUntil(eventTime time.Time) string {
	now := utils.GetNowLocalAdjusted()
	duration := eventTime.Sub(now)

	if duration < 0 {
		return "Event has already started"
	}

	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%d day(s) and %d hour(s)", days, hours)
		}
		return fmt.Sprintf("%d day(s)", days)
	}

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%d hour(s) and %d minute(s)", hours, minutes)
		}
		return fmt.Sprintf("%d hour(s)", hours)
	}

	if minutes > 0 {
		return fmt.Sprintf("%d minute(s)", minutes)
	}

	return "Less than a minute"
}

// GetTimeColor returns the appropriate color based on time remaining until event
func GetTimeColor(eventTime time.Time) lipgloss.Color {
	now := utils.GetNowLocalAdjusted()
	duration := eventTime.Sub(now)

	// Convert to minutes for easier comparison
	minutes := int(duration.Minutes())

	if minutes <= 15 {
		return lipgloss.Color("#FF0000") // Red - 15 minutes or less
	} else if minutes <= 60 {
		return lipgloss.Color("#FFFF00") // Yellow - 1 hour or less
	} else {
		return lipgloss.Color("#00FF00") // Green - more than 1 hour
	}
}

// DisplayNextMeeting shows the next meeting information with styled TUI
func DisplayNextMeeting() {
	nextEvent, err := GetNextMeeting()
	if err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Align(lipgloss.Center).
			Padding(2)
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		return
	}

	timeUntil := FormatTimeUntil(nextEvent.StartTime)
	timeColor := GetTimeColor(nextEvent.StartTime)

	// Title style - large text
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Align(lipgloss.Center).
		MarginTop(1).
		Width(60)

	// Time remaining style - extra large text, centered, colored
	timeStyle := lipgloss.NewStyle().
		Foreground(timeColor).
		Bold(true).
		Align(lipgloss.Center).
		Width(60)

	// Event details style
	detailsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Align(lipgloss.Center).
		Width(60)

	// Container style
	containerStyle := lipgloss.NewStyle().
		Align(lipgloss.Center)

	// Render the display
	title := titleStyle.Render("📅 " + nextEvent.Title)
	timeRemaining := timeStyle.Render("⏰ " + timeUntil)
	startTime := detailsStyle.Render("Starts: " + nextEvent.StartTime.Format("Monday, January 2, 2006 at 3:04 PM"))

	content := lipgloss.JoinVertical(lipgloss.Center, title, timeRemaining, startTime)
	display := containerStyle.Render(content)

	fmt.Println(display)
}

// NextMeetingModel represents the TUI model for the next meeting display
type NextMeetingModel struct {
	nextEvent  *CalendarEvent
	err        error
	lastUpdate time.Time
}

// tickMsg is sent every minute to update the display
type tickMsg time.Time

// doTick returns a command that sends a tick message every minute
func doTick() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// InitialNextMeetingModel creates the initial model for next meeting display
func InitialNextMeetingModel() NextMeetingModel {
	nextEvent, err := GetNextMeeting()
	return NextMeetingModel{
		nextEvent:  nextEvent,
		err:        err,
		lastUpdate: time.Now(),
	}
}

// Init initializes the next meeting model and starts the ticker
func (m NextMeetingModel) Init() tea.Cmd {
	return doTick()
}

// Update handles messages and updates the model
func (m NextMeetingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tickMsg:
		// Update the next meeting data every minute
		nextEvent, err := GetNextMeeting()
		m.nextEvent = nextEvent
		m.err = err
		m.lastUpdate = time.Time(msg)
		return m, doTick() // Schedule next tick
	}
	return m, nil
}

// View renders the next meeting display
func (m NextMeetingModel) View() string {
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Align(lipgloss.Center).
			Padding(2)
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.nextEvent == nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Align(lipgloss.Center).
			Padding(2)
		return errorStyle.Render("No upcoming events found")
	}

	timeUntil := FormatTimeUntil(m.nextEvent.StartTime)
	timeColor := GetTimeColor(m.nextEvent.StartTime)

	// Title style - large text
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Align(lipgloss.Center).
		MarginTop(1).
		Width(60)

	// Time remaining style - extra large text, centered, colored
	timeStyle := lipgloss.NewStyle().
		Foreground(timeColor).
		Bold(true).
		Align(lipgloss.Center).
		Width(60)

	// Event details style
	detailsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Align(lipgloss.Center).
		Width(60)

	// Last updated style
	lastUpdatedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Align(lipgloss.Center).
		Width(60).
		MarginTop(1)

	// Container style
	containerStyle := lipgloss.NewStyle().
		Align(lipgloss.Center)

	// Render the display
	title := titleStyle.Render("📅 " + m.nextEvent.Title)
	timeRemaining := timeStyle.Render("⏰ " + timeUntil)

	startTime := detailsStyle.Render("Starts: " + m.nextEvent.StartTime.Add(time.Hour*time.Duration(-7)).Format("Monday, January 2, 2006 at 3:04 PM"))
	lastUpdated := lastUpdatedStyle.Render("Last updated: " + m.lastUpdate.Format("3:04:05 PM"))
	footer := detailsStyle.Render("Press 'q' or Ctrl+C to quit")

	content := lipgloss.JoinVertical(lipgloss.Center, title, timeRemaining, startTime, lastUpdated, footer)
	display := containerStyle.Render(content)

	return display
}
