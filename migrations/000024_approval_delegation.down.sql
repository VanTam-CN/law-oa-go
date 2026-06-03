-- 000024_approval_delegation.down.sql

ALTER TABLE approval_records DROP COLUMN IF EXISTS original_approver_id;
ALTER TABLE approval_records DROP COLUMN IF EXISTS is_delegation;

ALTER TABLE approval_requests DROP COLUMN IF EXISTS escalated_to;
ALTER TABLE approval_requests DROP COLUMN IF EXISTS escalated_at;
ALTER TABLE approval_requests DROP COLUMN IF EXISTS escalated;
ALTER TABLE approval_requests DROP COLUMN IF EXISTS timeout_at;

DROP TABLE IF EXISTS approval_delegations;
