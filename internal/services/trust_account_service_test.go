package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// setupTrustTestDB 构造独立 SQLite + 代管款相关表
//
// 选择文件模式而非 :memory: + cache=shared 的原因：
//   - shared cache 模式下表锁立即冲突，无法测试 FOR UPDATE 的并发行为；
//   - 文件模式 + WAL + busy_timeout 让并发事务串行化，可复现 MySQL/PostgreSQL 的行锁语义。
//
// 每个测试一个独立 db 文件（在 t.TempDir 下），测试结束自动清理。
//
// 注：项目多个 model 复用 idx_status 索引名，SQLite 索引名全局唯一，
// AutoMigrate 对第二个 model 创建同名索引会返回 "already exists"，
// 这是已有 model 设计问题，与本任务无关，此处显式忽略以保证测试可跑。
func setupTrustTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("trust_%d_%d.db",
		time.Now().UnixNano(), rand.Int63()))
	dsn := fmt.Sprintf("file:%s?_busy_timeout=30000&_journal_mode=WAL&_locking_mode=NORMAL", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// 单独打开的连接无法看到彼此的 PRAGMA，因此通过 DSN 设置。
	// 显式设置一次以保险。
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=30000")

	err = db.AutoMigrate(
		&models.ClientTrustAccount{},
		&models.ClientTrustTransaction{},
		&models.User{},
		&models.Client{},
		&models.Case{},
	)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		require.NoError(t, err)
	}

	// WAL 模式下产生的 -wal/-shm 文件随 t.TempDir 清理
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		_ = os.Remove(dbPath)
		_ = os.Remove(dbPath + "-wal")
		_ = os.Remove(dbPath + "-shm")
	})
	return db
}

// seedTrustAccount 写入一个 active 账户并返回其 ID
func seedTrustAccount(t *testing.T, db *gorm.DB, balance, frozen float64) *models.ClientTrustAccount {
	t.Helper()
	openedAt := time.Now()
	acct := &models.ClientTrustAccount{
		ClientID:     1,
		AccountCode:  "TA-TEST-001",
		Balance:      balance,
		Currency:     "CNY",
		FrozenAmount: frozen,
		Status:       "active",
		OpenedAt:     &openedAt,
	}
	require.NoError(t, db.Create(acct).Error)
	return acct
}

// seedPendingWithdraw 写入一笔 pending 取款交易
func seedPendingWithdraw(t *testing.T, db *gorm.DB, accountID uint, amount float64) *models.ClientTrustTransaction {
	t.Helper()
	txn := &models.ClientTrustTransaction{
		AccountID:       accountID,
		TransactionCode: "TT-TEST-001",
		TransactionType: "withdraw",
		Amount:          amount,
		Description:     "test withdraw",
		Status:          "pending",
		CreatedBy:       1,
	}
	require.NoError(t, db.Create(txn).Error)
	return txn
}

// newTrustServiceFromDB 用真实 GORM 仓储装配服务（含 UnitOfWork）
func newTrustServiceFromDB(t *testing.T, db *gorm.DB) *TrustTransactionService {
	t.Helper()
	txRepo := repositories.NewTrustTransactionRepository(db)
	acctRepo := repositories.NewTrustAccountRepository(db)
	uow := repositories.NewTrustUnitOfWork(db)
	return NewTrustTransactionService(txRepo, acctRepo, nil, nil, WithTrustUnitOfWork(uow))
}

// -------------------------------------------------------------------
// Step 1 (RED): 计划要求的两个核心场景
// -------------------------------------------------------------------

// TestApproveTransactionRollsBackWhenTransactionUpdateFails
// 初始余额 100，withdraw 80；当交易状态条件更新失败（被外部并发改为 cancelled）时，
// 余额必须回滚回 100，且交易不得变为 completed。
func TestApproveTransactionRollsBackWhenTransactionUpdateFails(t *testing.T) {
	db := setupTrustTestDB(t)
	acct := seedTrustAccount(t, db, 100, 0)
	txn := seedPendingWithdraw(t, db, acct.ID, 80)

	// 模拟「状态条件更新」失败：在调用 Approve 之前，先把状态改为 cancelled
	// 这样 UpdateStatusIfPending (WHERE status='pending') 会返回 RowsAffected=0
	require.NoError(t, db.Model(&models.ClientTrustTransaction{}).
		Where("id = ?", txn.ID).
		Update("status", "cancelled").Error)

	svc := newTrustServiceFromDB(t, db)
	_, err := svc.ApproveTransaction(context.Background(), txn.ID, 99)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTransactionNotPending) || errors.Is(err, ErrTransactionAlreadyProcessed),
		"err 应为 ErrTransactionNotPending，实际: %v", err)

	// 余额必须保持 100，未被扣减（事务回滚）
	var finalAcct models.ClientTrustAccount
	require.NoError(t, db.First(&finalAcct, acct.ID).Error)
	assert.InDelta(t, 100.0, finalAcct.Balance, 0.0001,
		"余额必须回滚为 100，实际: %v", finalAcct.Balance)

	// 交易仍为 cancelled（未被 Approve 改回 completed）
	var finalTxn models.ClientTrustTransaction
	require.NoError(t, db.First(&finalTxn, txn.ID).Error)
	assert.Equal(t, "cancelled", finalTxn.Status)
}

