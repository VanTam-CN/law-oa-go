-- Prevent physical deletion of conflict evidence while keeping workflow state
-- transitions available where the application needs them.
DO $$
DECLARE
    target RECORD;
BEGIN
    FOR target IN
        SELECT * FROM (VALUES
            ('trg_conflict_checks_no_delete', 'conflict_checks', 'DELETE'),
            ('trg_conflict_details_append_only', 'conflict_details', 'UPDATE OR DELETE'),
            ('trg_conflict_check_records_no_delete', 'conflict_check_records', 'DELETE'),
            ('trg_conflict_cases_no_delete', 'conflict_cases', 'DELETE'),
            ('trg_conflict_reviewer_assignments_no_delete', 'conflict_reviewer_assignments', 'DELETE')
        ) AS guards(trigger_name, table_name, operation)
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I', target.trigger_name, target.table_name);
        EXECUTE format(
            'CREATE TRIGGER %I BEFORE %s ON %I FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation()',
            target.trigger_name, target.operation, target.table_name
        );
    END LOOP;
END $$;
