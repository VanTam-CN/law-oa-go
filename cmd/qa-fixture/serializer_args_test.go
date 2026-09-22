package main

import (
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

func TestRepeatSeedConflictCheckUpdateUsesJSONSerializerArgument(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	check := models.ConflictCheck{
		Status:         "REVIEW_REQUIRED",
		CheckedBy:      uintPtr(30),
		CheckedAt:      &now,
		Result:         &models.CheckResult{HasConflict: true, TotalConflicts: 1, CompletedAt: now, CoverageStatus: "COMPLETE"},
		ResultSummary:  "review required",
		TotalConflicts: 1,
		MediumCount:    1,
		CheckParams:    models.JSON{"source": "qa"},
	}
	expectedResult, err := json.Marshal(check.Result)
	require.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `conflict_checks`").
		WithArgs(
			sqlmock.AnyArg(), // updated_at
			"REVIEW_REQUIRED",
			uint64(30), // checked_by
			now,        // checked_at
			jsonString(string(expectedResult)),
			check.ResultSummary,
			int64(1),         // total_conflicts
			int64(1),         // medium_count
			sqlmock.AnyArg(), // check_params
			uint64(1),        // id
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, updateConflictCheckForRepeatSeed(gormDB, 1, check))
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

func uintPtr(value uint) *uint { return &value }
