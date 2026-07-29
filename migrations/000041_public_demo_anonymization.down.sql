-- Privacy cleanup is intentionally irreversible. Restoring private branding in
-- a down migration would reintroduce data that this migration is designed to remove.
SELECT 1;
