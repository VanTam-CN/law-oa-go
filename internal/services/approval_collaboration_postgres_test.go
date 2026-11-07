package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
)

// This adapter only forwards to the real service, matching the production
// handler adapter. It replaces no authorization, snapshot, binding or workflow
// logic; services cannot import handlers without creating an import cycle.
type approvalCollaborationPostgresAdapter struct{ *ApprovalService }

func (a *approvalCollaborationPostgresAdapter) ProcessApproval(userID, _ string, id string, req *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	return a.ProcessApprovalDecision(userID, id, req)
}

func (a *approvalCollaborationPostgresAdapter) BindApprovalTransaction(tx *gorm.DB) ApprovalServiceInterface {
	return &approvalCollaborationPostgresAdapter{NewApprovalService(tx)}
}

type approvalCollaborationPostgresFixture struct {
	db           *gorm.DB
	approval     *models.ApprovalRequest
	intakeID     string
	checkID      string
	clientID     uint
	clientDigest string
	partyDigest  string
}

func approvalCollaborationPostgresDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("LAW_OA_COLLAB_PG_DSN")
	if dsn == "" {
		t.Skip("NOT_RUN: LAW_OA_COLLAB_PG_DSN unset; PostgreSQL coverage must not be reported PASS")
	}
	parsed, err := pgconn.ParseConfig(dsn)
	require.NoError(t, err)
	require.Equal(t, "/private/tmp/lawoa-intake-collaboration-runtime/socket", parsed.Host)
	require.EqualValues(t, 55437, parsed.Port)
	require.Equal(t, "lawoa_intake_mvp_test", parsed.Database)
	require.Equal(t, "mac", parsed.User)
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("p", 32))
	config := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	admin, err := gorm.Open(postgres.Open(dsn), config)
	require.NoError(t, err)
	schema := "approval_collab_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() {
		require.NoError(t, admin.Exec("DROP SCHEMA "+schema+" CASCADE").Error)
		conn, _ := admin.DB()
		_ = conn.Close()
	})
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema+" application_name=approval_collaboration_pg_test statement_timeout=15000 lock_timeout=8000"), config)
	require.NoError(t, err)
	conn, err := db.DB()
	require.NoError(t, err)
	conn.SetMaxOpenConns(12)
	t.Cleanup(func() { _ = conn.Close() })
	var serverVersion, currentSchema string
	require.NoError(t, db.Raw("SELECT version(), current_schema()").Row().Scan(&serverVersion, &currentSchema))
	require.Equal(t, schema, currentSchema)
	t.Logf("PostgreSQL fixture schema=%s server=%s", schema, serverVersion)
	require.NoError(t, db.AutoMigrate(
		&models.User{}, &models.Client{}, &models.Case{}, &models.CaseEthicalWallWhitelist{},
		&models.Entity{}, &models.CaseParty{}, &models.ConflictCheckRecord{}, &models.ConflictReview{},
		&models.ConflictReviewerAssignment{}, &models.ApprovalWorkflow{}, &models.ApprovalRequest{}, &models.ApprovalRecord{}, &models.ApprovalNotification{},
		&models.ApprovalSnapshot{}, &repositories.CaseCreationTracking{}, &repositories.ConflictAssociation{}, &models.InboxItem{},
	))
	// Minimal intake schema uses the production column types read by these
	// workflows; all protocol constraints below come from the actual migrations.
	require.NoError(t, db.Exec(`CREATE TABLE case_intakes (
		id VARCHAR(36) PRIMARY KEY, intake_code VARCHAR(60), client_id BIGINT NOT NULL,
		title TEXT NOT NULL, case_type VARCHAR(50) NOT NULL, status VARCHAR(40) NOT NULL,
		priority VARCHAR(20), description TEXT, metadata JSONB NOT NULL,
		created_by VARCHAR(36) NOT NULL, approval_claim_token TEXT,
		created_at TIMESTAMPTZ DEFAULT now(), updated_at TIMESTAMPTZ DEFAULT now()
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE case_intake_parties (
		id BIGSERIAL PRIMARY KEY, intake_id VARCHAR(36) NOT NULL, entity_name TEXT NOT NULL,
		entity_type VARCHAR(30), party_role VARCHAR(40), identity_type VARCHAR(30),
		identity_number_digest VARCHAR(64), identity_number_ciphertext TEXT, metadata JSONB NOT NULL
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE case_materials (id BIGSERIAL PRIMARY KEY, intake_id VARCHAR(36), metadata JSONB NOT NULL DEFAULT '{}')`).Error)
	// AutoMigrate installed the same model index; replace only this fresh
	// schema's index with the migration-owned definition before running 000084.
	require.NoError(t, db.Migrator().DropIndex(&models.InboxItem{}, "uq_inbox_workflow_task_key"))
	for _, name := range []string{"000081_intake_approval_unique_binding.up.sql", "000082_intake_submission_snapshots.up.sql", "000084_intake_workflow_inbox.up.sql"} {
		raw, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		require.NoError(t, err)
		require.NoError(t, db.Exec(string(raw)).Error, name)
	}
	return db
}

func (f *approvalCollaborationPostgresFixture) integrated() *approvalConflictIntegrationService {
	caseService := NewCaseService(repositories.NewCaseRepository(f.db), repositories.NewClientRepository(f.db), repositories.NewUserRepository(f.db))
	caseService.SetSubjectRecheckService(NewSubjectRecheckService(f.db, nil))
	return &approvalConflictIntegrationService{
		approvalService: &approvalCollaborationPostgresAdapter{NewApprovalService(f.db)},
		caseService:     caseService, integrationRepo: repositories.NewIntegrationRepository(f.db), db: f.db,
	}
}

func approvalCollaborationPostgresJSON(t *testing.T, value interface{}) string {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	return string(raw)
}

func newApprovalCollaborationPostgresIntakeFixture(t *testing.T) *approvalCollaborationPostgresFixture {
	t.Helper()
	return seedApprovalCollaborationPostgresIntake(t, approvalCollaborationPostgresDB(t))
}

