-- Migration: Add document fields to tasks table
-- This adds support for document upload when marking tasks as done

-- Add document-related columns
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS document_filename VARCHAR(255);
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS document_path VARCHAR(500);
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS document_summary TEXT;

-- Create index for querying tasks with documents
CREATE INDEX IF NOT EXISTS idx_tasks_document_filename ON tasks(document_filename) WHERE document_filename IS NOT NULL;
