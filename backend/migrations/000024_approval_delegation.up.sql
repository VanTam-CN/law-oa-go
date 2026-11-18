-- 000024_approval_delegation.up.sql
-- 审批代理配置表

CREATE TABLE IF NOT EXISTS approval_delegations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    delegator_id UUID NOT NULL,
    delegate_id UUID NOT NULL,
    valid_from TIMESTAMP NOT NULL,
    valid_until TIMESTAMP,
    is_active BOOLEAN DEFAULT true NOT NULL,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by UUID NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT uq_delegator_delegate_time UNIQUE (delegator_id, delegate_id, valid_from),
    CONSTRAINT chk_no_self_delegation CHECK (delegator_id != delegate_id)
);

COMMENT ON TABLE approval_delegations IS '审批代理配置表';

-- 索引
CREATE INDEX IF NOT EXISTS idx_delegations_delegator ON approval_delegations(delegator_id);
CREATE INDEX IF NOT EXISTS idx_delegations_delegate ON approval_delegations(delegate_id);
CREATE INDEX IF NOT EXISTS idx_delegations_active ON approval_delegations(is_active, valid_from, valid_until);

-- 为 approval_requests 表添加超时相关字段
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS timeout_at TIMESTAMP;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS escalated BOOLEAN DEFAULT false;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS escalated_at TIMESTAMP;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS escalated_to UUID;

-- 为 approval_records 表添加代理审批字段
ALTER TABLE approval_records ADD COLUMN IF NOT EXISTS is_delegation BOOLEAN DEFAULT false;
ALTER TABLE approval_records ADD COLUMN IF NOT EXISTS original_approver_id UUID;
