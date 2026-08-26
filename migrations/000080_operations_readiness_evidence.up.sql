CREATE TABLE IF NOT EXISTS operations_readiness_evidence (
    id VARCHAR(36) PRIMARY KEY,
    control VARCHAR(40) NOT NULL,
    scope VARCHAR(30) NOT NULL,
    result VARCHAR(20) NOT NULL,
    evidence_reference TEXT NOT NULL,
    reviewed_by BIGINT NOT NULL,
    reviewed_at TIMESTAMPTZ NOT NULL,
    notes TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_operations_evidence_control_scope UNIQUE (control, scope),
    CONSTRAINT chk_operations_evidence_scope CHECK (scope IN ('qa', 'controlled_pilot')),
    CONSTRAINT chk_operations_evidence_result CHECK (result = 'passed'),
    CONSTRAINT chk_operations_evidence_reference CHECK (length(trim(evidence_reference)) BETWEEN 8 AND 1000)
);

CREATE INDEX IF NOT EXISTS idx_operations_evidence_reviewer ON operations_readiness_evidence (reviewed_by);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_proc p
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
