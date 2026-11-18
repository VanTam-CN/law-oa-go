-- A case owner must retain access after enabling an ethical wall. Older rows
-- could enable the wall without recording the owner in the whitelist, which
-- made the owner lose access to the matter and its client profile.
INSERT INTO case_ethical_wall_whitelist (case_id, user_id, granted_by, granted_at, reason)
SELECT
    c.id,
    c.lawyer_id,
    COALESCE(c.ethical_wall_enabled_by, c.lawyer_id),
    COALESCE(c.ethical_wall_enabled_at, CURRENT_TIMESTAMP),
    '系统回填：案件承办律师保留工作访问权'
FROM cases c
WHERE c.deleted_at IS NULL
  AND c.ethical_wall_enabled = TRUE
  AND c.lawyer_id IS NOT NULL
ON CONFLICT (case_id, user_id) DO NOTHING;
