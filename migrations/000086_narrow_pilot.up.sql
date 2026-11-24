-- First slice: PostgreSQL only. Apply explicitly to the isolated database.
CREATE TABLE case_intake_source_snapshots (
    id bigserial PRIMARY KEY,
    intake_id uuid NOT NULL REFERENCES case_intakes(id),
    saved_revision text NOT NULL,
    serverpack jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT np_source_intake_revision_unique UNIQUE (intake_id, saved_revision)
);
CREATE TABLE case_intake_write_requests (
    id bigserial PRIMARY KEY,
    actor_id text NOT NULL,
    request_id uuid NOT NULL,
    action text NOT NULL CHECK (action = 'update'),
    intake_id uuid NOT NULL REFERENCES case_intakes(id),
    state text NOT NULL CHECK (state IN ('committed','no_change','abandoned')),
    request_digest text,
    result_revision text,
    source_revision text,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT np_write_actor_request_unique UNIQUE (actor_id,request_id),
    CHECK ((state = 'abandoned' AND request_digest IS NULL AND result_revision IS NULL AND source_revision IS NULL)
        OR (state IN ('committed','no_change') AND request_digest IS NOT NULL AND result_revision IS NOT NULL))
);
