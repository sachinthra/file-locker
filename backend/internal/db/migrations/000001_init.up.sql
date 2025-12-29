-- Migration: 000001_init.up.sql
-- Description: Initial schema for File Locker hybrid architecture (PostgreSQL + Redis)
-- Creates tables for Users and Files (permanent data)

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- USERS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) DEFAULT 'user' NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT users_role_check CHECK (role IN ('user', 'admin'))
);

-- Index for fast username lookup
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- =====================================================
-- FILES TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name VARCHAR(1024) NOT NULL,
    description TEXT,
    mime_type VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    encrypted_size BIGINT NOT NULL,
    minio_path VARCHAR(2048) NOT NULL,
    encryption_key TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    download_count INTEGER DEFAULT 0,
    tags TEXT[] DEFAULT '{}',
    
    -- Constraints
    CONSTRAINT files_size_positive CHECK (size >= 0),
    CONSTRAINT files_encrypted_size_positive CHECK (encrypted_size >= 0),
    CONSTRAINT files_download_count_non_negative CHECK (download_count >= 0)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_files_expires_at ON files(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_files_tags ON files USING GIN(tags);

-- Index for full-text search on file names and descriptions
CREATE INDEX IF NOT EXISTS idx_files_search ON files USING gin(to_tsvector('english', 
    COALESCE(file_name, '') || ' ' || 
    COALESCE(description, '')
));

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Trigger to update updated_at timestamp on users table
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- VIEWS (Optional - for common queries)
-- =====================================================

-- View for active (non-expired) files
CREATE OR REPLACE VIEW active_files AS
SELECT * FROM files
WHERE expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP;

-- View for user file statistics
CREATE OR REPLACE VIEW user_file_stats AS
SELECT 
    user_id,
    COUNT(*) as total_files,
    SUM(size) as total_size,
    SUM(download_count) as total_downloads,
    MAX(created_at) as last_upload
FROM files
GROUP BY user_id;

-- =====================================================
-- COMMENTS
-- =====================================================
COMMENT ON TABLE users IS 'Stores user account information';
COMMENT ON TABLE files IS 'Stores encrypted file metadata';
COMMENT ON COLUMN files.encryption_key IS 'Base64 encoded AES-256 key for file decryption';
COMMENT ON COLUMN files.minio_path IS 'Path/key in MinIO object storage';

-- =====================================================
-- PERSONAL ACCESS TOKENS TABLE
-- =====================================================
-- Create personal_access_tokens table for Personal Access Tokens (PATs)
CREATE TABLE IF NOT EXISTS personal_access_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  token_hash TEXT NOT NULL,
  last_used_at TIMESTAMP WITH TIME ZONE NULL,
  expires_at TIMESTAMP WITH TIME ZONE NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pat_user_id ON personal_access_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_pat_expires_at ON personal_access_tokens(expires_at);
