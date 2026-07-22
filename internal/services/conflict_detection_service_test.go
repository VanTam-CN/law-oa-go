package services

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func TestBuildCheckStatisticsDoesNotClaimCompleteHistoryWithoutCoverage(t *testing.T) {
	service := &conflictDetectionService{}

	statistics := service.buildCheckStatistics(context.Background(), &models.ConflictCheckRequest{}, nil)

	require.NotNil(t, statistics)
	assert.Equal(t, "系统已登记历史（覆盖完整性待确认，未登记档案需人工核查）", statistics.TimeRange)
}

func TestGetConflictReviewReturnsEmptyWithoutRecordNotFoundError(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.AutoMigrate(&models.ConflictReview{}))
	service := &conflictDetectionService{caseRepo: repositories.NewCaseRepository(db)}

	review, err := service.GetConflictReview(context.Background(), "missing-review")

	require.NoError(t, err)
	assert.Nil(t, review)
}

func TestConflictDetectionPartyNameMatchNormalizesCompanySuffix(t *testing.T) {
	service := &conflictDetectionService{}

	// 规范化后完全相等（去公司后缀）→ Exact → true
	// 注意：旧实现把 "示例科技" 也当命中（contains），但那是短子串误报，已收紧
	if !service.isPartyNameMatch("上海示例科技有限公司", "上海示例科技") {
		t.Fatal("expected exact normalized name to match")
	}
}

func TestConflictDetectionPartyNameMatchRejectsUnrelatedNames(t *testing.T) {
	service := &conflictDetectionService{}

	if service.isPartyNameMatch("上海示例科技有限公司", "北京无关贸易有限公司") {
		t.Fatal("expected unrelated names not to match")
	}
}

// Task 7 Step 1 — RED 边界测试：短子串不得视为完全匹配
//
// 修复前：isPartyNameMatch 使用 strings.Contains，"华" 会匹配 "华为技术有限公司"
// → 上层 isPartyNameMatch() 命中后升级为 CRITICAL → 错误拒绝接案。
// 修复后：只有规范化后完全相等才返回 true；候选匹配由 classifyPartyMatch 单独处理。
func TestPartyNameMatch_DoesNotPromoteShortSubstringToExact(t *testing.T) {
	service := &conflictDetectionService{}

	cases := []struct {
		name1 string
		name2 string
	}{
		{"华为技术有限公司", "华"},
		{"上海示例科技有限公司", "示例"},
		{"阿里巴巴（中国）网络技术有限公司", "阿里"},
		{"北京甲科技有限公司", "上海甲贸易有限公司"},
	}
	for _, c := range cases {
		if service.isPartyNameMatch(c.name1, c.name2) {
			t.Fatalf("短子串/相近名 (%q vs %q) 不应被视为完全匹配", c.name1, c.name2)
		}
	}
}

// TestPartyNameMatch_RejectsEmptyAndSuffixOnly
// 空值或仅剩公司类型词不构成有效匹配。
func TestPartyNameMatch_RejectsEmptyAndSuffixOnly(t *testing.T) {
	service := &conflictDetectionService{}
	cases := []struct {
		name1 string
		name2 string
	}{
		{"", ""},
		{"有限公司", "公司"},
		{"", "上海示例科技有限公司"},
	}
	for _, c := range cases {
		if service.isPartyNameMatch(c.name1, c.name2) {
			t.Fatalf("空值/纯后缀 (%q vs %q) 不应命中", c.name1, c.name2)
		}
	}
}

// TestClassifyPartyMatch
// 三态分类：Exact / Candidate / NoMatch。
//   - Exact：规范化后完全相等 → 可直接判 CRITICAL
//   - Candidate：单向/双向包含、简称 → 只能作为候选，最高 HIGH
//   - NoMatch：完全无关或无效输入
func TestClassifyPartyMatch(t *testing.T) {
	service := &conflictDetectionService{}

	assert.Equal(t, PartyExactNormalizedMatch,
		service.classifyPartyMatch("上海示例科技有限公司", "上海示例科技"))
	assert.Equal(t, PartyCandidateMatch,
		service.classifyPartyMatch("上海示例科技有限公司", "示例科技"))
	assert.Equal(t, PartyCandidateMatch,
		service.classifyPartyMatch("华为技术有限公司", "华为"))
	assert.Equal(t, PartyNoMatch,
		service.classifyPartyMatch("北京甲科技有限公司", "上海甲贸易有限公司"))
	assert.Equal(t, PartyNoMatch,
		service.classifyPartyMatch("", "任何"))
	assert.Equal(t, PartyNoMatch,
		service.classifyPartyMatch("有限公司", "公司"))
}

