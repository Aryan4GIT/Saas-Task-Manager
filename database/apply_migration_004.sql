-- Apply migration 004: Add document fields to tasks table

\c saas_db;

-- Add document-related columns
ALTER TABLE tasks 
    ADD COLUMN IF NOT EXISTS document_filename VARCHAR(255),
    ADD COLUMN IF NOT EXISTS document_path VARCHAR(500),
    ADD COLUMN IF NOT EXISTS document_summary TEXT;

-- Create index for querying tasks with documents
CREATE INDEX IF NOT EXISTS idx_tasks_document_filename ON tasks(document_filename) WHERE document_filename IS NOT NULL;

COMMIT;

-- Verify the migration
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'tasks' 
AND column_name IN ('document_filename', 'document_path', 'document_summary');
