//go:build integration

// Package trust_atomicity 集成测试：真实 MySQL 8 / PostgreSQL 下验证代管款审批的原子性与幂等性。
//
// 运行方式（需要本机或 CI 暴露 MySQL 8 + PostgreSQL）：
//
//	TRUST_INTEGRATION_MYSQL_DSN="law_oa_user:1q2w#E$R@tcp(127.0.0.1:3306)/law_oa_test?charset=utf8mb4&parseTime=true&loc=Local" \
//	TRUST_INTEGRATION_PG_DSN="host=127.0.0.1 user=law_oa_user password=1q2w#E$R dbname=law_oa_test sslmode=disable" \
//	go test -tags=integration ./tests/integration/trust_atomicity -run TrustTransactionAtomicity -count=1
//
// 本文件默认在 `go test ./...` 中不参与编译（build tag=integration）。
// 独立子目录以避开 tests/integration 主包的预存编译错误（计划文件 Task 5 基线已记录）。
package trust_atomicity

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// 对每种后端跑相同的并发幂等测试。SQLite 不参与——FOR UPDATE 在 SQLite 是 no-op，
// 单元测试 internal/services/trust_account_service_test.go 已覆盖逻辑正确性。
//
// 本集成测试专门验证：在 MySQL InnoDB / PostgreSQL MVCC 下，两并发 Approve
// 只有一个 RowsAffected==1，另一个收到 ErrTransactionNotPending。

type backend struct {
	name string
	dsn  string
	open func(string) (func(), *gorm.DB, error)
}

func backends(t *testing.T) []backend {
	t.Helper()
	cases := []backend{}

	if dsn := os.Getenv("TRUST_INTEGRATION_MYSQL_DSN"); dsn != "" {
		cases = append(cases, backend{
			name: "mysql",
			dsn:  dsn,
			open: func(d string) (func(), *gorm.DB, error) {
				db, err := gorm.Open(mysql.Open(d), &gorm.Config{})
				if err != nil {
					return nil, nil, err
				}
				return func() {
					if sqlDB, e := db.DB(); e == nil {
						_ = sqlDB.Close()
					}
				}, db, nil
			},
		})
	}

	if dsn := os.Getenv("TRUST_INTEGRATION_PG_DSN"); dsn != "" {
		cases = append(cases, backend{
			name: "postgres",
			dsn:  dsn,
			open: func(d string) (func(), *gorm.DB, error) {
				db, err := gorm.Open(postgres.Open(d), &gorm.Config{})
				if err != nil {
					return nil, nil, err
				}
				return func() {
					if sqlDB, e := db.DB(); e == nil {
						_ = sqlDB.Close()
					}
				}, db, nil
			},
		})
	}

	if len(cases) == 0 {
		t.Skip("未设置 TRUST_INTEGRATION_MYSQL_DSN / TRUST_INTEGRATION_PG_DSN，跳过真实数据库集成测试")
	}
	return cases
}

func trustSchema() []interface{} {
	return []interface{}{
		&models.ClientTrustAccount{},
		&models.ClientTrustTransaction{},
	}
}

func truncateAll(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec("DELETE FROM client_trust_transactions").Error)
	require.NoError(t, db.Exec("DELETE FROM client_trust_accounts").Error)
}

func seedRow(t *testing.T, db *gorm.DB, balance float64) (*models.ClientTrustAccount, *models.ClientTrustTransaction) {
	t.Helper()
	now := time.Now()
	acct := &models.ClientTrustAccount{
		ClientID: 1, AccountCode: fmt.Sprintf("TA-INT-%d", time.Now().UnixNano()),
		Balance: balance, Currency: "CNY", Status: "active", OpenedAt: &now,
	}
	require.NoError(t, db.Create(acct).Error)

	txn := &models.ClientTrustTransaction{
		AccountID: acct.ID, TransactionCode: fmt.Sprintf("TT-INT-%d", time.Now().UnixNano()),
		TransactionType: "withdraw", Amount: 80, Description: "integration test withdraw",
		Status: "pending", CreatedBy: 1,
	}
	require.NoError(t, db.Create(txn).Error)

	// 用行锁读取保证返回数据库当前快照（避免 SQLite cached row 误用）
	var fresh models.ClientTrustTransaction
	require.NoError(t, db.Clauses(clause.Locking{Strength: "SHARE"}).First(&fresh, txn.ID).Error)
	return acct, &fresh
}

