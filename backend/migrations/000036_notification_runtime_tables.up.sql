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
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    created_by BIGINT NOT NULL,
    approved_by BIGINT,
    approved_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    sent_retry_count INT NOT NULL DEFAULT 0,
    error_message TEXT,
    contains_sensitive_info BOOLEAN NOT NULL DEFAULT FALSE,
    auto_send BOOLEAN NOT NULL DEFAULT FALSE,
    external_message_id VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notification_queue_status
ON notification_queue(status);

CREATE INDEX IF NOT EXISTS idx_notification_queue_recipient
ON notification_queue(recipient_id, status);

CREATE INDEX IF NOT EXISTS idx_notification_queue_trigger
ON notification_queue(trigger_type, trigger_id);

CREATE INDEX IF NOT EXISTS idx_notification_queue_created
ON notification_queue(created_at);

CREATE TABLE IF NOT EXISTS notification_templates (
    id BIGSERIAL PRIMARY KEY,
    template_code VARCHAR(50) NOT NULL UNIQUE,
    template_name VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    recipient_type VARCHAR(20) NOT NULL,
    trigger_event VARCHAR(100) NOT NULL,
    subject_template VARCHAR(200),
    content_template TEXT NOT NULL,
    variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    auto_send BOOLEAN NOT NULL DEFAULT FALSE,
    requires_approval BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notification_templates_channel_event
ON notification_templates(channel, trigger_event);

INSERT INTO notification_templates (
    template_code,
    template_name,
    channel,
    recipient_type,
    trigger_event,
    subject_template,
    content_template,
    variables,
    auto_send,
    requires_approval
) VALUES
    (
        'SYSTEM_MAINTENANCE',
        '系统维护通知',
        'email',
        'lawyer',
        'system_maintenance',
        '系统维护通知',
        '尊敬的{name}，系统将于{start_time}至{end_time}进行维护，届时将暂停服务。',
        '["name", "start_time", "end_time"]'::jsonb,
        TRUE,
        FALSE
    ),
    (
        'PAYMENT_RECEIVED',
        '收款确认通知',
        'wechat',
        'client',
        'payment_received',
        '收款确认',
        '尊敬的客户，我们已收到您的付款{amount}元，感谢您的配合。',
        '["amount", "payment_date"]'::jsonb,
        TRUE,
        FALSE
    ),
    (
        'CASE_HEARING',
        '开庭提醒',
        'wechat',
        'client',
        'hearing_reminder',
        '开庭提醒',
        '尊敬的客户，您的案件"{case_title}"将于{hearing_date}在{court}开庭。',
        '["case_title", "hearing_date", "court"]'::jsonb,
        FALSE,
        TRUE
    ),
    (
        'CASE_PROGRESS',
        '案件进展通知',
        'wechat',
        'client',
        'case_progress',
        '案件进展',
        '尊敬的客户，您的案件"{case_title}"有新进展：{progress}。',
        '["case_title", "progress"]'::jsonb,
        FALSE,
        TRUE
    )
ON CONFLICT (template_code) DO UPDATE SET
    template_name = EXCLUDED.template_name,
    channel = EXCLUDED.channel,
    recipient_type = EXCLUDED.recipient_type,
    trigger_event = EXCLUDED.trigger_event,
    subject_template = EXCLUDED.subject_template,
    content_template = EXCLUDED.content_template,
    variables = EXCLUDED.variables,
    auto_send = EXCLUDED.auto_send,
    requires_approval = EXCLUDED.requires_approval,
    is_active = TRUE,
    updated_at = now();

INSERT INTO notification_queue (
    trigger_type,
    trigger_id,
    recipient_type,
    recipient_id,
    recipient_name,
    recipient_contact,
    channel,
    subject,
    content,
    template_id,
    status,
    priority,
    created_by,
    contains_sensitive_info,
    auto_send
) VALUES (
    'system_notice',
    1,
    'admin',
    26,
    '示例管理员',
    'demo.admin@example.test',
    'email',
    '示例律师事务所OA试用环境已就绪',
    '系统已接入真实 PostgreSQL 数据、权限管理和核心业务接口，可开始试用。',
    'SYSTEM_MAINTENANCE',
    'pending',
    'normal',
    26,
    FALSE,
    FALSE
)
ON CONFLICT DO NOTHING;
