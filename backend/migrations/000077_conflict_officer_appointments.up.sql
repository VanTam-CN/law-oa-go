CREATE TABLE IF NOT EXISTS conflict_officer_appointments (
    id VARCHAR(36) PRIMARY KEY,
    officer_id BIGINT NOT NULL,
    deputy_id BIGINT NULL,
    appointed_by BIGINT NOT NULL,
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ NOT NULL,
    recusal_declaration TEXT NOT NULL,
    external_mechanism_reference TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_conflict_officer_term CHECK (effective_to > effective_from),
    CONSTRAINT chk_conflict_officer_separation CHECK (
        officer_id <> appointed_by
        AND (deputy_id IS NULL OR (deputy_id <> officer_id AND deputy_id <> appointed_by))
    )
);

CREATE INDEX IF NOT EXISTS idx_conflict_officer_appointment_term
    ON conflict_officer_appointments (officer_id, effective_from, effective_to);

DROP TRIGGER IF EXISTS trg_conflict_officer_appointments_append_only
    ON conflict_officer_appointments;
CREATE TRIGGER trg_conflict_officer_appointments_append_only
    BEFORE UPDATE OR DELETE ON conflict_officer_appointments
    FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation();
