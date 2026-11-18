-- Repair historical cases created from conflict approvals. The conflict task
-- owner is the authoritative responsible lawyer for the generated case.

UPDATE cases case_record
SET lawyer_id = conflict_record.user_id,
    created_by = conflict_record.user_id::text,
    updated_at = CURRENT_TIMESTAMP
FROM approval_requests approval
JOIN conflict_check_records conflict_record
  ON conflict_record.check_id = approval.conflict_check_id
WHERE approval.type = 'conflict_approval'
  AND approval.created_case_id = case_record.id::text
  AND conflict_record.user_id IS NOT NULL
  AND (
      case_record.lawyer_id IS DISTINCT FROM conflict_record.user_id
      OR case_record.created_by IS DISTINCT FROM conflict_record.user_id::text
  );
