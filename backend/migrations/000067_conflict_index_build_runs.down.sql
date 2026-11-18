DROP INDEX idx_conflict_search_scopes_index_run;
ALTER TABLE conflict_search_scopes DROP COLUMN index_run_id;
DROP TABLE conflict_index_build_runs;
