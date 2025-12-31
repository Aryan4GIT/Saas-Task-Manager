-- Migration: RAG Documents table for vector search
-- Requires pgvector extension to be enabled

-- Enable pgvector extension if not already enabled
CREATE EXTENSION IF NOT EXISTS vector;

-- Create rag_documents table for storing embeddings
CREATE TABLE IF NOT EXISTS rag_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source_type TEXT NOT NULL CHECK (source_type IN ('task', 'issue', 'comment')),
    source_id UUID NOT NULL,
    content TEXT NOT NULL,
    embedding vector(768),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for org_id lookups
CREATE INDEX IF NOT EXISTS idx_rag_documents_org_id ON rag_documents(org_id);

-- Create index for source lookups
CREATE INDEX IF NOT EXISTS idx_rag_documents_source ON rag_documents(source_type, source_id);

-- Create unique constraint to ensure one embedding per source record
CREATE UNIQUE INDEX IF NOT EXISTS idx_rag_documents_unique_source ON rag_documents(org_id, source_type, source_id);

-- Create IVFFlat index for vector similarity search (cosine distance)
-- Using 100 lists for optimal performance with moderate dataset sizes
CREATE INDEX IF NOT EXISTS idx_rag_documents_embedding ON rag_documents 
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
