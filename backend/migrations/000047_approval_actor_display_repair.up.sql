-- Refresh denormalized approval actor labels after public-demo account
-- anonymization. Stable user IDs remain authoritative.
UPDATE approval_requests approval
SET applicant_name = actor.name,
    updated_at = CURRENT_TIMESTAMP
FROM users actor
WHERE approval.applicant_id = actor.id::text
  AND approval.applicant_name IS DISTINCT FROM actor.name;

UPDATE approval_requests approval
SET current_approver_name = actor.name,
    updated_at = CURRENT_TIMESTAMP
FROM users actor
WHERE approval.current_approver_id = actor.id::text
  AND approval.current_approver_name IS DISTINCT FROM actor.name;

UPDATE approval_requests
SET current_approver_name = '历史示例律师',
    updated_at = CURRENT_TIMESTAMP
WHERE current_approver_name LIKE '%@law-oa.local';

UPDATE approval_records record
SET approver_name = actor.name,
    updated_at = CURRENT_TIMESTAMP
FROM users actor
WHERE record.approver_id = actor.id::text
  AND record.approver_name IS DISTINCT FROM actor.name;

UPDATE approval_records
SET approver_name = '历史示例律师',
    updated_at = CURRENT_TIMESTAMP
WHERE approver_name LIKE '%@law-oa.local';
