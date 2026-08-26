package health

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func TestConflictP0ReadinessCheck_QAExplicitlySkipsProductionEvidence(t *testing.T) {
	check := NewConflictP0ReadinessCheck(nil, false)

	result := check.Check(context.Background())

	require.NotNil(t, result)
	assert.Equal(t, ConflictP0ReadinessCheckName, result.Name)
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "QA技术就绪")
	assert.Contains(t, result.Message, "不代表生产可用")

	details, ok := result.Details.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, details["production_database_prerequisites_ready"])
	for _, gate := range requiredProductionEvidenceGates {
		assert.Equal(t, false, details["external_evidence_gates"].(map[string]bool)[gate])
	}
	assert.Equal(t, "完成并签署 PD-01 至 PD-07 决策物，登记书面签署凭证", details["next_actions"].(map[string]string)["G0"])
}

func TestConflictP0ReadinessCheck_ProductionMissingDatabaseFailsClosed(t *testing.T) {
	check := NewConflictP0ReadinessCheck(nil, true)

	result := check.Check(context.Background())

	require.NotNil(t, result)
	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "数据库连接未初始化")

	details, ok := result.Details.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, details["production_database_prerequisites_ready"])
	for _, gate := range requiredProductionEvidenceGates {
		assert.Equal(t, false, details["external_evidence_gates"].(map[string]bool)[gate])
	}
	assert.NotEmpty(t, details["next_actions"])
}

func TestConflictP0ReadinessCheck_ProductionPrerequisitesWithoutExternalEvidenceFailsClosed(t *testing.T) {
	db, sqlDB := setupConflictP0ProductionDatabase(t)
	seedCompleteConflictP0ProductionEvidence(t, db)

	result := NewConflictP0ReadinessCheck(sqlDB, true).Check(context.Background())

	require.NotNil(t, result)
	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "生产外部证据门禁未完整登记或复核")
	for _, gate := range requiredProductionEvidenceGates {
		assert.Contains(t, result.Message, gate)
	}
	details, ok := result.Details.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, details["production_database_prerequisites_ready"])
}

func TestConflictP0ReadinessCheck_ProductionCompleteExplicitEvidencePasses(t *testing.T) {
	db, sqlDB := setupConflictP0ProductionDatabase(t)
	seedCompleteConflictP0ProductionEvidence(t, db)
	seedCompleteProductionExternalEvidence(t, db)

	result := NewConflictP0ReadinessCheck(sqlDB, true).Check(context.Background())

	require.NotNil(t, result)
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "生产数据库前置条件和 G0-G7 外部证据登记复核通过")
	details, ok := result.Details.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, details["production_database_prerequisites_ready"])
}

func TestConflictP0ReadinessCheck_ProductionMissingGovernanceAuditFailsClosed(t *testing.T) {
	db, sqlDB := setupConflictP0ProductionDatabase(t)
	seedCompleteConflictP0ProductionEvidence(t, db)
	require.NoError(t, db.Exec("DELETE FROM compliance_audit_events WHERE event_type = 'CONFLICT_POLICY_APPROVED'").Error)

	result := NewConflictP0ReadinessCheck(sqlDB, true).Check(context.Background())

	require.NotNil(t, result)
	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "CONFLICT_POLICY_APPROVED")
	assert.Contains(t, result.Message, "G5")
}

func TestConflictP0ReadinessCheck_ProductionMissingOneExternalEvidenceGateFailsClosed(t *testing.T) {
	db, sqlDB := setupConflictP0ProductionDatabase(t)
	seedCompleteConflictP0ProductionEvidence(t, db)
	seedCompleteProductionExternalEvidence(t, db)
	require.NoError(t, db.Exec("DELETE FROM production_external_evidence WHERE gate = 'G6'").Error)

	result := NewConflictP0ReadinessCheck(sqlDB, true).Check(context.Background())

	require.NotNil(t, result)
	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "生产外部证据门禁未完整登记或复核")
	assert.Contains(t, result.Message, "G6")
	assert.NotContains(t, result.Message, "G6(缺失) G7")
}

