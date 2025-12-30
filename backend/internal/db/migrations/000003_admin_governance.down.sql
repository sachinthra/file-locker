-- Migration: 000003_admin_governance.down.sql
-- Description: Rollback admin governance features

-- Drop indexes first
DROP INDEX IF EXISTS idx_audit_logs_target;
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_actor_id;

-- Drop audit_logs table
DROP TABLE IF EXISTS audit_logs;

-- Drop is_active column index
DROP INDEX IF EXISTS idx_users_is_active;

-- Drop is_active column from users
ALTER TABLE users DROP COLUMN IF EXISTS is_active;
