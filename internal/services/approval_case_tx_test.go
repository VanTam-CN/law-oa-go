package services

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ---- sqlite fixture -------------------------------------------------------

func setupApprovalCaseTxDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("case_tx_%d_%d.db", time.Now().UnixNano(), rand.Int63()))
	db, err := gorm.Open(sqlite.Open("file:"+dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.ApprovalRequest{},
		&models.Case{},
		&models.Client{},
		&models.User{},
		&models.Entity{},
		&models.CaseParty{},
		&models.ConflictCheckRecord{},
		&models.ConflictReview{},
		&repositories.CaseCreationTracking{},
	))
	return db
}

// seededFixture is the minimum consistent world for AutoCreateCaseFromApproval.
type seededFixture struct {
	db       *gorm.DB
	svc      ApprovalConflictIntegrationService
	tracker  *txProbeIntegrationRepo
	approval *models.ApprovalRequest
}

func seedApprovalCaseWorld(t *testing.T) *seededFixture {
	t.Helper()
	db := setupApprovalCaseTxDB(t)

	client := &models.Client{Name: "成案客户有限公司", Email: fmt.Sprintf("case-tx-%d@example.com", time.Now().UnixNano())}
	require.NoError(t, db.Create(client).Error)
	lawyer := &models.User{Username: fmt.Sprintf("lawyer-%d", time.Now().UnixNano()), Email: fmt.Sprintf("lawyer-%d@example.com", time.Now().UnixNano()), Password: "x", Status: "active"}
	require.NoError(t, db.Create(lawyer).Error)
	reviewer := &models.User{Username: fmt.Sprintf("reviewer-%d", time.Now().UnixNano()), Email: fmt.Sprintf("reviewer-%d@example.com", time.Now().UnixNano()), Password: "x", Status: "active"}
	require.NoError(t, db.Create(reviewer).Error)

	resp := models.ConflictCheckResponse{CheckID: "CHK-TX-1", HasConflict: false}
	check := &models.ConflictCheckRecord{
		CheckID: "CHK-TX-1", ClientID: fmt.Sprintf("%d", client.ID), ClientName: client.Name,
		CheckStatus: "COMPLETED", UserID: lawyer.ID,
		SearchParameters: models.JSON{"parties": []interface{}{
			map[string]interface{}{"name": "对手方主体有限公司", "role": "OPPOSING_PARTY", "entityType": "LEGAL_PERSON", "identifiers": map[string]interface{}{"id_card": "hmac-sha256:tx-opposing"}},
		}},
		CheckResult: models.JSON{"decision": map[string]interface{}{"coverageStatus": "COMPLETE"}, "hasConflict": false},
		CheckTime:   time.Now(),
	}
	require.NoError(t, db.Create(check).Error)
	review := &models.ConflictReview{
		CheckID: "CHK-TX-1", Decision: "no_conflict", Notes: "ok",
		ReviewerID: reviewer.ID, ReviewerName: "独立复核人", EvidenceHash: conflictEvidenceHash(&resp),
	}
	require.NoError(t, db.Create(review).Error)

	approval := &models.ApprovalRequest{
		ID: "APR-TX-1", RequestNumber: fmt.Sprintf("RN-%d", time.Now().UnixNano()), Title: "成案事务审批",
		Type: "CASE_CREATION", Content: "approval body", ApplicantID: "u1", ApplicantName: "申请人",
		DepartmentName: "业务部", Status: models.ApprovalStatusApproved, WorkflowType: "case_creation",
		Metadata: `{"conflict_check_id":"CHK-TX-1"}`,
	}
	require.NoError(t, db.Create(approval).Error)

	integrationRepo := repositories.NewIntegrationRepository(db)
	tracker := &txProbeIntegrationRepo{IntegrationRepository: integrationRepo}
	caseSvc := NewCaseService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db))
	caseSvc.SetSubjectRecheckService(NewSubjectRecheckService(db, nil))
	svc := &approvalConflictIntegrationService{
		approvalService: &txProbeApprovalService{approval: approval},
		caseService:     caseSvc,
		integrationRepo: tracker,
		db:              db,
	}
	return &seededFixture{db: db, svc: svc, tracker: tracker, approval: approval}
}

func caseTxData() map[string]interface{} {
	return map[string]interface{}{
		"title": "成案事务测试案件", "case_type": "CIVIL", "priority": "HIGH",
		"billing_method": "HOURLY", "client_id": 1, "lawyer_id": 1,
		"conflict_check_id": "CHK-TX-1", "conflict_task_id": "CHK-TX-1",
	}
}

