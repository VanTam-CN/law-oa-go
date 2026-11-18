package services

import (
	"errors"
	"testing"
	"time"
)

// Independent regression: equal whole seconds do not prove equal task
// versions. Two adjustments can happen inside one second.
func TestIndependentTaskTimestampRejectsSameSecondStaleValue(t *testing.T) {
	current := time.Date(2026, 10, 2, 8, 30, 0, 800000000, time.UTC)
	stale := current.Add(-600 * time.Millisecond).Format(time.RFC3339Nano)
	err := versionGuard(1, current, &CaseTaskChangeRequest{ExpectedUpdatedAt: &stale})
	if !errors.Is(err, ErrCaseTaskChangeStale) {
		t.Fatalf("same-second stale timestamp must be rejected, got %v", err)
	}
	exact := current.Format(time.RFC3339Nano)
	if err := versionGuard(1, current, &CaseTaskChangeRequest{ExpectedUpdatedAt: &exact}); err != nil {
		t.Fatalf("exact current timestamp should pass: %v", err)
	}
}
