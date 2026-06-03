-- Runtime schema compatibility for RBAC writes on migrated PostgreSQL databases.

ALTER TABLE user_roles
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

UPDATE user_roles
SET updated_at = created_at
WHERE updated_at IS NULL;
