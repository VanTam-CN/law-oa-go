package repositories

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
)

// TestIsDirectOpposingPartyClientMatchesNormalizedCompanyName
// 规范化后完全相等 → 应当直接命中（CRITICAL 级别匹配）
// 注意：旧实现把"示例科技"也当命中（contains），但那是短子串误报，已收紧。
func TestIsDirectOpposingPartyClientMatchesNormalizedCompanyName(t *testing.T) {
	if !isDirectOpposingPartyClient("上海示例科技有限公司", []string{"上海示例科技"}) {
		t.Fatal("expected exact normalized name to match opposing party")
	}
}

// TestIsDirectOpposingPartyClientIgnoresUnrelatedParty 原有纯函数测试保留
func TestIsDirectOpposingPartyClientIgnoresUnrelatedParty(t *testing.T) {
	if isDirectOpposingPartyClient("上海示例科技有限公司", []string{"北京无关贸易有限公司"}) {
		t.Fatal("expected unrelated company name not to match")
	}
}

// Task 7 Step 1 — RED 边界测试：短子串不应判 CRITICAL
//
// 修复前：isDirectOpposingPartyClient 使用 strings.Contains，"华" 会匹配
// "华为技术有限公司" → 升级为 CRITICAL → 错误拒绝接案。
// 修复后：只有规范化后完全相等才直接命中；包含关系应作为候选返回，由 caller 决定是否升级。
func TestIsDirectOpposingPartyClient_DoesNotMatchShortSubstring(t *testing.T) {
	cases := map[string][]string{
		"华为技术有限公司":         {"华"},
		"上海示例科技有限公司":       {"示例"},
		"阿里巴巴（中国）网络技术有限公司": {"阿里"},
	}
	for client, parties := range cases {
		for _, party := range parties {
			if isDirectOpposingPartyClient(client, []string{party}) {
				t.Fatalf("短子串 %q 不应直接命中 %q 升级为 CRITICAL", party, client)
			}
		}
	}
}

// TestIsDirectOpposingPartyClient_DoesNotMatchDifferentCompanyWithSimilarToken
// 名称相近但不同主体（共享通用词"甲"），不应判 CRITICAL。
func TestIsDirectOpposingPartyClient_DoesNotMatchDifferentCompanyWithSimilarToken(t *testing.T) {
	if isDirectOpposingPartyClient("北京甲科技有限公司", []string{"上海甲贸易有限公司"}) {
		t.Fatal("名称相近但不同主体不应直接命中 CRITICAL")
	}
}

// TestIsDirectOpposingPartyClient_RejectsEmptyAndSuffixOnly
// 空值、仅剩公司类型词不应命中。
func TestIsDirectOpposingPartyClient_RejectsEmptyAndSuffixOnly(t *testing.T) {
	cases := []struct {
		client string
		party  string
	}{
		{"", ""},
		{"有限公司", "公司"},
		{"上海示例科技有限公司", ""},
		{"", "上海示例科技有限公司"},
	}
	for _, c := range cases {
		if isDirectOpposingPartyClient(c.client, []string{c.party}) {
			t.Fatalf("空值/纯后缀 (%q vs %q) 不应命中", c.client, c.party)
		}
	}
}

// TestClassifyDirectOpposingPartyMatch
// 引入三态分类：Exact / Candidate / NoMatch。仅 Exact 可升级 CRITICAL。
func TestClassifyDirectOpposingPartyMatch(t *testing.T) {
	assert.Equal(t, PartyExactNormalizedMatch,
		classifyDirectOpposingPartyMatch("上海示例科技有限公司", "上海示例科技"))
	assert.Equal(t, PartyCandidateMatch,
		classifyDirectOpposingPartyMatch("上海示例科技有限公司", "示例科技"))
	assert.Equal(t, PartyCandidateMatch,
		classifyDirectOpposingPartyMatch("华为技术有限公司", "华为"))
	assert.Equal(t, PartyNoMatch,
		classifyDirectOpposingPartyMatch("北京甲科技有限公司", "上海甲贸易有限公司"))
	assert.Equal(t, PartyNoMatch,
		classifyDirectOpposingPartyMatch("", "任何"))
	assert.Equal(t, PartyNoMatch,
		classifyDirectOpposingPartyMatch("有限公司", "公司"))
}