// setupSQLiteForConflictDetection 构造一个独立文件的 SQLite + 冲突检测相关表。
// 文件模式避开 shared cache 下 idx_status 之类的索引名冲突。
func setupSQLiteForConflictDetection(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", "01234567890123456789012345678901")
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("svc_conflict_%d_%d.db",
		time.Now().UnixNano(), rand.Int63()))
	dsn := fmt.Sprintf("file:%s?_busy_timeout=30000&_journal_mode=WAL", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	if err := db.AutoMigrate(
		&models.Case{},
		&models.Client{},
		&models.User{},
		&models.ClientRelation{},
	); err != nil && err.Error() != "table cases already exists" {
		require.NoError(t, err)
	}
	return db
}

// TestCheckOpponentConflicts_PropagatesDBError
// 验证 checkOpponentConflicts 在底层查询失败时返回 error，而不是静默吞掉返回空切片。
//
// 修复前：Rows()、Scan 都用 `continue` 吞错误，调用方拿到空切片 → 落库为"无冲突"成功记录。
// 修复后：任何查询/扫描/迭代错误必须包装上下文向上传播，调用方据此判定检查失败。
func TestCheckOpponentConflicts_PropagatesDBError(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)

	// DROP clients 表让 JOIN 失败——模拟数据库故障
	require.NoError(t, db.Migrator().DropTable("clients"))

	svc := &conflictDetectionService{
		caseRepo: repositories.NewCaseRepository(db),
	}

	request := &models.ConflictCheckRequest{
		ClientID:     "1",
		UserID:       "1",
		OtherParties: []string{"acme corp"},
	}

	conflicts, err := svc.checkOpponentConflicts(context.Background(), request, time.Time{})

	// 关键断言：故障必须传播——禁止 (空切片, nil) 的"无冲突"假象
	require.Error(t, err, "数据库故障必须传播 error，不能返回 (nil, nil) 制造无冲突假象")
	assert.Nil(t, conflicts, "出错时不应返回部分结果，避免调用方误用")
}

// TestCheckOpponentConflicts_DoesNotScopeFirmSearchToLawyerID
// 对方主体检索是全所级检查，不应因为承办律师 ID 不可解析而跳过。
func TestCheckOpponentConflicts_DoesNotScopeFirmSearchToLawyerID(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)

	svc := &conflictDetectionService{
		caseRepo: repositories.NewCaseRepository(db),
	}

	request := &models.ConflictCheckRequest{
		ClientID:     "1",
		UserID:       "not-a-number",
		OtherParties: []string{"acme corp"},
	}

	conflicts, err := svc.checkOpponentConflicts(context.Background(), request, time.Time{})

	require.NoError(t, err)
	assert.Empty(t, conflicts)
}

func TestStructuredPartyConflictsMatchIdentityAcrossFirmArchive(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.AutoMigrate(&models.Entity{}, &models.EntityNameHistory{}, &models.EntityRelation{}, &models.CaseParty{}))
	require.NoError(t, db.Create(&models.Client{ID: 1, Name: "当前委托人", Type: "企业", Email: "current-client@example.test"}).Error)
	require.NoError(t, db.Create(&models.Client{ID: 2, Name: "历史客户", Type: "企业", Email: "archive-client@example.test"}).Error)
	require.NoError(t, db.Create(&models.User{ID: 7, Username: "archive-lawyer", Name: "历史承办律师", Email: "archive@example.test", Password: "hash", Role: "lawyer"}).Error)
	require.NoError(t, db.Create(&models.Case{ID: 101, CaseNumber: "CASE-ARCHIVE-101", Title: "历史事项", ClientID: 2, LawyerID: 7, CaseType: "商事", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.Entity{
		ID:             201,
		EntityType:     models.EntityTypeLegalPerson,
		Name:           "北辰供应链有限公司",
		Alias:          "北辰供应链（旧名）",
		IdentityType:   models.IdentityTypeSocialCredit,
		IdentityNumber: "91310000ARCHIVE01",
		Status:         models.EntityStatusActive,
	}).Error)
	require.NoError(t, db.Create(&models.CaseParty{CaseID: 101, EntityID: 201, Role: models.PartyRoleDefendant, PartyType: models.PartyTypeOpposing}).Error)

	service := &conflictDetectionService{caseRepo: repositories.NewCaseRepository(db)}
	conflicts, err := service.checkStructuredPartyConflicts(context.Background(), &models.ConflictCheckRequest{
		ClientID: "1",
		Parties: []models.ConflictPartyInfo{{
			Name:        "完全不同的录入名",
			Role:        "OPPOSING_PARTY",
			EntityType:  "COMPANY",
			Identifiers: map[string]string{"unified_social_credit_code": "91310000ARCHIVE01"},
		}},
		OtherParties: []string{"完全不同的录入名"},
	}, time.Time{})

	require.NoError(t, err)
	require.Len(t, conflicts, 1)
	assert.Equal(t, "STRUCTURED_IDENTITY_EXACT", conflicts[0].RuleCode)
	assert.Equal(t, "EXACT", conflicts[0].MatchType)
	assert.Equal(t, "CASE-ARCHIVE-101", conflicts[0].CaseNo)
}

