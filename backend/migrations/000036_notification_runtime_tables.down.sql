DELETE FROM notification_queue
WHERE trigger_type = 'system_notice'
  AND subject = '示例律师事务所OA试用环境已就绪';

DELETE FROM notification_templates
WHERE template_code IN (
    'SYSTEM_MAINTENANCE',
    'PAYMENT_RECEIVED',
    'CASE_HEARING',
    'CASE_PROGRESS'
);

DROP TABLE IF EXISTS notification_templates;
DROP TABLE IF EXISTS notification_queue;