// Shared data only: bootstrap tests supply a database initialized by the real
// production installer, without executing this file's protocol-specific DDL.
func seedApprovalCollaborationPostgresIntake(t *testing.T, db *gorm.DB) *approvalCollaborationPostgresFixture {
	t.Helper()
	f := &approvalCollaborationPostgresFixture{db: db, intakeID: uuid.NewString(), checkID: "COLLAB-PG-" + uuid.NewString()}
	for _, user := range []models.User{
		{ID: 11, Username: "pg-applicant", Name: "虚构申请律师", Email: "applicant@example.invalid", Password: "fixture-only", Role: "lawyer", Status: "active"},
		{ID: 22, Username: "pg-approver", Name: "虚构审批合伙人", Email: "approver@example.invalid", Password: "fixture-only", Role: "partner", Status: "active"},
		{ID: 33, Username: "pg-reviewer", Name: "虚构独立复核人", Email: "reviewer@example.invalid", Password: "fixture-only", Role: "conflict_officer", Status: "active"},
	} {
		require.NoError(t, db.Create(&user).Error)
	}
	client := &models.Client{Name: "虚构协作客户", Type: "个人", CreatedBy: 11, IdentityType: models.IdentityTypeIDCard, IdentityNumber: "310110199001010011"}
	require.NoError(t, repositories.NewClientRepository(db).Create(context.Background(), client))
	f.clientID, f.clientDigest = client.ID, client.IdentityNumberDigest
	var err error
	f.partyDigest, err = security.IdentityDigest("310110199001010022")
	require.NoError(t, err)
	now := time.Now().UTC().Truncate(time.Microsecond)
	metadata := map[string]interface{}{
		"draft_revision": "draft-r1", "lawyer_id": "11", "lawyer_facts_version": "facts-v1",
		"lawyer_facts_confirmed_by": "11", "lawyer_facts_confirmed_at": now.Add(-time.Minute).Format(time.RFC3339),
		"conflict_check_id": f.checkID, "billing_method": "hourly",
	}
	require.NoError(t, db.Exec(`INSERT INTO case_intakes(id,intake_code,client_id,title,case_type,status,priority,description,metadata,created_by)
		VALUES(?,?,?,'虚构协作案件','CIVIL','conflict_ready','medium','事务集成测试',?::jsonb,'11')`, f.intakeID, "PG-COLLAB", client.ID, approvalCollaborationPostgresJSON(t, metadata)).Error)
	require.NoError(t, db.Exec(`INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,identity_type,identity_number_digest,metadata)
		VALUES(?,'虚构对方','INDIVIDUAL','OPPOSING_PARTY','ID_CARD',?,'{"localId":"opposing-1"}'::jsonb)`, f.intakeID, f.partyDigest).Error)
	check := &models.ConflictCheckRecord{
		CheckID: f.checkID, ClientID: strconv.FormatUint(uint64(client.ID), 10), ClientName: client.Name,
		CaseName: "虚构协作案件", CaseType: "CIVIL", CheckStatus: "COMPLETED", UserID: 11, RiskLevel: "LOW", CheckTime: now,
		SearchParameters: models.JSON{"intakeId": f.intakeID, "clientType": "PERSON", "clientIdentifiers": map[string]interface{}{"id_card": "hmac-sha256:" + f.clientDigest},
			"parties": []interface{}{map[string]interface{}{"name": "虚构对方", "role": "OPPOSING_PARTY", "entityType": "INDIVIDUAL", "identifiers": map[string]interface{}{"id_card": "hmac-sha256:" + f.partyDigest}}}},
		CheckResult: models.JSON{"checkId": f.checkID, "hasConflict": false, "decision": map[string]interface{}{"coverageStatus": "COMPLETE"}},
	}
	require.NoError(t, db.Create(check).Error)
	var response models.ConflictCheckResponse
	require.NoError(t, json.Unmarshal([]byte(approvalCollaborationPostgresJSON(t, check.CheckResult)), &response))
	require.NoError(t, db.Create(&models.ConflictReviewerAssignment{ID: uuid.NewString(), CheckID: f.checkID, ReviewerID: 33, AssignedBy: 22, Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true, IndependenceReason: "虚构独立复核声明", CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.ConflictReview{CheckID: f.checkID, Decision: "no_conflict", Notes: "虚构完整覆盖复核", ReviewerID: 33, ReviewerName: "虚构独立复核人", EvidenceHash: conflictEvidenceHash(&response), CreatedAt: now.Add(time.Second)}).Error)
	workflow := &models.ApprovalWorkflow{
		WorkflowCode: "COLLAB_PG", WorkflowName: "集成测试审批", WorkflowType: "COLLAB_PG", Status: "active", CreatedBy: "22", EffectiveDate: now,
		ApplicableTypes: "[]", ApplicableDepartments: "[]", ApplicableRoles: "[]", Conditions: "{}", Timeouts: "{}", Permissions: "{}", Notifications: "{}",
		Stages: `[{"stage_name":"partner_review","stage_order":1,"approver_id":"22","required":true}]`,
	}
	require.NoError(t, db.Create(workflow).Error)
	return f
}

func newApprovalCollaborationPostgresFixture(t *testing.T) *approvalCollaborationPostgresFixture {
	t.Helper()
	f := newApprovalCollaborationPostgresIntakeFixture(t)
	result, err := f.integrated().CreateIntegratedApproval(context.Background(), "11", "虚构申请律师", f.newRequest())
	require.NoError(t, err, "real initial integrated submission must succeed before any fault is installed")
	f.approval = f.reload(t, result.ApprovalID)
	require.NotNil(t, f.approval.ActiveSubmissionSnapshotID)
	require.Equal(t, models.ApprovalStatusSubmitted, f.approval.Status)
	require.Equal(t, int64(1), f.count(t, "approval_case_creation_tracking", "intake_id = ?", f.intakeID))
	require.Equal(t, int64(1), f.count(t, "inbox_items", "workflow_purpose = ? AND is_completed = false", "decide_approval"))
	return f
}

func (f *approvalCollaborationPostgresFixture) newRequest() *IntegrationRequest {
	return &IntegrationRequest{Type: "case_creation", Title: "虚构协作案件审批", Content: "虚构补正及事务测试", ApplicantName: "虚构申请律师", DepartmentName: "虚构测试业务部", WorkflowType: "COLLAB_PG", Metadata: map[string]interface{}{"intake_id": f.intakeID, "conflict_check_id": f.checkID}}
}

func (f *approvalCollaborationPostgresFixture) reload(t *testing.T, id string) *models.ApprovalRequest {
	t.Helper()
	var row models.ApprovalRequest
	require.NoError(t, f.db.First(&row, "id = ?", id).Error)
	return &row
}

func (f *approvalCollaborationPostgresFixture) count(t *testing.T, table, where string, args ...interface{}) int64 {
	t.Helper()
	var n int64
	require.NoError(t, f.db.Table(table).Where(where, args...).Count(&n).Error)
	return n
}

func (f *approvalCollaborationPostgresFixture) state(t *testing.T) map[string]string {
	t.Helper()
	state := map[string]string{}
	for _, table := range []string{"approval_requests", "approval_records", "approval_notifications", "approval_snapshots", "approval_case_creation_tracking", "approval_conflict_associations", "inbox_items", "case_intakes", "case_intake_parties", "cases", "case_parties", "entities"} {
		var raw string
		require.NoError(t, f.db.Raw("SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text), '[]'::jsonb)::text FROM "+table+" r").Row().Scan(&raw))
		state[table] = raw
	}
	return state
}

func (f *approvalCollaborationPostgresFixture) decision(t *testing.T, id, decision string) *models.ApprovalDecisionRequest {
	t.Helper()
	a := f.reload(t, id)
	require.NotNil(t, a.ActiveSubmissionSnapshotID)
	return &models.ApprovalDecisionRequest{
		Decision: decision, DecisionReason: "请补充事项说明", ExpectedSubmissionSnapshotID: *a.ActiveSubmissionSnapshotID, ExpectedActionVersion: a.ActionVersion,
		ImposedRequirements: map[string]interface{}{"schema_version": 1, "supplement_requests": []interface{}{map[string]interface{}{"category": "matter_explanation", "field_key": "description", "description": "请补充事项背景", "required": true}}},
	}
}

func (f *approvalCollaborationPostgresFixture) returnForRevision(t *testing.T) *models.ResubmitApprovalRequest {
	t.Helper()
	_, err := NewApprovalService(f.db).ProcessApprovalDecision("22", f.approval.ID, f.decision(t, f.approval.ID, models.ApprovalDecisionRequestChanges))
	require.NoError(t, err)
	var record models.ApprovalRecord
	require.NoError(t, f.db.Where("approval_request_id = ? AND decision = ?", f.approval.ID, models.ApprovalDecisionRequestChanges).First(&record).Error)
	items, err := ReadSupplementRequests(&record)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.EqualValues(t, 0, f.count(t, "approval_case_creation_tracking", "intake_id = ?", f.intakeID))
	return &models.ResubmitApprovalRequest{ExpectedIntakeRevision: "draft-r1", ExpectedSubmissionSnapshotID: *f.approval.ActiveSubmissionSnapshotID, ReturnRecordID: record.ID,
		RevisionNote: "补充事项背景，主体事实不变", Responses: []models.SupplementResponse{{ItemID: items[0].ItemID, Response: "已补充虚构事项背景"}}}
}

// Faults attach only to database writes reached by the real workflow. Every
// caller asserts the callback was hit inside a SQL transaction, so an earlier
// validation failure can never masquerade as rollback coverage.
func approvalCollaborationPostgresFault(t *testing.T, db *gorm.DB, operation, table string) (func(), *atomic.Int32, string) {
	t.Helper()
	name := "approval-collab-pg:" + uuid.NewString()
	marker := "injected-" + operation + "-" + table
	hits := &atomic.Int32{}
	callback := func(tx *gorm.DB) {
		if tx.Statement.Table != table {
			return
		}
		if _, ok := tx.Statement.ConnPool.(gorm.TxCommitter); !ok {
			tx.AddError(errors.New("fault reached outside transaction"))
			return
		}
		hits.Add(1)
		tx.AddError(errors.New(marker))
	}
	var remove func() error
	switch operation {
	case "create":
		require.NoError(t, db.Callback().Create().Before("gorm:create").Register(name, callback))
		remove = func() error { return db.Callback().Create().Remove(name) }
	case "update":
		require.NoError(t, db.Callback().Update().Before("gorm:update").Register(name, callback))
		remove = func() error { return db.Callback().Update().Remove(name) }
	default:
		t.Fatalf("unsupported fault operation %s", operation)
	}
	var once sync.Once
	clear := func() { once.Do(func() { require.NoError(t, remove()) }) }
	t.Cleanup(clear)
	return clear, hits, marker
}

func TestApprovalCollaborationPostgresReturnRollback(t *testing.T) {
	for _, entry := range []string{"ordinary", "integrated"} {
		for _, op := range []string{"update", "create"} {
			t.Run(entry+"_inbox_"+op, func(t *testing.T) {
				f := newApprovalCollaborationPostgresFixture(t)
				req := f.decision(t, f.approval.ID, models.ApprovalDecisionRequestChanges)
				before := f.state(t)
				clear, hits, marker := approvalCollaborationPostgresFault(t, f.db, op, "inbox_items")
				decide := func() (*models.ApprovalRequest, error) {
					if entry == "integrated" {
						return f.integrated().ProcessApprovalWithConflict(context.Background(), "22", "虚构审批合伙人", f.approval.ID, req)
					}
					return NewApprovalService(f.db).ProcessApprovalDecision("22", f.approval.ID, req)
				}
				_, err := decide()
				require.ErrorContains(t, err, marker)
				require.Positive(t, hits.Load())
				require.Equal(t, before, f.state(t), "record/status/slot/inbox must ALL roll back")
				clear()
				result, err := decide()
				require.NoError(t, err)
				require.Equal(t, models.ApprovalStatusNeedsRevision, result.Status)
				require.EqualValues(t, 1, f.count(t, "approval_records", "approval_request_id = ?", f.approval.ID))
				require.EqualValues(t, 0, f.count(t, "approval_case_creation_tracking", "intake_id = ?", f.intakeID))
				require.EqualValues(t, 1, f.count(t, "inbox_items", "workflow_purpose = 'supplement' AND is_completed = false"))
				t.Logf("C13 return entry=%s fault=%s hits=%d rollback_exact=true retry_supplement_tasks=1", entry, op, hits.Load())
			})
		}
	}
}

func TestApprovalCollaborationPostgresResubmitRollback(t *testing.T) {
	for _, fault := range []struct{ operation, table string }{{"create", "approval_snapshots"}, {"update", "inbox_items"}, {"create", "inbox_items"}} {
		t.Run(fault.table+"_"+fault.operation, func(t *testing.T) {
			f := newApprovalCollaborationPostgresFixture(t)
			req := f.returnForRevision(t)
			before := f.state(t)
			clear, hits, marker := approvalCollaborationPostgresFault(t, f.db, fault.operation, fault.table)
			_, err := NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, req)
			require.ErrorContains(t, err, marker)
			require.Positive(t, hits.Load())
			require.Equal(t, before, f.state(t), "pointer/snapshot/slot/status/inbox must ALL roll back")
			clear()
			result, err := NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, req)
			require.NoError(t, err)
			require.Equal(t, models.ApprovalStatusResubmitted, result.Status)
			require.NotEqual(t, *f.approval.ActiveSubmissionSnapshotID, *result.ActiveSubmissionSnapshotID)
			require.EqualValues(t, 2, f.count(t, "approval_snapshots", "approval_request_id = ?", f.approval.ID))
			require.EqualValues(t, 1, f.count(t, "approval_case_creation_tracking", "intake_id = ?", f.intakeID))
			require.EqualValues(t, 1, f.count(t, "inbox_items", "workflow_purpose = 'decide_approval' AND is_completed = false"))
			require.EqualValues(t, 0, f.count(t, "inbox_items", "workflow_purpose = 'supplement' AND is_completed = false"))
			after := f.state(t)
			replayed, err := NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, req)
			require.NoError(t, err)
			require.Equal(t, result.ActiveSubmissionSnapshotID, replayed.ActiveSubmissionSnapshotID)
			require.Equal(t, after, f.state(t), "identical resubmit retry must perform no writes")
			changed := *req
			changed.RevisionNote = "同一退回记录的另一份修改说明"
			_, err = NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, &changed)
			var stale *IntakeCollaborationError
			require.ErrorAs(t, err, &stale)
			require.Equal(t, 409, stale.Status)
			require.Equal(t, after, f.state(t), "different resubmit payload cannot revive the old round")
			t.Logf("C13/C14 resubmit fault=%s/%s hits=%d rollback_exact=true snapshots=2 slot=1 active_decision_task=1 replay_no_writes=true", fault.table, fault.operation, hits.Load())
		})
	}
}