func TestStructuredPartyConflictsSearchFormerNamesAndRelatedEntities(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.AutoMigrate(&models.Entity{}, &models.EntityNameHistory{}, &models.EntityRelation{}, &models.CaseParty{}))
	require.NoError(t, db.Create(&models.Client{ID: 1, Name: "当前委托人", Type: "企业", Email: "current-client@example.test"}).Error)
	require.NoError(t, db.Create(&models.Client{ID: 2, Name: "历史客户", Type: "企业", Email: "archive-client@example.test"}).Error)
	require.NoError(t, db.Create(&models.User{ID: 7, Username: "archive-lawyer", Name: "历史承办律师", Email: "archive@example.test", Password: "hash", Role: "lawyer"}).Error)
	require.NoError(t, db.Create(&models.Case{ID: 102, CaseNumber: "CASE-ARCHIVE-102", Title: "历史关联事项", ClientID: 2, LawyerID: 7, CaseType: "商事", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.Entity{
		ID:           202,
		EntityType:   models.EntityTypeLegalPerson,
		Name:         "现用主体有限公司",
		IdentityType: models.IdentityTypeSocialCredit,
		Status:       models.EntityStatusActive,
	}).Error)
	require.NoError(t, db.Create(&models.Entity{
		ID:           203,
		EntityType:   models.EntityTypeLegalPerson,
		Name:         "关联控股有限公司",
		IdentityType: models.IdentityTypeSocialCredit,
		Status:       models.EntityStatusActive,
	}).Error)
	require.NoError(t, db.Create(&models.EntityNameHistory{
		EntityID: 202, OldName: "旧称主体有限公司", NewName: "现用主体有限公司",
		ChangeDate: time.Now().AddDate(-1, 0, 0), ChangeReason: "企业更名",
	}).Error)
	require.NoError(t, db.Create(&models.EntityRelation{
		SourceEntityID: 202, TargetEntityID: 203, RelationType: models.RelationTypeParentCompany,
		IsActive: true, DataSource: "law-firm-registry",
	}).Error)
	require.NoError(t, db.Create(&models.CaseParty{CaseID: 102, EntityID: 202, Role: models.PartyRoleDefendant, PartyType: models.PartyTypeOpposing}).Error)

	service := &conflictDetectionService{caseRepo: repositories.NewCaseRepository(db)}
	conflicts, err := service.checkStructuredPartyConflicts(context.Background(), &models.ConflictCheckRequest{
		ClientID:   "1",
		ClientName: "当前委托人",
		Parties: []models.ConflictPartyInfo{
			{Name: "旧称主体有限公司", Role: "OPPOSING_PARTY", EntityType: "COMPANY"},
			{Name: "关联控股有限公司", Role: "OPPOSING_PARTY", EntityType: "COMPANY"},
		},
	}, time.Time{})

	require.NoError(t, err)
	require.Len(t, conflicts, 2)

	byRule := make(map[string]*models.ConflictCase, len(conflicts))
	for _, conflict := range conflicts {
		byRule[conflict.RuleCode] = conflict
	}
	formerName := byRule["FORMER_NAME_CANDIDATE_REVIEW"]
	require.NotNil(t, formerName)
	assert.Equal(t, "ENTITY_NAME_HISTORY", formerName.Evidence[0].SourceType)
	assert.Equal(t, "CASE-ARCHIVE-102", formerName.CaseNo)

	relatedEntity := byRule["RELATED_ENTITY_ADVERSE_REVIEW"]
	require.NotNil(t, relatedEntity)
	assert.Equal(t, "CLIENT_RELATION", relatedEntity.Evidence[0].SourceType)
	assert.Equal(t, "RELATION", relatedEntity.MatchType)
}