func caseTxDataFor(clientID, lawyerID uint) map[string]interface{} {
	data := caseTxData()
	data["client_id"] = clientID
	data["lawyer_id"] = lawyerID
	return data
}

func trackingCompleted(db *gorm.DB, t *testing.T) bool {
	t.Helper()
	var tracking struct{ CreationStatus string }
	err := db.Table("approval_case_creation_tracking").
		Where("approval_request_id = ?", "APR-TX-1").
		Order("created_at DESC").First(&tracking).Error
	if err != nil {
		return false
	}
	return tracking.CreationStatus == "completed"
}

// TestCasePartiesMatchSnapshotUnit pins the matching contract directly:
// digest identity dominates and role sets must match exactly.
func TestCasePartiesMatchSnapshotUnit(t *testing.T) {
	tx := setupApprovalCaseTxDB(t)
	client := &models.Client{Name: "成案客户有限公司", Email: "match-client@example.com"}
	require.NoError(t, tx.Create(client).Error)
	case_ := &models.Case{CaseNumber: "CASE-MATCH-1", Title: "t", ClientID: client.ID, CaseType: "CIVIL", Status: "active"}
	require.NoError(t, tx.Create(case_).Error)
	entities := []*models.Entity{
		{EntityType: models.EntityTypeLegalPerson, Name: "成案客户有限公司", Status: models.EntityStatusActive},
		{EntityType: models.EntityTypeLegalPerson, Name: "对手方主体有限公司", IdentityNumberDigest: "tx-opposing-digest", Status: models.EntityStatusActive},
	}
	for _, e := range entities {
		require.NoError(t, tx.Create(e).Error)
	}
	parties := []*models.CaseParty{
		{CaseID: case_.ID, EntityID: entities[0].ID, Role: models.PartyRolePlaintiff, PartyType: models.PartyTypeClient},
		{CaseID: case_.ID, EntityID: entities[1].ID, Role: models.PartyRoleDefendant, PartyType: models.PartyTypeOpposing},
	}
	for _, p := range parties {
		require.NoError(t, tx.Create(p).Error)
	}
	snapshot := []models.ConflictPartyInfo{
		{Name: "成案客户有限公司", Role: "CLIENT", EntityType: "LEGAL_PERSON"},
		{Name: "对手方主体有限公司", Role: "OPPOSING_PARTY", EntityType: "LEGAL_PERSON", Identifiers: map[string]string{"id_card": "hmac-sha256:tx-opposing-digest"}},
	}
	var diagRows []models.CaseParty
	require.NoError(t, tx.Where("case_id = ?", case_.ID).Find(&diagRows).Error)
	assert.True(t, casePartiesMatchSnapshot(tx, case_.ID, snapshot))
	// A role mismatch must fail.
	swapped := []models.ConflictPartyInfo{
		{Name: "成案客户有限公司", Role: "OPPOSING_PARTY", EntityType: "LEGAL_PERSON"},
		{Name: "对手方主体有限公司", Role: "OPPOSING_PARTY", EntityType: "LEGAL_PERSON", Identifiers: map[string]string{"id_card": "hmac-sha256:tx-opposing-digest"}},
	}
	assert.False(t, casePartiesMatchSnapshot(tx, case_.ID, swapped))
}

// ---- probe doubles --------------------------------------------------------

type txProbeApprovalService struct {
	ApprovalServiceInterface
	approval *models.ApprovalRequest
}

func (s *txProbeApprovalService) GetApprovalByID(id string) (*models.ApprovalRequest, error) {
	return s.approval, nil
}

type txProbeIntegrationRepo struct {
	*repositories.IntegrationRepository
	failTrackingCreate    func(tracking *repositories.CaseCreationTracking) error
	failTrackingUpdate    func(id string, updates map[string]interface{}) error
	failAssociationUpdate func(approvalID string) error
}

func (r *txProbeIntegrationRepo) CreateCaseCreationTracking(ctx context.Context, tracking *repositories.CaseCreationTracking) error {
	if r.failTrackingCreate != nil {
		if err := r.failTrackingCreate(tracking); err != nil {
			return err
		}
	}
	return r.IntegrationRepository.CreateCaseCreationTracking(ctx, tracking)
}

