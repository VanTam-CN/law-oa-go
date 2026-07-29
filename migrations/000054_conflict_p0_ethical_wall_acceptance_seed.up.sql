-- Add a protected historical matter to the fictional A/B/reviewer acceptance fixture.
-- This lets the three-account browser acceptance prove that protected details are
-- retained for the independent reviewer but never disclosed to the requesting lawyer.

UPDATE cases
SET
    ethical_wall_enabled = TRUE,
    ethical_wall_description = '虚构验收数据：仅独立冲突核查人可复核历史事项证据。',
    updated_at = CURRENT_TIMESTAMP
WHERE case_number = 'DEMO-HIGH-2026-001';
