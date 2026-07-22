-- Intentionally irreversible. On MySQL these tables may have been created by
-- migration 000016, so dropping them while rolling back this compensating
-- migration could destroy pre-existing waiver history.
SELECT 1;