func (r *txProbeIntegrationRepo) UpdateCaseCreationTracking(ctx context.Context, id string, updates map[string]interface{}) error {
	if r.failTrackingUpdate != nil {
		if err := r.failTrackingUpdate(id, updates); err != nil {
			return err
		}
	}
	return r.IntegrationRepository.UpdateCaseCreationTracking(ctx, id, updates)
}

func (r *txProbeIntegrationRepo) UpdateApprovalCaseAssociation(ctx context.Context, approvalID string, caseResult *models.CaseCreationAssociation) error {
	if r.failAssociationUpdate != nil {
		if err := r.failAssociationUpdate(approvalID); err != nil {
			return err
		}
	}
	return r.IntegrationRepository.UpdateApprovalCaseAssociation(ctx, approvalID, caseResult)
}

func clientRepoForTx(db *gorm.DB) *clientTxRepo { return &clientTxRepo{db: db} }
func userRepoForTx(db *gorm.DB) *userTxRepo     { return &userTxRepo{db: db} }

type clientTxRepo struct{ db *gorm.DB }

func (r *clientTxRepo) Create(ctx context.Context, c *models.Client) error {
	return r.db.WithContext(ctx).Create(c).Error
}
func (r *clientTxRepo) FindByID(ctx context.Context, id uint) (*models.Client, error) {
	var c models.Client
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}
func (r *clientTxRepo) Update(ctx context.Context, c *models.Client) error {
	return r.db.WithContext(ctx).Save(c).Error
}
func (r *clientTxRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Client{}, id).Error
}

type userTxRepo struct{ db *gorm.DB }

func (r *userTxRepo) Create(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}
func (r *userTxRepo) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
func (r *userTxRepo) Update(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}
func (r *userTxRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// ---- tests ----------------------------------------------------------------

// Happy path: one transaction commits case + parties + tracking + approval
// linkage together; every write is durable and consistent afterwards.
func TestApprovalCaseTxHappyPath(t *testing.T) {
	f := seedApprovalCaseWorld(t)

	result, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "completed", result.Status)

	assert.True(t, trackingCompleted(f.db, t))
	var caseCount, partyCount int64
	require.NoError(t, f.db.Model(&models.Case{}).Count(&caseCount).Error)
	require.NoError(t, f.db.Model(&models.CaseParty{}).Count(&partyCount).Error)
	assert.EqualValues(t, 1, caseCount)
	assert.EqualValues(t, 2, partyCount, "client + opposing party from the reviewed snapshot must land")
	var approval models.ApprovalRequest
	require.NoError(t, f.db.First(&approval, "id = ?", "APR-TX-1").Error)
	assert.True(t, approval.CaseCreated)
	assert.NotEmpty(t, approval.CreatedCaseID)
}

// A tracking failure rolls the whole approval-case transaction back: no case,
// no parties, and the approval stays claimable for a retry.
func TestApprovalCaseTxTrackingFailureRollsBack(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	f.tracker.failTrackingCreate = func(tracking *repositories.CaseCreationTracking) error {
		return fmt.Errorf("injected tracking failure")
	}

	_, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.Error(t, err)

	var caseCount, partyCount int64
	require.NoError(t, f.db.Model(&models.Case{}).Count(&caseCount).Error)
	require.NoError(t, f.db.Model(&models.CaseParty{}).Count(&partyCount).Error)
	assert.EqualValues(t, 0, caseCount, "tracking failure must roll back the created case")
	assert.EqualValues(t, 0, partyCount, "tracking failure must roll back the party rows")
	var approval models.ApprovalRequest
	require.NoError(t, f.db.First(&approval, "id = ?", "APR-TX-1").Error)
	assert.False(t, approval.CaseCreated, "approval linkage must roll back with the transaction")
	assert.False(t, trackingCompleted(f.db, t))
}

// An approval linkage failure rolls back the case, the parties and the
// tracking completion in the same transaction.
func TestApprovalCaseTxAssociationFailureRollsBack(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	f.tracker.failAssociationUpdate = func(approvalID string) error {
		return fmt.Errorf("injected association failure")
	}

	_, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.Error(t, err)

	var caseCount, partyCount int64
	require.NoError(t, f.db.Model(&models.Case{}).Count(&caseCount).Error)
	require.NoError(t, f.db.Model(&models.CaseParty{}).Count(&partyCount).Error)
	assert.EqualValues(t, 0, caseCount)
	assert.EqualValues(t, 0, partyCount)
	assert.False(t, trackingCompleted(f.db, t))
}

