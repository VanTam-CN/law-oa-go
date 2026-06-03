-- 分成规则表
CREATE TABLE IF NOT EXISTS commission_rules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL,
    min_amount DECIMAL(15,2) DEFAULT 0,
    max_amount DECIMAL(15,2) DEFAULT 0,
    base_rate DECIMAL(5,2) NOT NULL,
    performance_rate DECIMAL(5,2) DEFAULT 0,
    priority INTEGER DEFAULT 0,
    active BOOLEAN DEFAULT TRUE,
    effective_date TIMESTAMP,
    expiry_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_commission_rules_role ON commission_rules(role);
CREATE INDEX idx_commission_rules_active ON commission_rules(active);
CREATE INDEX idx_commission_rules_priority ON commission_rules(priority DESC);

-- 插入默认分成规则
INSERT INTO commission_rules (name, role, min_amount, max_amount, base_rate, priority, active) VALUES
('案源-小额', 'source', 0, 10000, 10, 1, TRUE),
('案源-中额', 'source', 10000, 50000, 15, 2, TRUE),
('案源-大额', 'source', 50000, 100000, 20, 3, TRUE),
('案源-特大额', 'source', 100000, 0, 30, 4, TRUE),
('主办律师-小额', 'lawyer', 0, 10000, 20, 1, TRUE),
('主办律师-中额', 'lawyer', 10000, 50000, 30, 2, TRUE),
('主办律师-大额', 'lawyer', 50000, 100000, 40, 3, TRUE),
('主办律师-特大额', 'lawyer', 100000, 0, 50, 4, TRUE),
('协办律师-小额', 'assistant', 0, 10000, 5, 1, TRUE),
('协办律师-中额', 'assistant', 10000, 50000, 8, 2, TRUE),
('协办律师-大额', 'assistant', 50000, 100000, 12, 3, TRUE),
('协办律师-特大额', 'assistant', 100000, 0, 15, 4, TRUE);