// setupConflictSQLiteDB 构造独立 SQLite 文件库 + 冲突检测相关表。
// 文件模式避开 shared cache 下的索引名冲突。
func setupConflictSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("conflict_%d_%d.db",
		time.Now().UnixNano(), rand.Int63()))
	dsn := fmt.Sprintf("file:%s?_busy_timeout=30000&_journal_mode=WAL", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	if err := db.AutoMigrate(
		&models.Case{},
		&models.Client{},
		&models.User{},
	); err != nil && !strings.Contains(err.Error(), "already exists") {
		require.NoError(t, err)
	}
	return db
}

func TestDocumentRepositoryListRejectsRawSortBy(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Document{}))

	older := time.Now().Add(-time.Hour)
	newer := time.Now()
	docs := []*models.Document{
		{Name: "older", Filename: "older.pdf", Filepath: "/tmp/older.pdf", CreatedAt: older},
		{Name: "newer", Filename: "newer.pdf", Filepath: "/tmp/newer.pdf", CreatedAt: newer},
	}
	for _, doc := range docs {
		require.NoError(t, db.Create(doc).Error)
	}

	repo := NewDocumentRepository(db)
	got, total, err := repo.List(context.Background(), &DocumentListParams{
		Page:      1,
		PageSize:  10,
		SortBy:    "created_at; DROP TABLE documents; --",
		SortOrder: "asc",
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, got, 2)
	require.Equal(t, "newer", got[0].Name)
	require.True(t, db.Migrator().HasTable(&models.Document{}))
}

// TestGetPotentialConflicts_CaseInsensitiveMatchOnSQLite
// 验证 GetPotentialConflicts 的对方当事人子查询在 SQLite 下也能命中大小写不同的字符串。
//
// 修复前：partyQuery 使用 PostgreSQL 专属的 ILIKE，SQLite 不支持 → SQL 语法错误，
// 且错误被 `continue` 吞掉，调用方拿到 0 个冲突，产生"无冲突"假阴性。
//
// 修复后：partyQuery 改为 LOWER(col) LIKE LOWER(?) 跨库通用，
// SQLite/MySQL/PostgreSQL 都能命中。
func TestGetPotentialConflicts_CaseInsensitiveMatchOnSQLite(t *testing.T) {
	db := setupConflictSQLiteDB(t)

	lawyer := models.User{Name: "lawyer2", Email: "l2@example.com", Password: "x", Role: "lawyer"}
	require.NoError(t, db.Create(&lawyer).Error)

	client := models.Client{Name: "Client A", Type: "individual"}
	require.NoError(t, db.Create(&client).Error)

	// 案件标题用大写，对方当事人查询用小写——验证大小写不敏感
	require.NoError(t, db.Create(&models.Case{
		Title:    "Suit against ACME CORP for breach",
		CaseType: "民事",
		ClientID: client.ID,
		LawyerID: lawyer.ID,
	}).Error)

	repo := NewConflictRepository(db, nil)
	cases, err := repo.GetPotentialConflicts(context.Background(),
		fmt.Sprintf("%d", client.ID) /* lawyerID */, 999,
		[]string{"acme corp"}, time.Time{})
	require.NoError(t, err, "SQLite 下应当能跑通查询；当前 ILIKE 是 PostgreSQL 专属")

	// 至少命中对方当事人子查询
	assert.NotEmpty(t, cases, "大小写不同的对方当事人必须命中，否则跨库漏报")
}