// Retry after a rolled-back attempt creates exactly one case, and the
// completed fast path returns the same case without creating another.
func TestApprovalCaseTxRetryReusesSameCase(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	f.tracker.failTrackingCreate = func(tracking *repositories.CaseCreationTracking) error {
		return fmt.Errorf("injected tracking failure")
	}
	_, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.Error(t, err)
	f.tracker.failTrackingCreate = nil

	first, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err)
	firstID, parseErr := strconv.ParseUint(first.CaseID, 10, 64)
	require.NoError(t, parseErr)
	var partiesAfterFirst int64
	require.NoError(t, f.db.Model(&models.CaseParty{}).Where("case_id = ?", firstID).Count(&partiesAfterFirst).Error)
	assert.EqualValues(t, 2, partiesAfterFirst, "first successful attempt must materialize the reviewed parties")

	second, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err)
	assert.Equal(t, first.CaseID, second.CaseID, "retry and fast path must reuse the same case")

	var caseCount int64
	require.NoError(t, f.db.Model(&models.Case{}).Count(&caseCount).Error)
	assert.EqualValues(t, 1, caseCount)
}

// Two concurrent attempts for one approval must create at most one case.
func TestApprovalCaseTxConcurrentSingleCase(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	var wg sync.WaitGroup
	results := make([]*CaseCreationResult, 2)
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
		}(i)
	}
	wg.Wait()

	var caseCount int64
	require.NoError(t, f.db.Model(&models.Case{}).Count(&caseCount).Error)
	assert.EqualValues(t, 1, caseCount, "concurrent attempts must not create two cases")
	successCount := 0
	for i := range errs {
		if errs[i] == nil {
			successCount++
		}
	}
	assert.GreaterOrEqual(t, successCount, 1, "at least one attempt must succeed")
}

// A completed tracking row whose case lost its party rows must be
// reconciled (idempotent repair on the same case), never published as a
// fake completion and never re-created as a duplicate case.
func TestApprovalCaseTxCompletedMissingPartiesRepairs(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	first, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err)
	caseID, parseErr := strconv.ParseUint(first.CaseID, 10, 64)
	require.NoError(t, parseErr)
	require.NoError(t, f.db.Where("case_id = ?", caseID).Delete(&models.CaseParty{}).Error)

	second, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err, "the fast path must repair parties on the same case")
	assert.Equal(t, first.CaseID, second.CaseID, "repair must reuse the same case, never create another")
	var partyCount int64
	require.NoError(t, f.db.Model(&models.CaseParty{}).Where("case_id = ?", caseID).Count(&partyCount).Error)
	assert.EqualValues(t, 2, partyCount, "missing parties must be re-materialized from the reviewed snapshot")
	var caseCount int64
	require.NoError(t, f.db.Model(&models.Case{}).Count(&caseCount).Error)
	assert.EqualValues(t, 1, caseCount)
}

// A corrupted snapshot (no parties at all) must fail loudly on both the
// creation path and the completed fast path; neither may fake a completion.
func TestApprovalCaseTxCorruptedSnapshotRejected(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	require.NoError(t, f.db.Model(&models.ConflictCheckRecord{}).
		Where("check_id = ?", "CHK-TX-1").
		Update("search_parameters", models.JSON{}).Error)
	require.NoError(t, f.db.Model(&models.ConflictCheckRecord{}).
		Where("check_id = ?", "CHK-TX-1").
		Update("client_name", "").Error)

	_, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少当事人信息")
	var caseCount int64
	require.NoError(t, f.db.Model(&models.Case{}).Count(&caseCount).Error)
	assert.EqualValues(t, 0, caseCount, "no case may be created from a corrupted snapshot")
	// Corrupted snapshot must also poison the completed fast path.
	f2 := seedApprovalCaseWorld(t)
	first, err := f2.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err)
	require.NoError(t, f2.db.Model(&models.ConflictCheckRecord{}).
		Where("check_id = ?", "CHK-TX-1").
		Update("search_parameters", models.JSON{}).Error)
	require.NoError(t, f2.db.Model(&models.ConflictCheckRecord{}).
		Where("check_id = ?", "CHK-TX-1").
		Update("client_name", "").Error)
	_, err = f2.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.Error(t, err, "completed fast path must refuse to validate against a corrupted snapshot")
	assert.Contains(t, err.Error(), "缺少当事人信息")
	_ = first
}
