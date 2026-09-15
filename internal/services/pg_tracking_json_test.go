package services

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/repositories"
)

// PostgreSQL validates jsonb inputs strictly: a Go zero value ("") in any
// tracking JSON column fails the insert with SQLSTATE 22P02. The dry-run
// capture asserts the create path binds valid JSON defaults for every JSON
// column. This is SQL-generation evidence; the real PG failure reproduction
// remains Astra's independent log.
type trackingDryRunRepo struct {
	*repositories.IntegrationRepository
}

// GetLatestCaseCreationTracking reports "no tracking yet" without touching the
// database so the dry run exercises the create branch of updateCaseTracking.
func (r *trackingDryRunRepo) GetLatestCaseCreationTracking(context.Context, string) (*repositories.CaseCreationTracking, error) {
	return nil, nil
}

func TestUpdateCaseTrackingCreateBindsNoEmptyJSON(t *testing.T) {
	db, err := gorm.Open(postgres.Open("host=127.0.0.1 user=qa dbname=qa sslmode=disable"), &gorm.Config{
		DryRun:                 true,
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	var insertSQL string
	var insertVars []interface{}
	require.NoError(t, db.Callback().Create().After("gorm:create").Register("pgjson:capture", func(tx *gorm.DB) {
		insertSQL = tx.Statement.SQL.String()
		insertVars = append([]interface{}(nil), tx.Statement.Vars...)
	}))

	svc := &approvalConflictIntegrationService{integrationRepo: &trackingDryRunRepo{IntegrationRepository: repositories.NewIntegrationRepository(db)}}
	caseID := "42"
	caseNumber := "CASE-JSON-1"
	require.NoError(t, svc.updateCaseTracking(context.Background(), "APR-JSON-1", "completed", &caseID, &caseNumber, ""))

	require.True(t, strings.HasPrefix(insertSQL, `INSERT INTO "approval_case_creation_tracking"`),
		"expected a tracking insert, got: %s", insertSQL)
	for _, v := range insertVars {
		if s, ok := v.(string); ok && s == "" {
			t.Fatalf("tracking insert binds an empty string into a json column: %s", insertSQL)
		}
	}
	require.Contains(t, insertVars, "{}", "json columns must default to valid JSON")
}

// The sqlite round-trip pins the actual stored value: a tracking row created
// because the approval had none must carry valid JSON defaults, which is what
// the real PostgreSQL insert binds.
func TestUpdateCaseTrackingCreateDefaultsValidJSON(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tracking_json.db")
	db, err := gorm.Open(sqlite.Open("file:"+dbPath), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&repositories.CaseCreationTracking{}))

	svc := &approvalConflictIntegrationService{integrationRepo: repositories.NewIntegrationRepository(db)}
	caseID := "7"
	caseNumber := "CASE-JSON-2"
	require.NoError(t, svc.updateCaseTracking(context.Background(), "APR-JSON-2", "completed", &caseID, &caseNumber, ""))

	var row repositories.CaseCreationTracking
	require.NoError(t, db.Where("approval_request_id = ?", "APR-JSON-2").First(&row).Error)
	require.Equal(t, "completed", row.CreationStatus)
	require.NotNil(t, row.CaseID)
	require.Equal(t, "7", *row.CaseID)
	require.Equal(t, "{}", row.DataMapping)
	require.Equal(t, "{}", row.ErrorDetails)
	require.Equal(t, "{}", row.MappedFields)
	require.Equal(t, "{}", row.UnmappedFields)
	require.Equal(t, "{}", row.AppliedConditions)
	require.Equal(t, "{}", row.ImposedRequirements)
	require.Equal(t, "[]", row.WorkflowActions)
	require.True(t, json.Valid([]byte(row.DataMapping)))
}
