-- PostgreSQL migration, following the candidate's migration-driver boundary.
-- No raw key, request body, identity plaintext or response snapshot is stored.
CREATE TABLE client_creation_requests (
    id BIGSERIAL PRIMARY KEY,
    operator_id BIGINT NOT NULL CHECK (operator_id > 0),
    operation VARCHAR(32) NOT NULL,
    key_digest VARCHAR(64) NOT NULL,
    request_digest VARCHAR(64) NOT NULL,
    client_id BIGINT REFERENCES clients(id),
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_client_creation_scope UNIQUE (operator_id, operation, key_digest),
    CONSTRAINT ck_client_creation_result CHECK (
        (status = 'pending' AND client_id IS NULL) OR
        (status = 'completed' AND client_id IS NOT NULL)
    ),
    CONSTRAINT ck_client_creation_digests CHECK (
        length(key_digest) = 64 AND length(request_digest) = 64
    )
);
