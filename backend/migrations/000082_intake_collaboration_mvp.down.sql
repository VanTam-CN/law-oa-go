-- N1 回退：仅用于本轮隔离演练库。
-- 若已产生真实协作轮次，旧程序无法解释 active_submission_snapshot_id
-- 依据；生产回退须先停止本轮写入并采用前向修复，不直接降级（方案 §8/§14）。

DROP TABLE IF EXISTS client_creation_requests;

DROP INDEX IF EXISTS uq_approval_snapshots_intake_submission;

ALTER TABLE approval_requests
    DROP COLUMN IF EXISTS active_submission_snapshot_id;

DROP INDEX IF EXISTS uq_inbox_workflow_task_key;

ALTER TABLE inbox_items
    DROP COLUMN IF EXISTS workflow_task_key,
    DROP COLUMN IF EXISTS completion_mode,
    DROP COLUMN IF EXISTS source_ref;

