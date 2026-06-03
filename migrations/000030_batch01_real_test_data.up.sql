-- Batch 01 PostgreSQL real test data.
-- These rows back the current frontend through API/database queries instead of
-- hard-coded frontend arrays.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE users ADD COLUMN IF NOT EXISTS department VARCHAR(50) DEFAULT '综合部';
ALTER TABLE users ADD COLUMN IF NOT EXISTS seniority VARCHAR(20) DEFAULT '初级';
ALTER TABLE clients ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 1;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS case_number VARCHAR(50);
ALTER TABLE cases ADD COLUMN IF NOT EXISTS created_by VARCHAR(36);
ALTER TABLE cases ADD COLUMN IF NOT EXISTS ethical_wall_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS ethical_wall_description TEXT;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS ethical_wall_enabled_by BIGINT;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS ethical_wall_enabled_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_cases_case_number_unique
ON cases(case_number);

CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_email_unique
ON clients(email);

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

CREATE TABLE IF NOT EXISTS inbox_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_type VARCHAR(50) NOT NULL,
    source_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    due_date TIMESTAMPTZ,
    due_date_type VARCHAR(50),
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    is_completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    reminder_sent BOOLEAN DEFAULT FALSE,
    reminder_count INT DEFAULT 0,
    escalated BOOLEAN DEFAULT FALSE,
    escalated_at TIMESTAMPTZ,
    snoozed_until TIMESTAMPTZ,
    snoozed_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS conflict_check_records (
    check_id VARCHAR(100) PRIMARY KEY,
    client_id VARCHAR(36) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    case_name VARCHAR(255) NOT NULL,
    case_type VARCHAR(100) NOT NULL,
    check_status VARCHAR(30) NOT NULL DEFAULT 'COMPLETED',
    has_conflict BOOLEAN DEFAULT FALSE,
    risk_level VARCHAR(30) DEFAULT 'LOW',
    search_parameters JSONB DEFAULT '{}'::jsonb,
    check_result JSONB DEFAULT '{}'::jsonb,
    user_id BIGINT,
    check_time TIMESTAMPTZ DEFAULT now(),
    duration BIGINT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS client_relations (
    id VARCHAR(36) PRIMARY KEY,
    client_id VARCHAR(36) NOT NULL,
    related_client_id VARCHAR(36) NOT NULL,
    relation_type VARCHAR(30) NOT NULL,
    relation_detail VARCHAR(500),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (client_id, related_client_id, relation_type)
);

CREATE INDEX IF NOT EXISTS idx_batch01_client_relations_client_id
ON client_relations(client_id);

CREATE INDEX IF NOT EXISTS idx_batch01_client_relations_related_client_id
ON client_relations(related_client_id);

CREATE TABLE IF NOT EXISTS approval_requests (
    id VARCHAR(36) PRIMARY KEY,
    request_number VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    category VARCHAR(100),
    content TEXT NOT NULL,
    applicant_id VARCHAR(36) NOT NULL,
    applicant_name VARCHAR(255) NOT NULL,
    applicant_title VARCHAR(100),
    department_id VARCHAR(36),
    department_name VARCHAR(255),
    urgency VARCHAR(20) DEFAULT 'normal',
    priority VARCHAR(20) DEFAULT 'medium',
    expected_effective_date TIMESTAMPTZ,
    expected_expiry_date TIMESTAMPTZ,
    duration_days INT,
    status VARCHAR(20) DEFAULT 'draft',
    submission_date TIMESTAMPTZ,
    current_stage VARCHAR(100),
    current_approver_id VARCHAR(36),
    current_approver_name VARCHAR(255),
    workflow_type VARCHAR(100) NOT NULL,
    workflow_config JSONB DEFAULT '{}'::jsonb,
    attachments JSONB DEFAULT '[]'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_by VARCHAR(36) NOT NULL,
    updated_by VARCHAR(36),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS timeout_at TIMESTAMPTZ;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS escalated BOOLEAN DEFAULT false;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS escalated_at TIMESTAMPTZ;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS escalated_to VARCHAR(36);
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS conflict_check_id VARCHAR(100);
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS conflict_risk_level VARCHAR(30);
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS conflict_check_time TIMESTAMPTZ;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS conflict_result JSONB DEFAULT '{}'::jsonb;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS case_created BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS created_case_id VARCHAR(36);
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS case_creation_time TIMESTAMPTZ;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS case_creation_status VARCHAR(30);

CREATE INDEX IF NOT EXISTS idx_batch01_approval_requests_case_created
ON approval_requests(case_created);

CREATE INDEX IF NOT EXISTS idx_batch01_approval_requests_created_case_id
ON approval_requests(created_case_id);

CREATE TABLE IF NOT EXISTS approval_workflows (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    workflow_code VARCHAR(100) NOT NULL UNIQUE,
    workflow_name VARCHAR(255) NOT NULL,
    workflow_type VARCHAR(100) NOT NULL,
    applicable_types JSONB DEFAULT '[]'::jsonb,
    applicable_departments JSONB DEFAULT '[]'::jsonb,
    applicable_roles JSONB DEFAULT '[]'::jsonb,
    stages JSONB NOT NULL,
    conditions JSONB DEFAULT '{}'::jsonb,
    timeouts JSONB DEFAULT '{}'::jsonb,
    permissions JSONB DEFAULT '{}'::jsonb,
    notifications JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    version INT NOT NULL DEFAULT 1,
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    expiry_date DATE,
    created_by VARCHAR(36) NOT NULL DEFAULT 'batch01',
    updated_by VARCHAR(36),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_approval_workflows_type_status
ON approval_workflows(workflow_type, status);

CREATE TABLE IF NOT EXISTS approval_records (
    id VARCHAR(36) PRIMARY KEY,
    approval_request_id VARCHAR(36) NOT NULL,
    stage VARCHAR(100) NOT NULL,
    stage_order INT NOT NULL DEFAULT 1,
    approver_id VARCHAR(36) NOT NULL,
    approver_name VARCHAR(255) NOT NULL DEFAULT '',
    approver_title VARCHAR(100),
    approver_role VARCHAR(100),
    decision VARCHAR(20) NOT NULL,
    decision_reason TEXT NOT NULL,
    decision_comments TEXT,
    approved_conditions JSONB NOT NULL DEFAULT '{}'::jsonb,
    imposed_requirements JSONB NOT NULL DEFAULT '{}'::jsonb,
    follow_up_actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_delegation BOOLEAN NOT NULL DEFAULT false,
    original_approver_id VARCHAR(36),
    approval_date TIMESTAMPTZ NOT NULL DEFAULT now(),
    effective_date TIMESTAMPTZ,
    next_review_date TIMESTAMPTZ,
    supporting_documents JSONB NOT NULL DEFAULT '[]'::jsonb,
    evidence_references JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_approval_records_request
ON approval_records(approval_request_id);

CREATE TABLE IF NOT EXISTS approval_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_request_id VARCHAR(36) NOT NULL,
    snapshot_type VARCHAR(80) NOT NULL,
    snapshot_data JSONB NOT NULL,
    source_version INT NOT NULL DEFAULT 1,
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

CREATE TABLE IF NOT EXISTS notification_queue (
    id BIGSERIAL PRIMARY KEY,
    trigger_type VARCHAR(50) NOT NULL,
    trigger_id BIGINT NOT NULL,
    case_id BIGINT,
    recipient_type VARCHAR(20) NOT NULL,
    recipient_id BIGINT NOT NULL,
    recipient_name VARCHAR(100) NOT NULL,
    recipient_contact VARCHAR(200),
    channel VARCHAR(20) NOT NULL,
    subject VARCHAR(200),
    content TEXT NOT NULL,
    template_id VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',
    priority VARCHAR(20) DEFAULT 'normal',
    created_by BIGINT NOT NULL,
    approved_by BIGINT,
    approved_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    sent_retry_count INT DEFAULT 0,
    error_message TEXT,
    contains_sensitive_info BOOLEAN DEFAULT FALSE,
    auto_send BOOLEAN DEFAULT FALSE,
    external_message_id VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notification_queue_status
ON notification_queue(status);

CREATE INDEX IF NOT EXISTS idx_notification_queue_recipient
ON notification_queue(recipient_id, status);

CREATE TABLE IF NOT EXISTS approval_conflict_associations (
    id VARCHAR(100) PRIMARY KEY,
    approval_request_id VARCHAR(36) NOT NULL,
    conflict_check_id VARCHAR(100) NOT NULL,
    association_status VARCHAR(30) NOT NULL DEFAULT 'active',
    association_type VARCHAR(30) NOT NULL DEFAULT 'required',
    risk_level VARCHAR(30),
    risk_score NUMERIC(5,2),
    conflict_count INT NOT NULL DEFAULT 0,
    requires_approval BOOLEAN NOT NULL DEFAULT false,
    auto_approval BOOLEAN NOT NULL DEFAULT false,
    approval_conditions JSONB NOT NULL DEFAULT '{}'::jsonb,
    mitigation_measures JSONB NOT NULL DEFAULT '[]'::jsonb,
    data_mapping JSONB NOT NULL DEFAULT '{}'::jsonb,
    mapped_fields JSONB NOT NULL DEFAULT '[]'::jsonb,
    validation_errors JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_by VARCHAR(36) NOT NULL,
    updated_by VARCHAR(36),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_approval_conflict_assoc_approval
ON approval_conflict_associations(approval_request_id, association_status);

CREATE TABLE IF NOT EXISTS approval_case_creation_tracking (
    id VARCHAR(120) PRIMARY KEY,
    approval_request_id VARCHAR(36) NOT NULL,
    case_id VARCHAR(50),
    case_number VARCHAR(100),
    case_type VARCHAR(100),
    creation_status VARCHAR(30) NOT NULL DEFAULT 'pending',
    creation_step VARCHAR(100),
    progress_percentage NUMERIC(5,2) NOT NULL DEFAULT 0,
    error_code VARCHAR(100),
    error_message TEXT,
    error_details JSONB NOT NULL DEFAULT '{}'::jsonb,
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    data_mapping JSONB NOT NULL DEFAULT '{}'::jsonb,
    mapped_fields JSONB NOT NULL DEFAULT '{}'::jsonb,
    unmapped_fields JSONB NOT NULL DEFAULT '{}'::jsonb,
    applied_conditions JSONB NOT NULL DEFAULT '{}'::jsonb,
    imposed_requirements JSONB NOT NULL DEFAULT '{}'::jsonb,
    workflow_actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_by VARCHAR(36) NOT NULL,
    processed_by VARCHAR(36),
    processed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_approval_case_tracking_approval
ON approval_case_creation_tracking(approval_request_id, created_at);

INSERT INTO users (username, name, email, password, role, phone, status, department, seniority)
VALUES
    ('batch01_admin', '系统管理员', 'batch01.admin@example.test', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', 'admin', '13800001000', 'active', '管理部', '合伙人'),
    ('batch01_zhang', '张律师', 'batch01.zhang@example.test', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', 'lawyer', '13800001001', 'active', '公司业务部', '合伙人'),
    ('batch01_li', '李律师', 'batch01.li@example.test', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', 'lawyer', '13800001002', 'active', '争议解决部', '高级'),
    ('batch01_compliance', '刘合规', 'batch01.compliance@example.test', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', 'user', '13800001003', 'active', '合规风控部', '高级')
ON CONFLICT (email) DO UPDATE
SET name = EXCLUDED.name,
    password = EXCLUDED.password,
    phone = EXCLUDED.phone,
    status = EXCLUDED.status,
    department = EXCLUDED.department,
    seniority = EXCLUDED.seniority,
    updated_at = now();

INSERT INTO approval_workflows (
    workflow_code, workflow_name, workflow_type, applicable_types, stages, status, effective_date, created_by
)
VALUES
    ('BATCH01_CASE_APPROVAL', 'Batch01 新建案件审批', 'CASE_APPROVAL',
     '["case_creation"]'::jsonb,
     '[{"stage_name": "系统管理员审批", "stage_order": 1, "approver_id": "1", "required": true, "timeout_hours": 24}]'::jsonb,
     'active', CURRENT_DATE, 'batch01'),
    ('BATCH01_CONFLICT_APPROVAL', 'Batch01 冲突审查审批', 'CONFLICT_APPROVAL',
     '["conflict_approval"]'::jsonb,
     '[{"stage_name": "合规复核", "stage_order": 1, "approver_id": "4", "required": true, "timeout_hours": 24}]'::jsonb,
     'active', CURRENT_DATE, 'batch01'),
    ('BATCH01_ACCESS_REVIEW', 'Batch01 权限变更审批', 'ACCESS_REVIEW',
     '["permission_change"]'::jsonb,
     '[{"stage_name": "权限管理员复核", "stage_order": 1, "approver_id": "1", "required": true, "timeout_hours": 24}]'::jsonb,
     'active', CURRENT_DATE, 'batch01')
ON CONFLICT (workflow_code) DO UPDATE
SET workflow_name = EXCLUDED.workflow_name,
    workflow_type = EXCLUDED.workflow_type,
    applicable_types = EXCLUDED.applicable_types,
    stages = EXCLUDED.stages,
    status = EXCLUDED.status,
    effective_date = EXCLUDED.effective_date,
    updated_by = 'batch01',
    updated_at = now();

INSERT INTO clients (name, type, email, phone, address, company, industry, contact_person, contact_phone, source, notes, status, version)
VALUES
    ('红杉资本投资管理集团', '企业', 'batch01.sequoia@client.local', '021-61000001', '上海市浦东新区世纪大道 100 号', '红杉资本投资管理集团', '投资管理', '王总', '13000001001', '现有客户介绍', 'Batch01 接案闭环测试客户，资料完整。', 'active', 1),
    ('上海华信建设集团有限公司', '企业', 'batch01.huaxin@client.local', '021-61000002', '上海市浦东新区张江路 88 号', '上海华信建设集团有限公司', '建筑工程', '李经理', '13000001002', '对方当事人库', 'Batch01 冲突检测对方当事人。', 'active', 1),
    ('天恒科技有限公司', '企业', 'batch01.tianheng@client.local', '021-61000003', '上海市徐汇区漕溪北路 18 号', '天恒科技有限公司', '智能制造', '赵经理', '13000001003', '市场拓展', 'Batch01 高风险审批样例客户。', 'active', 1)
ON CONFLICT (email) DO UPDATE
SET name = EXCLUDED.name,
    phone = EXCLUDED.phone,
    address = EXCLUDED.address,
    company = EXCLUDED.company,
    industry = EXCLUDED.industry,
    contact_person = EXCLUDED.contact_person,
    contact_phone = EXCLUDED.contact_phone,
    source = EXCLUDED.source,
    notes = EXCLUDED.notes,
    status = EXCLUDED.status,
    version = clients.version + 1,
    updated_at = now();

INSERT INTO cases (case_number, title, description, client_id, lawyer_id, case_type, priority, status, start_date, created_by, created_at, updated_at)
SELECT 'B01-CASE-001', '红杉资本投资管理咨询合同纠纷案', '对赌条款及回购义务履行争议，需完成接案前冲突检查。', c.id, u.id, '商事', 'high', 'pending', now() - interval '3 days', u.id::text, now() - interval '3 days', now()
FROM clients c, users u
WHERE c.email = 'batch01.sequoia@client.local' AND u.email = 'batch01.zhang@example.test'
ON CONFLICT (case_number) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    client_id = EXCLUDED.client_id,
    lawyer_id = EXCLUDED.lawyer_id,
    priority = EXCLUDED.priority,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO cases (case_number, title, description, client_id, lawyer_id, case_type, priority, status, start_date, created_by, created_at, updated_at)
SELECT 'B01-CASE-002', '天恒科技并购项目', '并购目标涉及历史相对方，需要合规复核。', c.id, u.id, '商事', 'high', 'in_progress', now() - interval '8 days', u.id::text, now() - interval '8 days', now() - interval '1 hour'
FROM clients c, users u
WHERE c.email = 'batch01.tianheng@client.local' AND u.email = 'batch01.li@example.test'
ON CONFLICT (case_number) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    client_id = EXCLUDED.client_id,
    lawyer_id = EXCLUDED.lawyer_id,
    priority = EXCLUDED.priority,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO case_intakes (intake_code, client_id, title, case_type, status, priority, description, metadata, created_by)
SELECT 'B01-INTAKE-001', c.id, '红杉资本投资管理咨询合同纠纷案', '商事', 'conflict_ready', 'high', '拟代理客户与对方就投资协议发生争议。', '{"source": "batch01_real_seed", "billing_method": "hourly"}'::jsonb, u.id
FROM clients c, users u
WHERE c.email = 'batch01.sequoia@client.local' AND u.email = 'batch01.zhang@example.test'
ON CONFLICT (intake_code) DO UPDATE
SET title = EXCLUDED.title,
    status = EXCLUDED.status,
    priority = EXCLUDED.priority,
    description = EXCLUDED.description,
    metadata = EXCLUDED.metadata,
    updated_at = now();

DELETE FROM case_intake_parties
WHERE intake_id IN (SELECT id FROM case_intakes WHERE intake_code = 'B01-INTAKE-001')
  AND metadata::text LIKE '%batch01_real_seed%';

INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, relation_depth, metadata)
SELECT ci.id, party.entity_name, party.entity_type, party.party_role, party.relation_depth, party.metadata
FROM case_intakes ci
JOIN (
    VALUES
        ('红杉资本投资管理集团', 'company', 'client', 0, '{"source": "batch01_real_seed"}'::jsonb),
        ('上海华信建设集团有限公司', 'company', 'opposing_party', 0, '{"source": "batch01_real_seed"}'::jsonb),
        ('华信建设（江苏）有限公司', 'company', 'related_party', 1, '{"source": "batch01_real_seed", "relationship": "subsidiary"}'::jsonb)
) AS party(entity_name, entity_type, party_role, relation_depth, metadata) ON TRUE
WHERE ci.intake_code = 'B01-INTAKE-001'
ON CONFLICT DO NOTHING;

DELETE FROM case_materials
WHERE intake_id IN (SELECT id FROM case_intakes WHERE intake_code = 'B01-INTAKE-001')
  AND metadata::text LIKE '%batch01_real_seed%';

INSERT INTO case_materials (intake_id, name, material_type, status, required, metadata)
SELECT ci.id, material.name, material.material_type, material.status, material.required, '{"source": "batch01_real_seed"}'::jsonb
FROM case_intakes ci
JOIN (
    VALUES
        ('客户主体资料', 'identity', 'received', true),
        ('投资协议及补充协议', 'contract', 'received', true),
        ('初步证据目录', 'evidence', 'missing', true)
) AS material(name, material_type, status, required) ON TRUE
WHERE ci.intake_code = 'B01-INTAKE-001'
ON CONFLICT DO NOTHING;

INSERT INTO conflict_check_records (check_id, client_id, client_name, case_name, case_type, check_status, has_conflict, risk_level, search_parameters, check_result, user_id, check_time, duration, created_at, updated_at)
SELECT 'B01-CCT-001', c.id::text, c.name, '红杉资本投资管理咨询合同纠纷案', '商事', 'COMPLETED', true, 'HIGH',
       '{"source": "batch01_real_seed", "searchDepth": "STANDARD"}'::jsonb,
       '{"riskAssessment": {"overallRisk": "HIGH", "riskScore": 86}, "recommendations": ["提交合规审批", "冻结审批快照"]}'::jsonb,
       u.id, now() - interval '2 hours', 2340, now() - interval '2 hours', now() - interval '2 hours'
FROM clients c, users u
WHERE c.email = 'batch01.sequoia@client.local' AND u.email = 'batch01.zhang@example.test'
ON CONFLICT (check_id) DO UPDATE
SET client_id = EXCLUDED.client_id,
    client_name = EXCLUDED.client_name,
    case_name = EXCLUDED.case_name,
    check_status = EXCLUDED.check_status,
    has_conflict = EXCLUDED.has_conflict,
    risk_level = EXCLUDED.risk_level,
    check_result = EXCLUDED.check_result,
    updated_at = now();

INSERT INTO approval_requests (id, request_number, title, type, category, content, applicant_id, applicant_name, applicant_title, department_id, department_name, urgency, priority, status, submission_date, current_stage, current_approver_id, current_approver_name, workflow_type, workflow_config, attachments, metadata, created_by, created_at, updated_at)
SELECT 'B01-APP-001', 'B01-APR-001', '冲突审查审批 - 红杉资本投资管理咨询合同纠纷案', 'conflict_approval', 'case_intake',
       '接案前冲突检测发现高风险命中，提交合伙人审批。', u.id::text, u.name, '合伙人', 'risk', '合规风控部',
       'urgent', 'high', 'under_review', now() - interval '90 minutes', '合伙人决策', reviewer.id::text, reviewer.name,
       'CONFLICT_APPROVAL', '{}'::jsonb, '[]'::jsonb,
       jsonb_build_object(
           'source', 'batch01_real_seed',
           'client', jsonb_build_object('id', c.id, 'name', c.name),
           'parties', jsonb_build_array(jsonb_build_object('role', 'opposing_party', 'name', '上海华信建设集团有限公司')),
           'materials', jsonb_build_array(jsonb_build_object('name', '客户主体资料', 'material_type', 'identity', 'status', 'received')),
           'conflict_result', jsonb_build_object('riskAssessment', jsonb_build_object('overallRisk', 'HIGH', 'riskScore', 86), 'checkId', 'B01-CCT-001')
       ),
       u.id::text, now() - interval '90 minutes', now() - interval '30 minutes'
FROM clients c, users u, users reviewer
WHERE c.email = 'batch01.sequoia@client.local'
  AND u.email = 'batch01.zhang@example.test'
  AND reviewer.email = 'batch01.compliance@example.test'
ON CONFLICT (request_number) DO UPDATE
SET title = EXCLUDED.title,
    content = EXCLUDED.content,
    status = EXCLUDED.status,
    current_stage = EXCLUDED.current_stage,
    current_approver_id = EXCLUDED.current_approver_id,
    current_approver_name = EXCLUDED.current_approver_name,
    metadata = EXCLUDED.metadata,
    updated_at = now();

INSERT INTO approval_requests (id, request_number, title, type, category, content, applicant_id, applicant_name, applicant_title, department_id, department_name, urgency, priority, status, submission_date, current_stage, current_approver_id, current_approver_name, workflow_type, workflow_config, attachments, metadata, created_by, created_at, updated_at)
SELECT 'B01-APP-002', 'B01-APR-002', '权限变更审批 - 开通审批快照下载权限', 'permission_change', 'access_control',
       '合规复核人员申请开通审批快照下载权限，用于归档复核材料。', u.id::text, u.name, '合规专员', 'risk', '合规风控部',
       'normal', 'high', 'submitted', now() - interval '35 minutes', '权限管理员复核', reviewer.id::text, reviewer.name,
       'ACCESS_REVIEW', '{}'::jsonb, '[]'::jsonb,
       '{"source": "batch01_real_seed", "permission": "approval_snapshot_download"}'::jsonb,
       u.id::text, now() - interval '35 minutes', now() - interval '20 minutes'
FROM users u, users reviewer
WHERE u.email = 'batch01.compliance@example.test'
  AND reviewer.email = 'batch01.admin@example.test'
ON CONFLICT (request_number) DO UPDATE
SET title = EXCLUDED.title,
    content = EXCLUDED.content,
    status = EXCLUDED.status,
    current_stage = EXCLUDED.current_stage,
    current_approver_id = EXCLUDED.current_approver_id,
    current_approver_name = EXCLUDED.current_approver_name,
    metadata = EXCLUDED.metadata,
    updated_at = now();

DELETE FROM approval_snapshots
WHERE approval_request_id IN ('B01-APP-001', 'B01-APP-002');

INSERT INTO approval_snapshots (approval_request_id, snapshot_type, snapshot_data, source_version)
SELECT id,
       type,
       jsonb_build_object(
           'snapshot_type', type,
           'source', 'batch01_real_seed',
           'approval', jsonb_build_object(
               'id', id,
               'request_number', request_number,
               'title', title,
               'status', status,
               'current_stage', current_stage,
               'current_approver_name', current_approver_name
           ),
           'metadata', metadata,
           'client', metadata->'client',
           'parties', COALESCE(metadata->'parties', '[]'::jsonb),
           'materials', COALESCE(metadata->'materials', '[]'::jsonb),
           'conflict_result', metadata->'conflict_result'
       ),
       1
FROM approval_requests
WHERE id IN ('B01-APP-001', 'B01-APP-002');

INSERT INTO inbox_items (user_id, source_type, source_id, title, content, priority, due_date, due_date_type, is_read, is_completed, created_at, updated_at)
SELECT u.id, item.source_type, item.source_id, item.title, item.content, item.priority, item.due_date, item.due_date_type, false, false, now(), now()
FROM users u
JOIN (
    VALUES
        ('approval', 1::bigint, '冲突审查审批 - 红杉资本投资管理咨询合同纠纷案', '高风险冲突审批待处理', 'critical', now() + interval '6 hours', 'approval_sla'),
        ('task', 2::bigint, '补齐初步证据目录', '接案材料缺失项待补充', 'high', now() - interval '1 day', 'material_check'),
        ('conflict', 3::bigint, '复核 B01-CCT-001 冲突检测记录', '冲突检测已完成，需要合规复核', 'high', now() + interval '3 hours', 'conflict_review')
) AS item(source_type, source_id, title, content, priority, due_date, due_date_type) ON TRUE
WHERE u.email = 'batch01.compliance@example.test'
  AND NOT EXISTS (
      SELECT 1 FROM inbox_items existing
      WHERE existing.user_id = u.id
        AND existing.source_type = item.source_type
        AND existing.title = item.title
        AND existing.deleted_at IS NULL
  );

INSERT INTO risk_audit_events (event_type, actor_id, subject_type, subject_id, risk_level, summary, payload, created_at)
SELECT event.event_type, u.id, event.subject_type, event.subject_id, event.risk_level, event.summary, event.payload, event.created_at
FROM users u
JOIN (
    VALUES
        ('permission_review', 'approval_request', 'B01-APP-002', 'HIGH', '权限变更审批进入管理员复核。', '{"source": "batch01_real_seed"}'::jsonb, now() - interval '18 minutes'),
        ('conflict_snapshot', 'approval_request', 'B01-APP-001', 'HIGH', '高风险冲突审批快照已落库。', '{"source": "batch01_real_seed"}'::jsonb, now() - interval '28 minutes')
) AS event(event_type, subject_type, subject_id, risk_level, summary, payload, created_at) ON TRUE
WHERE u.email = 'batch01.admin@example.test'
  AND NOT EXISTS (
      SELECT 1 FROM risk_audit_events existing
      WHERE existing.subject_id = event.subject_id
        AND existing.event_type = event.event_type
        AND existing.payload::text LIKE '%batch01_real_seed%'
  );

INSERT INTO system_settings (setting_key, setting_value, category, description, updated_by)
SELECT item.setting_key, item.setting_value, item.category, item.description, u.id
FROM users u
JOIN (
    VALUES
        ('batch01.approval.partner_final_review', '{"enabled": true, "source": "batch01_real_seed"}'::jsonb, 'approval', '冲突高风险必须合伙人终审'),
        ('batch01.approval.immutable_snapshot', '{"enabled": true, "source": "batch01_real_seed"}'::jsonb, 'approval', '审批提交后生成不可变快照'),
        ('batch01.notification.approval_sla', '{"enabled": true, "channel": "站内信 + 邮件", "source": "batch01_real_seed"}'::jsonb, 'notification', '审批即将超时提前提醒'),
        ('batch01.notification.material_completed', '{"enabled": true, "channel": "站内信", "source": "batch01_real_seed"}'::jsonb, 'notification', '资料补正完成通知律师与合规'),
        ('batch01.audit.snapshot_persist', '{"enabled": true, "source": "batch01_real_seed"}'::jsonb, 'audit', '审批快照落盘并记录版本'),
        ('batch01.security.sensitive_masking', '{"enabled": true, "source": "batch01_real_seed"}'::jsonb, 'security', '敏感字段按角色脱敏'),
        ('batch01.file.archive_policy', '{"enabled": true, "source": "batch01_real_seed"}'::jsonb, 'file', '审批通过后自动归档关键材料')
) AS item(setting_key, setting_value, category, description) ON TRUE
WHERE u.email = 'batch01.admin@example.test'
ON CONFLICT (setting_key) DO UPDATE
SET setting_value = EXCLUDED.setting_value,
    category = EXCLUDED.category,
    description = EXCLUDED.description,
    updated_by = EXCLUDED.updated_by,
    updated_at = now();