func TestNormalizeConflictSubjectsCanonicalizesPunctuationSuffixAndAliases(t *testing.T) {
	request := &models.ConflictCheckRequest{
		ClientName: " 上海示例科技（集团）有限公司 ", ClientType: "COMPANY",
		ClientIdentifiers: map[string]string{"unified_social_credit_code": " 91360000TEST "},
		ClientAliases:     []string{"上海示例科技集团", "示例科技"},
		Parties:           []models.ConflictPartyInfo{{Name: "北辰数字服务有限公司", Role: "OPPOSING_PARTY", EntityType: "COMPANY"}},
	}
	subjects := normalizeConflictSubjects(request)
	require.Len(t, subjects, 2)
	assert.Equal(t, "上海示例科技(集团)", subjects[0].NormalizedName)
	assert.Equal(t, "91360000TEST", subjects[0].Identifiers["unified_social_credit_code"])
	assert.Contains(t, subjects[0].Aliases, "示例科技")
}

func TestNormalizeConflictSubjectsDeduplicatesByAuthoritativeRole(t *testing.T) {
	request := &models.ConflictCheckRequest{
		ClientName:    "上海星河科技有限公司",
		ClientType:    "COMPANY",
		ClientAliases: []string{"STAR-RIVER Tech"},
		Parties: []models.ConflictPartyInfo{
			{Name: "上海星河科技公司", Role: "OPPOSING_PARTY", EntityType: "COMPANY"},
			{Name: "ACME.Legal", Role: "RELATED_PARTY", EntityType: "COMPANY"},
			{Name: "acme-legal", Role: "OPPOSING_PARTY", EntityType: "COMPANY"},
			{Name: "共同投资人", Role: "待确认", EntityType: ""},
			{Name: "独立相关方", Role: "", EntityType: ""},
		},
		OtherParties: []string{
			"上海星河科技有限公司",
			"ACME Legal",
			"共同投资人",
			"独立对方",
		},
	}

	subjects := normalizeConflictSubjects(request)
	require.Len(t, subjects, 5)

	byName := make(map[string]models.ConflictNormalizedSubject, len(subjects))
	for _, subject := range subjects {
		assert.NotEmpty(t, subject.Role)
		assert.NotEmpty(t, subject.EntityType)
		assert.NotEqual(t, "待确认", subject.Role)
		byName[subject.NormalizedName] = subject
	}

	assert.Equal(t, "CLIENT", byName["上海星河科技"].Role, "客户与其他来源重复时必须保留客户角色")
	assert.Equal(t, "OPPOSING_PARTY", byName["acmelegal"].Role, "同一权威来源内对方角色优先于相关方")
	assert.Equal(t, "RELATED_PARTY", byName["共同投资人"].Role, "OtherParties 不得覆盖显式相关方角色")
	assert.Equal(t, "ANY", byName["共同投资人"].EntityType)
	assert.Equal(t, "RELATED_PARTY", byName["独立相关方"].Role)
	assert.Equal(t, "ANY", byName["独立相关方"].EntityType)
	assert.Equal(t, "OPPOSING_PARTY", byName["独立对方"].Role, "真正独立的 OtherParties 主体必须保留")
}

func TestNormalizeConflictSubjectsDeduplicatesClientAlias(t *testing.T) {
	request := &models.ConflictCheckRequest{
		ClientName:        "上海星河科技有限公司",
		ClientType:        "COMPANY",
		ClientIdentifiers: map[string]string{" unified_social_credit_code ": " 91310000ABC "},
		ClientAliases:     []string{"STAR-RIVER Tech"},
		OtherParties:      []string{"star river tech"},
	}

	subjects := normalizeConflictSubjects(request)
	require.Len(t, subjects, 1)
	assert.Equal(t, "CLIENT", subjects[0].Role)
	assert.Equal(t, "91310000ABC", subjects[0].Identifiers["unified_social_credit_code"])
}

func TestP0PolicyTreatsNameOnlyExactMatchAsReviewRequired(t *testing.T) {
	svc := &conflictDetectionService{caseRepo: repositories.NewCaseRepository(setupSQLiteForConflictDetection(t))}
	request := &models.ConflictCheckRequest{}
	candidate := &models.ConflictCase{
		ID: "candidate", CaseID: "candidate", ConflictType: "名称相似待核实", RiskLevel: "HIGH",
		MatchType: "CANDIDATE", RuleCode: "SUBJECT_CANDIDATE_REVIEW", OpposingParties: []string{"示例科技"},
	}
	decision, assessment := svc.applyP0ConflictPolicy(context.Background(), request, []*models.ConflictCase{candidate})
	assert.Equal(t, "REVIEW_REQUIRED", decision.Status)
	assert.Equal(t, "MEDIUM", assessment.OverallRisk)
	assert.True(t, candidate.RequiresManualReview)

	exact := &models.ConflictCase{
		ID: "exact", CaseID: "exact", ConflictType: "对方当事人直接冲突", RiskLevel: "MEDIUM",
		MatchType: "EXACT", RuleCode: "DIRECT_ADVERSE_CURRENT_CLIENT", OpposingParties: []string{"上海示例科技有限公司"},
	}
	decision, assessment = svc.applyP0ConflictPolicy(context.Background(), request, []*models.ConflictCase{exact})
	assert.Equal(t, "REVIEW_REQUIRED", decision.Status)
	assert.Equal(t, "HIGH", assessment.OverallRisk)
	assert.True(t, exact.RequiresManualReview)
}

