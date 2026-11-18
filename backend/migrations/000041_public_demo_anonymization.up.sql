-- Anonymize legacy public-demo fixtures without retaining the former private
-- brand as a source-code literal. Rows are selected only by stable demo markers.

UPDATE users
SET username = 'legacy_demo_' || id::text,
    name = CASE role
        WHEN 'admin' THEN '历史示例管理员'
        WHEN 'assistant' THEN '历史示例助理'
        WHEN 'finance' THEN '历史示例财务'
        ELSE '历史示例律师'
    END,
    email = 'legacy.demo.' || id::text || '@example.test',
    updated_at = CURRENT_TIMESTAMP
WHERE email LIKE '%@law-oa.local';

-- Replace the old primary demo client name inside denormalized audit payloads
-- before updating the client master row itself.
WITH legacy_client AS (
    SELECT id::text AS id, name AS old_name, replace(name, '上海', '') AS old_short_name
    FROM clients
    WHERE source = 'trial_seed' AND email = 'browser.qa@example.com'
    LIMIT 1
)
UPDATE conflict_check_records record
SET client_name = CASE WHEN record.client_id = legacy_client.id THEN '上海示例科技有限公司' ELSE record.client_name END,
    search_parameters = replace(
        replace(COALESCE(record.search_parameters::text, '{}'), legacy_client.old_name, '上海示例科技有限公司'),
        legacy_client.old_short_name,
        '示例科技有限公司'
    )::jsonb,
    check_result = replace(
        replace(COALESCE(record.check_result::text, '{}'), legacy_client.old_name, '上海示例科技有限公司'),
        legacy_client.old_short_name,
        '示例科技有限公司'
    )::jsonb,
    updated_at = CURRENT_TIMESTAMP
FROM legacy_client
WHERE record.client_id = legacy_client.id
   OR COALESCE(record.search_parameters::text, '') LIKE '%' || legacy_client.old_name || '%'
   OR COALESCE(record.search_parameters::text, '') LIKE '%' || legacy_client.old_short_name || '%'
   OR COALESCE(record.check_result::text, '') LIKE '%' || legacy_client.old_name || '%'
   OR COALESCE(record.check_result::text, '') LIKE '%' || legacy_client.old_short_name || '%';

WITH legacy_client AS (
    SELECT id::text AS id, name AS old_name, replace(name, '上海', '') AS old_short_name
    FROM clients
    WHERE source = 'trial_seed' AND email = 'browser.qa@example.com'
    LIMIT 1
)
UPDATE conflict_cases conflict
SET case_no = CASE WHEN case_no ~ '^[A-Z]{2}-' THEN 'DEMO-' || substring(case_no FROM 4) ELSE case_no END,
    case_name = replace(replace(case_name, legacy_client.old_name, '上海示例科技有限公司'), legacy_client.old_short_name, '示例科技有限公司'),
    description = replace(replace(COALESCE(description, ''), legacy_client.old_name, '上海示例科技有限公司'), legacy_client.old_short_name, '示例科技有限公司'),
    conflict_details = replace(replace(COALESCE(conflict_details, ''), legacy_client.old_name, '上海示例科技有限公司'), legacy_client.old_short_name, '示例科技有限公司'),
    opposing_parties = replace(
        replace(COALESCE(opposing_parties::text, '[]'), legacy_client.old_name, '上海示例科技有限公司'),
        legacy_client.old_short_name,
        '示例科技有限公司'
    )::jsonb
FROM legacy_client
WHERE case_no ~ '^[A-Z]{2}-'
   OR case_name LIKE '%' || legacy_client.old_name || '%'
   OR case_name LIKE '%' || legacy_client.old_short_name || '%'
   OR COALESCE(description, '') LIKE '%' || legacy_client.old_name || '%'
   OR COALESCE(description, '') LIKE '%' || legacy_client.old_short_name || '%'
   OR COALESCE(conflict_details, '') LIKE '%' || legacy_client.old_name || '%'
   OR COALESCE(conflict_details, '') LIKE '%' || legacy_client.old_short_name || '%'
   OR COALESCE(opposing_parties::text, '') LIKE '%' || legacy_client.old_name || '%'
   OR COALESCE(opposing_parties::text, '') LIKE '%' || legacy_client.old_short_name || '%';

