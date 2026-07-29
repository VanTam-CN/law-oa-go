//go:build integration

// Package conflict_dialects 集成测试：在真实 MySQL 8 / PostgreSQL 下验证冲突检测查询
// 使用跨库通用的 LOWER(col) LIKE LOWER(?) 而非 PostgreSQL 专属 ILIKE。
//
// 运行方式（需要本机或 CI 暴露 MySQL 8 + PostgreSQL）：
//
//	CONFLICT_INTEGRATION_MYSQL_DSN="law_oa_user:1q2w#E$R@tcp(127.0.0.1:3306)/law_oa_test?charset=utf8mb4&parseTime=true&loc=Local" \
//	CONFLICT_INTEGRATION_PG_DSN="host=127.0.0.1 user=law_oa_user password=1q2w#E$R dbname=law_oa_test sslmode=disable" \
//	go test -tags=integration ./tests/integration/conflict_dialects -run TestConflictDetectionDialects -count=1
//
// 本文件默认在 `go test ./...` 中不参与编译（build tag=integration）。
// 独立子目录避开 tests/integration 主包的预存编译错误。
package conflict_dialects

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// 对每种后端跑相同的查询测试。SQLite 已在单元测试中覆盖（ILIKE → LOWER LIKE 修复）。
// 本集成测试专门验证：MySQL 8 / PostgreSQL 下 GetPotentialConflicts / checkOpponentConflicts
// 都能命中大小写不同的对方当事人，不依赖 ILIKE。

type backend struct {
	name string
	dsn  string
	open func(string) (func(), *gorm.DB, error)
}

func backends(t *testing.T) []backend {
	t.Helper()
	cases := []backend{}

	if dsn := os.Getenv("CONFLICT_INTEGRATION_MYSQL_DSN"); dsn != "" {
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

	if dsn := os.Getenv("CONFLICT_INTEGRATION_PG_DSN"); dsn != "" {
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
		t.Skip("未设置 CONFLICT_INTEGRATION_MYSQL_DSN / CONFLICT_INTEGRATION_PG_DSN，跳过真实数据库集成测试")
	}
	return cases
}

func conflictSchema() []interface{} {
	return []interface{}{
		&models.Case{},
		&models.Client{},
		&models.User{},
	}
}

func truncateAll(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec("DELETE FROM cases").Error)
	require.NoError(t, db.Exec("DELETE FROM clients").Error)
	require.NoError(t, db.Exec("DELETE FROM users").Error)
}

func seed(t *testing.T, db *gorm.DB) (clientID uint, lawyerID uint) {
	t.Helper()
	lawyer := models.User{Name: "integ_lawyer", Email: fmt.Sprintf("integ_l_%d@example.com", time.Now().UnixNano()), Password: "x", Role: "lawyer"}
	require.NoError(t, db.Create(&lawyer).Error)
	client := models.Client{Name: "Integ Client", Type: "individual"}
	require.NoError(t, db.Create(&client).Error)
	// 案件标题用大写；对方当事人查询用小写——验证跨库大小写不敏感
	require.NoError(t, db.Create(&models.Case{
		Title:    "Suit against ACME CORP for breach",
		CaseType: "民事",
		ClientID: client.ID,
		LawyerID: lawyer.ID,
	}).Error)
	return client.ID, lawyer.ID
}

// TestConflictDetectionDialects_CaseInsensitiveMatchOnRealDBs
// 在 MySQL/PostgreSQL 上验证：GetPotentialConflicts 的对方当事人子查询
// 能命中大小写不同的字符串；底层使用 LOWER(col) LIKE LOWER(?) 而非 ILIKE。
func TestConflictDetectionDialects_CaseInsensitiveMatchOnRealDBs(t *testing.T) {
	for _, b := range backends(t) {
		b := b
		t.Run(b.name, func(t *testing.T) {
			cleanup, db, err := b.open(b.dsn)
			require.NoError(t, err, "连接 %s 失败", b.name)
			defer cleanup()

			require.NoError(t, db.AutoMigrate(conflictSchema()...))
			truncateAll(t, db)
			clientID, lawyerID := seed(t, db)

			repo := repositories.NewConflictRepository(db, nil)
			cases, err := repo.GetPotentialConflicts(context.Background(),
				fmt.Sprintf("%d", clientID), lawyerID,
				[]string{"acme corp"}, time.Time{})
			require.NoError(t, err, "%s 下应当能跑通 LOWER LIKE 查询", b.name)
			assert.NotEmpty(t, cases, "%s 下大小写不同的对方当事人必须命中，否则跨库漏报", b.name)
		})
	}
}

// TestConflictDetectionDialects_ServicePropagatesDBError
// 在 MySQL/PostgreSQL 上验证：手动 DROP clients 表后，service 层
// checkOpponentConflicts 必须返回 error，而不是吞掉错误伪造"无冲突"。
func TestConflictDetectionDialects_ServicePropagatesDBError(t *testing.T) {
	for _, b := range backends(t) {
		b := b
		t.Run(b.name, func(t *testing.T) {
			cleanup, db, err := b.open(b.dsn)
			require.NoError(t, err)
			defer cleanup()

			require.NoError(t, db.AutoMigrate(conflictSchema()...))
			truncateAll(t, db)

			// 手动 DROP clients 表，让 JOIN 失败
			require.NoError(t, db.Migrator().DropTable("clients"))

			svc := services.NewConflictDetectionService(
				repositories.NewConflictRepository(db, nil),
				nil, nil, nil,
				repositories.NewCaseRepository(db),
			)
			req := &models.ConflictCheckRequest{
				ClientID:     "1",
				UserID:       "1",
				OtherParties: []string{"acme corp"},
			}

			// 调用 public API PerformConflictCheck；它内部调用 checkOpponentConflicts
			_, svcErr := svc.PerformConflictCheck(context.Background(), req)
			require.Error(t, svcErr, "%s 下数据库故障必须传播到 service 调用方", b.name)
		})
	}
}
