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

// TestIsDirectOpposingPartyClientMatchesNormalizedCompanyName 原有纯函数测试保留
func TestIsDirectOpposingPartyClientMatchesNormalizedCompanyName(t *testing.T) {
	if !isDirectOpposingPartyClient("上海示例科技有限公司", []string{"示例科技"}) {
		t.Fatal("expected normalized company name to match opposing party")
	}
}

// TestIsDirectOpposingPartyClientIgnoresUnrelatedParty 原有纯函数测试保留
func TestIsDirectOpposingPartyClientIgnoresUnrelatedParty(t *testing.T) {
	if isDirectOpposingPartyClient("上海示例科技有限公司", []string{"北京无关贸易有限公司"}) {
		t.Fatal("expected unrelated company name not to match")
	}
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
		fmt.Sprintf("%d", client.ID), /* lawyerID */ 999,
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