func TestApprovalCollaborationPostgresStaleDecisionNoWrites(t *testing.T) {
	for _, entry := range []string{"ordinary", "integrated"} {
		for _, stale := range []string{"snapshot", "node"} {
			t.Run(entry+"_"+stale, func(t *testing.T) {
				f := newApprovalCollaborationPostgresFixture(t)
				req := f.decision(t, f.approval.ID, models.ApprovalDecisionRequestChanges)
				resubmit := f.returnForRevision(t)
				_, err := NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, resubmit)
				require.NoError(t, err)
				current := f.reload(t, f.approval.ID)
				require.NotEqual(t, req.ExpectedActionVersion, current.ActionVersion)
				if stale == "snapshot" {
					req.ExpectedActionVersion = current.ActionVersion
				} else {
					req.ExpectedSubmissionSnapshotID = *current.ActiveSubmissionSnapshotID
				}
				before := f.state(t)
				if entry == "ordinary" {
					_, err = NewApprovalService(f.db).ProcessApprovalDecision("22", f.approval.ID, req)
				} else {
					_, err = f.integrated().ProcessApprovalWithConflict(context.Background(), "22", "虚构审批合伙人", f.approval.ID, req)
				}
				var gate *ApprovalConflictGateError
				require.ErrorAs(t, err, &gate, "handler maps this typed gate to HTTP 409")
				expected := "APPROVAL_SUBMISSION_VERSION_CONFLICT"
				if stale == "node" {
					expected = "APPROVAL_ACTION_VERSION_CONFLICT"
				}
				require.Equal(t, expected, gate.Code)
				require.Equal(t, before, f.state(t))
				t.Logf("C11 entry=%s stale=%s typed_409=%s writes=0", entry, stale, gate.Code)
			})
		}
	}
}

