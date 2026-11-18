package services

// C14 (PG isolation): a database-level fault in the final case-party
// materialization must abort the whole approval-driven case transaction —
// no case row, no party rows, no fake completed tracking, approval still
// claimable. After the fault is removed, the identical request retries
// successfully and materializes exactly one case whose party set equals the
// reviewed snapshot (client + opposing party); a further replay takes the
// completed fast path and reuses the same case.
//
// Runs against the isolated trial database; skipped unless
// INTAKE_MVP_PG_DSN is set (same contract as approval_resubmit_pg_test.go).

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

type pgCaseTxWorld struct {
	t            *testing.T
	db           *gorm.DB
	svc          *approvalConflictIntegrationService
	seed         string
	approvalID   string
	checkID      string
	clientName   string
	opposingName string
	clientEmail  string
	lawyerName   string
	reviewerName string
	opposingDgst string
	clientID     uint
	lawyerID     uint
}

func newPGCaseTxWorld(t *testing.T) *pgCaseTxWorld {
	t.Helper()
	dsn := os.Getenv("INTAKE_MVP_PG_DSN")
	if dsn == "" {
		t.Skip("INTAKE_MVP_PG_DSN not set; PG isolation evidence runs separately")
	}
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	w := &pgCaseTxWorld{
		t:            t,
		db:           db,
		seed:         strings.ReplaceAll(uuid.NewString(), "-", "")[:16],
		clientName:   "C14PG委托方-",
		opposingName: "C14PG对手方-",
	}
	w.clientName += w.seed
	w.opposingName += w.seed
	w.approvalID = "APR-C14PG-" + w.seed
	w.checkID = "CHK-C14PG-" + w.seed
	w.clientEmail = "c14pg-" + w.seed + "@example.com"
	w.lawyerName = "c14pg-lawyer-" + w.seed
	w.reviewerName = "c14pg-reviewer-" + w.seed
	w.opposingDgst = "c14pg-" + w.seed

	// t.Cleanup runs LIFO: registering the full mutable purge first lets
	// any later fixture-level cleanups run before it. conflict_check_records
	// and conflict_reviews are append-only evidence (DB trigger rejects
	// UPDATE/DELETE, SQLSTATE P0001); this test never deletes that evidence
	// and never disables the protection. The DSN must point at a disposable
	// database: the synthetic evidence rows are discarded together with the
	// throwaway clone after the suite (C14 leak fix 2026-09-28).
	t.Cleanup(func() { w.purgeMutable(t) })

	client := &models.Client{Name: w.clientName, Email: w.clientEmail}
	require.NoError(t, db.Create(client).Error)
	w.clientID = client.ID
	lawyer := &models.User{Username: w.lawyerName, Email: w.lawyerName + "@example.com", Password: "x", Status: "active"}
	require.NoError(t, db.Create(lawyer).Error)
	w.lawyerID = lawyer.ID
	reviewer := &models.User{Username: w.reviewerName, Email: w.reviewerName + "@example.com", Password: "x", Status: "active"}
	require.NoError(t, db.Create(reviewer).Error)

	// The GORM postgres driver sends driver.Valuer []byte payloads as bytea
	// literals, which the json column rejects. Seed through raw SQL with an
	// explicit ::jsonb cast, the same contract as the approval seed below.
	require.NoError(t, db.Exec(
		"INSERT INTO conflict_check_records (check_id, client_id, client_name, check_status, search_parameters, check_result, user_id, check_time) "+
			"VALUES (?, ?, ?, 'COMPLETED', ?::jsonb, ?::jsonb, ?, NOW())",
		w.checkID, strconv.FormatUint(uint64(client.ID), 10), client.Name,
		fmt.Sprintf(`{"parties":[{"name":%q,"role":"OPPOSING_PARTY","entityType":"LEGAL_PERSON","identifiers":{"id_card":%q}}]}`,
			w.opposingName, "hmac-sha256:"+w.opposingDgst),
		`{"decision":{"coverageStatus":"COMPLETE"},"hasConflict":false}`,
		lawyer.ID).Error)
	review := &models.ConflictReview{
		CheckID: w.checkID, Decision: "no_conflict", Notes: "C14 PG evidence",
		ReviewerID: reviewer.ID, EvidenceHash: conflictEvidenceHash(&models.ConflictCheckResponse{CheckID: w.checkID, HasConflict: false}),
	}
	require.NoError(t, db.Create(review).Error)
	t.Logf("c14 world ids: client=%d approval=%s check=%s lawyer=%s reviewer=%s clientName=%s opposingName=%s",
		w.clientID, w.approvalID, w.checkID, w.lawyerName, w.reviewerName, w.clientName, w.opposingName)

	require.NoError(t, db.Exec(
		"INSERT INTO approval_requests (id, request_number, title, type, content, applicant_id, applicant_name, status, workflow_type, created_by, metadata, workflow_config, attachments, conflict_result, created_at, updated_at) "+
			"VALUES (?, ?, 'C14', 'CASE_CREATION', 'C14 PG isolation', 'u-lawyer', '承办律师', 'approved', ?, 'u-lawyer', ?::jsonb, '{}', '[]', '{}', NOW(), NOW())",
		w.approvalID, "C14-"+w.seed, "C14 case creation",
		fmt.Sprintf(`{"conflict_check_id":%q}`, w.checkID)).Error)

	integrationRepo := repositories.NewIntegrationRepository(db)
	caseSvc := NewCaseService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db))
	caseSvc.SetSubjectRecheckService(NewSubjectRecheckService(db, nil))
	w.svc = &approvalConflictIntegrationService{
		approvalService: &txProbeApprovalService{approval: &models.ApprovalRequest{ID: w.approvalID, Status: models.ApprovalStatusApproved}},
		caseService:     caseSvc,
		integrationRepo: integrationRepo,
		db:              db,
	}
	return w
}

