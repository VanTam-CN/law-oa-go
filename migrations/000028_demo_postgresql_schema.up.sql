-- Demo 聚合接口所需的 PostgreSQL 业务补全表。
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS case_intakes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    intake_code VARCHAR(50) NOT NULL UNIQUE,
    client_id BIGINT,
    title VARCHAR(255) NOT NULL,
    case_type VARCHAR(100),
    status VARCHAR(40) NOT NULL DEFAULT 'draft',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    description TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS case_intake_parties (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT,
    intake_id UUID,
    entity_name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(50) NOT NULL DEFAULT 'company',
    party_role VARCHAR(80) NOT NULL,
    relation_depth INT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS case_materials (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT,
    intake_id UUID,
    name VARCHAR(255) NOT NULL,
    material_type VARCHAR(80),
    status VARCHAR(40) NOT NULL DEFAULT 'missing',
    required BOOLEAN NOT NULL DEFAULT true,
    storage_url TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS approval_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_request_id VARCHAR(36) NOT NULL,
    snapshot_type VARCHAR(80) NOT NULL,
    snapshot_data JSONB NOT NULL,
    source_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS conflict_evidence_items (
    id BIGSERIAL PRIMARY KEY,
    conflict_check_id VARCHAR(100) NOT NULL,
    hit_entity_name VARCHAR(255) NOT NULL,
    evidence_type VARCHAR(80) NOT NULL,
    source_table VARCHAR(80),
    source_id VARCHAR(100),
    summary TEXT NOT NULL,
    confidence NUMERIC(5, 4) NOT NULL DEFAULT 0.8000,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS risk_audit_events (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(80) NOT NULL,
    actor_id BIGINT,
    subject_type VARCHAR(80) NOT NULL,
    subject_id VARCHAR(100) NOT NULL,
    risk_level VARCHAR(30),
    summary TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS system_settings (
    id BIGSERIAL PRIMARY KEY,
    setting_key VARCHAR(120) NOT NULL UNIQUE,
    setting_value JSONB NOT NULL DEFAULT '{}'::jsonb,
    category VARCHAR(80) NOT NULL DEFAULT 'general',
    description TEXT,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_case_intakes_status ON case_intakes(status);
CREATE INDEX IF NOT EXISTS idx_case_intakes_client_id ON case_intakes(client_id);
CREATE INDEX IF NOT EXISTS idx_case_intake_parties_intake_id ON case_intake_parties(intake_id);
CREATE INDEX IF NOT EXISTS idx_case_materials_intake_id ON case_materials(intake_id);
CREATE INDEX IF NOT EXISTS idx_approval_snapshots_request_id ON approval_snapshots(approval_request_id);
CREATE INDEX IF NOT EXISTS idx_conflict_evidence_check_id ON conflict_evidence_items(conflict_check_id);
CREATE INDEX IF NOT EXISTS idx_risk_audit_subject ON risk_audit_events(subject_type, subject_id);

INSERT INTO case_intakes (intake_code, client_id, title, case_type, status, priority, description, metadata)
VALUES
    ('INT-DEMO-001', 1, '华东新能源并购专项', '并购重组', 'conflict_ready', 'high', '收购标的涉及多层股权穿透与历史相对方复核。', '{"estimated_fee": 680000, "source": "demo_seed"}'::jsonb),
    ('INT-DEMO-002', 2, '员工股权激励合规审查', '公司合规', 'materials_pending', 'medium', '客户需补充董事会决议和员工名册。', '{"estimated_fee": 180000, "source": "demo_seed"}'::jsonb)
ON CONFLICT (intake_code) DO NOTHING;

INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, relation_depth, metadata)
SELECT ci.id, party.entity_name, party.entity_type, party.party_role, party.relation_depth, party.metadata
FROM case_intakes ci
JOIN (
    VALUES
        ('INT-DEMO-001', '华东新能源有限公司', 'company', 'client', 0, '{}'::jsonb),
        ('INT-DEMO-001', '北辰资本管理有限公司', 'company', 'opposing_party', 1, '{"risk_flag": "historical_opposing_party"}'::jsonb),
        ('INT-DEMO-001', '上海青澜投资合伙企业', 'partnership', 'related_party', 2, '{"ownership": "18.5%"}'::jsonb),
        ('INT-DEMO-002', '云启科技有限公司', 'company', 'client', 0, '{}'::jsonb)
) AS party(intake_code, entity_name, entity_type, party_role, relation_depth, metadata)
ON ci.intake_code = party.intake_code
WHERE NOT EXISTS (
    SELECT 1 FROM case_intake_parties cp
    WHERE cp.intake_id = ci.id AND cp.entity_name = party.entity_name AND cp.party_role = party.party_role
);

INSERT INTO case_materials (intake_id, name, material_type, status, required, metadata)
SELECT ci.id, material.name, material.material_type, material.status, material.required, '{}'::jsonb
FROM case_intakes ci
JOIN (
    VALUES
        ('INT-DEMO-001', '营业执照', 'identity', 'received', true),
        ('INT-DEMO-001', '股权结构图', 'ownership', 'received', true),
        ('INT-DEMO-001', '董事会决议', 'approval', 'missing', true),
        ('INT-DEMO-002', '员工名册', 'hr', 'missing', true)
) AS material(intake_code, name, material_type, status, required)
ON ci.intake_code = material.intake_code
WHERE NOT EXISTS (
    SELECT 1 FROM case_materials cm
    WHERE cm.intake_id = ci.id AND cm.name = material.name
);

INSERT INTO system_settings (setting_key, setting_value, category, description)
VALUES
    ('conflict.default_search_depth', '{"value": "deep", "years": 5}'::jsonb, 'conflict', '默认冲突检索深度'),
    ('approval.snapshot_required', '{"value": true}'::jsonb, 'approval', '审批提交时冻结快照'),
    ('notification.retry_policy', '{"max_attempts": 3, "interval_minutes": 10}'::jsonb, 'notification', '通知队列重试策略')
ON CONFLICT (setting_key) DO NOTHING;
