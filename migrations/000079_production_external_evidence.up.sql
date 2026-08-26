CREATE TABLE IF NOT EXISTS production_external_evidence (
    id BIGSERIAL PRIMARY KEY,
    gate VARCHAR(8) NOT NULL UNIQUE,
    evidence_reference TEXT NOT NULL,
    reviewed_by VARCHAR(120) NOT NULL,
    reviewer_role VARCHAR(80) NOT NULL,
    review_result VARCHAR(20) NOT NULL,
    reviewed_at TEXT NOT NULL,
    integrity_hash VARCHAR(64) NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CONSTRAINT chk_production_external_gate CHECK (gate IN ('G0','G1','G2','G3','G4','G5','G6','G7')),
    CONSTRAINT chk_production_external_result CHECK (review_result IN ('PASSED','FAILED')),
    CONSTRAINT chk_production_external_review_time CHECK (reviewed_at <= updated_at)
);

DROP TRIGGER IF EXISTS trg_production_external_evidence_append_only
    ON production_external_evidence;
CREATE TRIGGER trg_production_external_evidence_append_only
    BEFORE UPDATE OR DELETE
    ON production_external_evidence
    FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation();
