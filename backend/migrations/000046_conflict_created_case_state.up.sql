-- Align historical formal cases with the approval and conflict evidence that
-- produced them. A case that has already been generated must not re-enter the
-- pre-approval conflict-review stage.

UPDATE cases case_record
SET description = approval.content,
    priority = approval.priority::text,
    status = 'active',
    updated_at = CURRENT_TIMESTAMP
FROM approval_requests approval
WHERE approval.type = 'conflict_approval'
  AND approval.created_case_id = case_record.id::text
  AND approval.case_created = TRUE;

UPDATE approval_requests
SET status = 'approved',
    current_stage = '已完成',
    case_creation_status = 'completed',
    updated_at = CURRENT_TIMESTAMP
WHERE case_created = TRUE
  AND created_case_id IS NOT NULL
  AND created_case_id <> ''
  AND status <> 'approved';
