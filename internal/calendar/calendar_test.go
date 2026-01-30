package calendar

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"google.golang.org/api/calendar/v3"
)

func TestGetColorValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected lipgloss.Color
	}{
		{"aqua color", "aqua", "#00FFFF"},
		{"teal color", "teal", "#008080"},
		{"green color", "green", "#00FF00"},
		{"red color", "red", "#FF0000"},
		{"unknown color defaults to orange", "unknown", "#FFA500"},
		{"empty string defaults to orange", "", "#FFA500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetColorValue(tt.input)
			if result != tt.expected {
				t.Errorf("GetColorValue(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRoundToNearestHalfHour(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "round down to :00 from :05",
			input:    time.Date(2026, 1, 31, 10, 5, 0, 0, time.UTC),
			expected: time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC),
		},
		{
			name:     "round down to :00 from :14",
			input:    time.Date(2026, 1, 31, 10, 14, 59, 0, time.UTC),
			expected: time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC),
		},
		{
			name:     "round to :30 from :15",
			input:    time.Date(2026, 1, 31, 10, 15, 0, 0, time.UTC),
			expected: time.Date(2026, 1, 31, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "round to :30 from :25",
			input:    time.Date(2026, 1, 31, 10, 25, 0, 0, time.UTC),
			expected: time.Date(2026, 1, 31, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "round to :30 from :44",
			input:    time.Date(2026, 1, 31, 10, 44, 59, 0, time.UTC),
			expected: time.Date(2026, 1, 31, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "round up to next hour from :45",
			input:    time.Date(2026, 1, 31, 10, 45, 0, 0, time.UTC),
			expected: time.Date(2026, 1, 31, 11, 0, 0, 0, time.UTC),
		},
		{
			name:     "round up to next hour from :55",
			input:    time.Date(2026, 1, 31, 10, 55, 0, 0, time.UTC),
			expected: time.Date(2026, 1, 31, 11, 0, 0, 0, time.UTC),
		},
		{
			name:     "already at :00",
			input:    time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC),
			expected: time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC),
		},
		{
			name:     "already at :30",
			input:    time.Date(2026, 1, 31, 10, 30, 0, 0, time.UTC),
			expected: time.Date(2026, 1, 31, 10, 30, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roundToNearestHalfHour(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("roundToNearestHalfHour(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseCalendars(t *testing.T) {
	tests := []struct {
		name        string
		color       string
		events      *calendar.Events
		expectError bool
		expectCount int
	}{
		{
			name:  "parse single timed event",
			color: "blue",
			events: &calendar.Events{
				Items: []*calendar.Event{
					{
						Summary: "Test Meeting",
						Start: &calendar.EventDateTime{
							DateTime: "2026-01-31T10:00:00Z",
						},
						End: &calendar.EventDateTime{
							DateTime: "2026-01-31T11:00:00Z",
						},
					},
				},
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name:  "parse multiple events",
			color: "green",
			events: &calendar.Events{
				Items: []*calendar.Event{
					{
						Summary: "Meeting 1",
						Start: &calendar.EventDateTime{
							DateTime: "2026-01-31T10:00:00Z",
						},
						End: &calendar.EventDateTime{
							DateTime: "2026-01-31T11:00:00Z",
						},
					},
					{
						Summary: "Meeting 2",
						Start: &calendar.EventDateTime{
							DateTime: "2026-01-31T14:00:00Z",
						},
						End: &calendar.EventDateTime{
							DateTime: "2026-01-31T15:00:00Z",
						},
					},
				},
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:  "empty events list",
			color: "red",
			events: &calendar.Events{
				Items: []*calendar.Event{},
			},
			expectError: false,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCalendars(tt.color, tt.events)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(result) != tt.expectCount {
				t.Errorf("Expected %d events, got %d", tt.expectCount, len(result))
			}

			// Verify color is set correctly
			for _, event := range result {
				if event.Color != tt.color {
					t.Errorf("Expected color %q, got %q", tt.color, event.Color)
				}
			}
		})
	}
}
