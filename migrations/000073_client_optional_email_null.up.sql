-- Client email is optional. Keep the unique index for non-NULL values while
-- allowing more than one client to omit an email address.
UPDATE clients
SET email = NULL
WHERE TRIM(email) = '';
