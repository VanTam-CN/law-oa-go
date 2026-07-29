CREATE TABLE IF NOT EXISTS conflict_reviews (
    id BIGSERIAL PRIMARY KEY,
    check_id VARCHAR(100) NOT NULL,
    decision VARCHAR(40) NOT NULL,
    notes TEXT NOT NULL,
    reviewer_id BIGINT NOT NULL,
    reviewer_name VARCHAR(100) NOT NULL,
    evidence_hash VARCHAR(64) NOT NULL,
    next_review_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_conflict_review_decision CHECK (decision IN (
        'no_conflict', 'confirmed_conflict', 'false_positive',
        'insufficient_information', 'waiver_requested'
    ))
);

CREATE INDEX IF NOT EXISTS idx_conflict_reviews_check_created
    ON conflict_reviews (check_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_conflict_reviews_reviewer
    ON conflict_reviews (reviewer_id, created_at DESC);

COMMENT ON TABLE conflict_reviews IS 'Immutable professional review conclusions over frozen conflict evidence';
COMMENT ON COLUMN conflict_reviews.evidence_hash IS 'SHA-256 of the evidence snapshot reviewed by the professional';
