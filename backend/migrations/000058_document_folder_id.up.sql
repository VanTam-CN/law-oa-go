-- Link uploaded documents to the case-folder tree so folder counts are real.
ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS folder_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_documents_folder_id
    ON documents (folder_id);
