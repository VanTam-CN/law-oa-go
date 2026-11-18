-- Approval timestamps are legal audit evidence. Legacy schemas used
-- timestamp without time zone while the application wrote China wall-clock
-- values, causing browsers to add another eight hours. Convert only legacy
-- columns and interpret their existing values as Asia/Shanghai time.
DO $$
DECLARE
    target RECORD;
BEGIN
    FOR target IN
        SELECT table_name, column_name
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND data_type = 'timestamp without time zone'
          AND (
            (table_name = 'approval_delegations' AND column_name IN ('valid_from', 'valid_until'))
            OR (table_name = 'approval_notifications' AND column_name IN ('scheduled_at', 'sent_at', 'read_at', 'created_at', 'updated_at'))
            OR (table_name = 'approval_records' AND column_name IN ('approval_date', 'effective_date', 'next_review_date', 'created_at', 'updated_at'))
            OR (table_name = 'approval_requests' AND column_name IN ('expected_effective_date', 'expected_expiry_date', 'submission_date', 'created_at', 'updated_at', 'deleted_at'))
            OR (table_name = 'approval_templates' AND column_name IN ('last_used_date', 'created_at', 'updated_at'))
            OR (table_name = 'approval_workflows' AND column_name IN ('created_at', 'updated_at'))
          )
    LOOP
        EXECUTE format(
            'ALTER TABLE %I ALTER COLUMN %I TYPE TIMESTAMPTZ USING %I AT TIME ZONE %L',
            target.table_name,
            target.column_name,
            target.column_name,
            'Asia/Shanghai'
        );
    END LOOP;
END $$;