func TestConflictReviewRejectsTaskOwner(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.AutoMigrate(&models.ConflictCheckRecord{}, &models.ConflictReview{}))
	conflictRepo := repositories.NewConflictRepository(db, nil)
	require.NoError(t, conflictRepo.SaveCheckRecord(context.Background(), &models.ConflictCheckRecord{
		CheckID: "self-review", UserID: 7,
		CheckResult: models.JSON{"checkId": "self-review", "conflictCases": []interface{}{}},
	}))
	service := &conflictDetectionService{
		caseRepo:     repositories.NewCaseRepository(db),
		conflictRepo: conflictRepo,
	}
	_, err := service.ReviewConflict(context.Background(), "self-review", "no_conflict", "本人不得自行复核。", 7, "申请律师", nil)
	require.ErrorContains(t, err, "不得自行复核")
}

func TestConflictReviewRejectsDuplicateDecision(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.AutoMigrate(
		&models.ConflictCheckRecord{}, &models.ConflictReview{},
		&models.ConflictReviewerAssignment{}, &models.ComplianceAuditEvent{},
	))
	now := time.Now()
	require.NoError(t, db.Create(&models.User{
		ID: 7, Username: "applicant-7", Name: "申请律师", Email: "applicant-7@example.test", Role: "lawyer", Status: "active",
	}).Error)
	require.NoError(t, db.Create(&models.User{
		ID: 8, Username: "reviewer-8", Name: "独立核查人", Email: "reviewer-8@example.test", Role: "conflict_officer", Status: "active",
	}).Error)
	require.NoError(t, db.Create(&models.Case{
		ID: 100, CaseNumber: "CASE-100", Title: "重复复核测试", LawyerID: 7, CreatedBy: "7", SubjectState: models.SubjectStateEffective,
	}).Error)
	require.NoError(t, db.Create(&models.ConflictCheckRecord{
		CheckID: "duplicate-review", UserID: 7, CheckStatus: "COMPLETED",
		SearchParameters: models.JSON{"subjectCaseId": "100"},
		CheckResult:      models.JSON{"decision": map[string]interface{}{"coverageStatus": "COMPLETE"}, "conflictCases": []interface{}{}},
	}).Error)
	require.NoError(t, db.Create(&models.ConflictReviewerAssignment{
		ID: "assignment-duplicate", CheckID: "duplicate-review", ReviewerID: 8, AssignedBy: 8,
		Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true,
		EffectiveFrom: &now, CreatedAt: now, UpdatedAt: now,
	}).Error)

	service := &conflictDetectionService{
		caseRepo:     repositories.NewCaseRepository(db),
		conflictRepo: repositories.NewConflictRepository(db, nil),
	}
	_, err := service.ReviewConflict(context.Background(), "duplicate-review", "confirmed_conflict", "首轮独立复核", 8, "独立核查人", nil)
	require.NoError(t, err)
	_, err = service.ReviewConflict(context.Background(), "duplicate-review", "confirmed_conflict", "重复提交", 8, "独立核查人", nil)
	require.ErrorContains(t, err, "已有复核结论")

	var count int64
	require.NoError(t, db.Model(&models.ConflictReview{}).Where("check_id = ?", "duplicate-review").Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestConflictReviewRejectsLimitedCoverage(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.AutoMigrate(
		&models.ConflictCheckRecord{}, &models.ConflictReview{},
		&models.ConflictReviewerAssignment{}, &models.ComplianceAuditEvent{},
	))
	now := time.Now()
	require.NoError(t, db.Create(&models.User{
		ID: 7, Username: "limited-applicant", Name: "申请律师", Email: "limited-applicant@example.test", Role: "lawyer", Status: "active",
	}).Error)
	require.NoError(t, db.Create(&models.User{
		ID: 8, Username: "limited-reviewer", Name: "独立核查人", Email: "limited-reviewer@example.test", Role: "conflict_officer", Status: "active",
	}).Error)
	require.NoError(t, db.Create(&models.Case{
		ID: 101, CaseNumber: "CASE-LIMITED", Title: "范围受限复核测试", LawyerID: 7, CreatedBy: "7", SubjectState: models.SubjectStateEffective,
	}).Error)
	require.NoError(t, db.Create(&models.ConflictCheckRecord{
		CheckID: "limited-review", UserID: 7, CheckStatus: "COMPLETED",
		SearchParameters: models.JSON{"subjectCaseId": "101"},
		CheckResult:      models.JSON{"decision": map[string]interface{}{"coverageStatus": "COVERAGE_LIMITED"}, "conflictCases": []interface{}{}},
	}).Error)
	require.NoError(t, db.Create(&models.ConflictReviewerAssignment{
		ID: "assignment-limited", CheckID: "limited-review", ReviewerID: 8, AssignedBy: 8,
		Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true,
		EffectiveFrom: &now, CreatedAt: now, UpdatedAt: now,
	}).Error)

	service := &conflictDetectionService{
		caseRepo:     repositories.NewCaseRepository(db),
		conflictRepo: repositories.NewConflictRepository(db, nil),
	}
	_, err := service.ReviewConflict(context.Background(), "limited-review", "no_conflict", "范围受限不得形成无冲突结论", 8, "独立核查人", nil)
	if code := reviewerErrorCode(t, err); code != "COVERAGE_LIMITED" {
		t.Fatalf("expected coverage gate, got %s", code)
	}

	var count int64
	require.NoError(t, db.Model(&models.ConflictReview{}).Where("check_id = ?", "limited-review").Count(&count).Error)
	require.Zero(t, count)
}

