-- Keep the fictional acceptance accounts on the same local-only password hash.
-- This is not a production credential; it aligns the reviewer seed with the
-- existing lawyer trial accounts so the three-account acceptance can run.
UPDATE users
SET password = '$2a$10$Vle060p9FOiSeUVW7sf33u0sAJXnOJdk/orzAFPeMcSCcVQyT02XO',
    updated_at = CURRENT_TIMESTAMP
WHERE email IN (
    'demo.lawyer@example.test',
    'demo.lawyer.b@example.test',
    'demo.conflict.officer@example.test'
);
