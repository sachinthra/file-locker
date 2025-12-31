-- Migration: 000004_enhanced_admin_features.down.sql
-- Description: Rollback enhanced admin features

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS user_announcement_dismissals;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS settings;

-- Drop account_status column from users
ALTER TABLE users DROP COLUMN IF EXISTS account_status;

-- Drop account_status enum type
DROP TYPE IF EXISTS account_status;

-- Drop indexes
DROP INDEX IF EXISTS idx_users_account_status;
DROP INDEX IF EXISTS idx_settings_updated_at;
DROP INDEX IF EXISTS idx_announcements_is_active;
DROP INDEX IF EXISTS idx_announcements_expires_at;
DROP INDEX IF EXISTS idx_announcements_created_at;
DROP INDEX IF EXISTS idx_announcements_target_user_ids;
DROP INDEX IF EXISTS idx_dismissals_user_id;
DROP INDEX IF EXISTS idx_dismissals_announcement_id;
DROP INDEX IF EXISTS idx_dismissals_dismissed_at;