func setupConflictP0ProductionDatabase(t *testing.T) (*gorm.DB, *sql.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "conflict-p0-production.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{}, &models.Client{}, &models.Case{}, &models.Entity{}, &models.EntityRelation{},
		&models.EntityNameHistory{}, &models.CaseParty{}, &models.ConflictSubjectVersion{},
		&models.ConflictSubjectIdentifier{}, &models.ConflictMatchEvidenceV2{}, &models.ConflictIndexBuildRun{},
		&models.ConflictSearchScope{}, &models.ComplianceAuditEvent{},
		&models.LawFirmCompliancePolicyProfile{}, &models.LawFirmCompliancePolicyPackage{}, &models.LawFirmCompliancePolicyEndorsement{},
		&models.ConflictOfficerAppointment{}, &models.CaseSubjectRevision{},
	))
	rawDB, err := db.DB()
	require.NoError(t, err)
	return db, rawDB
}

func seedCompleteConflictP0ProductionEvidence(t *testing.T, db *gorm.DB) {
	t.Helper()
	now := time.Now().UTC()
	sum := sha256.Sum256([]byte("production evidence fixture"))
	hash := hex.EncodeToString(sum[:])
	users := []models.User{
		{ID: 1, Username: "director", Name: "管理合伙人", Email: "director@example.com", Role: "director", Status: "active"},
		{ID: 2, Username: "compliance", Name: "合规负责人", Email: "compliance@example.com", Role: "compliance", Status: "active"},
		{ID: 3, Username: "officer", Name: "冲突核查人", Email: "officer@example.com", Role: "conflict_officer", Status: "active"},
		{ID: 4, Username: "lawyer", Name: "律师", Email: "lawyer@example.com", Role: "lawyer", Status: "active"},
	}
	for index := range users {
		users[index].CreatedAt = now
		users[index].UpdatedAt = now
		require.NoError(t, db.Create(&users[index]).Error)
	}

	packageID := "POLICY-PROD-001"
	policyPackage := models.LawFirmCompliancePolicyPackage{
		ID: packageID, PolicyVersion: "prod-v1", Jurisdiction: "中国", ApplicableRuleName: "利益冲突治理规则",
		ApplicableRuleVersion: "v1", ApplicableRuleAuthority: "律所管理委员会", ApplicableRuleReference: "rule://prod",
		DataSourcePolicyReference: "policy://sources", PrivacyBasisMatrixReference: "policy://privacy", RetentionPolicyReference: "policy://retention",
		WaiverPolicyReference: "policy://waiver", ControlledActionsReference: "policy://actions", ExternalReviewReference: "policy://external-review",
		EffectiveAt: now.Add(-time.Hour), NextReviewAt: now.AddDate(1, 0, 0), IntegrityHash: hash, CreatedBy: 1, CreatedAt: now,
	}
	require.NoError(t, db.Create(&policyPackage).Error)
	require.NoError(t, db.Create(&models.LawFirmCompliancePolicyEndorsement{
		ID: "policy-management-endorsement", PolicyPackageID: packageID, EndorsementType: services.PolicyEndorsementManagement,
		EndorsedBy: 1, EndorserRole: "director", PackageIntegrityHash: hash, CreatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&models.LawFirmCompliancePolicyEndorsement{
		ID: "policy-compliance-endorsement", PolicyPackageID: packageID, EndorsementType: services.PolicyEndorsementCompliance,
		EndorsedBy: 2, EndorserRole: "compliance", PackageIntegrityHash: hash, CreatedAt: now,
	}).Error)
	approvedAt := now
	effectiveAt := policyPackage.EffectiveAt
	nextReviewAt := policyPackage.NextReviewAt
	require.NoError(t, db.Create(&models.LawFirmCompliancePolicyProfile{
		ID: packageID, PolicyVersion: policyPackage.PolicyVersion, Status: "APPROVED", Jurisdiction: policyPackage.Jurisdiction,
		ApplicableRuleName: policyPackage.ApplicableRuleName, ApplicableRuleVersion: policyPackage.ApplicableRuleVersion,
		ApplicableRuleAuthority: policyPackage.ApplicableRuleAuthority, ApplicableRuleReference: policyPackage.ApplicableRuleReference,
		DataSourcePolicyReference: policyPackage.DataSourcePolicyReference, PrivacyBasisMatrixReference: policyPackage.PrivacyBasisMatrixReference,
		RetentionPolicyReference: policyPackage.RetentionPolicyReference, WaiverPolicyReference: policyPackage.WaiverPolicyReference,
		ControlledActionsReference: policyPackage.ControlledActionsReference, ExternalReviewReference: policyPackage.ExternalReviewReference,
		ManagementApprovedBy: 1, ComplianceApprovedBy: 2, ApprovedAt: &approvedAt, EffectiveAt: &effectiveAt,
		NextReviewAt: &nextReviewAt, IntegrityHash: hash, CreatedAt: now, UpdatedAt: now,
	}).Error)

	require.NoError(t, db.Create(&models.ConflictOfficerAppointment{
		ID: "officer-appointment", OfficerID: 3, DeputyID: ptrUint(2), AppointedBy: 1,
		EffectiveFrom: now.Add(-time.Hour), EffectiveTo: now.AddDate(1, 0, 0),
		RecusalDeclaration: "生产任命回避声明", ExternalMechanismReference: "evidence://appointment", CreatedAt: now,
	}).Error)

	client := models.Client{ID: 1, Name: "生产证据客户", Type: "企业", Email: "client@example.com", Status: "active", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&client).Error)
	matter := models.Case{ID: 1, CaseNumber: "CASE-PROD-1", Title: "生产证据案件", ClientID: 1, LawyerID: 4, CaseType: "commercial", Status: "active", CreatedBy: "4", SubjectState: models.SubjectStateEffective, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&matter).Error)
	entity := models.Entity{ID: 1, EntityType: models.EntityTypeLegalPerson, Name: "生产证据主体", IdentityType: models.IdentityTypeBusinessLicense, Status: models.EntityStatusActive, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&entity).Error)
	require.NoError(t, db.Create(&models.CaseParty{CaseID: 1, EntityID: 1, Role: models.PartyRolePlaintiff, PartyType: models.PartyTypeClient, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.EntityRelation{SourceEntityID: 1, TargetEntityID: 1, RelationType: models.RelationTypeParentCompany, IsActive: true, DataSource: "evidence://relation", CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.EntityNameHistory{EntityID: 1, OldName: "生产证据主体旧名", NewName: entity.Name, ChangeDate: now, CreatedAt: now, UpdatedAt: now}).Error)

	reconciler, ok := repositories.NewConflictRepository(db, nil).(repositories.ConflictP0SubjectIndexReconciler)
	require.True(t, ok)
	runs, err := reconciler.ReconcileConflictSubjectIndex(context.Background(), 3, "evidence://production-index", true)
	require.NoError(t, err)
	require.Len(t, runs, len(requiredConflictScopeTypes))
	for _, run := range runs {
		qualityOwnerID := uint(3)
		scopeID := "prod-scope-" + run.ScopeType
		coveredFrom := now.AddDate(-10, 0, 0)
		coveredTo := now
		lastSync := now
		reviewedAt := now
		scope := models.ConflictSearchScope{
			ID: scopeID, ScopeType: run.ScopeType, Status: services.ConflictScopeActive, CoverageStatus: services.ConflictCoverageComplete,
			SourceVersion: run.SourceVersion, EvidenceReference: run.EvidenceReference, CoveredFrom: &coveredFrom, CoveredTo: &coveredTo,
			MissingSources: "[]", IndexRunID: run.ID, ApprovedBy: &qualityOwnerID, ApprovedAt: &reviewedAt, SourceOfTruth: true,
			SyncMode: "BATCH", MaxSyncLagMinutes: 1440, LastSuccessfulSyncAt: &lastSync, MinimumFieldCoverageBPS: 10000,
			MeasuredFieldCoverageBPS: 10000, MaximumDuplicateRateBPS: 0, MeasuredDuplicateRateBPS: 0,
			QualityOwnerID: &qualityOwnerID, QualityReviewedAt: &reviewedAt, MaxQualityReviewAgeDays: 365,
			FailureAlertReference: "evidence://sync-alert", CorrectionProcedureReference: "evidence://correction", CreatedAt: now, UpdatedAt: now,
		}
		require.NoError(t, db.Create(&scope).Error)
		require.NoError(t, createGovernanceAudit(db, "CONFLICT_SCOPE_UPDATED", scopeID, scope))
	}
	require.NoError(t, createGovernanceAudit(db, "CONFLICT_POLICY_PACKAGE_CREATED", packageID, policyPackage))
	require.NoError(t, createGovernanceAudit(db, "CONFLICT_POLICY_ENDORSED", packageID, policyPackage))
	require.NoError(t, createGovernanceAudit(db, "CONFLICT_POLICY_APPROVED", packageID, policyPackage))
	require.NoError(t, createGovernanceAudit(db, "CONFLICT_OFFICER_APPOINTED", "officer-appointment", nil))
}

func createGovernanceAudit(db *gorm.DB, eventType, objectID string, payload interface{}) error {
	if payload == nil {
		payload = map[string]string{"objectID": objectID}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	return db.Create(&models.ComplianceAuditEvent{
		ID: eventType + "-" + objectID, ActorID: ptrUint(1), ActorRole: "director", EventType: eventType,
		ObjectType: "CONFLICT_P0_PRODUCTION_EVIDENCE", ObjectID: objectID, ToState: "ACTIVE",
		Payload: string(raw), IntegrityHash: hex.EncodeToString(sum[:]), CreatedAt: time.Now().UTC(),
	}).Error
}

func seedCompleteProductionExternalEvidence(t *testing.T, db *gorm.DB) {
	t.Helper()
	rawDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, ensureProductionExternalEvidenceTable(context.Background(), rawDB))
	now := time.Now().UTC()
	for _, gate := range requiredProductionEvidenceGates {
		payload := gate + ":approved:test-registration"
		sum := sha256.Sum256([]byte(payload))
		reviewer := "测试复核人"
		if gate == "G7" {
			reviewer = "运维负责人|合规负责人"
		}
		_, err := rawDB.Exec(`
			INSERT INTO production_external_evidence
				(id, gate, evidence_reference, reviewed_by, reviewer_role, review_result, reviewed_at, integrity_hash, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 'PASSED', ?, ?, ?, ?)
		`, gate[1], gate, "evidence://"+gate, reviewer, "test-reviewer", now.Format(time.RFC3339Nano), hex.EncodeToString(sum[:]), now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
		require.NoError(t, err)
	}
}

func TestConflictP0ReadinessCheck_ProductionG7RequiresDistinctFinalReviewers(t *testing.T) {
	db, sqlDB := setupConflictP0ProductionDatabase(t)
	seedCompleteConflictP0ProductionEvidence(t, db)
	seedCompleteProductionExternalEvidence(t, db)
	require.NoError(t, db.Exec("UPDATE production_external_evidence SET reviewed_by = '同一复核人' WHERE gate = 'G7'").Error)

	result := NewConflictP0ReadinessCheck(sqlDB, true).Check(context.Background())

	require.NotNil(t, result)
	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "G7最终复核记录未由两名不同责任人共同签署")
}

func ptrUint(value uint) *uint { return &value }
