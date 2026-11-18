package integration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"law-oa-go/internal/database"
)

func TestPostgreSQL_ProductionExternalEvidenceMigrationAndBootstrap(t *testing.T) {
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

	t.Run("migration is idempotent and append-only", func(t *testing.T) {
		schema := fmt.Sprintf("production_evidence_migration_test_%d", os.Getpid())
		dropPostgresEvidenceSchema(t, db, schema)
		t.Cleanup(func() { dropPostgresEvidenceSchema(t, db, schema) })
		if _, err := db.Exec("CREATE SCHEMA " + schema); err != nil {
			t.Fatalf("创建 PostgreSQL 测试 schema 失败: %v", err)
		}

		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("获取 PostgreSQL 连接失败: %v", err)
		}
		defer conn.Close()
		setPostgresEvidenceSearchPath(t, ctx, conn, schema)
		for _, table := range []string{"compliance_audit_events", "conflict_reviews", "case_subject_revisions"} {
			if _, err := conn.ExecContext(ctx, "CREATE TABLE "+table+" (id INTEGER PRIMARY KEY)"); err != nil {
				t.Fatalf("创建迁移前置测试表 %s 失败: %v", table, err)
			}
		}
		executeMigrationFiles(t, ctx, conn,
			"000062_append_only_evidence_guards.up.sql",
			"000079_production_external_evidence.up.sql",
		)
		seedExternalEvidenceReview(t, ctx, conn, "G7", "operations-owner", "operations", "2026-08-26T08:00:00Z")
		seedExternalEvidenceReview(t, ctx, conn, "G7", "compliance-owner", "compliance", "2026-08-26T08:00:00Z")
		executeMigrationFiles(t, ctx, conn, "000079_production_external_evidence.up.sql")

		assertExternalEvidenceDuplicateRejected(t, ctx, conn, "G7", "operations-owner")
		assertAppendOnlyMutationRejected(t, ctx, conn,
			"UPDATE production_external_evidence SET evidence_reference = 'tampered://evidence' WHERE gate = 'G7'",
		)
		assertAppendOnlyMutationRejected(t, ctx, conn,
			"DELETE FROM production_external_evidence WHERE gate = 'G7'",
		)
		assertExternalEvidenceRowCount(t, ctx, conn, "G7", 2)
	})

	t.Run("fresh bootstrap installs readiness contract", func(t *testing.T) {
		schema := fmt.Sprintf("production_evidence_bootstrap_test_%d", os.Getpid())
		dropPostgresEvidenceSchema(t, db, schema)
		t.Cleanup(func() { dropPostgresEvidenceSchema(t, db, schema) })
		if _, err := db.Exec("CREATE SCHEMA " + schema); err != nil {
			t.Fatalf("创建 PostgreSQL 测试 schema 失败: %v", err)
		}

		pgxDB, err := sql.Open("pgx", os.Getenv("LAW_OA_POSTGRES_TEST_DSN"))
		if err != nil {
			t.Fatalf("初始化 pgx PostgreSQL 连接失败: %v", err)
		}
		defer pgxDB.Close()
		conn, err := pgxDB.Conn(ctx)
		if err != nil {
			t.Fatalf("获取 pgx PostgreSQL 连接失败: %v", err)
		}
		defer conn.Close()
		setPostgresEvidenceSearchPath(t, ctx, conn, schema)
		gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: conn, DriverName: "pgx"}), &gorm.Config{DisableAutomaticPing: true})
		if err != nil {
			t.Fatalf("初始化 GORM PostgreSQL 连接失败: %v", err)
		}
		if err := database.BootstrapProductionSchema(gormDB.Session(&gorm.Session{NewDB: true, Logger: gormDB.Logger})); err != nil {
			t.Fatalf("fresh PostgreSQL bootstrap 失败: %v", err)
		}

		assertProductionSchemaVersion(t, ctx, conn, database.ProductionSchemaVersion)
		assertExternalEvidenceUniqueIndex(t, ctx, conn)
		seedExternalEvidenceReview(t, ctx, conn, "G2", "release-manager", "operations", "2026-08-27T08:00:00Z")
		assertExternalEvidenceDuplicateRejected(t, ctx, conn, "G2", "release-manager")
		assertAppendOnlyMutationRejected(t, ctx, conn,
			"UPDATE production_external_evidence SET evidence_reference = 'tampered://evidence' WHERE gate = 'G2'",
		)
		assertAppendOnlyMutationRejected(t, ctx, conn,
			"DELETE FROM production_external_evidence WHERE gate = 'G2'",
		)
		assertExternalEvidenceReadinessQuery(t, ctx, conn, "G2", "release-manager")
		assertExternalEvidenceRowCount(t, ctx, conn, "G2", 1)
	})
}

func executeMigrationFiles(t *testing.T, ctx context.Context, conn *sql.Conn, filenames ...string) {
	t.Helper()
	for _, filename := range filenames {
		path := filepath.Join("..", "..", "migrations", filename)
		statement, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("读取迁移文件 %s 失败: %v", filename, err)
		}
		if _, err := conn.ExecContext(ctx, string(statement)); err != nil {
			t.Fatalf("执行迁移文件 %s 失败: %v", filename, err)
		}
	}
}

