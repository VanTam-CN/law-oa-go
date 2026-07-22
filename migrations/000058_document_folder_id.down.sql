DROP INDEX IF EXISTS idx_documents_folder_id;

ALTER TABLE documents
    DROP COLUMN IF EXISTS folder_id;