// TestTrustTransactionAtomicity_ConcurrentApprovesOnlyOneSucceeds
// 在 MySQL/PostgreSQL 上验证：两个 goroutine 同时审批同一笔交易，
// 只有一个 ApproveTransaction 返回 nil；另一个返回 ErrTransactionNotPending；
// 最终账户余额恰好为 balance - amount，交易状态为 completed。
func TestTrustTransactionAtomicity_ConcurrentApprovesOnlyOneSucceeds(t *testing.T) {
	for _, b := range backends(t) {
		b := b
		t.Run(b.name, func(t *testing.T) {
			cleanup, db, err := b.open(b.dsn)
			require.NoError(t, err, "连接 %s 失败", b.name)
			defer cleanup()

			require.NoError(t, db.AutoMigrate(trustSchema()...))
			truncateAll(t, db)

			_, txn := seedRow(t, db, 100)

			uow := repositories.NewTrustUnitOfWork(db)
			svc := services.NewTrustTransactionService(
				repositories.NewTrustTransactionRepository(db),
				repositories.NewTrustAccountRepository(db),
				nil, nil,
				services.WithTrustUnitOfWork(uow),
			)

			const concurrency = 2
			var wg sync.WaitGroup
			wg.Add(concurrency)
			var successCount int32
			var notPendingCount int32
			start := make(chan struct{})

			for i := 0; i < concurrency; i++ {
				go func() {
					defer wg.Done()
					<-start
					_, err := svc.ApproveTransaction(context.Background(), txn.ID, 99)
					switch {
					case err == nil:
						atomic.AddInt32(&successCount, 1)
					case errors.Is(err, services.ErrTransactionNotPending) ||
						errors.Is(err, services.ErrTransactionAlreadyProcessed):
						atomic.AddInt32(&notPendingCount, 1)
					default:
						t.Errorf("意外错误: %v", err)
					}
				}()
			}
			close(start)
			wg.Wait()

			assert.Equal(t, int32(1), atomic.LoadInt32(&successCount),
				"成功审批次数必须恰好为 1")
			assert.Equal(t, int32(1), atomic.LoadInt32(&notPendingCount),
				"并发冲突方必须收到 ErrTransactionNotPending")

			var finalAcct models.ClientTrustAccount
			require.NoError(t, db.First(&finalAcct, txn.AccountID).Error)
			assert.InDelta(t, 20.0, finalAcct.Balance, 0.0001,
				"余额必须恰好为 20（100-80），不能重复扣款")

			var finalTxn models.ClientTrustTransaction
			require.NoError(t, db.First(&finalTxn, txn.ID).Error)
			assert.Equal(t, "completed", finalTxn.Status)
		})
	}
}

// TestTrustTransactionAtomicity_RollsBackOnInsufficientBalance
// 在 MySQL/PostgreSQL 上验证：取款超过可用余额时拒绝审批；事务回滚后余额与交易状态保持原样。
func TestTrustTransactionAtomicity_RollsBackOnInsufficientBalance(t *testing.T) {
	for _, b := range backends(t) {
		b := b
		t.Run(b.name, func(t *testing.T) {
			cleanup, db, err := b.open(b.dsn)
			require.NoError(t, err)
			defer cleanup()

			require.NoError(t, db.AutoMigrate(trustSchema()...))
			truncateAll(t, db)

			_, txn := seedRow(t, db, 50) // 余额 50，取款 80 必然不足

			uow := repositories.NewTrustUnitOfWork(db)
			svc := services.NewTrustTransactionService(
				repositories.NewTrustTransactionRepository(db),
				repositories.NewTrustAccountRepository(db),
				nil, nil,
				services.WithTrustUnitOfWork(uow),
			)

			_, err = svc.ApproveTransaction(context.Background(), txn.ID, 7)
			require.Error(t, err)
			assert.True(t, errors.Is(err, services.ErrInsufficientBalance),
				"应为 ErrInsufficientBalance，实际=%v", err)

			var finalAcct models.ClientTrustAccount
			require.NoError(t, db.First(&finalAcct, txn.AccountID).Error)
			assert.InDelta(t, 50.0, finalAcct.Balance, 0.0001, "余额必须仍为 50")

			var finalTxn models.ClientTrustTransaction
			require.NoError(t, db.First(&finalTxn, txn.ID).Error)
			assert.Equal(t, "pending", finalTxn.Status, "交易状态必须仍为 pending")
		})
	}
}
