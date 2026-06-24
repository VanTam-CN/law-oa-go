package services

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
	"law-oa-go/internal/repositories"
)

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
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("svc_conflict_%d_%d.db",
		time.Now().UnixNano(), rand.Int63()))
	dsn := fmt.Sprintf("file:%s?_busy_timeout=30000&_journal_mode=WAL", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	if err := db.AutoMigrate(
		&models.Case{},
		&models.Client{},
		&models.User{},
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

// TestCheckOpponentConflicts_RejectsInvalidUserID
// 验证 checkOpponentConflicts 对非法 UserID 提前返回 error，而不是循环内 log+continue。
func TestCheckOpponentConflicts_RejectsInvalidUserID(t *testing.T) {
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

	require.Error(t, err, "非法 UserID 必须返回错误")
	assert.Nil(t, conflicts)
}
