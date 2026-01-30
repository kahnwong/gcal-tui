package utils

import (
	"testing"
	"time"
)

func TestGetNowLocalAdjusted(t *testing.T) {
	// Get the adjusted time
	adjusted := GetNowLocalAdjusted()

	// Get the actual current time
	now := time.Now()

	// The adjusted time should be approximately 7 hours ahead
	expectedDiff := time.Hour * 7
	actualDiff := adjusted.Sub(now)

	// Allow for a small margin of error (1 second) due to execution time
	margin := time.Second

	if actualDiff < expectedDiff-margin || actualDiff > expectedDiff+margin {
		t.Errorf("Expected time difference of %v, got %v", expectedDiff, actualDiff)
	}
}

func TestGetNowLocalAdjustedReturnsTime(t *testing.T) {
	// Simply verify the function returns a valid time.Time object
	result := GetNowLocalAdjusted()

	if result.IsZero() {
		t.Error("Expected non-zero time, got zero time")
	}
}