func TestConflictEvidenceHashIsStableAcrossCaseOrder(t *testing.T) {
	first := &models.ConflictCase{Evidence: []models.ConflictEvidence{{RuleCode: "B", MatchType: "TEXT", Summary: "b"}}}
	second := &models.ConflictCase{Evidence: []models.ConflictEvidence{{RuleCode: "A", MatchType: "EXACT", Summary: "a"}}}
	left := conflictEvidenceHash(&models.ConflictCheckResponse{ConflictCases: []*models.ConflictCase{first, second}})
	right := conflictEvidenceHash(&models.ConflictCheckResponse{ConflictCases: []*models.ConflictCase{second, first}})
	assert.Equal(t, left, right)
	assert.Len(t, left, 64)
}

func TestEthicalWallEvidenceIsRestrictedUnlessActorIsWhitelisted(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS case_ethical_wall_whitelist (case_id integer NOT NULL, user_id integer NOT NULL)`).Error)
	require.NoError(t, db.Create(&models.Case{ID: 99, CaseNumber: "SECRET-99", Title: "受限并购事项", EthicalWallEnabled: true}).Error)

	newConflict := func() *models.ConflictCase {
		return &models.ConflictCase{
			ID: "restricted", CaseID: "99", CaseNo: "SECRET-99", CaseName: "受限并购事项",
			Description: "包含受限客户及交易细节",
			Evidence:    []models.ConflictEvidence{{SourceCaseID: "99", SourceCaseNumber: "SECRET-99", SourceCaseName: "受限并购事项", MatchedEntity: "受限客户", Summary: "受限证据"}},
		}
	}

	svc := &conflictDetectionService{caseRepo: repositories.NewCaseRepository(db)}
	restricted := newConflict()
	assert.True(t, svc.redactEthicalWallEvidence(context.Background(), 7, restricted))
	assert.True(t, restricted.Restricted)
	assert.True(t, restricted.Evidence[0].Restricted)
	assert.Equal(t, "SECRET-99", restricted.CaseNo, "审计记录必须保留原始证据，供独立核查人复核")
	assert.Equal(t, "受限客户", restricted.Evidence[0].MatchedEntity)

	require.NoError(t, db.Exec(`INSERT INTO case_ethical_wall_whitelist (case_id, user_id) VALUES (?, ?)`, 99, 7).Error)
	allowed := newConflict()
	assert.False(t, svc.redactEthicalWallEvidence(context.Background(), 7, allowed))
	assert.Equal(t, "SECRET-99", allowed.CaseNo)
}

func TestClientRelationConflictRequiresAdversePartyMatch(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.Create(&models.User{ID: 7, Username: "lawyer7", Name: "示例律师", Email: "lawyer7@example.test", Role: "lawyer", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.Client{ID: 1, Name: "当前客户", Type: "企业", Email: "current@example.test"}).Error)
	require.NoError(t, db.Create(&models.Client{ID: 2, Name: "关联控股公司", Type: "企业", Email: "related@example.test"}).Error)
	require.NoError(t, db.Create(&models.Case{ID: 101, CaseNumber: "REL-101", Title: "关联公司历史事项", ClientID: 2, LawyerID: 7, CaseType: "commercial", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.ClientRelation{ID: "relation-1", ClientID: "1", RelatedClientID: "2", RelationType: "CONTROLLER", RelationDetail: "同一控制人", Active: true}).Error)

	svc := &conflictDetectionService{
		caseRepo:     repositories.NewCaseRepository(db),
		conflictRepo: repositories.NewConflictRepository(db, nil),
	}
	request := &models.ConflictCheckRequest{ClientID: "1", UserID: "7", OtherParties: []string{"完全无关公司"}}
	assert.Empty(t, svc.checkClientRelationConflicts(context.Background(), request, time.Time{}), "仅存在客户关系不能自动构成冲突")

	request.OtherParties = []string{"关联控股公司"}
	conflicts := svc.checkClientRelationConflicts(context.Background(), request, time.Time{})
	require.Len(t, conflicts, 1)
	assert.Equal(t, "RELATED_ENTITY_ADVERSE_REVIEW", conflicts[0].RuleCode)
	assert.Equal(t, "RELATION", conflicts[0].MatchType)
	assert.True(t, conflicts[0].RequiresManualReview)
	require.Len(t, conflicts[0].Evidence, 1)
	assert.Equal(t, "CLIENT_RELATION", conflicts[0].Evidence[0].SourceType)
}

func TestWaiverWorkflowPersistsWithActualRepositories(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.AutoMigrate(
		&models.ConflictCheckRecord{},
		&models.WaiverApplication{},
		&models.WaiverApprovalRecord{},
		&models.WaiverSignature{},
		&models.WaiverMonitoringRecord{},
	))
	conflictRepo := repositories.NewConflictRepository(db, nil)
	waiverRepo := repositories.NewEnhancedConflictRepository(db)
	require.NoError(t, conflictRepo.SaveCheckRecord(context.Background(), &models.ConflictCheckRecord{
		CheckID: "actual-waiver-check", ClientID: "1", UserID: 7, HasConflict: true,
		CheckResult: models.JSON{"decision": map[string]interface{}{"status": "BLOCKED"}},
	}))
	service := NewWaiverWorkflowService(waiverRepo, conflictRepo, nil, nil)
	days := 90
	application, err := service.CreateWaiver(context.Background(), "", "7", "申请律师", "lawyer", &CreateWaiverRequest{
		ConflictCheckID: "actual-waiver-check", Rationale: "已取得知情同意，申请按隔离和监督条件进行独立评估。",
		AssignedReviewer: "9", ProposedConditions: []string{"建立信息隔离墙", "每月合规复核"}, DurationDays: &days,
	})
	require.NoError(t, err)
	stored, err := service.GetConflictWaiver(context.Background(), "actual-waiver-check", "7", "lawyer")
	require.NoError(t, err)
	assert.Equal(t, application.ID, stored.ID)
	assert.Equal(t, WaiverStatusUnderReview, stored.Status)

	decided, err := service.DecideWaiver(context.Background(), application.ID, "9", "合规复核人", "compliance", &WaiverDecisionRequest{
		Decision: "APPROVE", DecisionReason: "书面同意和隔离措施完整，可以按条件批准。",
	})
	require.NoError(t, err)
	assert.Equal(t, WaiverStatusApproved, decided.Status)
	record, err := conflictRepo.GetCheckRecord(context.Background(), "actual-waiver-check")
	require.NoError(t, err)
	assert.Equal(t, "WAIVED", record.CheckResult["decision"].(map[string]interface{})["status"])
}

func TestWaiverWorkflowRejectsConfirmedDirectConflict(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.AutoMigrate(
		&models.ConflictCheckRecord{},
		&models.WaiverApplication{},
	))
	conflictRepo := repositories.NewConflictRepository(db, nil)
	waiverRepo := repositories.NewEnhancedConflictRepository(db)
	require.NoError(t, conflictRepo.SaveCheckRecord(context.Background(), &models.ConflictCheckRecord{
		CheckID: "direct-conflict-no-waiver", UserID: 7, HasConflict: true,
		CheckResult: models.JSON{
			"decision": map[string]interface{}{"status": "BLOCKED"},
			"conflictCases": []interface{}{map[string]interface{}{
				"ruleCode":     "DIRECT_ADVERSE_CURRENT_CLIENT",
				"matchType":    "EXACT",
				"conflictType": "对方当事人直接冲突",
			}},
		},
	}))

	service := NewWaiverWorkflowService(waiverRepo, conflictRepo, nil, nil)
	_, err := service.CreateWaiver(context.Background(), "", "7", "申请律师", "lawyer", &CreateWaiverRequest{
		ConflictCheckID:    "direct-conflict-no-waiver",
		Rationale:          "测试直接冲突不得申请豁免。",
		AssignedReviewer:   "9",
		ProposedConditions: []string{"隔离墙"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrWaiverForbidden)
}

func TestAuditSafeConflictRequestUsesKeyedIdentityDigest(t *testing.T) {
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("s", 32))
	request := &models.ConflictCheckRequest{
		ClientIdentifiers: map[string]string{"id_card": "110101199001010011"},
		Parties: []models.ConflictPartyInfo{{
			Name:        "测试对方",
			Identifiers: map[string]string{"unified_social_credit_code": "91310000TEST0001"},
		}},
	}

	safe := auditSafeConflictRequest(request)
	require.NotNil(t, safe)
	require.NotContains(t, safe.ClientIdentifiers["id_card"], "110101199001010011")
	assert.True(t, strings.HasPrefix(safe.ClientIdentifiers["id_card"], "hmac-sha256:"))
	assert.True(t, strings.HasPrefix(safe.Parties[0].Identifiers["unified_social_credit_code"], "hmac-sha256:"))

	t.Setenv("SUBJECT_DATA_KEY", "")
	protected := auditSafeConflictRequest(request)
	assert.Equal(t, "已保护标识", protected.ClientIdentifiers["id_card"])
}

func TestStructuredPartyConflictsSearchesHistoricalClientArchive(t *testing.T) {
	db := setupSQLiteForConflictDetection(t)
	require.NoError(t, db.AutoMigrate(&models.Entity{}, &models.EntityNameHistory{}, &models.EntityRelation{}, &models.CaseParty{}))
	require.NoError(t, db.Create(&models.User{ID: 9, Username: "archive-lawyer", Name: "历史承办律师", Email: "archive-client-search@example.test", Password: "hash", Role: "lawyer"}).Error)
	require.NoError(t, db.Create(&models.Client{ID: 1, Name: "拟接案客户", Type: "企业", Email: "current-client-search@example.test", IDCard: "110101199001010011"}).Error)
	require.NoError(t, db.Create(&models.Client{ID: 2, Name: "远景重工有限公司", Type: "企业", Email: "historical-client-search@example.test"}).Error)
	require.NoError(t, db.Create(&models.Client{ID: 3, Name: "身份历史客户", Type: "个人", Email: "historical-identity-search@example.test", IDCard: "110101199001010022"}).Error)
	require.NoError(t, db.Create(&models.Case{ID: 201, CaseNumber: "CLIENT-ARCHIVE-201", Title: "历史客户主档事项", ClientID: 2, LawyerID: 9, CaseType: "商事", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.Case{ID: 202, CaseNumber: "CLIENT-ARCHIVE-202", Title: "身份历史事项", ClientID: 3, LawyerID: 9, CaseType: "民事", Status: "active"}).Error)

	svc := &conflictDetectionService{caseRepo: repositories.NewCaseRepository(db)}
	nameConflicts, err := svc.checkStructuredPartyConflicts(context.Background(), &models.ConflictCheckRequest{
		ClientID:   "1",
		ClientName: "远景重工",
		ClientType: "COMPANY",
	}, time.Time{})
	require.NoError(t, err)
	require.Len(t, nameConflicts, 1)
	assert.Equal(t, "CLIENT_ARCHIVE_NAME_CANDIDATE", nameConflicts[0].RuleCode)
	assert.Equal(t, "CLIENT_ARCHIVE", nameConflicts[0].Evidence[0].SourceType)

	identityConflicts, err := svc.checkStructuredPartyConflicts(context.Background(), &models.ConflictCheckRequest{
		ClientID:          "1",
		ClientName:        "完全不同的录入名",
		ClientType:        "PERSON",
		ClientIdentifiers: map[string]string{"id_card": "110101199001010022"},
	}, time.Time{})
	require.NoError(t, err)
	require.Len(t, identityConflicts, 1)
	assert.Equal(t, "STRUCTURED_IDENTITY_EXACT", identityConflicts[0].RuleCode)
	assert.Equal(t, "EXACT", identityConflicts[0].MatchType)
	assert.Equal(t, "CLIENT-ARCHIVE-202", identityConflicts[0].CaseNo)
}
