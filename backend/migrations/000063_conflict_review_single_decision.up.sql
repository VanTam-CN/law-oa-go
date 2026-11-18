-- A check is a frozen evidence snapshot and may have at most one professional
-- conclusion. Existing duplicate rows must be reconciled by the firm before
-- this constraint can be installed; they must never be physically deleted by
-- a migration.
CREATE UNIQUE INDEX IF NOT EXISTS uq_conflict_reviews_check_id
    ON conflict_reviews (check_id);
