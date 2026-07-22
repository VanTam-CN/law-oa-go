-- Normalize historical localized case-type values to the API's canonical enum.
UPDATE cases
SET case_type = CASE case_type
    WHEN '商事' THEN 'commercial'
    WHEN '民事' THEN 'civil'
    WHEN '民事诉讼' THEN 'civil_litigation'
    WHEN '建设工程' THEN 'construction'
    WHEN '劳动' THEN 'labor'
    WHEN '知识产权' THEN 'intellectual'
    WHEN '刑事' THEN 'criminal'
    WHEN '行政' THEN 'administrative'
    ELSE case_type
END,
updated_at = CURRENT_TIMESTAMP
WHERE case_type IN ('商事', '民事', '民事诉讼', '建设工程', '劳动', '知识产权', '刑事', '行政');

UPDATE approval_requests
SET metadata = jsonb_set(
        metadata,
        '{case_creation_config,case_type}',
        to_jsonb(CASE metadata #>> '{case_creation_config,case_type}'
            WHEN '商事' THEN 'commercial'
            WHEN '民事' THEN 'civil'
            WHEN '民事诉讼' THEN 'civil_litigation'
            WHEN '建设工程' THEN 'construction'
            WHEN '劳动' THEN 'labor'
            WHEN '知识产权' THEN 'intellectual'
            WHEN '刑事' THEN 'criminal'
            WHEN '行政' THEN 'administrative'
        END),
        true
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE metadata #>> '{case_creation_config,case_type}' IN
    ('商事', '民事', '民事诉讼', '建设工程', '劳动', '知识产权', '刑事', '行政');
