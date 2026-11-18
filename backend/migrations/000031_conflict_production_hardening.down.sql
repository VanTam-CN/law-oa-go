ALTER TABLE conflict_cases DROP CONSTRAINT IF EXISTS fk_conflict_cases_check;
ALTER TABLE conflict_cases DROP CONSTRAINT IF EXISTS chk_conflict_cases_risk_level;
ALTER TABLE conflict_check_records DROP CONSTRAINT IF EXISTS chk_conflict_check_records_status;
ALTER TABLE conflict_check_records DROP CONSTRAINT IF EXISTS chk_conflict_check_records_risk_level;

DROP INDEX IF EXISTS idx_client_relations_related;
DROP INDEX IF EXISTS idx_client_relations_client_active;
DROP INDEX IF EXISTS idx_conflict_cases_client_risk;
DROP INDEX IF EXISTS idx_conflict_cases_check_id;
DROP INDEX IF EXISTS idx_conflict_check_records_risk;
DROP INDEX IF EXISTS idx_conflict_check_records_status_updated;
DROP INDEX IF EXISTS idx_conflict_check_records_client_time;

DROP TABLE IF EXISTS conflict_cases;
