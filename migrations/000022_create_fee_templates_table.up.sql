CREATE TABLE IF NOT EXISTS fee_templates (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(100) NOT NULL,
    case_type VARCHAR(50) NOT NULL,
    billing_type VARCHAR(50) NOT NULL,
    base_rates JSONB NOT NULL,
    performance_bonus_rate DECIMAL(5,2) DEFAULT 0,
    min_amount DECIMAL(15,2) DEFAULT 0,
    max_amount DECIMAL(15,2) DEFAULT 0,
    cost_rate DECIMAL(5,2) DEFAULT 0,
    active BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_fee_templates_case_type ON fee_templates(case_type);
CREATE INDEX IF NOT EXISTS idx_fee_templates_active ON fee_templates(active);

COMMENT ON COLUMN fee_templates.name IS '模板名称';
COMMENT ON COLUMN fee_templates.case_type IS '适用案件类型: litigation/non_litigation/consulting';
COMMENT ON COLUMN fee_templates.billing_type IS '计费模式: hourly/fixed/hybrid/retainer';
COMMENT ON COLUMN fee_templates.base_rates IS '按角色定义基础费率: {source: 0.15, lawyer: 0.30, assistant: 0.10}';
COMMENT ON COLUMN fee_templates.performance_bonus_rate IS '绩效奖金比例';
COMMENT ON COLUMN fee_templates.min_amount IS '最小适用金额';
COMMENT ON COLUMN fee_templates.max_amount IS '最大适用金额';
COMMENT ON COLUMN fee_templates.cost_rate IS '成本扣除比例';
COMMENT ON COLUMN fee_templates.active IS '是否启用';
