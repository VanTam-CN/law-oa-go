DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = current_schema() AND table_name = 'conflict_check_records'
    ) THEN
        DROP INDEX IF EXISTS idx_conflict_check_records_status_updated;

        ALTER TABLE conflict_check_records
            DROP CONSTRAINT IF EXISTS chk_conflict_check_records_status;

        ALTER TABLE conflict_check_records
            ADD CONSTRAINT chk_conflict_check_records_status
            CHECK (check_status IN ('PROCESSING', 'COMPLETED', 'FAILED'));

        ALTER TABLE conflict_check_records
            ALTER COLUMN check_status SET DEFAULT 'PROCESSING';
    END IF;
END $$;