func seedExternalEvidenceReview(t *testing.T, ctx context.Context, conn *sql.Conn, gate, reviewer, role, timestamp string) {
	t.Helper()
	canonical := strings.Join([]string{gate, "evidence://" + gate, reviewer, role, "PASSED", timestamp, timestamp}, "\n")
	hash := sha256.Sum256([]byte(canonical))
	query := `
		INSERT INTO production_external_evidence
			(gate, evidence_reference, reviewed_by, reviewer_role, review_result, reviewed_at, integrity_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'PASSED', $5, $6, $5, $5)`
	if _, err := conn.ExecContext(ctx, query, gate, "evidence://"+gate, reviewer, role, timestamp, hex.EncodeToString(hash[:])); err != nil {
		t.Fatalf("插入外部证据登记失败: %v", err)
	}
}

func dropPostgresEvidenceSchema(t *testing.T, db *sql.DB, schema string) {
	t.Helper()
	if _, err := db.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE"); err != nil {
		t.Errorf("清理 PostgreSQL 测试 schema %s 失败: %v", schema, err)
	}
}

func setPostgresEvidenceSearchPath(t *testing.T, ctx context.Context, conn *sql.Conn, schema string) {
	t.Helper()
	if _, err := conn.ExecContext(ctx, "SET search_path TO "+schema); err != nil {
		t.Fatalf("设置 PostgreSQL 测试 search_path 失败: %v", err)
	}
}

func assertProductionSchemaVersion(t *testing.T, ctx context.Context, conn *sql.Conn, expected string) {
	t.Helper()
	var actual string
	if err := conn.QueryRowContext(ctx, "SELECT version FROM schema_bootstrap_state WHERE id = 1").Scan(&actual); err != nil {
		t.Fatalf("读取 fresh bootstrap 版本失败: %v", err)
	}
	if actual != expected {
		t.Fatalf("fresh bootstrap 版本期望 %s，实际 %s", expected, actual)
	}
}

func assertExternalEvidenceUniqueIndex(t *testing.T, ctx context.Context, conn *sql.Conn) {
	t.Helper()
	var definition string
	err := conn.QueryRowContext(ctx, `
		SELECT indexdef
		FROM pg_indexes
		WHERE schemaname = current_schema()
		  AND tablename = 'production_external_evidence'
		  AND indexname = 'uq_production_external_evidence_gate_reviewer'
	`).Scan(&definition)
	if err != nil {
		t.Fatalf("读取 production_external_evidence 唯一索引失败: %v", err)
	}
	actual := strings.ToLower(definition)
	for _, term := range []string{"unique", "gate", "reviewed_by"} {
		if !strings.Contains(actual, term) {
			t.Fatalf("唯一索引定义缺少 %s: %s", term, definition)
		}
	}
}

func assertExternalEvidenceDuplicateRejected(t *testing.T, ctx context.Context, conn *sql.Conn, gate, reviewer string) {
	t.Helper()
	canonical := strings.Join([]string{gate, "evidence://" + gate + "-duplicate", reviewer, "operations", "PASSED", "2026-08-27T08:00:00Z", "2026-08-27T08:00:00Z"}, "\n")
	hash := sha256.Sum256([]byte(canonical))
	query := `
		INSERT INTO production_external_evidence
			(gate, evidence_reference, reviewed_by, reviewer_role, review_result, reviewed_at, integrity_hash, created_at, updated_at)
		VALUES ($1, $2, $3, 'operations', 'PASSED', $4, $5, $4, $4)`
	_, err := conn.ExecContext(ctx, query, gate, "evidence://"+gate+"-duplicate", reviewer, "2026-08-27T08:00:00Z", hex.EncodeToString(hash[:]))
	if err == nil {
		t.Fatalf("重复 (gate, reviewed_by) 登记未被唯一索引拒绝")
	}
}

func assertExternalEvidenceRowCount(t *testing.T, ctx context.Context, conn *sql.Conn, gate string, expected int) {
	t.Helper()
	var actual int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM production_external_evidence WHERE gate = $1", gate).Scan(&actual); err != nil {
		t.Fatalf("读取外部证据登记失败: %v", err)
	}
	if actual != expected {
		t.Fatalf("gate %s append-only 记录数期望 %d，实际 %d", gate, expected, actual)
	}
}

func assertExternalEvidenceReadinessQuery(t *testing.T, ctx context.Context, conn *sql.Conn, gate, reviewer string) {
	t.Helper()
	var actualGate, actualReviewer string
	err := conn.QueryRowContext(ctx, `
		SELECT gate, reviewed_by
		FROM production_external_evidence
		WHERE gate = $1 AND reviewed_by = $2
	`, gate, reviewer).Scan(&actualGate, &actualReviewer)
	if err != nil {
		t.Fatalf("readiness 查询 production_external_evidence 失败: %v", err)
	}
	if actualGate != gate || actualReviewer != reviewer {
		t.Fatalf("readiness 查询返回 (%s, %s)，期望 (%s, %s)", actualGate, actualReviewer, gate, reviewer)
	}
}

func assertAppendOnlyMutationRejected(t *testing.T, ctx context.Context, conn *sql.Conn, statement string) {
	t.Helper()
	if _, err := conn.ExecContext(ctx, statement); err == nil {
		t.Fatalf("append-only 语句未被拒绝: %s", statement)
	} else if !strings.Contains(err.Error(), "append-only evidence cannot be updated or deleted") {
		t.Fatalf("append-only 语句返回意外错误: %v", err)
	}
}
