-- Apply the corrected default-approver rule to case approvals created while the
-- reserved system account was still selected by creation time.
UPDATE approval_requests approval
SET current_approver_id = trial_director.id::text,
    current_approver_name = trial_director.name,
    updated_at = CURRENT_TIMESTAMP
FROM users trial_director
WHERE trial_director.email = 'demo.admin@example.test'
  AND trial_director.status = 'active'
  AND approval.workflow_type = 'CASE_APPROVAL'
  AND approval.status IN ('submitted', 'under_review', 'resubmitted')
  AND approval.current_approver_name = '系统管理员';
