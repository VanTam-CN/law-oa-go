DROP TRIGGER IF EXISTS trg_law_firm_policy_profiles_append_only
    ON law_firm_compliance_policy_profiles;
DROP INDEX IF EXISTS idx_law_firm_compliance_policy_status;
DROP TABLE IF EXISTS law_firm_compliance_policy_profiles;

ALTER TABLE conflict_search_scopes
    DROP COLUMN IF EXISTS correction_procedure_reference,
    DROP COLUMN IF EXISTS failure_alert_reference,
    DROP COLUMN IF EXISTS max_quality_review_age_days,
    DROP COLUMN IF EXISTS quality_reviewed_at,
    DROP COLUMN IF EXISTS quality_owner_id,
    DROP COLUMN IF EXISTS measured_duplicate_rate_bps,
    DROP COLUMN IF EXISTS maximum_duplicate_rate_bps,
    DROP COLUMN IF EXISTS measured_field_coverage_bps,
    DROP COLUMN IF EXISTS minimum_field_coverage_bps,
    DROP COLUMN IF EXISTS last_successful_sync_at,
    DROP COLUMN IF EXISTS max_sync_lag_minutes,
    DROP COLUMN IF EXISTS sync_mode,
    DROP COLUMN IF EXISTS source_of_truth;
