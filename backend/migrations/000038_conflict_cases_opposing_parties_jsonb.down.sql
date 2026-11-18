DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'conflict_cases'
          AND column_name = 'opposing_parties'
          AND udt_name = 'jsonb'
    ) THEN
        ALTER TABLE conflict_cases
            ALTER COLUMN opposing_parties DROP DEFAULT,
            ALTER COLUMN opposing_parties DROP NOT NULL,
            ALTER COLUMN opposing_parties TYPE TEXT[]
            USING ARRAY[]::TEXT[];
    END IF;
END $$;
