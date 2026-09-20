-- Durable duplicate-approval guard: one intake can be bound to at most one
-- approval through the case-creation path. A plain unique index over the
-- nullable intake_id lets any number of NULL rows through while allowing at
-- most one non-NULL value per intake; that NULL semantics was verified
-- identical on MySQL 8.0.46 and PostgreSQL 15. The MySQL migration path in
-- this repository is fail-closed by design (validateMigrationFiles), so this
-- file targets the PostgreSQL migration driver; on that driver CREATE UNIQUE
-- INDEX IF NOT EXISTS is legal and idempotent.
CREATE UNIQUE INDEX IF NOT EXISTS idx_case_creation_tracking_intake_unique
    ON approval_case_creation_tracking (intake_id);