WITH legacy_client AS (
    SELECT name AS old_name, replace(name, '上海', '') AS old_short_name
    FROM clients
    WHERE source = 'trial_seed' AND email = 'browser.qa@example.com'
    LIMIT 1
)
UPDATE approval_requests approval
SET title = replace(replace(title, legacy_client.old_name, '上海示例科技有限公司'), legacy_client.old_short_name, '示例科技有限公司'),
    content = replace(replace(COALESCE(content, ''), legacy_client.old_name, '上海示例科技有限公司'), legacy_client.old_short_name, '示例科技有限公司'),
    metadata = replace(
        replace(COALESCE(metadata::text, '{}'), legacy_client.old_name, '上海示例科技有限公司'),
        legacy_client.old_short_name,
        '示例科技有限公司'
    )::jsonb,
    conflict_result = replace(
        replace(COALESCE(conflict_result::text, '{}'), legacy_client.old_name, '上海示例科技有限公司'),
        legacy_client.old_short_name,
        '示例科技有限公司'
    )::jsonb,
    updated_at = CURRENT_TIMESTAMP
FROM legacy_client
WHERE title LIKE '%' || legacy_client.old_name || '%'
   OR title LIKE '%' || legacy_client.old_short_name || '%'
   OR COALESCE(content, '') LIKE '%' || legacy_client.old_name || '%'
   OR COALESCE(content, '') LIKE '%' || legacy_client.old_short_name || '%'
   OR COALESCE(metadata::text, '') LIKE '%' || legacy_client.old_name || '%'
   OR COALESCE(metadata::text, '') LIKE '%' || legacy_client.old_short_name || '%'
   OR COALESCE(conflict_result::text, '') LIKE '%' || legacy_client.old_name || '%'
   OR COALESCE(conflict_result::text, '') LIKE '%' || legacy_client.old_short_name || '%';

UPDATE clients
SET name = CASE
        WHEN source = 'trial_seed' AND email = 'browser.qa@example.com' THEN '上海示例科技有限公司'
        WHEN source = 'lawyer_trial_acceptance_seed' AND email LIKE '%@law-oa.local' THEN '历史示例隔离客户B'
        ELSE name
    END,
    company = CASE
        WHEN source = 'trial_seed' AND email = 'browser.qa@example.com' THEN '上海示例科技有限公司'
        WHEN source = 'lawyer_trial_acceptance_seed' AND email LIKE '%@law-oa.local' THEN '历史示例隔离客户B'
        ELSE company
    END,
    email = CASE
        WHEN source = 'lawyer_trial_acceptance_seed' AND email LIKE '%@law-oa.local' THEN 'legacy.demo.client.' || id::text || '@example.test'
        ELSE email
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE (source = 'trial_seed' AND email = 'browser.qa@example.com')
   OR (source = 'lawyer_trial_acceptance_seed' AND email LIKE '%@law-oa.local');

UPDATE cases
SET case_number = CASE WHEN case_number ~ '^[A-Z]{2}-' THEN 'DEMO-' || substring(case_number FROM 4) ELSE case_number END,
    updated_at = CURRENT_TIMESTAMP
WHERE case_number ~ '^[A-Z]{2}-';

-- Repair conflict-approval ownership using the authoritative conflict task owner.
UPDATE approval_requests approval
SET applicant_id = owner.id::text,
    applicant_name = owner.name,
    created_by = owner.id::text,
    updated_at = CURRENT_TIMESTAMP
FROM conflict_check_records record
JOIN users owner ON owner.id = record.user_id
WHERE approval.conflict_check_id = record.check_id
  AND approval.type = 'conflict_approval'
  AND approval.applicant_id IS DISTINCT FROM owner.id::text;

INSERT INTO approval_snapshots (
    approval_request_id, snapshot_type, snapshot_data, source_version, created_at
)
SELECT
    approval.id,
    approval.type,
    jsonb_build_object(
        'snapshot_type', approval.type,
        'source', 'public_demo_repair',
        'approval', jsonb_build_object(
            'id', approval.id,
            'request_number', approval.request_number,
            'title', approval.title,
            'status', approval.status,
            'current_stage', approval.current_stage,
            'current_approver_name', approval.current_approver_name
        ),
        'metadata', approval.metadata,
        'conflict_result', approval.conflict_result
    ),
    1,
    CURRENT_TIMESTAMP
FROM approval_requests approval
WHERE approval.id = 'LAWYER-TRIAL-APPROVAL-001'
  AND NOT EXISTS (
      SELECT 1 FROM approval_snapshots snapshot
      WHERE snapshot.approval_request_id = approval.id
  );
