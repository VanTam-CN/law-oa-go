CREATE TABLE IF NOT EXISTS operations_readiness_evidence (
    id VARCHAR(36) PRIMARY KEY,
    control VARCHAR(40) NOT NULL,
    scope VARCHAR(30) NOT NULL,
    result VARCHAR(20) NOT NULL,
    evidence_reference TEXT NOT NULL,
    reviewed_by BIGINT NOT NULL,
    reviewed_at TIMESTAMPTZ NOT NULL,
    notes TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Existing QA databases can already contain the pre-hash-chain table. Upgrade
-- the storage contract without rewriting append-only evidence rows.
-- The reference constraints are NOT VALID so historical evidence:// values and
-- the former 1000-character limit remain preserved; PostgreSQL still enforces
-- both checks on every newly inserted or updated row.
ALTER TABLE operations_readiness_evidence
    ADD COLUMN IF NOT EXISTS previous_evidence_id VARCHAR(36) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS integrity_hash VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_operations_evidence_reviewer
    ON operations_readiness_evidence (reviewed_by);
CREATE INDEX IF NOT EXISTS idx_operations_evidence_control_scope_time
    ON operations_readiness_evidence (control, scope, reviewed_at, created_at);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_proc p
        JOIN pg_namespace n ON n.oid = p.pronamespace
        WHERE n.nspname = current_schema()
          AND p.proname = 'law_oa_reject_append_only_mutation'
    ) THEN
        CREATE FUNCTION law_oa_reject_append_only_mutation()
        RETURNS trigger
        LANGUAGE plpgsql
        AS $fn$
        BEGIN
            RAISE EXCEPTION 'append-only evidence cannot be updated or deleted: %', TG_TABLE_NAME;
        END;
        $fn$;
    END IF;
END;
$$;

DROP TRIGGER IF EXISTS trg_operations_readiness_evidence_append_only
    ON operations_readiness_evidence;
CREATE TRIGGER trg_operations_readiness_evidence_append_only
    BEFORE UPDATE OR DELETE ON operations_readiness_evidence
    FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation();

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_operations_evidence_scope'
          AND conrelid = 'operations_readiness_evidence'::regclass
    ) THEN
        ALTER TABLE operations_readiness_evidence
            ADD CONSTRAINT chk_operations_evidence_scope
            CHECK (scope IN ('qa', 'controlled_pilot'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_operations_evidence_result'
          AND conrelid = 'operations_readiness_evidence'::regclass
    ) THEN
        ALTER TABLE operations_readiness_evidence
            ADD CONSTRAINT chk_operations_evidence_result
            CHECK (result = 'passed');
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_operations_evidence_reference'
          AND conrelid = 'operations_readiness_evidence'::regclass
    ) THEN
        ALTER TABLE operations_readiness_evidence
            ADD CONSTRAINT chk_operations_evidence_reference
            CHECK (length(trim(evidence_reference)) BETWEEN 8 AND 512)
            NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_operations_evidence_reference_scheme'
          AND conrelid = 'operations_readiness_evidence'::regclass
    ) THEN
        ALTER TABLE operations_readiness_evidence
            ADD CONSTRAINT chk_operations_evidence_reference_scheme CHECK (
                evidence_reference ~ '^(archive|ticket|qa|controlled-pilot)://[^/[:space:]]+'
                OR evidence_reference ~ '^https://[^/[:space:]]+/[^[:space:]]+$'
            ) NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_operations_evidence_chain'
          AND conrelid = 'operations_readiness_evidence'::regclass
    ) THEN
        ALTER TABLE operations_readiness_evidence
            ADD CONSTRAINT chk_operations_evidence_chain CHECK (
                (integrity_hash = '' AND previous_evidence_id = '')
                OR (
                    integrity_hash ~ '^[0-9a-f]{64}$'
                    AND previous_evidence_id ~ '^[0-9a-f]{32}([0-9a-f]{4})?$'
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_operations_evidence_id_link'
          AND conrelid = 'operations_readiness_evidence'::regclass
    ) THEN
        ALTER TABLE operations_readiness_evidence
            ADD CONSTRAINT chk_operations_evidence_id_link
            CHECK (previous_evidence_id = '' OR previous_evidence_id <> id);
    END IF;
END;
$$;
