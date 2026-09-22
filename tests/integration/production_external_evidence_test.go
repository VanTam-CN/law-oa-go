package integration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
	dsn := os.Getenv("LAW_OA_POSTGRES_TEST_DSN")
	if strings.Contains(dsn, "?") && strings.Contains(strings.ToLower(dsn), "sslmode=") == false {
		t.Fatalf("LAW_OA_POSTGRES_TEST_DSN must include an explicit sslmode setting")
	}
	db, err := sql.Open("pgx", dsn)
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

	t.Run("v14 bootstrap state is not silently upgraded", func(t *testing.T) {
		schema := fmt.Sprintf("production_evidence_v14_state_test_%d", os.Getpid())
		dropPostgresEvidenceSchema(t, db, schema)
		t.Cleanup(func() { dropPostgresEvidenceSchema(t, db, schema) })
		if _, err := db.Exec("CREATE SCHEMA " + schema); err != nil {
			t.Fatalf("创建 PostgreSQL 测试 schema 失败: %v", err)
		}
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("获取 PostgreSQL 测试连接失败: %v", err)
		}
		defer conn.Close()
		setPostgresEvidenceSearchPath(t, ctx, conn, schema)
		if _, err := conn.ExecContext(ctx, `
CREATE TABLE schema_bootstrap_state (
	id INTEGER PRIMARY KEY,
	version VARCHAR(120) NOT NULL,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO schema_bootstrap_state (id, version) VALUES (1, 'postgres-mvp-2026-08-27-v14');
		`); err != nil {
			t.Fatalf("预置 v14 bootstrap 状态失败: %v", err)
		}

		gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: conn, DriverName: "pgx"}), &gorm.Config{PrepareStmt: true, DisableAutomaticPing: true})
		if err != nil {
			t.Fatalf("初始化 GORM PostgreSQL 连接失败: %v", err)
		}
		err = database.BootstrapProductionSchema(gormDB.Session(&gorm.Session{NewDB: true, Logger: gormDB.Logger}))
		if err == nil || !strings.Contains(err.Error(), "拒绝升级生产 schema bootstrap") {
			t.Fatalf("v14 bootstrap 状态未被明确拒绝，错误为: %v", err)
		}
		var version string
		if err := conn.QueryRowContext(ctx, "SELECT version FROM schema_bootstrap_state WHERE id = 1").Scan(&version); err != nil {
			t.Fatalf("复读 v14 bootstrap 状态失败: %v", err)
		}
		if version != "postgres-mvp-2026-08-27-v14" {
			t.Fatalf("v14 bootstrap 状态被改写为 %s", version)
		}
	})

	t.Run("fresh bootstrap installs readiness contract", func(t *testing.T) {
		schema := fmt.Sprintf("production_evidence_bootstrap_test_%d", os.Getpid())
		dropPostgresEvidenceSchema(t, db, schema)
		t.Cleanup(func() { dropPostgresEvidenceSchema(t, db, schema) })
		if _, err := db.Exec("CREATE SCHEMA " + schema); err != nil {
			t.Fatalf("创建 PostgreSQL 测试 schema 失败: %v", err)
		}

		pgxDB, err := sql.Open("pgx", dsn)
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
		gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: conn, DriverName: "pgx"}), &gorm.Config{PrepareStmt: true, DisableAutomaticPing: true})
		if err != nil {
			t.Fatalf("初始化 GORM PostgreSQL 连接失败: %v", err)
		}
		if err := database.BootstrapProductionSchema(gormDB.Session(&gorm.Session{NewDB: true, Logger: gormDB.Logger})); err != nil {
			t.Fatalf("fresh PostgreSQL bootstrap 失败: %v", err)
		}
		if err := database.BootstrapProductionSchema(gormDB.Session(&gorm.Session{NewDB: true, Logger: gormDB.Logger})); err != nil {
			t.Fatalf("同版本幂等 bootstrap 失败: %v", err)
		}

		assertProductionSchemaVersion(t, ctx, conn, database.ProductionSchemaVersion)
		assertExternalEvidenceUniqueIndex(t, ctx, conn)
		assertOperationsReadinessEvidenceContract(t, ctx, db, schema)
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

	t.Run("operations evidence migration and v15 bootstrap are equivalent", func(t *testing.T) {
		if database.ProductionSchemaVersion != "postgres-mvp-2026-08-27-v15" {
			t.Fatalf("production schema version must remain deterministic, got %s", database.ProductionSchemaVersion)
		}

		migrationSchema := fmt.Sprintf("operations_evidence_migration_v15_%d", os.Getpid())
		bootstrapSchema := fmt.Sprintf("operations_evidence_bootstrap_v15_%d", os.Getpid())
		dropPostgresEvidenceSchema(t, db, migrationSchema)
		dropPostgresEvidenceSchema(t, db, bootstrapSchema)
		t.Cleanup(func() {
			dropPostgresEvidenceSchema(t, db, migrationSchema)
			dropPostgresEvidenceSchema(t, db, bootstrapSchema)
		})
		if _, err := db.Exec("CREATE SCHEMA " + migrationSchema); err != nil {
			t.Fatalf("创建迁移测试 schema 失败: %v", err)
		}
		if _, err := db.Exec("CREATE SCHEMA " + bootstrapSchema); err != nil {
			t.Fatalf("创建 bootstrap 测试 schema 失败: %v", err)
		}

		migrationConn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("获取迁移测试连接失败: %v", err)
		}
		defer migrationConn.Close()
		setPostgresEvidenceSearchPath(t, ctx, migrationConn, migrationSchema)
		for _, table := range []string{"compliance_audit_events", "conflict_reviews", "case_subject_revisions"} {
			if _, err := migrationConn.ExecContext(ctx, "CREATE TABLE "+table+" (id INTEGER PRIMARY KEY)"); err != nil {
				t.Fatalf("创建迁移前置测试表 %s 失败: %v", table, err)
			}
		}
		executeMigrationFiles(t, ctx, migrationConn,
			"000062_append_only_evidence_guards.up.sql",
			"000080_operations_readiness_evidence.up.sql",
		)
		migrationShape := readOperationsEvidenceShape(t, ctx, db, migrationSchema)

		pgxDB, err := sql.Open("pgx", dsn)
		if err != nil {
			t.Fatalf("初始化 pgx PostgreSQL 连接失败: %v", err)
		}
		defer pgxDB.Close()
		bootstrapConn, err := pgxDB.Conn(ctx)
		if err != nil {
			t.Fatalf("获取 bootstrap pgx 连接失败: %v", err)
		}
		defer bootstrapConn.Close()
		setPostgresEvidenceSearchPath(t, ctx, bootstrapConn, bootstrapSchema)
		gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: bootstrapConn, DriverName: "pgx"}), &gorm.Config{PrepareStmt: true, DisableAutomaticPing: true})
		if err != nil {
			t.Fatalf("初始化 GORM PostgreSQL 连接失败: %v", err)
		}
		if err := database.BootstrapProductionSchema(gormDB.Session(&gorm.Session{NewDB: true, Logger: gormDB.Logger})); err != nil {
			t.Fatalf("fresh PostgreSQL bootstrap 失败: %v", err)
		}
		bootstrapShape := readOperationsEvidenceShape(t, ctx, db, bootstrapSchema)

		if !reflect.DeepEqual(migrationShape, bootstrapShape) {
			t.Fatalf("fresh bootstrap 与 migration 000080 的 operations_readiness_evidence 关键 schema 不一致:\nmigration: %#v\nbootstrap: %#v", migrationShape, bootstrapShape)
		}
		assertAppendOnlyMutationRejected(t, ctx, migrationConn,
			"INSERT INTO operations_readiness_evidence (id, control, scope, result, evidence_reference, reviewed_by, reviewed_at, created_at) VALUES ('migration-trigger-check', 'backup', 'qa', 'passed', 'qa://migration', 1, '2026-08-27T08:00:00Z', '2026-08-27T08:00:00Z'); UPDATE operations_readiness_evidence SET notes = 'tampered' WHERE id = 'migration-trigger-check'",
		)
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

