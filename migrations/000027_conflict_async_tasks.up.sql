-- 异步冲突检测任务状态扩展（PostgreSQL）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = current_schema() AND table_name = 'conflict_check_records'
    ) THEN
        ALTER TABLE conflict_check_records
            ALTER COLUMN check_status TYPE VARCHAR(20),
            ALTER COLUMN check_status SET DEFAULT 'QUEUED';

        ALTER TABLE conflict_check_records
            DROP CONSTRAINT IF EXISTS chk_conflict_check_records_status;

        ALTER TABLE conflict_check_records
            ADD CONSTRAINT chk_conflict_check_records_status
            CHECK (check_status IN ('QUEUED', 'RUNNING', 'PROCESSING', 'COMPLETED', 'FAILED'));

        CREATE INDEX IF NOT EXISTS idx_conflict_check_records_status_updated
            ON conflict_check_records(check_status, updated_at);
    END IF;
END $$;