// TestGetPotentialConflicts_PropagatesDBError
// 验证 GetPotentialConflicts 的对方当事人子查询失败时不会被静默吞掉。
//
// 修复前：partyRows, err := ...; if err != nil { continue } —— 错误被吞，
// 调用方拿到 0 个冲突。
// 修复后：错误必须包装上下文向上传播，调用方据此判定检查失败而非"无冲突"。
func TestGetPotentialConflicts_PropagatesDBError(t *testing.T) {
	db := setupConflictSQLiteDB(t)

	// 故意构造一个会让 partyQuery 报错的场景：DROP clients 表，
	// 让 JOIN 失败；partyQuery 不会再返回 0+nil
	require.NoError(t, db.Migrator().DropTable("clients"))

	repo := NewConflictRepository(db, nil)
	cases, err := repo.GetPotentialConflicts(context.Background(), "1", 999,
		[]string{"acme"}, time.Time{})

	// 错误必须显式返回——不能既没冲突又没错误（那就是漏报）
	if len(cases) == 0 {
		require.Error(t, err, "数据库故障必须传播 error，不能返回 (nil, nil) 制造无冲突假象")
	}
}

func TestGetPotentialConflicts_IgnoresUnrelatedCasesOwnedBySameLawyer(t *testing.T) {
	db := setupConflictSQLiteDB(t)
	lawyer := models.User{Name: "承办律师", Username: "owner", Email: "owner@example.test", Password: "x", Role: "lawyer", Status: "active"}
	require.NoError(t, db.Create(&lawyer).Error)
	currentClient := models.Client{Name: "当前客户有限公司", Type: "企业", Email: "current@example.test"}
	historicalClient := models.Client{Name: "无关历史客户有限公司", Type: "企业", Email: "history@example.test"}
	require.NoError(t, db.Create(&currentClient).Error)
	require.NoError(t, db.Create(&historicalClient).Error)
	require.NoError(t, db.Create(&models.Case{
		CaseNumber: "CASE-UNRELATED-001",
		Title:      "无关劳动咨询",
		CaseType:   "commercial",
		ClientID:   historicalClient.ID,
		LawyerID:   lawyer.ID,
		Status:     "active",
	}).Error)

	repo := NewConflictRepository(db, nil)
	cases, err := repo.GetPotentialConflicts(context.Background(), fmt.Sprint(currentClient.ID), lawyer.ID, []string{"另一家公司有限公司"}, time.Time{})
	require.NoError(t, err)
	assert.Empty(t, cases, "同一律师承办无关客户的案件不能单独构成代理冲突")
}

func TestGetPotentialConflicts_ClassifiesExactAndCandidateOpponentEvidence(t *testing.T) {
	db := setupConflictSQLiteDB(t)
	lawyer := models.User{Name: "承办律师", Username: "owner2", Email: "owner2@example.test", Password: "x", Role: "lawyer", Status: "active"}
	require.NoError(t, db.Create(&lawyer).Error)
	currentClient := models.Client{Name: "当前客户有限公司", Type: "企业", Email: "current2@example.test"}
	historicalClient := models.Client{Name: "上海示例科技有限公司", Type: "企业", Email: "example-tech@example.test"}
	require.NoError(t, db.Create(&currentClient).Error)
	require.NoError(t, db.Create(&historicalClient).Error)
	require.NoError(t, db.Create(&models.Case{
		CaseNumber: "CASE-HISTORY-001",
		Title:      "历史商事案件",
		CaseType:   "commercial",
		ClientID:   historicalClient.ID,
		LawyerID:   lawyer.ID,
		Status:     "active",
	}).Error)

	repo := NewConflictRepository(db, nil)
	exact, err := repo.GetPotentialConflicts(context.Background(), fmt.Sprint(currentClient.ID), lawyer.ID, []string{"上海示例科技"}, time.Time{})
	require.NoError(t, err)
	require.Len(t, exact, 1)
	assert.Equal(t, "CRITICAL", exact[0].RiskLevel)
	assert.Equal(t, "对方当事人直接冲突", exact[0].ConflictType)

	candidate, err := repo.GetPotentialConflicts(context.Background(), fmt.Sprint(currentClient.ID), lawyer.ID, []string{"示例科技"}, time.Time{})
	require.NoError(t, err)
	require.Len(t, candidate, 1)
	assert.Equal(t, "MEDIUM", candidate[0].RiskLevel)
	assert.Equal(t, "名称相似待核实", candidate[0].ConflictType)
}

