package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"law-oa-go/internal/database"
)

// TestPostgreSQL_ProductionSchemaIntakeClaimToken locks the fresh-bootstrap
// contract for migrations/000080 + 000081: the intake duplicate-approval
// guard column and the intake-to-approval unique binding must exist on a
// clean production bootstrap, and the contract validator must accept the
// resulting schema. Regression for the C4 armed run in which a fresh
// bootstrap (schema contract before this fix) reached runtime without
// case_intakes.approval_claim_token and the submit gate failed with
// SQLSTATE 42703 before any approval row was created.
func TestPostgreSQL_ProductionSchemaIntakeClaimToken(t *testing.T) {
	if os.Getenv("LAW_OA_RUN_POSTGRES_EVIDENCE_INTEGRATION") != "1" || os.Getenv("LAW_OA_POSTGRES_TEST_DSN") == "" {
		t.Skip("需要专用 LAW_OA_POSTGRES_TEST_DSN，并设置 LAW_OA_RUN_POSTGRES_EVIDENCE_INTEGRATION=1")
	}

	ctx := context.Background()
	db, err := sql.Open("postgres", os.Getenv("LAW_OA_POSTGRES_TEST_DSN"))
	if err != nil {
		t.Fatalf("连接专用 PostgreSQL 测试库失败: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("专用 PostgreSQL 测试库不可用: %v", err)
	}

	schema := fmt.Sprintf("production_schema_intake_claim_test_%d", os.Getpid())
	dropPostgresEvidenceSchema(t, db, schema)
	// The shared handle is closed by defers before t.Cleanup runs, so the
	// schema cleanup must open its own short-lived connection.
	t.Cleanup(func() {
		cleanupDB, err := sql.Open("postgres", os.Getenv("LAW_OA_POSTGRES_TEST_DSN"))
		if err != nil {
			t.Errorf("清理连接创建失败: %v", err)
			return
		}
		defer cleanupDB.Close()
		dropPostgresEvidenceSchema(t, cleanupDB, schema)
	})
	if _, err := db.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatalf("创建 PostgreSQL 测试 schema 失败: %v", err)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("获取 PostgreSQL 连接失败: %v", err)
	}
	defer conn.Close()
	setPostgresEvidenceSearchPath(t, ctx, conn, schema)

	// Mirror the reviewed fresh-bootstrap test setup: the GORM connection
	// rides on a pgx stdlib conn (lib/pq's extended protocol disagrees with
	// GORM's parameterized LIMIT), and foreign-key DDL stays disabled like
	// InitWithConfig so the test does not strengthen the schema contract
	// beyond the reviewed bootstrap.
	pgxDB, err := sql.Open("pgx", os.Getenv("LAW_OA_POSTGRES_TEST_DSN"))
	if err != nil {
		t.Fatalf("初始化 pgx PostgreSQL 连接失败: %v", err)
	}
	defer pgxDB.Close()
	pgxConn, err := pgxDB.Conn(ctx)
	if err != nil {
		t.Fatalf("获取 pgx PostgreSQL 连接失败: %v", err)
	}
	defer pgxConn.Close()
	setPostgresEvidenceSearchPath(t, ctx, pgxConn, schema)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: pgxConn, DriverName: "pgx"}), &gorm.Config{
		DisableAutomaticPing:                     true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("初始化 GORM PostgreSQL 连接失败: %v", err)
	}
	if err := database.BootstrapProductionSchema(gormDB.Session(&gorm.Session{NewDB: true, Logger: gormDB.Logger})); err != nil {
		t.Fatalf("fresh PostgreSQL bootstrap 失败: %v", err)
	}

	t.Run("claim token column exists with reviewed type", func(t *testing.T) {
		var columnType string
		if err := conn.QueryRowContext(ctx, `
			SELECT data_type || ':' || COALESCE(character_maximum_length::text, '')
			FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = 'case_intakes'
			  AND column_name = 'approval_claim_token'
		`).Scan(&columnType); err != nil {
			t.Fatalf("fresh bootstrap 缺少 approval_claim_token 列: %v", err)
		}
		if columnType != "character varying:64" {
			t.Fatalf("approval_claim_token 类型期望 character varying(64)，实际 %s", columnType)
		}
	})

	t.Run("intake unique binding rejects duplicates and allows nulls", func(t *testing.T) {
		var uniqueIndexCount int
		if err := conn.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM pg_indexes
			WHERE schemaname = current_schema()
			  AND tablename = 'approval_case_creation_tracking'
			  AND indexname = 'idx_case_creation_tracking_intake_unique'
			  AND indexdef LIKE 'CREATE UNIQUE INDEX%'
		`).Scan(&uniqueIndexCount); err != nil {
			t.Fatalf("读取成案跟踪唯一索引失败: %v", err)
		}
		if uniqueIndexCount < 1 {
			t.Fatalf("fresh bootstrap 缺少 approval_case_creation_tracking(intake_id) 唯一索引")
		}

		insertTracking := func(id string, intakeID *string) error {
			_, err := conn.ExecContext(ctx, `
				INSERT INTO approval_case_creation_tracking (id, approval_request_id, intake_id, created_by)
				VALUES ($1, 'schema-test-approval', $2, 'schema-test')
			`, id, intakeID)
			return err
		}
		bound := "schema-test-intake-binding"
		if err := insertTracking("track-bound-first", &bound); err != nil {
			t.Fatalf("首次绑定插入失败: %v", err)
		}
		t.Cleanup(func() {
			_, _ = conn.ExecContext(ctx, "DELETE FROM approval_case_creation_tracking WHERE id IN ('track-bound-first', 'track-bound-second', 'track-null-a', 'track-null-b')")
		})
		if err := insertTracking("track-bound-second", &bound); err == nil {
			t.Fatalf("同一 intake 的第二个绑定未被唯一索引拒绝")
		}
		for _, id := range []string{"track-null-a", "track-null-b"} {
			if err := insertTracking(id, nil); err != nil {
				t.Fatalf("NULL intake_id 行 %s 应不受唯一索引限制: %v", id, err)
			}
		}
	})

	t.Run("runtime claim and release predicates round-trip", func(t *testing.T) {
		var intakeID string
		if err := conn.QueryRowContext(ctx, `
			INSERT INTO case_intakes (intake_code, title, status)
			VALUES ('SCHEMA-TEST-CLAIM-0001', 'claim token contract', 'conflict_ready')
			RETURNING id
		`).Scan(&intakeID); err != nil {
			t.Fatalf("播种冲突就绪接案失败: %v", err)
		}
		t.Cleanup(func() {
			_, _ = conn.ExecContext(ctx, "DELETE FROM case_intakes WHERE id = $1", intakeID)
		})

		const claimToken = "schema-test-claim-token"
		claim := `
			UPDATE case_intakes
			SET status = 'approval_submitting',
			    updated_at = CURRENT_TIMESTAMP,
			    approval_claim_token = $2
			WHERE id = $1
			  AND (status = 'conflict_ready'
			       OR (status = 'approval_submitting' AND updated_at < CURRENT_TIMESTAMP - interval '5 minutes'))
		`
		result, err := conn.ExecContext(ctx, claim, intakeID, claimToken)
		if err != nil {
			t.Fatalf("接案锁定更新失败: %v", err)
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			t.Fatalf("接案锁定影响行数期望 1，实际 %d", affected)
		}

		var storedToken *string
		if err := conn.QueryRowContext(ctx,
			"SELECT approval_claim_token FROM case_intakes WHERE id = $1", intakeID,
		).Scan(&storedToken); err != nil {
			t.Fatalf("读取锁定令牌失败: %v", err)
		}
		if storedToken == nil || *storedToken != claimToken {
			t.Fatalf("锁定令牌回读期望 %s，实际 %v", claimToken, storedToken)
		}

		release := `
			UPDATE case_intakes
			SET status = 'conflict_ready',
			    updated_at = CURRENT_TIMESTAMP,
			    approval_claim_token = NULL
			WHERE id = $1 AND status = 'approval_submitting' AND approval_claim_token = $2
		`
		result, err = conn.ExecContext(ctx, release, intakeID, claimToken)
		if err != nil {
			t.Fatalf("接案释放更新失败: %v", err)
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			t.Fatalf("接案释放影响行数期望 1，实际 %d", affected)
		}
	})

	t.Run("contract validator accepts fresh schema", func(t *testing.T) {
		if err := database.ValidateProductionSchemaContract(gormDB); err != nil {
			t.Fatalf("生产 schema 契约校验拒绝 fresh bootstrap: %v", err)
		}
	})
}
