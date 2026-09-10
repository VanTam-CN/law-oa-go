package repositories

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// setupIntakeCommitDB 构造 CommitIntakeConflictResult 所需的最小表集合：
// 冲突检测记录、冲突案例与接案记录。文件模式与 setupConflictSQLiteDB 一致，
// 避开 shared cache 下的索引名冲突。
func setupIntakeCommitDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("intake_commit_%d_%d.db",
		time.Now().UnixNano(), rand.Int63()))
	dsn := fmt.Sprintf("file:%s?_busy_timeout=30000&_journal_mode=WAL", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.ConflictCheckRecord{}, &models.ConflictCase{}, &models.Client{}))
	require.NoError(t, db.Exec("CREATE TABLE case_intakes (" +
		"id TEXT PRIMARY KEY, client_id INTEGER, status TEXT, metadata TEXT, updated_at DATETIME)").Error)
	return db
}

func seedCommitIntake(t *testing.T, db *gorm.DB, id string, clientID uint, status, metadata string) {
	t.Helper()
	require.NoError(t, db.Exec(
		"INSERT INTO case_intakes (id, client_id, status, metadata, updated_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)",
		id, clientID, status, metadata).Error)
}

func buildCommitRecord(checkID string, clientID uint) *models.ConflictCheckRecord {
	return &models.ConflictCheckRecord{
		CheckID:     checkID,
		ClientID:    fmt.Sprint(clientID),
		ClientName:  "提交客户",
		CaseName:    "提交测试案件",
		CaseType:    "民事",
		CheckStatus: "COMPLETED",
		HasConflict: false,
		RiskLevel:   "LOW",
		CheckTime:   time.Now(),
	}
}

func commitAssociation(checkID, intakeID string, clientID uint) ConflictSubjectAssociation {
	return ConflictSubjectAssociation{
		CheckID:        checkID,
		IntakeID:       intakeID,
		ClientID:       fmt.Sprint(clientID),
		CoverageStatus: "COMPLETE",
		CheckedAt:      time.Now(),
	}
}

func commitIntakeRow(t *testing.T, db *gorm.DB, id string) (string, string) {
	t.Helper()
	var row struct {
		Status   string
		Metadata string
	}
	require.NoError(t, db.Table("case_intakes").
		Select("status, metadata").Where("id = ?", id).Take(&row).Error)
	return row.Status, row.Metadata
}

func countCommitRecords(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&models.ConflictCheckRecord{}).Count(&count).Error)
	return count
}

// 首次正式检测：审计记录、接案状态与 metadata 关联必须同事务落库。
func TestCommitIntakeConflictResultPersistsRecordAndIntakeLink(t *testing.T) {
	db := setupIntakeCommitDB(t)
	client := models.Client{Name: "提交客户", Type: "企业"}
	require.NoError(t, db.Create(&client).Error)
	seedCommitIntake(t, db, "intake-commit-001", client.ID, "lawyer_facts_confirmed",
		"{\"lawyer_facts_confirmed_at\":\"2026-09-10T01:00:00Z\"}")

	record := buildCommitRecord("CCT-COMMIT-001", client.ID)
	response := &models.ConflictCheckResponse{CheckID: "CCT-COMMIT-001", HasConflict: false}
	require.NoError(t, CommitIntakeConflictResult(context.Background(), db, record, nil, nil, response,
		commitAssociation("CCT-COMMIT-001", "intake-commit-001", client.ID), "2026-09-10T01:00:00Z"))

	status, metadata := commitIntakeRow(t, db, "intake-commit-001")
	assert.Equal(t, "conflict_ready", status)
	assert.Contains(t, metadata, "CCT-COMMIT-001")
	var stored models.ConflictCheckRecord
	require.NoError(t, db.Where("check_id = ?", "CCT-COMMIT-001").Take(&stored).Error)
	assert.Equal(t, "COMPLETED", stored.CheckStatus)
}

