-- Intentionally irreversible: converting multiple NULL emails back to the
-- same empty string would violate the existing unique email index.
SELECT 1;
