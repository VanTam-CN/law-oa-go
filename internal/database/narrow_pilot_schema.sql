-- Shared by the offline migration and the router before it registers business routes.
WITH required_columns(table_name, column_name, type_name, required) AS (
    VALUES
        ('case_intake_source_snapshots', 'id', 'bigint', true),
        ('case_intake_source_snapshots', 'intake_id', 'uuid', true),
        ('case_intake_source_snapshots', 'saved_revision', 'text', true),
        ('case_intake_source_snapshots', 'serverpack', 'jsonb', true),
        ('case_intake_source_snapshots', 'created_at', 'timestamp with time zone', true),
        ('case_intake_write_requests', 'id', 'bigint', true),
        ('case_intake_write_requests', 'actor_id', 'text', true),
        ('case_intake_write_requests', 'request_id', 'uuid', true),
        ('case_intake_write_requests', 'action', 'text', true),
        ('case_intake_write_requests', 'intake_id', 'uuid', true),
        ('case_intake_write_requests', 'state', 'text', true),
        ('case_intake_write_requests', 'request_digest', 'text', false),
        ('case_intake_write_requests', 'result_revision', 'text', false),
        ('case_intake_write_requests', 'source_revision', 'text', false),
        ('case_intake_write_requests', 'created_at', 'timestamp with time zone', true)
), required_constraints(table_name, constraint_name, definition) AS (
    VALUES
        ('case_intake_source_snapshots', 'case_intake_source_snapshots_pkey', 'PRIMARY KEY (id)'),
        ('case_intake_write_requests', 'case_intake_write_requests_pkey', 'PRIMARY KEY (id)'),
        ('case_intake_source_snapshots', 'np_source_intake_revision_unique', 'UNIQUE (intake_id, saved_revision)'),
        ('case_intake_write_requests', 'np_write_actor_request_unique', 'UNIQUE (actor_id, request_id)'),
        ('case_intake_source_snapshots', 'case_intake_source_snapshots_intake_id_fkey', 'FOREIGN KEY (intake_id) REFERENCES case_intakes(id)'),
        ('case_intake_write_requests', 'case_intake_write_requests_intake_id_fkey', 'FOREIGN KEY (intake_id) REFERENCES case_intakes(id)'),
        ('case_intake_write_requests', 'case_intake_write_requests_action_check', $$CHECK ((action = 'update'::text))$$),
        ('case_intake_write_requests', 'case_intake_write_requests_state_check', $$CHECK ((state = ANY (ARRAY['committed'::text, 'no_change'::text, 'abandoned'::text])))$$),
        ('case_intake_write_requests', 'case_intake_write_requests_check', $$CHECK ((((state = 'abandoned'::text) AND (request_digest IS NULL) AND (result_revision IS NULL) AND (source_revision IS NULL)) OR ((state = ANY (ARRAY['committed'::text, 'no_change'::text])) AND (request_digest IS NOT NULL) AND (result_revision IS NOT NULL))))$$)
)
SELECT current_schema() = 'public'
    AND to_regclass('case_intakes') = to_regclass('public.case_intakes')
    AND NOT EXISTS (
        SELECT FROM required_columns r
        WHERE NOT EXISTS (
            SELECT FROM pg_attribute a JOIN pg_class t ON t.oid = a.attrelid
            WHERE t.oid = to_regclass('public.' || r.table_name)
              AND t.oid = to_regclass(r.table_name) AND t.relkind = 'r'
              AND a.attname = r.column_name AND NOT a.attisdropped
              AND format_type(a.atttypid, a.atttypmod) = r.type_name
              AND a.attnotnull = r.required
        )
    )
    AND NOT EXISTS (
        SELECT FROM required_constraints r
        WHERE NOT EXISTS (
            SELECT FROM pg_constraint c
            WHERE c.conrelid = to_regclass('public.' || r.table_name)
              AND c.conname = r.constraint_name AND c.convalidated AND NOT c.condeferrable
              AND pg_get_constraintdef(c.oid) = r.definition
              AND (c.contype NOT IN ('u', 'p') OR EXISTS (
                  SELECT FROM pg_index i WHERE i.indexrelid = c.conindid
                    AND i.indisunique AND i.indisvalid AND i.indisready AND i.indimmediate
              ))
        )
    ) AS ready