type operationsEvidenceShape struct {
	Columns     map[string]string
	PrimaryKey  []string
	Indexes     map[string]string
	Constraints map[string]string
	Trigger     string
}

func assertOperationsReadinessEvidenceContract(t *testing.T, ctx context.Context, db *sql.DB, schema string) {
	t.Helper()
	shape := readOperationsEvidenceShape(t, ctx, db, schema)
	for _, name := range []string{
		"id", "control", "scope", "result", "evidence_reference", "reviewed_by", "reviewed_at",
		"notes", "previous_evidence_id", "integrity_hash", "created_at",
	} {
		if _, ok := shape.Columns[name]; !ok {
			t.Fatalf("operations_readiness_evidence 缺少关键列 %s", name)
		}
	}
	for _, name := range []string{
		"idx_operations_evidence_reviewer", "idx_operations_evidence_control_scope_time",
	} {
		if _, ok := shape.Indexes[name]; !ok {
			t.Fatalf("operations_readiness_evidence 缺少关键索引 %s", name)
		}
	}
	for _, name := range []string{
		"chk_operations_evidence_scope", "chk_operations_evidence_result", "chk_operations_evidence_reference",
		"chk_operations_evidence_reference_scheme", "chk_operations_evidence_chain", "chk_operations_evidence_id_link",
	} {
		if _, ok := shape.Constraints[name]; !ok {
			t.Fatalf("operations_readiness_evidence 缺少关键约束 %s", name)
		}
	}
	if shape.Trigger != "trg_operations_readiness_evidence_append_only" {
		t.Fatalf("operations_readiness_evidence 缺少 append-only trigger，实际 %q", shape.Trigger)
	}
}

