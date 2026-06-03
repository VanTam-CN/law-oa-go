-- Roll back Batch 01 PostgreSQL real test data.

DELETE FROM inbox_items
WHERE title IN (
    '冲突审查审批 - 红杉资本投资管理咨询合同纠纷案',
    '补齐初步证据目录',
    '复核 B01-CCT-001 冲突检测记录'
);

DELETE FROM approval_snapshots
WHERE approval_request_id = 'B01-APP-001';

DELETE FROM approval_requests
WHERE id IN ('B01-APP-001', 'B01-APP-002')
   OR request_number IN ('B01-APR-001', 'B01-APR-002')
   OR metadata::text LIKE '%batch01_real_seed%';

DELETE FROM risk_audit_events
WHERE payload::text LIKE '%batch01_real_seed%'
   OR subject_id IN ('B01-APP-001', 'B01-APP-002');

DELETE FROM system_settings
WHERE setting_key LIKE 'batch01.%'
   OR setting_value::text LIKE '%batch01_real_seed%';

DELETE FROM conflict_check_records
WHERE check_id = 'B01-CCT-001'
   OR search_parameters::text LIKE '%batch01_real_seed%';

DELETE FROM case_materials
WHERE metadata::text LIKE '%batch01_real_seed%';

DELETE FROM case_intake_parties
WHERE metadata::text LIKE '%batch01_real_seed%';

DELETE FROM case_intakes
WHERE intake_code = 'B01-INTAKE-001'
   OR metadata::text LIKE '%batch01_real_seed%';

DELETE FROM cases
WHERE case_number IN ('B01-CASE-001', 'B01-CASE-002');

DELETE FROM clients
WHERE email IN (
    'batch01.sequoia@client.local',
    'batch01.huaxin@client.local',
    'batch01.tianheng@client.local'
);

DELETE FROM users
WHERE email IN (
    'batch01.admin@example.test',
    'batch01.zhang@example.test',
    'batch01.li@example.test',
    'batch01.compliance@example.test'
);
