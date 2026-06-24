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

	if !service.isPartyNameMatch("上海示例科技有限公司", "示例科技") {
		t.Fatal("expected service party matching to normalize common company suffixes")
	}
}

func TestConflictDetectionPartyNameMatchRejectsUnrelatedNames(t *testing.T) {
	service := &conflictDetectionService{}

	if service.isPartyNameMatch("上海示例科技有限公司", "北京无关贸易有限公司") {
		t.Fatal("expected unrelated names not to match")
	}
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
