# Database Initialization Scripts

## 📋 Overview

This directory contains SQL scripts for initializing and managing the File Locker database.

---

## 📄 Files

### `001_init.sql`
**Purpose**: Complete database schema for fresh installations

**What it creates**:
- ✅ Users table (with role column: 'user' or 'admin')
- ✅ Files table (with encryption metadata)
- ✅ Personal access tokens table
- ✅ All necessary indexes
- ✅ Triggers for automatic timestamp updates
- ✅ Views for common queries
- ✅ All constraints and relationships

**When to use**: 
- First-time database setup
- Recreating the database from scratch

### `002_set_admin.sql`
**Purpose**: Promote a user to admin role

**What it does**:
- Updates an existing user's role to 'admin'
- Provides options to update by username, email, or ID
- Verifies the update

**When to use**:
- After creating your first user account
- When you need to grant admin access to someone

---

## 🚀 Quick Start

### Default Admin Account

The initialization script automatically creates a default admin user:

- **Username**: `admin`
- **Password**: `password123`
- **Email**: `admin@filelocker.local`
- **⚠️ IMPORTANT**: Change this password immediately after first login!

### For Fresh Installation:

1. **Start PostgreSQL** (via Docker Compose):
   ```bash
   docker compose up -d postgres
   ```

2. **Run the initialization script**:
   ```bash
   psql -U postgres -d file_locker -f backend/init-db/001_init.sql
   ```

3. **Login with default admin**:
   - Go to http://localhost:9010
   - Login with `admin` / `password123`
   - Go to Settings → Change Password

4. **(Optional) Create additional admin users**:
   ```bash
   # Edit 002_set_admin.sql first to set your username
   psql -U postgres -d file_locker -f backend/init-db/002_set_admin.sql
   ```

### For Existing Database:

If your database was created before the admin feature:

1. **Check if role column exists**:
   ```bash
   psql -U postgres -d file_locker -c "\d users"
   ```

2. **If role column is missing**, add it manually:
   ```sql
   ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'user' NOT NULL;
   ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('user', 'admin'));
   CREATE INDEX idx_users_role ON users(role);
   ```

3. **Create default admin user OR promote existing user**:
   
   **Option A - Create default admin**:
   ```sql
   INSERT INTO users (id, username, email, password_hash, role) 
   VALUES (
       'a0000000-0000-0000-0000-000000000001',
       'admin',
       'admin@filelocker.local',
       '$2a$10$rKjXVY7Y.6J3LHqM5JXNxOqCh4CxL2kF4LwHlUdGzNKCYKx3qYdZG',
       'admin'
   )
   ON CONFLICT (username) DO NOTHING;
   ```
   
   **Option B - Promote existing user**:
   ```bash
   psql -U postgres -d file_locker -f backend/init-db/002_set_admin.sql
   ```

---

## 🔐 Default User Roles

- **`user`**: Default role for all new registrations
  - Can upload, download, and manage their own files
  - Can create personal access tokens
  - Cannot access admin features

- **`admin`**: Administrative role
  - All user permissions
  - Can view system statistics
  - Can list all users
  - Can delete users and their files
  - Access to `/admin` dashboard

---

## 📊 Database Connection

Using Docker Compose (default):
```bash
Host: localhost
Port: 5433
Database: file_locker
User: postgres
Password: postgres
```

Using psql:
```bash
psql -h localhost -p 5433 -U postgres -d file_locker
```

---

## ⚠️ Important Notes

1. **001_init.sql is idempotent**: Uses `IF NOT EXISTS` clauses, safe to run multiple times
2. **Backup first**: Always backup your database before running scripts on production
3. **Role security**: The role column has a CHECK constraint - only 'user' or 'admin' values allowed
4. **First admin**: Make sure to create at least one admin user after initial setup
5. **Password security**: Admin users still need strong passwords - role doesn't bypass authentication

---

## 🧪 Verification

After setup, verify everything:

```sql
-- Check users table structure
\d users

-- List all users with roles
SELECT username, email, role, created_at FROM users;

-- Verify admin exists
SELECT username, role FROM users WHERE role = 'admin';

-- Check constraints
\d+ users
```

---

## 🔄 Schema Evolution

**Current Version**: v1.0 (with roles)

Future migrations will be numbered sequentially:
- `003_*.sql` - Next feature
- `004_*.sql` - Another feature
- etc.

Keep track of which migrations have been applied to your database.

---

## 📝 Quick Reference

| Script | Purpose | When to Run |
|--------|---------|-------------|
| `001_init.sql` | Fresh install | Once (first time) |
| `002_set_admin.sql` | Promote user to admin | After user registration |

---

## 🆘 Troubleshooting

### "relation 'users' already exists"
- Normal if running 001_init.sql multiple times
- Uses `IF NOT EXISTS` clauses

### "column 'role' does not exist"
- Your database was created before the admin feature
- Add the role column manually (see instructions above)

### "no rows updated" when setting admin
- Check the username/email/ID in the UPDATE statement
- Verify user exists: `SELECT * FROM users;`

### Connection refused
- Ensure PostgreSQL container is running: `docker compose ps`
- Check port mapping: `docker compose port postgres 5432`

---

## 📚 Additional Resources

- PostgreSQL Documentation: https://www.postgresql.org/docs/
- Docker Compose: https://docs.docker.com/compose/
- Project Documentation: See `/Docs` directory
