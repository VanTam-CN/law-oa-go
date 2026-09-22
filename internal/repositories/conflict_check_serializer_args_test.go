package repositories

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func TestUpdateConflictCheckStatusUsesJSONSerializerArgument(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	result := &models.CheckResult{
		HasConflict:    true,
		TotalConflicts: 1,
		CompletedAt:    now,
		CoverageStatus: "COMPLETE",
	}
	expectedResult, err := json.Marshal(result)
	require.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `conflict_checks`").
		WithArgs(
			sqlmock.AnyArg(), // updated_at
			"REVIEW_REQUIRED",
			jsonString(string(expectedResult)),
			uint64(7),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repository := NewConflictCheckRepository(gormDB)
	require.NoError(t, repository.UpdateConflictCheckStatus(context.Background(), 7, "REVIEW_REQUIRED", result))
	require.NoError(t, mock.ExpectationsWereMet())
}

type jsonStringArgument struct {
	expected string
}

func jsonString(value string) sqlmock.Argument {
	return jsonStringArgument{expected: value}
}

func (arg jsonStringArgument) Match(value driver.Value) bool {
	text, ok := value.(string)
	return ok && text == arg.expected
}