// Fingerprints in logs make old immutable evidence comparisons reviewable
// without printing full frozen subject documents.
func approvalCollaborationPostgresFingerprint(value string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}

func (f *approvalCollaborationPostgresFixture) assertOneCompleteCase(t *testing.T, approvalID string) string {
	t.Helper()
	a := f.reload(t, approvalID)
	require.True(t, a.CaseCreated)
	require.Equal(t, "completed", a.CaseCreationStatus)
	require.NotEmpty(t, a.CreatedCaseID)
	require.EqualValues(t, 1, f.count(t, "cases", "deleted_at IS NULL"))
	var matter models.Case
	require.NoError(t, f.db.First(&matter, "id = ?", a.CreatedCaseID).Error)
	require.Equal(t, f.checkID, matter.ConflictCheckID)
	require.Equal(t, f.clientID, matter.ClientID)
	require.EqualValues(t, 11, matter.LawyerID)
	var parties []struct {
		Digest    string
		PartyType string
	}
	require.NoError(t, f.db.Raw(`SELECT e.identity_number_digest AS digest, p.party_type
		FROM case_parties p JOIN entities e ON e.id = p.entity_id
		WHERE p.case_id = ? AND p.deleted_at IS NULL AND e.deleted_at IS NULL ORDER BY p.party_type`, matter.ID).Scan(&parties).Error)
	require.Len(t, parties, 2)
	require.Equal(t, f.clientDigest, parties[0].Digest)
	require.Equal(t, "CLIENT", parties[0].PartyType)
	require.Equal(t, f.partyDigest, parties[1].Digest)
	require.Equal(t, "OPPOSING", parties[1].PartyType)
	return a.CreatedCaseID
}

