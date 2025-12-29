-- Migration: 000002_add_admin.up.sql
-- Description: Add default admin user
-- IMPORTANT: Change this password immediately after first login!

INSERT INTO users (id, username, email, password_hash, role) 
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'admin',
    'admin@filelocker.local',
    '$2a$10$Ak5bcSPwUWFPITX.F/onhewwBPSx2uLeIHLnD./1vX/ZW4oGB2P3W',
    'admin'
)
ON CONFLICT (username) DO NOTHING;
