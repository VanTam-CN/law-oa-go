-- Create testing related tables for comprehensive testing framework

-- Test suites table - stores test suite definitions
CREATE TABLE test_suites (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    name VARCHAR(255) NOT NULL,
    type ENUM('api', 'ui', 'performance', 'integration') NOT NULL,
    description TEXT,
    config JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR(36),
    is_active BOOLEAN DEFAULT TRUE,
    schedule_cron VARCHAR(100),
    environment VARCHAR(50) DEFAULT 'test',
    INDEX idx_type (type),
    INDEX idx_created_by (created_by),
    INDEX idx_created_at (created_at),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Test executions table - stores test execution records
CREATE TABLE test_executions (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    suite_id VARCHAR(36) NOT NULL,
    status ENUM('pending', 'running', 'completed', 'failed', 'cancelled') NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    duration_ms INT DEFAULT 0,
    environment VARCHAR(50),
    trigger_type ENUM('manual', 'scheduled', 'api') DEFAULT 'manual',
    triggered_by VARCHAR(36),
    result JSON,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (suite_id) REFERENCES test_suites(id) ON DELETE CASCADE,
    INDEX idx_suite_id (suite_id),
    INDEX idx_status (status),
    INDEX idx_started_at (started_at),
    INDEX idx_created_at (created_at),
    INDEX idx_trigger_type (trigger_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Test results table - stores detailed test results
CREATE TABLE test_results (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    execution_id VARCHAR(36) NOT NULL,
    test_name VARCHAR(255) NOT NULL,
    test_type VARCHAR(50) NOT NULL,
    status ENUM('passed', 'failed', 'skipped', 'error') NOT NULL,
    duration_ms INT DEFAULT 0,
    error_message TEXT,
    assertion_results JSON,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (execution_id) REFERENCES test_executions(id) ON DELETE CASCADE,
    INDEX idx_execution_id (execution_id),
    INDEX idx_status (status),
    INDEX idx_test_type (test_type),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User events table - stores user behavior tracking data
CREATE TABLE user_events (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id VARCHAR(36),
    session_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    element VARCHAR(255),
    element_selector VARCHAR(500),
    page_url VARCHAR(500),
    page_title VARCHAR(255),
    referrer VARCHAR(500),
    user_agent TEXT,
    ip_address VARCHAR(45),
    timestamp TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_session_id (session_id),
    INDEX idx_event_type (event_type),
    INDEX idx_timestamp (timestamp),
    INDEX idx_page_url (page_url),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User sessions table - stores user session information
CREATE TABLE user_sessions (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id VARCHAR(36),
    started_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    ended_at TIMESTAMP(3) NULL,
    duration_ms INT DEFAULT 0,
    page_views INT DEFAULT 0,
    events_count INT DEFAULT 0,
    browser VARCHAR(100),
    browser_version VARCHAR(50),
    os VARCHAR(100),
    os_version VARCHAR(50),
    device_type VARCHAR(50),
    screen_resolution VARCHAR(20),
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_started_at (started_at),
    INDEX idx_device_type (device_type),
    INDEX idx_ip_address (ip_address)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Analytics reports table - stores generated analytics reports
CREATE TABLE analytics_reports (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    report_type ENUM('user_behavior', 'system_performance', 'test_summary', 'custom') NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    period_start TIMESTAMP NULL,
    period_end TIMESTAMP NULL,
    data JSON NOT NULL,
    file_path VARCHAR(500),
    status ENUM('generating', 'completed', 'failed') DEFAULT 'generating',
    generated_by VARCHAR(36),
    generated_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_report_type (report_type),
    INDEX idx_status (status),
    INDEX idx_period_start (period_start),
    INDEX idx_generated_by (generated_by),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- System metrics table - stores system monitoring metrics
CREATE TABLE system_metrics (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    metric_name VARCHAR(100) NOT NULL,
    metric_type ENUM('counter', 'gauge', 'histogram', 'summary') NOT NULL,
    value DECIMAL(15,4) NOT NULL,
    labels JSON,
    timestamp TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_metric_name (metric_name),
    INDEX idx_metric_type (metric_type),
    INDEX idx_timestamp (timestamp),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Alerts table - stores alert notifications
CREATE TABLE alerts (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity ENUM('info', 'warning', 'error', 'critical') NOT NULL,
    source VARCHAR(100) NOT NULL,
    status ENUM('active', 'acknowledged', 'resolved') DEFAULT 'active',
    rule_name VARCHAR(255),
    rule_id VARCHAR(36),
    metadata JSON,
    triggered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP NULL,
    acknowledged_by VARCHAR(36),
    resolved_at TIMESTAMP NULL,
    resolved_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_severity (severity),
    INDEX idx_status (status),
    INDEX idx_source (source),
    INDEX idx_triggered_at (triggered_at),
    INDEX idx_rule_id (rule_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Alert notifications table - tracks sent alert notifications
CREATE TABLE alert_notifications (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    alert_id VARCHAR(36) NOT NULL,
    channel_type ENUM('email', 'webhook', 'sms', 'in_app') NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    status ENUM('pending', 'sent', 'failed') DEFAULT 'pending',
    sent_at TIMESTAMP NULL,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    INDEX idx_alert_id (alert_id),
    INDEX idx_channel_type (channel_type),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert test suite example
INSERT INTO test_suites (id, name, type, description, created_by) VALUES
(
    UUID(),
    'Core API Health Check',
    'api',
    'Basic health check tests for core API endpoints',
    'system'
);