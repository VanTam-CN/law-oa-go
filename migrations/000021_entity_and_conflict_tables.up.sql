-- ============================================================
-- Law OA Go 数据库迁移文件
-- 版本: 000021
-- 描述: 创建实体管理、实体关系、案件当事人及冲突检测核心表
-- ============================================================

-- 1. 创建实体表 (entities)
-- 统一管理个人和企业的法律实体信息
CREATE TABLE IF NOT EXISTS entities (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    -- 基本信息
    entity_type     VARCHAR(20) NOT NULL,                        -- PERSON, COMPANY, GOVERNMENT, NGO, OTHER
    name            VARCHAR(200) NOT NULL,
    alias           VARCHAR(500),
    identity_type   VARCHAR(30),                                 -- ID_CARD, PASSPORT, BUSINESS_LICENSE, etc.
    identity_number VARCHAR(100),
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',       -- ACTIVE, INACTIVE, SUSPENDED

    -- 个人/企业详情
    gender          VARCHAR(10),
    nationality     VARCHAR(50),
    birth_date      TIMESTAMPTZ,
    legal_representative VARCHAR(100),
    registered_capital   NUMERIC(20,2) DEFAULT 0,
    establish_date       TIMESTAMPTZ,
    business_scope       TEXT,

    -- 联系方式
    address         TEXT,
    phone           VARCHAR(20),
    email           VARCHAR(100),
    contact_person  VARCHAR(100),

    -- 备注
    notes           TEXT
);

-- entities 索引
CREATE INDEX IF NOT EXISTS idx_entities_created_at ON entities(created_at);
CREATE INDEX IF NOT EXISTS idx_entities_deleted_at ON entities(deleted_at);
CREATE INDEX IF NOT EXISTS idx_entities_entity_type ON entities(entity_type);
CREATE INDEX IF NOT EXISTS idx_entities_name ON entities(name);
CREATE INDEX IF NOT EXISTS idx_entities_identity_type ON entities(identity_type);
CREATE INDEX IF NOT EXISTS idx_entities_identity_number ON entities(identity_number);
CREATE UNIQUE INDEX IF NOT EXISTS uq_entities_identity ON entities(identity_type, identity_number) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_entities_status ON entities(status);

-- 2. 创建实体关系表 (entity_relations)
-- 管理实体之间的多维关系（股权、关联、对立等）
CREATE TABLE IF NOT EXISTS entity_relations (
    id                BIGSERIAL PRIMARY KEY,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,

    source_entity_id  BIGINT NOT NULL,
    target_entity_id  BIGINT NOT NULL,
    relation_type     VARCHAR(30) NOT NULL,                       -- SHAREHOLDER, PARENT, SUBSIDIARY, SISTER, ADVERSE, etc.
    shareholding_ratio NUMERIC(8,4),                              -- 持股比例 0.0000 - 1.0000
    description       TEXT,
    start_date        TIMESTAMPTZ,
    end_date          TIMESTAMPTZ,
    is_active         BOOLEAN DEFAULT TRUE,
    data_source       VARCHAR(100),

    -- 外键约束
    CONSTRAINT fk_entity_relations_source
        FOREIGN KEY (source_entity_id) REFERENCES entities(id) ON DELETE CASCADE,
    CONSTRAINT fk_entity_relations_target
        FOREIGN KEY (target_entity_id) REFERENCES entities(id) ON DELETE CASCADE
);