func TestApprovalCollaborationPostgresResubmitVersusNewApproval(t *testing.T) {
	for _, sequence := range []string{"simultaneous-1", "simultaneous-2", "simultaneous-3", "resubmit-then-new", "new-then-resubmit"} {
		t.Run(sequence, func(t *testing.T) {
			f := newApprovalCollaborationPostgresFixture(t)
			req := f.returnForRevision(t)
			type outcome struct {
				label, id string
				err       error
				elapsed   time.Duration
			}
			resubmit := func() outcome {
				start := time.Now()
				a, err := NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, req)
				id := ""
				if a != nil {
					id = a.ID
				}
				return outcome{"resubmit", id, err, time.Since(start)}
			}
			fresh := func() outcome {
				start := time.Now()
				a, err := f.integrated().CreateIntegratedApproval(context.Background(), "11", "虚构申请律师", f.newRequest())
				id := ""
				if a != nil {
					id = a.ApprovalID
				}
				return outcome{"new", id, err, time.Since(start)}
			}
			var outcomes []outcome
			if strings.HasPrefix(sequence, "simultaneous") {
				ready := make(chan struct{}, 2)
				start := make(chan struct{})
				results := make(chan outcome, 2)
				for _, fn := range []func() outcome{resubmit, fresh} {
					go func(run func() outcome) { ready <- struct{}{}; <-start; results <- run() }(fn)
				}
				<-ready
				<-ready
				close(start)
				for i := 0; i < 2; i++ {
					select {
					case result := <-results:
						outcomes = append(outcomes, result)
					case <-time.After(35 * time.Second):
						t.Fatal("concurrent transaction did not finish within bound")
					}
				}
			} else if sequence == "resubmit-then-new" {
				outcomes = []outcome{resubmit(), fresh()}
			} else {
				outcomes = []outcome{fresh(), resubmit()}
			}
			winner := ""
			successes := 0
			for _, out := range outcomes {
				t.Logf("C12 sequence=%s actor=%s approval=%s elapsed=%s error=%v", sequence, out.label, out.id, out.elapsed, out.err)
				if out.err == nil {
					successes++
					winner = out.id
					continue
				}
				var gate *ApprovalConflictGateError
				require.ErrorAs(t, out.err, &gate, "loser must be the controlled conflict, not a raw deadlock/database failure")
				require.Equal(t, "INTAKE_APPROVAL_ALREADY_IN_PROGRESS", gate.Code)
			}
			require.Equal(t, 1, successes)
			require.NotEmpty(t, winner)
			require.EqualValues(t, 1, f.count(t, "approval_case_creation_tracking", "intake_id = ?", f.intakeID))
			require.EqualValues(t, 1, f.count(t, "approval_case_creation_tracking", "intake_id = ? AND approval_request_id = ?", f.intakeID, winner))
			require.EqualValues(t, 1, f.count(t, "approval_requests", "status IN ?", []string{models.ApprovalStatusSubmitted, models.ApprovalStatusUnderReview, models.ApprovalStatusResubmitted}))
			require.EqualValues(t, 2, f.count(t, "approval_snapshots", "snapshot_type = ?", models.IntakeSubmissionSnapshotType))
			require.EqualValues(t, 1, f.count(t, "inbox_items", "workflow_purpose = 'decide_approval' AND is_completed = false"))
			_, err := f.integrated().ProcessApprovalWithConflict(context.Background(), "22", "虚构审批合伙人", winner, f.decision(t, winner, models.ApprovalDecisionApprove))
			require.NoError(t, err, "winning slot must finish the real approval/case workflow")
			caseID := f.assertOneCompleteCase(t, winner)
			if winner != f.approval.ID {
				_, err := f.integrated().AutoCreateCaseFromApproval(context.Background(), f.approval.ID, nil)
				require.Error(t, err, "losing original approval cannot create a case")
			}
			replay, err := f.integrated().AutoCreateCaseFromApproval(context.Background(), winner, nil)
			require.NoError(t, err)
			require.Equal(t, caseID, replay.CaseID)
			f.assertOneCompleteCase(t, winner)
			t.Logf("C12 sequence=%s slots=1 valid_approvals=1 cases=1 complete_parties=2 replay_case=%s", sequence, caseID)
		})
	}
}

func TestApprovalCollaborationPostgresCasePartyRollback(t *testing.T) {
	f := newApprovalCollaborationPostgresFixture(t)
	req := f.returnForRevision(t)
	oldDigest := f.partyDigest
	oldSnapshot := f.state(t)["approval_snapshots"]
	f.reviseReviewedSubject(t, req)
	resubmitted, err := NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, req)
	require.NoError(t, err)
	// Authorize the decision through the actual ordinary decision service;
	// the case transaction is then exercised independently with fault injection.
	_, err = NewApprovalService(f.db).ProcessApprovalDecision("22", f.approval.ID, f.decision(t, f.approval.ID, models.ApprovalDecisionApprove))
	require.NoError(t, err)
	before := f.state(t)
	clear, hits, marker := approvalCollaborationPostgresFault(t, f.db, "create", "case_parties")
	_, err = f.integrated().AutoCreateCaseFromApproval(context.Background(), f.approval.ID, nil)
	require.ErrorContains(t, err, marker)
	require.Positive(t, hits.Load())
	require.Equal(t, before, f.state(t), "case/party/claim/completion/association must ALL roll back")
	require.EqualValues(t, 0, f.count(t, "cases", "1=1"))
	require.EqualValues(t, 0, f.count(t, "case_parties", "1=1"))
	require.EqualValues(t, 0, f.count(t, "approval_case_creation_tracking", "creation_status = 'completed'"))
	clear()
	result, err := f.integrated().AutoCreateCaseFromApproval(context.Background(), f.approval.ID, nil)
	require.NoError(t, err)
	require.Equal(t, "completed", result.Status)
	caseID := f.assertOneCompleteCase(t, f.approval.ID)
	replay, err := f.integrated().AutoCreateCaseFromApproval(context.Background(), f.approval.ID, nil)
	require.NoError(t, err)
	require.Equal(t, caseID, replay.CaseID)
	f.assertOneCompleteCase(t, f.approval.ID)
	require.Equal(t, resubmitted.ActiveSubmissionSnapshotID, f.reload(t, f.approval.ID).ActiveSubmissionSnapshotID)
	require.NotEqual(t, oldDigest, f.partyDigest)
	var originalSnapshot models.ApprovalSnapshot
	require.NoError(t, f.db.First(&originalSnapshot, "id = ?", *f.approval.ActiveSubmissionSnapshotID).Error)
	require.Contains(t, originalSnapshot.SnapshotData, oldDigest)
	require.NotContains(t, originalSnapshot.SnapshotData, f.partyDigest)
	require.Contains(t, oldSnapshot, oldDigest)
	t.Logf("C14 case-party fault_hits=%d rollback_exact=true snapshots_sha256_before=%s cases=1 parties=2 active_snapshot=%s retry_case=%s", hits.Load(), approvalCollaborationPostgresFingerprint(before["approval_snapshots"]), *resubmitted.ActiveSubmissionSnapshotID, caseID)
}