// purgeMutable removes every mutable row this world created, in
// foreign-key-safe child-before-parent order, scoped to this world's exact
// generated IDs only. Append-only evidence (conflict_check_records,
// conflict_reviews) is intentionally left in place: the DB trigger forbids
// deletes and the rows are synthetic audit data discarded with the
// disposable clone. Every delete error fails the test.
func (w *pgCaseTxWorld) purgeMutable(t *testing.T) {
	t.Helper()
	del := func(label, query string, args ...interface{}) int64 {
		res := w.db.Exec(query, args...)
		require.NoError(t, res.Error, "c14 cleanup %s delete failed", label)
		return res.RowsAffected
	}
	inbox := del("inbox", "DELETE FROM inbox_items WHERE source_ref = ?", w.approvalID)
	// purgeMutable is registered before the client insert assigns
	// w.clientID; a failed insert leaves the Go zero value 0. Deleting by
	// client_id=0 would not be a precise generated ID, so the case/party
	// sweep is skipped entirely on that path. Email/name/UUID-scoped
	// deletes below stay: they target this world's exact generated values.
	var parties, cases int64
	if w.clientID != 0 {
		parties = del("case-parties", "DELETE FROM case_parties WHERE case_id IN (SELECT id FROM cases WHERE client_id = ?)", w.clientID)
		cases = del("cases", "DELETE FROM cases WHERE client_id = ?", w.clientID)
	} else {
		t.Log("c14 purge skips case/party deletes: clientID=0 (client insert failed before assignment)")
	}
	entities := del("entities", "DELETE FROM entities WHERE name = ? OR name = ? OR identity_number_digest = ?", w.clientName, w.opposingName, w.opposingDgst)
	tracking := del("tracking", "DELETE FROM approval_case_creation_tracking WHERE approval_request_id = ?", w.approvalID)
	records := del("approval-record", "DELETE FROM approval_records WHERE approval_request_id = ?", w.approvalID)
	approvals := del("approval", "DELETE FROM approval_requests WHERE id = ?", w.approvalID)
	users := del("users", "DELETE FROM users WHERE username = ? OR username = ?", w.lawyerName, w.reviewerName)
	clients := del("clients", "DELETE FROM clients WHERE email = ?", w.clientEmail)
	t.Logf("c14 purge rows: inbox=%d parties=%d cases=%d entities=%d tracking=%d records=%d approvals=%d users=%d clients=%d (immutable conflict evidence check=%s stays for the disposable clone)",
		inbox, parties, cases, entities, tracking, records, approvals, users, clients, w.checkID)

	count := func(label, query string, args ...interface{}) {
		var n int64
		require.NoError(t, w.db.Raw("SELECT COUNT(*) FROM "+query, args...).Scan(&n).Error,
			"c14 cleanup %s verify query failed", label)
		require.Zero(t, n, "c14 cleanup residue in %s", label)
	}
	count("inbox", "inbox_items WHERE source_ref = ?", w.approvalID)
	if w.clientID != 0 {
		count("case-parties", "case_parties WHERE case_id IN (SELECT id FROM cases WHERE client_id = ?)", w.clientID)
		count("cases", "cases WHERE client_id = ?", w.clientID)
	}
	count("entities", "entities WHERE name = ? OR name = ? OR identity_number_digest = ?", w.clientName, w.opposingName, w.opposingDgst)
	count("tracking", "approval_case_creation_tracking WHERE approval_request_id = ?", w.approvalID)
	count("approval-record", "approval_records WHERE approval_request_id = ?", w.approvalID)
	count("approval", "approval_requests WHERE id = ?", w.approvalID)
	count("users", "users WHERE username = ? OR username = ?", w.lawyerName, w.reviewerName)
	count("clients", "clients WHERE email = ?", w.clientEmail)
}

