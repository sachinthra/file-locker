-- Migration: 000001_init.down.sql
-- Description: Rollback initial schema

DROP VIEW IF EXISTS user_file_stats;
DROP VIEW IF EXISTS active_files;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS personal_access_tokens;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "uuid-ossp";