-- entity_relations 索引
CREATE INDEX IF NOT EXISTS idx_entity_relations_created_at ON entity_relations(created_at);
CREATE INDEX IF NOT EXISTS idx_entity_relations_deleted_at ON entity_relations(deleted_at);
CREATE INDEX IF NOT EXISTS idx_entity_relations_source_entity_id ON entity_relations(source_entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_relations_target_entity_id ON entity_relations(target_entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_relations_relation_type ON entity_relations(relation_type);

-- 3. 创建实体名称变更历史表 (entity_name_history)
-- 记录实体名称变更记录
CREATE TABLE IF NOT EXISTS entity_name_history (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    entity_id       BIGINT NOT NULL,
    old_name        VARCHAR(200) NOT NULL,
    new_name        VARCHAR(200) NOT NULL,
    change_date     TIMESTAMPTZ NOT NULL,
    change_reason   VARCHAR(500),

    -- 外键约束
    CONSTRAINT fk_entity_name_history_entity
        FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);

-- entity_name_history 索引
CREATE INDEX IF NOT EXISTS idx_entity_name_history_deleted_at ON entity_name_history(deleted_at);
CREATE INDEX IF NOT EXISTS idx_entity_name_history_entity_id ON entity_name_history(entity_id);

-- 4. 创建案件当事人表 (case_parties)
-- 管理案件中各方的当事人信息，关联实体表
CREATE TABLE IF NOT EXISTS case_parties (
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    case_id         BIGINT NOT NULL,
    entity_id       BIGINT NOT NULL,
    role            VARCHAR(30) NOT NULL,                         -- PLAINTIFF, DEFENDANT, APPLICANT, RESPONDENT, etc.
    party_type      VARCHAR(20) NOT NULL,                         -- INDIVIDUAL, ORGANIZATION
    description     TEXT,
    display_order   INT DEFAULT 0,

    -- 外键约束
    CONSTRAINT fk_case_parties_case
        FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    CONSTRAINT fk_case_parties_entity
        FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);

-- case_parties 索引
CREATE INDEX IF NOT EXISTS idx_case_parties_created_at ON case_parties(created_at);
CREATE INDEX IF NOT EXISTS idx_case_parties_deleted_at ON case_parties(deleted_at);
CREATE INDEX IF NOT EXISTS idx_case_parties_case_id ON case_parties(case_id);
CREATE INDEX IF NOT EXISTS idx_case_parties_entity_id ON case_parties(entity_id);
CREATE INDEX IF NOT EXISTS idx_case_parties_role ON case_parties(role);

-- 5. 创建冲突检查表 (conflict_checks)
-- GORM V2 冲突检查主表，存储每次冲突检查的整体信息
CREATE TABLE IF NOT EXISTS conflict_checks (
    id                BIGSERIAL PRIMARY KEY,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,

    case_id           BIGINT NOT NULL,
    status            VARCHAR(50) NOT NULL DEFAULT 'PENDING',    -- PENDING, PROCESSING, COMPLETED, FAILED, CANCELLED
    requested_by      BIGINT NOT NULL,
    requested_at      TIMESTAMPTZ NOT NULL,
    checked_by        BIGINT,
    checked_at        TIMESTAMPTZ,
    result            JSONB,                                      -- GORM serializer:json -> 存储完整检测结果
    result_summary    TEXT,
    total_conflicts   INT DEFAULT 0,
    critical_count    INT DEFAULT 0,
    high_count        INT DEFAULT 0,
    medium_count      INT DEFAULT 0,
    low_count         INT DEFAULT 0,
    check_params      JSONB,                                      -- 检查参数
    report_path       VARCHAR(500),
    report_generated_at TIMESTAMPTZ,

    -- 外键约束
    CONSTRAINT fk_conflict_checks_case
        FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    CONSTRAINT fk_conflict_checks_requested_by
        FOREIGN KEY (requested_by) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT fk_conflict_checks_checked_by
        FOREIGN KEY (checked_by) REFERENCES users(id) ON DELETE SET NULL
);

-- conflict_checks 索引
CREATE INDEX IF NOT EXISTS idx_conflict_checks_deleted_at ON conflict_checks(deleted_at);
CREATE INDEX IF NOT EXISTS idx_conflict_checks_case_id ON conflict_checks(case_id);
CREATE INDEX IF NOT EXISTS idx_conflict_checks_status ON conflict_checks(status);
CREATE INDEX IF NOT EXISTS idx_conflict_checks_requested_by ON conflict_checks(requested_by);
CREATE INDEX IF NOT EXISTS idx_conflict_checks_created_at ON conflict_checks(created_at);

-- 6. 创建冲突详情表 (conflict_details)
-- 存储每条冲突检测的详细结果
CREATE TABLE IF NOT EXISTS conflict_details (
    id                BIGSERIAL PRIMARY KEY,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,

    conflict_check_id BIGINT NOT NULL,
    matched_entity_id BIGINT NOT NULL,
    matched_case_id   BIGINT,
    conflict_type     VARCHAR(50) NOT NULL,                      -- ADVERSE_PARTY, CLIENT_CONFLICT, etc.
    risk_level        VARCHAR(20) NOT NULL,                       -- CRITICAL, HIGH, MEDIUM, LOW
    description       TEXT NOT NULL,
    evidence          TEXT,
    recommendation    TEXT,
    is_waived         BOOLEAN DEFAULT FALSE,
    waived_by         BIGINT,
    waived_at         TIMESTAMPTZ,
    waive_reason      VARCHAR(500),
    match_reason      VARCHAR(500),

    -- 外键约束
    CONSTRAINT fk_conflict_details_check
        FOREIGN KEY (conflict_check_id) REFERENCES conflict_checks(id) ON DELETE CASCADE,
    CONSTRAINT fk_conflict_details_matched_entity
        FOREIGN KEY (matched_entity_id) REFERENCES entities(id) ON DELETE RESTRICT,
    CONSTRAINT fk_conflict_details_matched_case
        FOREIGN KEY (matched_case_id) REFERENCES cases(id) ON DELETE SET NULL,
    CONSTRAINT fk_conflict_details_waived_by
        FOREIGN KEY (waived_by) REFERENCES users(id) ON DELETE SET NULL
);

-- conflict_details 索引
CREATE INDEX IF NOT EXISTS idx_conflict_details_created_at ON conflict_details(created_at);
CREATE INDEX IF NOT EXISTS idx_conflict_details_deleted_at ON conflict_details(deleted_at);
CREATE INDEX IF NOT EXISTS idx_conflict_details_conflict_check_id ON conflict_details(conflict_check_id);
CREATE INDEX IF NOT EXISTS idx_conflict_details_matched_entity_id ON conflict_details(matched_entity_id);
CREATE INDEX IF NOT EXISTS idx_conflict_details_matched_case_id ON conflict_details(matched_case_id);
CREATE INDEX IF NOT EXISTS idx_conflict_details_conflict_type ON conflict_details(conflict_type);
CREATE INDEX IF NOT EXISTS idx_conflict_details_risk_level ON conflict_details(risk_level);
