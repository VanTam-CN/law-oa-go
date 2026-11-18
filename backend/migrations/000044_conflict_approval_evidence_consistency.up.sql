-- Keep every representation of conflict evidence aligned with the approval's
-- authoritative conflict_check_id. Historical fixtures may have corrected the
-- structured result without updating denormalized metadata and snapshots.

UPDATE approval_requests
SET metadata = jsonb_set(
        jsonb_set(
            COALESCE(metadata, '{}'::jsonb),
            '{conflict_task_id}',
            to_jsonb(conflict_check_id),
            TRUE
        ),
        '{conflict_result}',
        COALESCE(conflict_result, '{}'::jsonb),
        TRUE
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE type = 'conflict_approval'
  AND conflict_check_id IS NOT NULL
  AND conflict_result IS NOT NULL;

UPDATE approval_snapshots snapshot
SET snapshot_data = jsonb_set(
        jsonb_set(
            jsonb_set(
                jsonb_set(
                    COALESCE(snapshot.snapshot_data, '{}'::jsonb),
                    '{conflict_task_id}',
                    to_jsonb(approval.conflict_check_id),
                    TRUE
                ),
                '{conflict_result}',
                approval.conflict_result,
                TRUE
            ),
            '{metadata,conflict_task_id}',
            to_jsonb(approval.conflict_check_id),
            TRUE
        ),
        '{metadata,conflict_result}',
        approval.conflict_result,
        TRUE
    )
FROM approval_requests approval
WHERE snapshot.approval_request_id = approval.id
  AND approval.type = 'conflict_approval'
  AND approval.conflict_check_id IS NOT NULL
  AND approval.conflict_result IS NOT NULL;
