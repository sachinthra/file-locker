-- Migration: 000004_enhanced_admin_features.up.sql
-- Description: Add account status, settings, announcements, and dismissals

-- =================================================================
-- 1. ADD ACCOUNT STATUS ENUM AND COLUMN
-- =================================================================

-- Create account_status enum type
DO $$ BEGIN
    CREATE TYPE account_status AS ENUM ('pending', 'active', 'rejected', 'suspended');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Add account_status column to users table (default to 'active' for existing users)
ALTER TABLE users ADD COLUMN IF NOT EXISTS account_status account_status DEFAULT 'active' NOT NULL;

-- Create index for faster queries on account status
CREATE INDEX IF NOT EXISTS idx_users_account_status ON users(account_status);

-- Update existing users to have proper status based on is_active
-- This migration handles the transition from is_active to account_status
UPDATE users 
SET account_status = CASE 
    WHEN is_active = false THEN 'suspended'::account_status
    ELSE 'active'::account_status
END
WHERE account_status = 'active'; -- Only update if not already modified

-- =================================================================
-- 2. SETTINGS TABLE FOR GLOBAL CONFIGURATION
-- =================================================================

CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Insert default settings
INSERT INTO settings (key, value, description) 
VALUES 
    ('registration_auto_approve', 'false', 'Automatically approve new user registrations'),
    ('max_file_size_bytes', '104857600', 'Maximum file size in bytes (default 100MB)'),
    ('storage_quota_per_user_bytes', '1073741824', 'Storage quota per user in bytes (default 1GB)')
ON CONFLICT (key) DO NOTHING;

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_settings_updated_at ON settings(updated_at DESC);

-- =================================================================
-- 3. ANNOUNCEMENTS TABLE
-- =================================================================

CREATE TABLE IF NOT EXISTS announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'info', -- 'info', 'warning', 'critical'
    target_type VARCHAR(20) NOT NULL DEFAULT 'all', -- 'all', 'specific_users'
    target_user_ids UUID[], -- Array of user IDs if target_type = 'specific_users'
    is_active BOOLEAN DEFAULT true NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for announcement queries
CREATE INDEX IF NOT EXISTS idx_announcements_is_active ON announcements(is_active);
CREATE INDEX IF NOT EXISTS idx_announcements_expires_at ON announcements(expires_at);
CREATE INDEX IF NOT EXISTS idx_announcements_created_at ON announcements(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_announcements_target_user_ids ON announcements USING GIN(target_user_ids);

-- =================================================================
-- 4. USER ANNOUNCEMENT DISMISSALS TABLE
-- =================================================================

CREATE TABLE IF NOT EXISTS user_announcement_dismissals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    dismissed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, announcement_id)
);

-- Create indexes for dismissal queries
CREATE INDEX IF NOT EXISTS idx_dismissals_user_id ON user_announcement_dismissals(user_id);
CREATE INDEX IF NOT EXISTS idx_dismissals_announcement_id ON user_announcement_dismissals(announcement_id);
CREATE INDEX IF NOT EXISTS idx_dismissals_dismissed_at ON user_announcement_dismissals(dismissed_at DESC);

-- =================================================================
-- 5. ADD CONSTRAINT CHECK FOR ANNOUNCEMENT TYPES
-- =================================================================

ALTER TABLE announcements 
ADD CONSTRAINT check_announcement_type 
CHECK (type IN ('info', 'warning', 'critical'));

ALTER TABLE announcements 
ADD CONSTRAINT check_target_type 
CHECK (target_type IN ('all', 'specific_users'));

-- =================================================================
-- MIGRATION COMPLETE
-- =================================================================
