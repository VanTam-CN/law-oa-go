package services

// Regression for defect 6 (2026-09-26 audit): generateUUID returned a
// nanosecond hex string, which PostgreSQL uuid columns reject with 22P02,
// making every approval-creation path fail with HTTP 500.

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGenerateUUIDIsRFC4122(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		id := generateUUID()
		parsed, err := uuid.Parse(id)
		require.NoError(t, err, "generateUUID must return a parseable UUID, got %q", id)
		require.Equal(t, uuid.Version(4), parsed.Version(), "expected RFC 4122 version 4 random UUID")
		require.False(t, seen[id], "generateUUID must not repeat under rapid successive calls")
		seen[id] = true
	}
}