func TestApprovalCollaborationPostgresApprovedSnapshotSurvivesCheckMutation(t *testing.T) {
	f := newApprovalCollaborationPostgresFixture(t)
	req := f.returnForRevision(t)
	f.reviseReviewedSubject(t, req)
	resubmitted, err := NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, req)
	require.NoError(t, err)
	_, err = NewApprovalService(f.db).ProcessApprovalDecision("22", f.approval.ID, f.decision(t, f.approval.ID, models.ApprovalDecisionApprove))
	require.NoError(t, err)
	frozenBefore := f.state(t)["approval_snapshots"]
	var approvedSnapshot models.ApprovalSnapshot
	require.NoError(t, f.db.First(&approvedSnapshot, "id = ?", *resubmitted.ActiveSubmissionSnapshotID).Error)
	require.Contains(t, approvedSnapshot.SnapshotData, f.partyDigest)
	var check models.ConflictCheckRecord
	require.NoError(t, f.db.First(&check, "check_id = ?", f.checkID).Error)
	mutateLiveCheck := func(identity string) string {
		t.Helper()
		digest, err := security.IdentityDigest(identity)
		require.NoError(t, err)
		var parameters map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(approvalCollaborationPostgresJSON(t, check.SearchParameters)), &parameters))
		parameters["parties"] = []interface{}{map[string]interface{}{
			"name": "批准后被修改的虚构主体", "role": "THIRD_PARTY", "entityType": "INDIVIDUAL",
			"identifiers": map[string]interface{}{"id_card": "hmac-sha256:" + digest},
		}}
		// Deliberately model later database drift in the mutable formal check.
		// Do not modify either approved snapshot or substitute any case logic.
		require.NoError(t, f.db.Exec("UPDATE conflict_check_records SET search_parameters = ?::jsonb WHERE check_id = ?", approvalCollaborationPostgresJSON(t, parameters), f.checkID).Error)
		var persisted models.ConflictCheckRecord
		require.NoError(t, f.db.First(&persisted, "check_id = ?", f.checkID).Error)
		require.Contains(t, approvalCollaborationPostgresJSON(t, persisted.SearchParameters), digest)
		require.NotContains(t, approvalCollaborationPostgresJSON(t, persisted.SearchParameters), f.partyDigest)
		return digest
	}
	changedBeforeCreation := mutateLiveCheck("310110199001010044")
	created, err := f.integrated().AutoCreateCaseFromApproval(context.Background(), f.approval.ID, nil)
	require.NoError(t, err, "initial case creation must use approved snapshot parties despite mutable-check drift")
	require.Equal(t, "completed", created.Status)
	caseID := f.assertOneCompleteCase(t, f.approval.ID)
	changedBeforeReplay := mutateLiveCheck("310110199001010055")
	replayed, err := f.integrated().AutoCreateCaseFromApproval(context.Background(), f.approval.ID, nil)
	require.NoError(t, err, "completed replay must use the same approved snapshot, not reread mutable check parties")
	require.Equal(t, caseID, replayed.CaseID)
	require.Equal(t, "completed", replayed.Status)
	f.assertOneCompleteCase(t, f.approval.ID)
	require.EqualValues(t, 0, f.count(t, "entities", "identity_number_digest IN ?", []string{changedBeforeCreation, changedBeforeReplay}))
	require.Equal(t, frozenBefore, f.state(t)["approval_snapshots"], "all submission snapshots must remain byte-for-byte unchanged")
	require.Equal(t, resubmitted.ActiveSubmissionSnapshotID, f.reload(t, f.approval.ID).ActiveSubmissionSnapshotID)
	t.Logf("C14 mutable_check_changed_twice=true cases=1 approved_parties=2 mutated_entities=0 snapshot_round=%d snapshots_unchanged=true replay_case=%s", approvedSnapshot.SourceVersion, caseID)
}

// Seed a second valid, independently reviewed evidence set as fixture input.
// The actual resubmit still executes every fact/review/snapshot guard itself.
func (f *approvalCollaborationPostgresFixture) reviseReviewedSubject(t *testing.T, req *models.ResubmitApprovalRequest) {
	t.Helper()
	var check models.ConflictCheckRecord
	require.NoError(t, f.db.First(&check, "check_id = ?", f.checkID).Error)
	f.checkID = "COLLAB-PG-REVISED-" + uuid.NewString()
	var err error
	f.partyDigest, err = security.IdentityDigest("310110199001010033")
	require.NoError(t, err)
	now := time.Now().UTC().Truncate(time.Microsecond)
	check.CheckID, check.CheckTime, check.CreatedAt, check.UpdatedAt = f.checkID, now, now, now
	check.CheckResult["checkId"] = f.checkID
	check.SearchParameters["parties"] = []interface{}{map[string]interface{}{"name": "虚构对方", "role": "OPPOSING_PARTY", "entityType": "INDIVIDUAL", "identifiers": map[string]interface{}{"id_card": "hmac-sha256:" + f.partyDigest}}}
	require.NoError(t, f.db.Create(&check).Error)
	require.NoError(t, f.db.Create(&models.ConflictReviewerAssignment{ID: uuid.NewString(), CheckID: f.checkID, ReviewerID: 33, AssignedBy: 22, Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true, IndependenceReason: "补正后独立核验", CreatedAt: now, UpdatedAt: now}).Error)
	var response models.ConflictCheckResponse
	require.NoError(t, json.Unmarshal([]byte(approvalCollaborationPostgresJSON(t, check.CheckResult)), &response))
	require.NoError(t, f.db.Create(&models.ConflictReview{CheckID: f.checkID, Decision: "no_conflict", Notes: "补正后重新复核", ReviewerID: 33, ReviewerName: "虚构独立复核人", EvidenceHash: conflictEvidenceHash(&response), CreatedAt: now.Add(time.Second)}).Error)
	require.NoError(t, f.db.Exec(`UPDATE case_intakes SET metadata = metadata || ?::jsonb WHERE id = ?`, approvalCollaborationPostgresJSON(t, map[string]interface{}{"draft_revision": "draft-r2", "lawyer_facts_version": "facts-v2", "lawyer_facts_confirmed_at": now.Add(-time.Second).Format(time.RFC3339), "conflict_check_id": f.checkID}), f.intakeID).Error)
	require.NoError(t, f.db.Exec("UPDATE case_intake_parties SET identity_number_digest = ? WHERE intake_id = ?", f.partyDigest, f.intakeID).Error)
	req.ExpectedIntakeRevision = "draft-r2"
}

// ApprovalCollaborationPostgresHTTPFixture exposes only test data to the
// external test package, allowing use of the real handlers without an import cycle.
func ApprovalCollaborationPostgresHTTPFixture(t *testing.T) (*gorm.DB, *models.ApprovalRequest, *models.ApprovalRequest, func() map[string]string) {
	t.Helper()
	f := newApprovalCollaborationPostgresFixture(t)
	old := f.reload(t, f.approval.ID)
	req := f.returnForRevision(t)
	_, err := NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, req)
	require.NoError(t, err)
	return f.db, old, f.reload(t, f.approval.ID), func() map[string]string { return f.state(t) }
}

type approvalCollaborationPostgresScheduleKey struct{}

// External-package HTTP tests start before any approval exists. The handler
// must run its real claim, transaction binder, freeze and inbox path itself.
func ApprovalCollaborationPostgresInitialHTTPFixture(t *testing.T) (*gorm.DB, *IntegrationRequest, func() map[string]string) {
	t.Helper()
	f := newApprovalCollaborationPostgresIntakeFixture(t)
	return f.db, f.newRequest(), func() map[string]string { return f.state(t) }
}