func (w *pgCaseTxWorld) caseData() map[string]interface{} {
	return map[string]interface{}{
		"title": "C14成案事务测试案件", "case_type": "CIVIL", "priority": "HIGH",
		"billing_method": "HOURLY", "client_id": w.clientID, "lawyer_id": w.lawyerID,
		"conflict_check_id": w.checkID, "conflict_task_id": w.checkID,
	}
}

func (w *pgCaseTxWorld) caseCount() int64 {
	var n int64
	require.NoError(w.t, w.db.Model(&models.Case{}).Where("client_id = ?", w.clientID).Count(&n).Error)
	return n
}

func (w *pgCaseTxWorld) partyCount(caseID uint) int64 {
	var n int64
	require.NoError(w.t, w.db.Model(&models.CaseParty{}).Where("case_id = ?", caseID).Count(&n).Error)
	return n
}

func (w *pgCaseTxWorld) partyRoleCount(caseID uint, role string) int64 {
	var n int64
	require.NoError(w.t, w.db.Model(&models.CaseParty{}).Where("case_id = ? AND role = ?", caseID, role).Count(&n).Error)
	return n
}

func TestIntakeMVP_PG_CasePartyFaultRollsBackThenRetrySucceeds(t *testing.T) {
	w := newPGCaseTxWorld(t)
	const funcName = "intake_mvp_c14_fault"
	const triggerName = "trg_intake_mvp_c14_fault"
	require.NoError(t, w.db.Exec(
		"CREATE FUNCTION "+funcName+"() RETURNS trigger AS $$ BEGIN RAISE EXCEPTION 'intake-mvp-c14 fault'; END; $$ LANGUAGE plpgsql").Error)
	require.NoError(t, w.db.Exec(
		"CREATE TRIGGER "+triggerName+" BEFORE INSERT ON case_parties FOR EACH ROW EXECUTE FUNCTION "+funcName+"()").Error)
	t.Cleanup(func() {
		w.db.Exec("DROP TRIGGER IF EXISTS " + triggerName + " ON case_parties")
		w.db.Exec("DROP FUNCTION IF EXISTS " + funcName + "()")
	})

	// Fault window: the final party write aborts and the whole case
	// transaction rolls back.
	_, err := w.svc.AutoCreateCaseFromApproval(context.Background(), w.approvalID, w.caseData())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "写入案件当事人失败")
	assert.Contains(t, err.Error(), "intake-mvp-c14 fault")
	assert.EqualValues(t, 0, w.caseCount(), "party fault must roll back the case row")
	var approval models.ApprovalRequest
	require.NoError(t, w.db.First(&approval, "id = ?", w.approvalID).Error)
	assert.False(t, approval.CaseCreated, "approval linkage must roll back with the transaction")
	assert.Empty(t, approval.CreatedCaseID)
	var trackingCount int64
	require.NoError(t, w.db.Table("approval_case_creation_tracking").
		Where("approval_request_id = ?", w.approvalID).Count(&trackingCount).Error)
	assert.EqualValues(t, 0, trackingCount, "no completed tracking may survive a party fault")

	// Clear the fault and replay the identical request.
	require.NoError(t, w.db.Exec("DROP TRIGGER "+triggerName+" ON case_parties").Error)
	require.NoError(t, w.db.Exec("DROP FUNCTION "+funcName+"()").Error)
	first, err := w.svc.AutoCreateCaseFromApproval(context.Background(), w.approvalID, w.caseData())
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Equal(t, "completed", first.Status)
	assert.NotEmpty(t, first.CaseNumber)
	assert.EqualValues(t, 1, w.caseCount(), "retry must create exactly one case")
	caseID64, parseErr := strconv.ParseUint(first.CaseID, 10, 64)
	require.NoError(t, parseErr)
	caseID := uint(caseID64)
	assert.EqualValues(t, 2, w.partyCount(caseID), "client + opposing party from the reviewed snapshot must land")
	assert.EqualValues(t, 1, w.partyRoleCount(caseID, string(models.PartyRolePlaintiff)))
	assert.EqualValues(t, 1, w.partyRoleCount(caseID, string(models.PartyRoleDefendant)))

	// Object-level identity check: the live case-party set (joined with its
	// entities) must equal the reviewed snapshot of THIS check record as a
	// (digest/name, role, partyType) set in both directions. Counts alone
	// cannot prove the right subjects landed.
	record := &models.ConflictCheckRecord{}
	require.NoError(t, w.db.First(record, "check_id = ?", w.checkID).Error)
	expectedParties, buildErr := buildReviewedParties(record)
	require.NoError(t, buildErr)
	require.NotEmpty(t, expectedParties)
	type partyIdentity struct {
		digest    string
		name      string
		role      string
		partyType string
	}
	snapshotIdentity := func(party models.ConflictPartyInfo) partyIdentity {
		digest := ""
		for _, key := range []string{"id_card", "social_credit_code", "unified_social_credit_code"} {
			if strings.HasPrefix(party.Identifiers[key], "hmac-sha256:") {
				digest = strings.TrimPrefix(party.Identifiers[key], "hmac-sha256:")
				break
			}
		}
		role, partyType := "", "RELATED"
		switch repositories.NormalizeConflictPartyRole(party.Role) {
		case "CLIENT":
			role, partyType = string(models.PartyRolePlaintiff), string(models.PartyTypeClient)
		case "OPPOSING_PARTY":
			role, partyType = string(models.PartyRoleDefendant), string(models.PartyTypeOpposing)
		}
		name := ""
		if digest == "" {
			name = strings.ToLower(strings.TrimSpace(party.Name))
		}
		return partyIdentity{digest: digest, name: name, role: role, partyType: partyType}
	}
	expected := map[partyIdentity]bool{}
	for _, party := range expectedParties {
		expected[snapshotIdentity(party)] = true
	}
	require.Len(t, expected, 2, "reviewed snapshot must contain client + opposing party")
	var partyRows []models.CaseParty
	require.NoError(t, w.db.Where("case_id = ? AND deleted_at IS NULL", caseID).Find(&partyRows).Error)
	actual := map[partyIdentity]bool{}
	for _, row := range partyRows {
		var entity models.Entity
		require.NoError(t, w.db.First(&entity, row.EntityID).Error)
		digest := strings.TrimSpace(entity.IdentityNumberDigest)
		name := ""
		if digest == "" {
			name = strings.ToLower(strings.TrimSpace(entity.Name))
		}
		actual[partyIdentity{digest: digest, name: name, role: string(row.Role), partyType: string(row.PartyType)}] = true
	}
	assert.Equal(t, expected, actual, "case parties must equal the reviewed snapshot of this check record")

	var trackingRows []repositories.CaseCreationTracking
	require.NoError(t, w.db.Where("approval_request_id = ?", w.approvalID).Find(&trackingRows).Error)
	require.Len(t, trackingRows, 1, "completed creation must leave exactly one tracking row")
	assert.Equal(t, "completed", trackingRows[0].CreationStatus)
	require.NotNil(t, trackingRows[0].CaseID)
	assert.Equal(t, first.CaseID, *trackingRows[0].CaseID)
	require.NoError(t, w.db.First(&approval, "id = ?", w.approvalID).Error)
	assert.True(t, approval.CaseCreated)
	assert.Equal(t, first.CaseID, approval.CreatedCaseID)

	// Completed fast path: another identical request reuses the same case
	// without creating a second one or extra parties.
	second, err := w.svc.AutoCreateCaseFromApproval(context.Background(), w.approvalID, w.caseData())
	require.NoError(t, err)
	assert.Equal(t, first.CaseID, second.CaseID, "fast path must reuse the same case")
	assert.EqualValues(t, 1, w.caseCount())
	assert.EqualValues(t, 2, w.partyCount(caseID))
}