func readOperationsEvidenceShape(t *testing.T, ctx context.Context, db *sql.DB, schema string) operationsEvidenceShape {
	t.Helper()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("获取 schema 结构读取连接失败: %v", err)
	}
	defer conn.Close()
	setPostgresEvidenceSearchPath(t, ctx, conn, schema)
	shape := operationsEvidenceShape{
		Columns:     map[string]string{},
		PrimaryKey:  []string{},
		Indexes:     map[string]string{},
		Constraints: map[string]string{},
	}
	rows, err := conn.QueryContext(ctx, `
SELECT column_name, data_type, coalesce(character_maximum_length::text, '')
FROM information_schema.columns
WHERE table_schema = current_schema() AND table_name = 'operations_readiness_evidence'
`)
	if err != nil {
		t.Fatalf("读取 operations_readiness_evidence 列结构失败: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, dataType, length string
		if err := rows.Scan(&name, &dataType, &length); err != nil {
			t.Fatalf("解析 operations_readiness_evidence 列结构失败: %v", err)
		}
		shape.Columns[name] = strings.TrimSpace(dataType + " " + length)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("遍历 operations_readiness_evidence 列结构失败: %v", err)
	}

	rows, err = conn.QueryContext(ctx, `
SELECT a.attname
FROM pg_index i
JOIN pg_class c ON c.oid = i.indrelid
JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY(i.indkey)
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = current_schema() AND c.relname = 'operations_readiness_evidence' AND i.indisprimary
ORDER BY a.attname
`)
	if err != nil {
		t.Fatalf("读取 operations_readiness_evidence 主键失败: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("解析 operations_readiness_evidence 主键失败: %v", err)
		}
		shape.PrimaryKey = append(shape.PrimaryKey, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("遍历 operations_readiness_evidence 主键失败: %v", err)
	}

	rows, err = conn.QueryContext(ctx, `
SELECT indexname, lower(indexdef)
FROM pg_indexes
WHERE schemaname = current_schema() AND tablename = 'operations_readiness_evidence'
ORDER BY indexname
`)
	if err != nil {
		t.Fatalf("读取 operations_readiness_evidence 索引失败: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			t.Fatalf("解析 operations_readiness_evidence 索引失败: %v", err)
		}
		definition = strings.Replace(definition, schema+".", "", 1)
		shape.Indexes[name] = strings.Join(strings.Fields(definition), " ")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("遍历 operations_readiness_evidence 索引失败: %v", err)
	}

	rows, err = conn.QueryContext(ctx, `
SELECT conname, pg_get_constraintdef(oid, true)
FROM pg_constraint
WHERE connamespace = current_schema()::regnamespace
  AND conrelid = 'operations_readiness_evidence'::regclass
  AND contype = 'c'
ORDER BY conname
`)
	if err != nil {
		t.Fatalf("读取 operations_readiness_evidence 约束失败: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			t.Fatalf("解析 operations_readiness_evidence 约束失败: %v", err)
		}
		shape.Constraints[name] = strings.Join(strings.Fields(definition), " ")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("遍历 operations_readiness_evidence 约束失败: %v", err)
	}

	if err := conn.QueryRowContext(ctx, `
SELECT t.tgname
FROM pg_trigger t
JOIN pg_class c ON c.oid = t.tgrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = current_schema()
  AND c.relname = 'operations_readiness_evidence'
  AND t.tgname = 'trg_operations_readiness_evidence_append_only'
  AND NOT t.tgisinternal
`).Scan(&shape.Trigger); err != nil {
		t.Fatalf("读取 operations_readiness_evidence append-only trigger 失败: %v", err)
	}
	return shape
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
