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
	"time"

	_ "github.com/lib/pq"
)

func TestPostgreSQL_ProductionExternalEvidenceAppendOnly(t *testing.T) {
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

	// Run the exact migration SQL in an isolated schema. Migration 000062
	// refreshes triggers on older evidence tables, so minimal stand-ins keep
	// this test independent of unrelated objects in the dedicated database.
	schema := fmt.Sprintf("evidence_append_only_test_%d", os.Getpid())
	_, err = db.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE")
	if err != nil {
		t.Fatalf("清理专用 PostgreSQL 测试 schema 失败: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE"); err != nil {
			t.Fatalf("清理专用 PostgreSQL 测试 schema 失败: %v", err)
		}
	})
	if _, err := db.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatalf("创建专用 PostgreSQL 测试 schema 失败: %v", err)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("获取 PostgreSQL 连接失败: %v", err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, "SET search_path TO "+schema); err != nil {
		t.Fatalf("设置 PostgreSQL 测试 search_path 失败: %v", err)
	}
	for _, table := range []string{"compliance_audit_events", "conflict_reviews", "case_subject_revisions"} {
		if _, err := conn.ExecContext(ctx, "CREATE TABLE "+table+" (id INTEGER PRIMARY KEY)"); err != nil {
			t.Fatalf("创建迁移前置测试表 %s 失败: %v", table, err)
		}
	}
	executeMigrationFiles(t, ctx, conn,
		"000062_append_only_evidence_guards.up.sql",
		"000079_production_external_evidence.up.sql",
	)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	seedExternalEvidenceReview(t, ctx, conn, "G7", "operations-owner", "operations", now)
	seedExternalEvidenceReview(t, ctx, conn, "G7", "compliance-owner", "compliance", now)

	assertAppendOnlyMutationRejected(t, ctx, conn,
		"UPDATE production_external_evidence SET evidence_reference = 'tampered://evidence' WHERE gate = 'G7'",
	)
	assertAppendOnlyMutationRejected(t, ctx, conn,
		"DELETE FROM production_external_evidence WHERE gate = 'G7'",
	)

	var count int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM production_external_evidence WHERE gate = 'G7'").Scan(&count); err != nil {
		t.Fatalf("读取外部证据登记失败: %v", err)
	}
	if count != 2 {
		t.Fatalf("append-only 记录数应保持为 2，实际为 %d", count)
	}
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

func assertAppendOnlyMutationRejected(t *testing.T, ctx context.Context, conn *sql.Conn, statement string) {
	t.Helper()
	if _, err := conn.ExecContext(ctx, statement); err == nil {
		t.Fatalf("append-only 语句未被拒绝: %s", statement)
	} else if !strings.Contains(err.Error(), "append-only evidence cannot be updated or deleted") {
		t.Fatalf("append-only 语句返回意外错误: %v", err)
	}
}
