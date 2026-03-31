package filter

import (
	"testing"
	"time"
)

func TestIsSafe(t *testing.T) {
	now := time.Now().UTC()
	tenDaysAgo := now.AddDate(0, 0, -10).Format(time.RFC3339)
	twoDaysAgo := now.AddDate(0, 0, -2).Format(time.RFC3339)

	if !IsSafe(tenDaysAgo) {
		t.Errorf("Expected 10 days ago to be safe")
	}
	if IsSafe(twoDaysAgo) {
		t.Errorf("Expected 2 days ago to be unsafe")
	}
}
