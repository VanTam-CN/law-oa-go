package handlers

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// Shared fixture for the conflict queue review-state projection tests: two
// completed checks (one with a saved independent review, one without) plus
// the ethics-wall tables the visibility filter expects.
func newConflictQueueReviewStateDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE conflict_check_records (
			check_id TEXT PRIMARY KEY, user_id INTEGER, case_name TEXT, client_id TEXT,
			client_name TEXT, case_type TEXT, check_status TEXT, risk_level TEXT,
			has_conflict BOOLEAN, check_result TEXT, search_parameters TEXT,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE conflict_reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT, check_id TEXT, decision TEXT,
			notes TEXT, reviewer_id INTEGER, reviewer_name TEXT, evidence_hash TEXT,
			created_at DATETIME
		)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.AutoMigrate(&models.Client{}, &models.Case{}, &models.CaseEthicalWallWhitelist{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Client{ID: 1, Name: "新案客户", Type: "企业", Status: "active", Version: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Case{
		ID: 910, CaseNumber: "CASE-910", Title: "复核状态验证案件（虚构）", ClientID: 1,
		LawyerID: 34, CaseType: "commercial", Status: "active",
	}).Error; err != nil {
		t.Fatal(err)
	}
	frozenResult := `{"decision":{"status":"REVIEW_REQUIRED","coverageStatus":"COMPLETE","evidenceCount":0,"restrictedCount":0},"riskAssessment":{"overallRisk":"MEDIUM"}}`
	if err := db.Exec(`INSERT INTO conflict_check_records
		(check_id, user_id, case_name, client_id, client_name, case_type, check_status,
		 risk_level, has_conflict, check_result, search_parameters)
		VALUES ('CHECK-REVIEWED', 34, '复核完成案件（虚构）', '1', '新案客户', 'commercial',
		 'COMPLETED', 'MEDIUM', 0, ?, '{"subjectCaseId":"910"}'),
		('CHECK-PENDING', 34, '待复核案件（虚构）', '1', '新案客户', 'commercial',
		 'COMPLETED', 'MEDIUM', 0, ?, '{"subjectCaseId":"910"}')`, frozenResult, frozenResult).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO conflict_reviews
		(check_id, decision, notes, reviewer_id, reviewer_name, evidence_hash, created_at)
		VALUES ('CHECK-REVIEWED', 'no_conflict', '独立复核完成（虚构）', 35, '独立冲突核查人', 'hash-a', CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}
