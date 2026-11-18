-- N1 接案协作 MVP：契约与存储扩展（方案 §5/§6/§8）
-- 加法兼容迁移；不改写历史行语义。所有新列可空或带默认值。

-- 1) inbox_items：来源字符串引用 + 完成语义 + 协作任务幂等键
ALTER TABLE inbox_items
    ADD COLUMN IF NOT EXISTS source_ref VARCHAR(100),
    ADD COLUMN IF NOT EXISTS completion_mode VARCHAR(20) NOT NULL DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS workflow_task_key VARCHAR(200);

COMMENT ON COLUMN inbox_items.source_ref IS '真实业务对象 ID（审批 UUID 等），替代无法承载字符串的 source_id';
COMMENT ON COLUMN inbox_items.completion_mode IS 'workflow=业务转移才能完成; manual=个人提醒可完成; legacy_unresolved=仅可归档';
COMMENT ON COLUMN inbox_items.workflow_task_key IS '协作待办幂等键：业务类型+对象ID+轮次/节点+接收人+用途';

ALTER TABLE inbox_items ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_inbox_items_deleted_at ON inbox_items (deleted_at);

CREATE UNIQUE INDEX IF NOT EXISTS uq_inbox_workflow_task_key
    ON inbox_items (workflow_task_key)
    WHERE workflow_task_key IS NOT NULL AND deleted_at IS NULL;

-- 2) approval_requests：当前权威提交快照指针
ALTER TABLE approval_requests
    ADD COLUMN IF NOT EXISTS active_submission_snapshot_id UUID;

COMMENT ON COLUMN approval_requests.active_submission_snapshot_id IS '当前权威 intake_submission 快照；决定/成案/重提统一读取';

-- 3) approval_snapshots：新快照类型唯一约束（不动历史重复行）
CREATE UNIQUE INDEX IF NOT EXISTS uq_approval_snapshots_intake_submission
    ON approval_snapshots (approval_request_id, snapshot_type, source_version)
    WHERE snapshot_type = 'intake_submission';

-- 4) 客户创建幂等窄表（操作者 + key 作用域；不保存身份原文）
CREATE TABLE IF NOT EXISTS client_creation_requests (
    id BIGSERIAL PRIMARY KEY,
    operator_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    idempotency_key VARCHAR(120) NOT NULL,
    request_digest VARCHAR(128) NOT NULL,
    client_id BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE client_creation_requests IS '客户创建幂等记录：同 key 同规范化请求返回同一客户；同 key 不同请求 409';
COMMENT ON COLUMN client_creation_requests.request_digest IS '受保护 HMAC 摘要，不含身份明文';

CREATE UNIQUE INDEX IF NOT EXISTS uq_client_creation_requests_operator_key
    ON client_creation_requests (operator_id, idempotency_key);

