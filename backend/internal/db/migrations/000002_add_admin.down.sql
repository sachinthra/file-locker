-- Migration: 000002_add_admin.down.sql
-- Description: Remove default admin user

DELETE FROM users WHERE id = 'a0000000-0000-0000-0000-000000000001';