func TestLinkConflictCheckToCaseRequiresExplicitConsistentContext(t *testing.T) {
	db := setupConflictSQLiteDB(t)
	client := models.Client{Name: "当前客户", Type: "企业"}
	require.NoError(t, db.Create(&client).Error)
	caseModel := models.Case{
		CaseNumber: "CASE-LINK-001",
		Title:      "待复核案件",
		CaseType:   "commercial",
		ClientID:   client.ID,
		LawyerID:   1,
		Status:     "pending",
	}
	require.NoError(t, db.Create(&caseModel).Error)
	require.NoError(t, db.Exec(`CREATE TABLE conflict_check_records (
		check_id TEXT PRIMARY KEY, client_id TEXT, search_parameters TEXT
	)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO conflict_check_records (check_id, client_id, search_parameters) VALUES (?, ?, ?)`,
		"CCT-LINK-001", fmt.Sprint(client.ID), fmt.Sprintf(`{"clientId":"%d","subjectCaseId":"%d","intakeId":"%s"}`, client.ID, caseModel.ID, "intake-link-001")).Error)
	require.NoError(t, db.Exec(`CREATE TABLE case_intakes (id TEXT PRIMARY KEY, client_id TEXT, status TEXT, metadata TEXT, updated_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO case_intakes (id, client_id, status, metadata, updated_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`, "intake-link-001", client.ID, "draft", `{"source":"test"}`).Error)

	repo := NewConflictRepository(db, nil)
	require.NoError(t, repo.(ConflictSubjectLinker).LinkConflictCheckToCase(context.Background(), ConflictSubjectAssociation{
		CheckID:           "CCT-LINK-001",
		SubjectCaseID:     fmt.Sprint(caseModel.ID),
		SubjectCaseNumber: caseModel.CaseNumber,
		IntakeID:          "intake-link-001",
		ClientID:          fmt.Sprint(client.ID),
		CoverageStatus:    "COMPLETE",
		CheckedAt:         time.Now(),
	}))

	var linked models.Case
	require.NoError(t, db.First(&linked, caseModel.ID).Error)
	assert.Equal(t, "CCT-LINK-001", linked.ConflictCheckID)
	assert.Equal(t, "COMPLETE", linked.ConflictCoverageStatus)
	var intake struct {
		Status   string
		Metadata string
	}
	require.NoError(t, db.Table("case_intakes").Select("status, metadata").Where("id = ?", "intake-link-001").Take(&intake).Error)
	assert.Equal(t, "conflict_ready", intake.Status)
	assert.Contains(t, intake.Metadata, "CCT-LINK-001")

	err := repo.(ConflictSubjectLinker).LinkConflictCheckToCase(context.Background(), ConflictSubjectAssociation{
		CheckID:           "CCT-LINK-002",
		SubjectCaseID:     fmt.Sprint(caseModel.ID),
		SubjectCaseNumber: "CASE-WRONG-001",
		ClientID:          fmt.Sprint(client.ID),
	})
	require.Error(t, err)
	var unchanged models.Case
	require.NoError(t, db.First(&unchanged, caseModel.ID).Error)
	assert.Equal(t, "CCT-LINK-001", unchanged.ConflictCheckID, "a stale case number must not overwrite the existing association")

	err = repo.(ConflictSubjectLinker).LinkConflictCheckToCase(context.Background(), ConflictSubjectAssociation{
		CheckID:       "CCT-LINK-001",
		SubjectCaseID: fmt.Sprint(caseModel.ID),
		ClientID:      "different-client",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "客户")

	err = repo.(ConflictSubjectLinker).LinkConflictCheckToCase(context.Background(), ConflictSubjectAssociation{
		CheckID:           "CCT-LINK-001",
		SubjectCaseID:     fmt.Sprint(caseModel.ID + 1),
		SubjectCaseNumber: caseModel.CaseNumber,
		ClientID:          fmt.Sprint(client.ID),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "案件 ID")
}
