-- Script: 002_set_admin.sql
-- Description: Promote a user to admin role
-- Note: This assumes you've already run 001_init.sql which creates the role column

-- Option 1: Update by username
-- IMPORTANT: Change 'your_username' to your actual username
UPDATE users SET role = 'admin' WHERE username = 'admin';

-- Option 2: Update by email
-- UPDATE users SET role = 'admin' WHERE email = 'your_email@example.com';

-- Option 3: Update by user ID
-- UPDATE users SET role = 'admin' WHERE id = 'your-user-id-here';

-- Verify the update
SELECT id, username, email, role, created_at 
FROM users 
WHERE role = 'admin'
ORDER BY created_at;
