-- Migration: Add workflow fields to tasks table
-- This adds support for the verification/approval workflow:
-- todo -> in_progress -> done -> verified -> approved

-- Add new status values to the constraint
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_status_check;
ALTER TABLE tasks ADD CONSTRAINT tasks_status_check 
    CHECK (status IN ('todo', 'in_progress', 'done', 'verified', 'approved', 'blocked'));

-- Add verified_by and approved_by columns
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS verified_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP WITH TIME ZONE;

-- Create indexes for the new columns
CREATE INDEX IF NOT EXISTS idx_tasks_verified_by ON tasks(verified_by);
CREATE INDEX IF NOT EXISTS idx_tasks_approved_by ON tasks(approved_by);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
