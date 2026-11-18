UPDATE cases
SET
    ethical_wall_enabled = FALSE,
    ethical_wall_description = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE case_number = 'DEMO-HIGH-2026-001';
