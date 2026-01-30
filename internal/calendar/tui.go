package calendar

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kahnwong/gcal-tui/internal/utils"
)

type Model struct {
	Events      []CalendarEvent
	StartDate   time.Time // Starting date (Monday for week view, specific date for today view)
	ColumnCount int       // Number of columns (1 for today, 7 for week)
	ColWidth    int       // Width of each column
}

func GetColorValue(name string) lipgloss.Color {
	switch name {
	case "aqua":
		return "#00FFFF"
	case "teal":
		return "#008080"
	case "green":
		return "#00FF00"
	case "red":
		return "#FF0000"
	default:
		return "#FFA500" // fallback color
	}
}

// Styles for rendering
var (
	EventStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Bold(true)
	EventStyleActive = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	EmptyStyle       = lipgloss.NewStyle().Background(lipgloss.Color("#222")).Foreground(lipgloss.Color("#888"))
	HeaderStyle      = lipgloss.NewStyle().
				Background(lipgloss.Color("#FFF")).
				Foreground(lipgloss.Color("#000")).
				Align(lipgloss.Center).
				Bold(true)
	TimeLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0FF"))
	BorderStyle    = lipgloss.NewStyle().Border(lipgloss.HiddenBorder()).Padding(0, 1)
	SeparatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555"))
)

// NewModel creates a new calendar model with specified column count and width
func NewModel(columnCount int, colWidth int) Model {
	now := time.Now()
	var startDate time.Time

	if columnCount == 7 {
		// Week view: find Monday of current week
		offset := int(now.Weekday()) - 1
		if offset < 0 {
			offset = 6 // Sunday
		}
		startDate = now.AddDate(0, 0, -offset).Truncate(24 * time.Hour)
	} else {
		// Today view: use current date
		startDate = now.Truncate(24 * time.Hour)
	}

	events, err := FetchAllEvents(startDate)
	if err != nil {
		// Return model with empty events and store error for display
		return Model{
			Events:      []CalendarEvent{},
			StartDate:   startDate,
			ColumnCount: columnCount,
			ColWidth:    colWidth,
		}
	}
	return Model{
		Events:      events,
		StartDate:   startDate,
		ColumnCount: columnCount,
		ColWidth:    colWidth,
	}
}

// InitialModel creates a week view model (7 columns, 20 width) - for backward compatibility
func InitialModel() Model {
	return NewModel(7, 20)
}

// InitialTodayModel creates a today view model (1 column, 20 width) - for backward compatibility
func InitialTodayModel() Model {
	return NewModel(1, 20)
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left":
			if m.ColumnCount == 7 {
				// Previous week
				m.StartDate = m.StartDate.AddDate(0, 0, -7)
			} else {
				// Previous day
				m.StartDate = m.StartDate.AddDate(0, 0, -1)
			}
			events, err := FetchAllEvents(m.StartDate)
			if err == nil {
				m.Events = events
			}
		case "right":
			if m.ColumnCount == 7 {
				// Next week
				m.StartDate = m.StartDate.AddDate(0, 0, 7)
			} else {
				// Next day
				m.StartDate = m.StartDate.AddDate(0, 0, 1)
			}
			events, err := FetchAllEvents(m.StartDate)
			if err == nil {
				m.Events = events
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	// Time slots: 8am to 11pm, 30-minute intervals
	startHour, endHour := 8, 23
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

	var headerParts []string
	headerParts = append(headerParts, TimeLabelStyle.Width(7).Render("")) // time label column for alignment

	for d := range m.ColumnCount {
		dayDate := m.StartDate.AddDate(0, 0, d)
		var dayLabel string
		if m.ColumnCount == 1 {
			dayLabel = dayDate.Format("Monday 01/02")
		} else {
			dayLabel = fmt.Sprintf("%s %02d/%02d", days[d], dayDate.Month(), dayDate.Day())
		}
		headerParts = append(headerParts, HeaderStyle.Width(m.ColWidth).Render(dayLabel))
		if d < m.ColumnCount-1 {
			headerParts = append(headerParts, SeparatorStyle.Width(1).Render("|"))
		}
	}

	var tableRows []string
	tableRows = append(tableRows, strings.Join(headerParts, ""))

	// Time rows (30-minute intervals)
	now := utils.GetNowLocalAdjusted()
	for hour := startHour; hour < endHour; hour++ {
		for min := 0; min < 60; min += 30 {
			var timeLabel string
			if min == 0 {
				timeLabel = fmt.Sprintf("%02d:00", hour)
			} else {
				timeLabel = ""
			}

			var rowParts []string
			rowParts = append(rowParts, TimeLabelStyle.Width(7).Render(timeLabel))

			for d := range m.ColumnCount {
				cellTime := m.StartDate.AddDate(0, 0, d).Add(time.Hour * time.Duration(hour)).Add(time.Minute * time.Duration(min))
				cell := EmptyStyle.Width(m.ColWidth).Render("")

				for _, e := range m.Events {
					// Check if cellTime is within event duration
					if cellTime.Equal(e.StartTime) || (cellTime.After(e.StartTime) && cellTime.Before(e.EndTime)) {
						eventStart := e.StartTime
						eventEnd := e.EndTime
						totalSlots := int(eventEnd.Sub(eventStart).Minutes()) / 30
						slotIndex := int(cellTime.Sub(eventStart).Minutes()) / 30
						maxTitleLen := m.ColWidth - 2
						title := e.Title

						// Split title into chunks
						var chunks []string
						for i := 0; i < len(title); i += maxTitleLen {
							end := i + maxTitleLen
							if end > len(title) {
								end = len(title)
							}
							chunks = append(chunks, title[i:end])
						}

						// Render chunk if within event duration and title length
						if slotIndex < len(chunks) && slotIndex < totalSlots {
							if now.After(e.StartTime) && now.Before(e.EndTime) {
								cell = EventStyleActive.Background(GetColorValue(e.Color)).Width(m.ColWidth).Render(chunks[slotIndex])
							} else {
								cell = EventStyle.Background(GetColorValue(e.Color)).Width(m.ColWidth).Render(chunks[slotIndex])
							}
						} else if slotIndex < totalSlots {
							cell = EventStyle.Background(GetColorValue(e.Color)).Width(m.ColWidth).Render("")
						}
						break
					}
				}

				rowParts = append(rowParts, cell)
				if d < m.ColumnCount-1 {
					rowParts = append(rowParts, SeparatorStyle.Width(1).Render("|"))
				}
			}
			tableRows = append(tableRows, strings.Join(rowParts, ""))
		}
	}

	// Footer
	var footerText string
	if m.ColumnCount == 1 {
		footerText = "\n←/→: Prev/Next day   q: Quit\n"
	} else {
		footerText = "\n←/→: Prev/Next week   q: Quit\n"
	}

	return BorderStyle.Render(strings.Join(tableRows, "\n") + footerText)
}