// 重检语义：同一接案可再次提交，metadata 指向最新检测；
// 旧审计记录保留（append-only），冲突案例允许整行替换。
func TestCommitIntakeConflictResultRecheckReplacesAssociation(t *testing.T) {
	db := setupIntakeCommitDB(t)
	client := models.Client{Name: "重检客户", Type: "企业"}
	require.NoError(t, db.Create(&client).Error)
	seedCommitIntake(t, db, "intake-commit-002", client.ID, "lawyer_facts_confirmed",
		"{\"lawyer_facts_confirmed_at\":\"2026-09-10T01:00:00Z\"}")

	firstResponse := &models.ConflictCheckResponse{CheckID: "CCT-COMMIT-002A", HasConflict: true}
	firstCases := []*models.ConflictCase{{
		ID:        "CASE-RECHECK-1",
		CheckID:   "CCT-COMMIT-002A",
		CaseName:  "对手方案件",
		RiskLevel: "HIGH",
	}}
	require.NoError(t, CommitIntakeConflictResult(context.Background(), db,
		buildCommitRecord("CCT-COMMIT-002A", client.ID), firstCases, nil, firstResponse,
		commitAssociation("CCT-COMMIT-002A", "intake-commit-002", client.ID), "2026-09-10T01:00:00Z"))

	secondResponse := &models.ConflictCheckResponse{CheckID: "CCT-COMMIT-002B", HasConflict: true}
	secondCases := []*models.ConflictCase{{
		ID:        "CASE-RECHECK-1",
		CheckID:   "CCT-COMMIT-002B",
		CaseName:  "对手方案件",
		RiskLevel: "LOW",
	}}
	require.NoError(t, CommitIntakeConflictResult(context.Background(), db,
		buildCommitRecord("CCT-COMMIT-002B", client.ID), secondCases, nil, secondResponse,
		commitAssociation("CCT-COMMIT-002B", "intake-commit-002", client.ID), "2026-09-10T01:00:00Z"))

	status, metadata := commitIntakeRow(t, db, "intake-commit-002")
	assert.Equal(t, "conflict_ready", status)
	assert.Contains(t, metadata, "CCT-COMMIT-002B")
	assert.NotContains(t, metadata, "CCT-COMMIT-002A")
	assert.EqualValues(t, 2, countCommitRecords(t, db), "append-only audit history must keep both records")
	var replaced models.ConflictCase
	require.NoError(t, db.Where("id = ?", "CASE-RECHECK-1").Take(&replaced).Error)
	assert.Equal(t, "LOW", replaced.RiskLevel)
}

// 事实时间戳不匹配：接案资料在检测期间被修改时必须整体拒绝，
// 不留下半提交的审计记录或错误状态。
func TestCommitIntakeConflictResultRejectsStaleFactsConfirmation(t *testing.T) {
	db := setupIntakeCommitDB(t)
	client := models.Client{Name: "过期事实客户", Type: "企业"}
	require.NoError(t, db.Create(&client).Error)
	seedCommitIntake(t, db, "intake-commit-003", client.ID, "lawyer_facts_confirmed",
		"{\"lawyer_facts_confirmed_at\":\"2026-09-10T01:00:00Z\"}")

	record := buildCommitRecord("CCT-COMMIT-003", client.ID)
	response := &models.ConflictCheckResponse{CheckID: "CCT-COMMIT-003", HasConflict: false}
	err := CommitIntakeConflictResult(context.Background(), db, record, nil, nil, response,
		commitAssociation("CCT-COMMIT-003", "intake-commit-003", client.ID), "2026-09-10T02:00:00Z")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "事实")

	status, metadata := commitIntakeRow(t, db, "intake-commit-003")
	assert.Equal(t, "lawyer_facts_confirmed", status, "status must stay unchanged on rejection")
	assert.NotContains(t, metadata, "CCT-COMMIT-003")
	assert.EqualValues(t, 0, countCommitRecords(t, db), "no orphan audit record may survive a failed commit")
}

// 客户不匹配：关联客户与接案客户不一致是硬错误。
func TestCommitIntakeConflictResultRejectsClientMismatch(t *testing.T) {
	db := setupIntakeCommitDB(t)
	client := models.Client{Name: "真实客户", Type: "企业"}
	require.NoError(t, db.Create(&client).Error)
	seedCommitIntake(t, db, "intake-commit-004", client.ID, "lawyer_facts_confirmed",
		"{\"lawyer_facts_confirmed_at\":\"2026-09-10T01:00:00Z\"}")

	wrongClientID := client.ID + 999
	record := buildCommitRecord("CCT-COMMIT-004", wrongClientID)
	response := &models.ConflictCheckResponse{CheckID: "CCT-COMMIT-004", HasConflict: false}
	err := CommitIntakeConflictResult(context.Background(), db, record, nil, nil, response,
		commitAssociation("CCT-COMMIT-004", "intake-commit-004", wrongClientID), "2026-09-10T01:00:00Z")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "客户")

	status, _ := commitIntakeRow(t, db, "intake-commit-004")
	assert.Equal(t, "lawyer_facts_confirmed", status)
	assert.EqualValues(t, 0, countCommitRecords(t, db))
}