// Simultaneous starts alone need not overlap the critical locks. These
// callbacks only pause after real SQL locks/inserts; they do not replace any
// business operation, forge a result, or change PostgreSQL locking behavior.
func TestApprovalCollaborationPostgresOverlappingLockOrder(t *testing.T) {
	f := newApprovalCollaborationPostgresFixture(t)
	req := f.returnForRevision(t)
	intakeLocked := make(chan struct{})
	newSlotInserted := make(chan struct{})
	resumeResubmit := make(chan struct{})
	newBackend := make(chan int, 1)
	var intakeOnce, slotOnce, queryOnce, resumeOnce sync.Once
	release := func() { resumeOnce.Do(func() { close(resumeResubmit) }) }
	defer release()
	queryName, createName, beforeQueryName := "collab-pg-observe-intake-lock", "collab-pg-observe-slot", "collab-pg-observe-new-query"
	require.NoError(t, f.db.Callback().Query().After("gorm:query").Register(queryName, func(tx *gorm.DB) {
		if tx.Error != nil || tx.Statement.Table != "case_intakes" || tx.Statement.Context.Value(approvalCollaborationPostgresScheduleKey{}) != "resubmit" {
			return
		}
		if _, locking := tx.Statement.Clauses["FOR"]; !locking {
			return
		}
		intakeOnce.Do(func() {
			close(intakeLocked)
			select {
			case <-resumeResubmit:
			case <-time.After(10 * time.Second):
				tx.AddError(errors.New("test schedule did not release the intake lock"))
			}
		})
	}))
	require.NoError(t, f.db.Callback().Query().Before("gorm:query").Register(beforeQueryName, func(tx *gorm.DB) {
		if tx.Statement.Table != "case_intakes" || tx.Statement.Context.Value(approvalCollaborationPostgresScheduleKey{}) != "new" {
			return
		}
		if _, locking := tx.Statement.Clauses["FOR"]; !locking {
			return
		}
		queryOnce.Do(func() {
			var pid int
			if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Row().Scan(&pid); err != nil {
				tx.AddError(err)
				return
			}
			newBackend <- pid
		})
	}))
	require.NoError(t, f.db.Callback().Create().After("gorm:create").Register(createName, func(tx *gorm.DB) {
		if tx.Error == nil && tx.Statement.Table == "approval_case_creation_tracking" && tx.Statement.Context.Value(approvalCollaborationPostgresScheduleKey{}) == "new" {
			slotOnce.Do(func() { close(newSlotInserted) })
		}
	}))
	t.Cleanup(func() {
		require.NoError(t, f.db.Callback().Query().Remove(queryName))
		require.NoError(t, f.db.Callback().Query().Remove(beforeQueryName))
		require.NoError(t, f.db.Callback().Create().Remove(createName))
	})
	type outcome struct {
		label, id string
		err       error
	}
	results := make(chan outcome, 2)
	var workers sync.WaitGroup
	// Drain workers before callbacks/schema are removed, including fatal paths
	// where a real business error prevents either locking observation.
	defer func() {
		release()
		workers.Wait()
	}()
	workers.Add(1)
	go func() {
		defer workers.Done()
		ctx := context.WithValue(context.Background(), approvalCollaborationPostgresScheduleKey{}, "resubmit")
		a, err := NewApprovalService(f.db.WithContext(ctx)).ResubmitIntakeApproval("11", f.approval.ID, req)
		id := ""
		if a != nil {
			id = a.ID
		}
		results <- outcome{"resubmit", id, err}
	}()
	select {
	case <-intakeLocked:
	case out := <-results:
		t.Fatalf("%s ended before acquiring the intake lock: %v", out.label, out.err)
	case <-time.After(8 * time.Second):
		t.Fatal("resubmit did not acquire its real intake lock")
	}
	workers.Add(1)
	go func() {
		defer workers.Done()
		ctx := context.WithValue(context.Background(), approvalCollaborationPostgresScheduleKey{}, "new")
		a, err := f.integrated().CreateIntegratedApproval(ctx, "11", "虚构申请律师", f.newRequest())
		id := ""
		if a != nil {
			id = a.ApprovalID
		}
		results <- outcome{"new", id, err}
	}()
	var blockedPID int
	select {
	case blockedPID = <-newBackend:
	case out := <-results:
		t.Fatalf("%s ended before the competing intake locking query: %v", out.label, out.err)
	case <-time.After(5 * time.Second):
		t.Fatal("new approval did not reach the real intake locking query")
	}
	blocked := false
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		require.NoError(t, f.db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE pid = ? AND wait_event_type = 'Lock' AND cardinality(pg_blocking_pids(pid)) > 0)", blockedPID).Row().Scan(&blocked))
		if blocked {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.True(t, blocked, "must observe a genuine PostgreSQL lock wait, not only goroutine startup")
	select {
	case <-newSlotInserted:
		t.Error("new approval inserted its intake slot before waiting on the intake lock; lock order is reversed")
	default:
	}
	t.Logf("C12 lock evidence: new_backend=%d wait_event_type=Lock blockers_present=true slot_not_inserted_while_waiting=true", blockedPID)
	release()
	winner := ""
	successes := 0
	for i := 0; i < 2; i++ {
		select {
		case out := <-results:
			t.Logf("C12 controlled_overlap actor=%s approval=%s error=%v", out.label, out.id, out.err)
			if out.err == nil {
				successes++
				winner = out.id
				continue
			}
			var gate *ApprovalConflictGateError
			if !errors.As(out.err, &gate) {
				var pgErr *pgconn.PgError
				state := "not-a-pg-error"
				if errors.As(out.err, &pgErr) {
					state = pgErr.Code
				}
				t.Errorf("C12 overlapping intake/slot locks returned uncontrolled error (SQLSTATE=%s), expected INTAKE_APPROVAL_ALREADY_IN_PROGRESS: %v", state, out.err)
			} else {
				require.Equal(t, "INTAKE_APPROVAL_ALREADY_IN_PROGRESS", gate.Code)
			}
		case <-time.After(25 * time.Second):
			t.Fatal("controlled overlap did not finish")
		}
	}
	require.Equal(t, 1, successes)
	require.EqualValues(t, 1, f.count(t, "approval_case_creation_tracking", "intake_id = ?", f.intakeID))
	require.EqualValues(t, 1, f.count(t, "approval_requests", "status IN ?", []string{models.ApprovalStatusSubmitted, models.ApprovalStatusResubmitted, models.ApprovalStatusUnderReview}))
	_, err := f.integrated().ProcessApprovalWithConflict(context.Background(), "22", "虚构审批合伙人", winner, f.decision(t, winner, models.ApprovalDecisionApprove))
	require.NoError(t, err)
	f.assertOneCompleteCase(t, winner)
	t.Log("C12 controlled_overlap slots=1 valid_approvals=1 cases=1 complete_parties=2; error contract evaluated separately")
}

func TestApprovalCollaborationPostgresCaseCompletionInboxRollback(t *testing.T) {
	for _, phase := range []string{"initial_completion", "completed_replay"} {
		for _, operation := range []string{"update", "create"} {
			t.Run(phase+"_inbox_"+operation, func(t *testing.T) {
				f := newApprovalCollaborationPostgresFixture(t)
				_, err := NewApprovalService(f.db).ProcessApprovalDecision("22", f.approval.ID, f.decision(t, f.approval.ID, models.ApprovalDecisionApprove))
				require.NoError(t, err)
				var resultTask models.InboxItem
				require.NoError(t, f.db.Where("workflow_group = ? AND workflow_purpose = 'result'", "approval:"+f.approval.ID).First(&resultTask).Error)
				require.True(t, resultTask.IsCompleted)
				require.NotEmpty(t, resultTask.WorkflowTaskKey)
				inboxBeforeCompletion := f.state(t)["inbox_items"]
				if phase == "completed_replay" {
					_, err = f.integrated().AutoCreateCaseFromApproval(context.Background(), f.approval.ID, nil)
					require.NoError(t, err)
					f.assertOneCompleteCase(t, f.approval.ID)
				}
				before := f.state(t)
				clear, hits, marker := approvalCollaborationPostgresFault(t, f.db, operation, "inbox_items")
				// The decision has already committed. These callbacks can only
				// fire in the actual case-completion/replay transaction.
				_, err = f.integrated().AutoCreateCaseFromApproval(context.Background(), f.approval.ID, nil)
				require.ErrorContains(t, err, marker)
				require.Positive(t, hits.Load(), "case completion must reach its transactional inbox handoff")
				require.Equal(t, before, f.state(t), "case/parties/tracking/association/inbox must all roll back on completion handoff failure")
				if phase == "initial_completion" {
					require.EqualValues(t, 0, f.count(t, "cases", "1=1"))
					require.EqualValues(t, 0, f.count(t, "case_parties", "1=1"))
					require.EqualValues(t, 0, f.count(t, "approval_case_creation_tracking", "creation_status = 'completed'"))
					require.False(t, f.reload(t, f.approval.ID).CaseCreated)
				} else {
					f.assertOneCompleteCase(t, f.approval.ID)
				}
				clear()
				created, err := f.integrated().AutoCreateCaseFromApproval(context.Background(), f.approval.ID, nil)
				require.NoError(t, err)
				require.Equal(t, "completed", created.Status)
				caseID := f.assertOneCompleteCase(t, f.approval.ID)
				replayed, err := f.integrated().AutoCreateCaseFromApproval(context.Background(), f.approval.ID, nil)
				require.NoError(t, err)
				require.Equal(t, caseID, replayed.CaseID)
				f.assertOneCompleteCase(t, f.approval.ID)
				require.EqualValues(t, 1, f.count(t, "inbox_items", "workflow_group = ? AND workflow_purpose = 'result'", "approval:"+f.approval.ID))
				var afterTask models.InboxItem
				require.NoError(t, f.db.Where("workflow_task_key = ?", resultTask.WorkflowTaskKey).First(&afterTask).Error)
				require.Equal(t, resultTask.ID, afterTask.ID)
				require.Equal(t, inboxBeforeCompletion, f.state(t)["inbox_items"], "completion and replay must retain the same result task, including its stored state")
				t.Logf("C13/C14 case_completion phase=%s fault=inbox_%s hits=%d rollback_exact=true cases=1 parties=2 result_tasks=1 same_result_key=true replay_case=%s", phase, operation, hits.Load(), caseID)
			})
		}
	}
}

func TestApprovalCollaborationPostgresLegacyMissingIntakeLinkNoWrites(t *testing.T) {
	f := newApprovalCollaborationPostgresIntakeFixture(t)
	originalMetadata := map[string]interface{}{"intake_id": f.intakeID, "conflict_check_id": f.checkID}
	draft, err := NewApprovalService(f.db).CreateApproval("11", "虚构申请律师", &models.CreateApprovalRequest{
		Type: "case_creation", Title: "虚构历史草稿", Content: "遗留元数据关联删除回归", WorkflowType: "COLLAB_PG", Metadata: originalMetadata,
	})
	require.NoError(t, err)
	require.Nil(t, draft.ActiveSubmissionSnapshotID)
	// Simulate an already-corrupted legacy draft, preserving the durable
	// formal check. The real Submit/Freeze must not treat it as non-intake.
	require.NoError(t, f.db.Table("approval_requests").Where("id = ?", draft.ID).Update("metadata", approvalCollaborationPostgresJSON(t, map[string]interface{}{"conflict_check_id": f.checkID})).Error)
	before := f.state(t)
	_, err = NewApprovalService(f.db).SubmitApproval("11", draft.ID)
	var gate *IntakeCollaborationError
	require.ErrorAs(t, err, &gate)
	require.Equal(t, 409, gate.Status)
	require.Equal(t, "APPROVAL_SNAPSHOT_UNRESOLVED", gate.Code)
	require.Equal(t, before, f.state(t), "missing legacy metadata association cannot bypass Freeze through its early no-op")
	require.Equal(t, models.ApprovalStatusDraft, f.reload(t, draft.ID).Status)
	require.NoError(t, f.db.Table("approval_requests").Where("id = ?", draft.ID).Update("metadata", approvalCollaborationPostgresJSON(t, originalMetadata)).Error)
	submitted, err := NewApprovalService(f.db).SubmitApproval("11", draft.ID)
	require.NoError(t, err, "restoring the original association must permit real protocol submission")
	require.NotNil(t, submitted.ActiveSubmissionSnapshotID)
	require.EqualValues(t, 1, f.count(t, "approval_snapshots", "approval_request_id = ?", draft.ID))
	require.EqualValues(t, 1, f.count(t, "approval_case_creation_tracking", "intake_id = ? AND approval_request_id = ?", f.intakeID, draft.ID))
	t.Log("C11 legacy_missing_metadata_link=typed_409 writes=0 restored_link_submit=success snapshot=1 slot=1")
}

// Only business data is shared with the production-bootstrap HTTP smoke.
func ApprovalBootstrapPostgresHTTPFixture(t *testing.T, db *gorm.DB) (*IntegrationRequest, func() map[string]string) {
	t.Helper()
	f := seedApprovalCollaborationPostgresIntake(t, db)
	return f.newRequest(), func() map[string]string { return f.state(t) }
}