// TestApproveTransactionConcurrentOnlyAppliesOnce
// 两 goroutine 同时审批同一笔 withdraw；最终余额必须为 20 且只扣款一次。
func TestApproveTransactionConcurrentOnlyAppliesOnce(t *testing.T) {
	db := setupTrustTestDB(t)
	// 打开 SQLite WAL + busy_timeout，让事务串行化、避免 lock contention 直接失败
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=10000")

	acct := seedTrustAccount(t, db, 100, 0)
	txn := seedPendingWithdraw(t, db, acct.ID, 80)

	svc := newTrustServiceFromDB(t, db)

	const concurrency = 2
	var wg sync.WaitGroup
	wg.Add(concurrency)
	results := make([]error, concurrency)
	approvalOK := make([]bool, concurrency)

	start := make(chan struct{})
	for i := 0; i < concurrency; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start
			_, err := svc.ApproveTransaction(context.Background(), txn.ID, 99)
			results[i] = err
			approvalOK[i] = err == nil
		}()
	}
	close(start)
	wg.Wait()

	// 断言：成功数量恰好为 1
	successCount := 0
	for _, ok := range approvalOK {
		if ok {
			successCount++
		}
	}
	assert.Equal(t, 1, successCount,
		"并发审批必须只有 1 次成功，实际: %d (errors=%v)", successCount, results)

	// 最终余额必须为 20（100 - 80），未被重复扣减
	var finalAcct models.ClientTrustAccount
	require.NoError(t, db.First(&finalAcct, acct.ID).Error)
	assert.InDelta(t, 20.0, finalAcct.Balance, 0.0001,
		"余额必须为 20，实际: %v", finalAcct.Balance)

	// 交易状态最终为 completed
	var finalTxn models.ClientTrustTransaction
	require.NoError(t, db.First(&finalTxn, txn.ID).Error)
	assert.Equal(t, "completed", finalTxn.Status)
}

// -------------------------------------------------------------------
// 补充场景：审批成功、余额不足、重复审批
// -------------------------------------------------------------------

// TestApproveTransaction_Succeeds_Deposit 存款审批成功 → 余额增加
func TestApproveTransaction_Succeeds_Deposit(t *testing.T) {
	db := setupTrustTestDB(t)
	acct := seedTrustAccount(t, db, 100, 0)

	txn := &models.ClientTrustTransaction{
		AccountID:       acct.ID,
		TransactionCode: "TT-DEP-001",
		TransactionType: "deposit",
		Amount:          50,
		Description:     "deposit",
		Status:          "pending",
		CreatedBy:       1,
	}
	require.NoError(t, db.Create(txn).Error)

	svc := newTrustServiceFromDB(t, db)
	resp, err := svc.ApproveTransaction(context.Background(), txn.ID, 7)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "completed", resp.Status)

	var finalAcct models.ClientTrustAccount
	require.NoError(t, db.First(&finalAcct, acct.ID).Error)
	assert.InDelta(t, 150.0, finalAcct.Balance, 0.0001)
}

// TestApproveTransaction_Fails_InsufficientBalance 取款金额超过可用余额时拒绝
func TestApproveTransaction_Fails_InsufficientBalance(t *testing.T) {
	db := setupTrustTestDB(t)
	acct := seedTrustAccount(t, db, 50, 0)
	txn := seedPendingWithdraw(t, db, acct.ID, 80)

	svc := newTrustServiceFromDB(t, db)
	_, err := svc.ApproveTransaction(context.Background(), txn.ID, 9)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInsufficientBalance),
		"应为 ErrInsufficientBalance，实际: %v", err)

	// 余额未变，状态仍 pending
	var finalAcct models.ClientTrustAccount
	require.NoError(t, db.First(&finalAcct, acct.ID).Error)
	assert.InDelta(t, 50.0, finalAcct.Balance, 0.0001)

	var finalTxn models.ClientTrustTransaction
	require.NoError(t, db.First(&finalTxn, txn.ID).Error)
	assert.Equal(t, "pending", finalTxn.Status)
}

// TestApproveTransaction_Fails_AlreadyCompleted 重复审批必须显式失败而非再次扣款
func TestApproveTransaction_Fails_AlreadyCompleted(t *testing.T) {
	db := setupTrustTestDB(t)
	acct := seedTrustAccount(t, db, 100, 0)
	txn := seedPendingWithdraw(t, db, acct.ID, 30)

	svc := newTrustServiceFromDB(t, db)

	// 第一次审批必须成功，余额变为 70
	_, err := svc.ApproveTransaction(context.Background(), txn.ID, 9)
	require.NoError(t, err)

	// 第二次审批必须返回 ErrTransactionNotPending，不得再次扣款
	_, err = svc.ApproveTransaction(context.Background(), txn.ID, 9)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTransactionNotPending) || errors.Is(err, ErrTransactionAlreadyProcessed),
		"重复审批应返回 ErrTransactionNotPending，实际: %v", err)

	var finalAcct models.ClientTrustAccount
	require.NoError(t, db.First(&finalAcct, acct.ID).Error)
	assert.InDelta(t, 70.0, finalAcct.Balance, 0.0001)
}
