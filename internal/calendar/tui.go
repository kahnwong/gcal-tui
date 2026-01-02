package calendar

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kahnwong/gcal-tui/internal/utils"
)

type Model struct {
	Events    []CalendarEvent
	WeekStart time.Time // Monday of the current week
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
const ColWidth = 20

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

func InitialModel() Model {
	now := time.Now()
	// Find Monday of current week
	offset := int(now.Weekday()) - 1
	if offset < 0 {
		offset = 6 // Sunday
	}
	weekStart := now.AddDate(0, 0, -offset).Truncate(24 * time.Hour)
	events := FetchAllEvents(weekStart)
	return Model{Events: events, WeekStart: weekStart}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left":
			// Previous week
			m.WeekStart = m.WeekStart.AddDate(0, 0, -7)
			m.Events = FetchAllEvents(m.WeekStart)
		case "right":
			// Next week
			m.WeekStart = m.WeekStart.AddDate(0, 0, 7)
			m.Events = FetchAllEvents(m.WeekStart)
		}
	}
	return m, nil
}

func (m Model) View() string {
	// Time slots: 8am to 6pm, 30-minute intervals
	startHour, endHour := 8, 23
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

	// Header row
	header := TimeLabelStyle.Width(7).Render("") // time label column for alignment
	for d := 0; d < 7; d++ {
		dayDate := m.WeekStart.AddDate(0, 0, d)
		header += HeaderStyle.Width(ColWidth).Render(fmt.Sprintf("%s %02d/%02d", days[d], dayDate.Month(), dayDate.Day()))
		if d < 6 {
			header += SeparatorStyle.Width(1).Render("|")
		}
	}
	table := header + "\n"

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
			row := TimeLabelStyle.Width(7).Render(timeLabel)
			for d := 0; d < 7; d++ {
				cellTime := m.WeekStart.AddDate(0, 0, d).Add(time.Hour * time.Duration(hour)).Add(time.Minute * time.Duration(min))
				cell := EmptyStyle.Width(ColWidth).Render("")
				for _, e := range m.Events {
					// Check if cellTime is within event duration
					if cellTime.Equal(e.StartTime) || (cellTime.After(e.StartTime) && cellTime.Before(e.EndTime)) {
						eventStart := e.StartTime
						eventEnd := e.EndTime
						totalSlots := int(eventEnd.Sub(eventStart).Minutes()) / 30
						slotIndex := int(cellTime.Sub(eventStart).Minutes()) / 30
						maxTitleLen := ColWidth - 2
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
								cell = EventStyleActive.Background(GetColorValue(e.Color)).Width(ColWidth).Render(chunks[slotIndex])
							} else {
								cell = EventStyle.Background(GetColorValue(e.Color)).Width(ColWidth).Render(chunks[slotIndex])
							}
						} else if slotIndex < totalSlots {
							cell = EventStyle.Background(GetColorValue(e.Color)).Width(ColWidth).Render("")
						}
						break
					}
				}
				row += cell
				if d < 6 {
					row += SeparatorStyle.Width(1).Render("|")
				}
			}
			table += row + "\n"
		}
	}

	// Footer
	table += "\n←/→: Prev/Next week   q: Quit\n"
	return BorderStyle.Render(table)
}
