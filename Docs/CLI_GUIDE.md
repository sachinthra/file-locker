# File Locker CLI Guide

The File Locker CLI (`fl`) provides command-line access to all File Locker API functionality. It's ideal for automation, scripting, and CI/CD integration.

## Installation

Build the CLI from source:

```bash
cd backend
go build -o fl cmd/cli/main.go
sudo mv fl /usr/local/bin/  # Optional: install globally
```

## Configuration

The CLI stores configuration in `~/.filelocker.json`:

```json
{
  "base_url": "http://localhost:9010/api/v1",
  "token": "your-auth-token"
}
```

## Authentication

### Login with Personal Access Token (Recommended)

```bash
fl login --token fl_abc123...
```

### Login with Username/Password

```bash
fl login -u username -p password
```

### Set Custom Server URL

```bash
fl login --host https://filelocker.example.com
```

### Logout

```bash
fl logout
```

### Show Current User

```bash
fl me
# or
fl whoami
```

**Output:**
```
User ID:      uuid-here
Username:     john.doe
Email:        john@example.com
Role:         admin
Member Since: 2024-01-15
```

---

## File Operations

### List Files

```bash
# Table format (default)
fl ls

# JSON format (for scripts)
fl ls --json
```

### Upload File

```bash
# Basic upload
fl upload document.pdf

# With tags
fl upload report.pdf --tags work,quarterly,finance

# With expiration (hours)
fl upload temp.zip --expire 24
```

### Download File

```bash
# Download with original name
fl download file-id-here

# Download with custom name
fl download file-id-here -o myfile.pdf
```

### Delete File

```bash
fl rm file-id-here
```

### Search Files

```bash
fl search "project files"
fl search quarterly
```

**Output:**
```
ID          NAME              SIZE      TAGS
e3f5a2...   project-plan.pdf  2.3 MB    work, planning
b7c8d1...   quarterly.xlsx    512 KB    finance, quarterly
```

### Export All Files

```bash
# Export to default name (filelocker-export.zip)
fl export

# Custom output filename
fl export -o backup-2024.zip
```

Shows progress bar during download.

### Update File Metadata

```bash
# Update tags
fl update file-id --tags new,tags,here

# Rename file
fl update file-id --name newfilename.pdf

# Both
fl update file-id --tags important --name report-final.pdf
```

---

## Personal Access Tokens

### List Tokens

```bash
fl tokens list
```

**Output:**
```
ID          NAME             CREATED        EXPIRES    LAST USED
a1b2c3...   CI/CD Pipeline   2 weeks ago    Never      5 hours ago
d4e5f6...   Dev Laptop       1 month ago    in 5 days  2 days ago
```

### Create Token

```bash
# Never expires
fl tokens create "My New Token"

# With expiration date
fl tokens create "Temporary Token" --expire 2024-12-31
```

**Output:**
```
✅ Token created successfully!
Name:  My New Token
Token: fl_abc123def456ghi789...

⚠️  Save this token now - you won't be able to see it again!
```

### Revoke Token

```bash
fl tokens revoke token-id-here
```

---

## User Management

### Change Password

```bash
# Non-interactive (for scripts)
fl password --old oldpass123 --new newpass456
```

---

## Announcements

### List Announcements

```bash
fl announcements
```

**Output:**
```
ℹ️ [info] System Maintenance
   Scheduled maintenance on Saturday 2AM-4AM UTC
   ID: announcement-id-1

⚠️ [warning] Storage Quota Warning
   You're using 95% of your storage quota
   ID: announcement-id-2
```

### Dismiss Announcement

```bash
fl announcements dismiss announcement-id-here
```

---

## Admin Commands

⚠️ **Admin commands require admin role**

### System Statistics

```bash
fl admin stats
```

**Output:**
```
📊 System Statistics
Total Users:     42
Total Files:     1,337
Storage Used:    127.5 GB
Active Sessions: 8
```

### User Management

#### List Users

```bash
# All users
fl admin users

# Filter by status
fl admin users --status pending
fl admin users --status active
fl admin users --status inactive
```

#### Approve/Reject Pending Users

```bash
fl admin users approve user-id
fl admin users reject user-id
```

#### Delete User

```bash
fl admin users delete user-id
```

#### Update User Status

```bash
fl admin users user-id status active
fl admin users user-id status inactive
```

#### Update User Role

```bash
fl admin users user-id role admin
fl admin users user-id role user
```

#### Reset User Password

```bash
fl admin users user-id reset-password
```

**Output:**
```
✅ Password reset successfully
Temporary Password: temp_abc123def
```

#### Force Logout User

```bash
fl admin users user-id logout
```

### Settings Management

#### View All Settings

```bash
fl admin settings
```

**Output:**
```
⚙️  System Settings:
max_file_size: 104857600
allow_registration: true
require_email_verification: false
default_file_expiry_hours: 720
```

#### Update Setting

```bash
fl admin settings max_file_size 209715200
fl admin settings allow_registration false
```

### File Management

#### List All Files

```bash
fl admin files
```

Shows all files from all users.

#### Delete Any File

```bash
fl admin files delete file-id
```

### Storage Management

#### Analyze Storage

```bash
fl admin storage analyze
```

**Output:**
```
💾 Storage Analysis:
Total Files:     1,337
Total Size:      127.5 GB
Orphaned Files:  3
Expired Files:   12
```

#### Cleanup Storage

```bash
fl admin storage cleanup
```

Removes orphaned and expired files.

**Output:**
```
🧹 Cleanup completed!
Files Deleted:  15
Space Freed:    2.3 GB
```

### Audit Logs

#### View All Logs

```bash
fl admin logs
```

#### Filter by Action

```bash
fl admin logs --action login
fl admin logs --action file_upload
fl admin logs --action file_delete
```

#### Filter by User

```bash
fl admin logs --user_id user-uuid-here
```

**Output:**
```
[2024-01-20 14:32:15] file_upload - a1b2c3d4... (192.168.1.100) - Uploaded: report.pdf
[2024-01-20 14:30:22] login - a1b2c3d4... (192.168.1.100) - Successful login
```

### Announcement Management

#### List All Announcements (Admin View)

```bash
fl admin announcements
```

Shows all announcements including metadata.

#### Create Announcement

```bash
fl admin announcements create \
  --title "System Maintenance" \
  --message "Scheduled maintenance on Saturday" \
  --severity info
```

Severity options: `info`, `warning`, `error`

#### Delete Announcement

```bash
fl admin announcements delete announcement-id
```

---

## Scripting Examples

### Automated Backup Script

```bash
#!/bin/bash
# Daily backup of File Locker

# Login with token
fl login --token "$FILELOCKER_TOKEN"

# Export all files
BACKUP_FILE="backup-$(date +%Y%m%d).zip"
fl export -o "$BACKUP_FILE"

# Upload to S3 (or other backup storage)
aws s3 cp "$BACKUP_FILE" "s3://backups/filelocker/"

# Cleanup
rm "$BACKUP_FILE"
```

### CI/CD File Upload

```bash
#!/bin/bash
# Upload build artifacts to File Locker

fl login --token "$CI_FILELOCKER_TOKEN"

for file in dist/*.zip; do
  echo "Uploading $file..."
  fl upload "$file" --tags ci,build,$(git rev-parse --short HEAD)
done
```

### User Approval Automation

```bash
#!/bin/bash
# Auto-approve users from specific domain

fl login --token "$ADMIN_TOKEN"

# Get pending users (requires jq for JSON parsing)
PENDING_USERS=$(fl admin users --status pending | grep "@example.com")

# Parse and approve
# (Additional parsing logic would be needed for full automation)
```

### Storage Monitoring

```bash
#!/bin/bash
# Monitor storage and alert if threshold exceeded

fl login --token "$MONITORING_TOKEN"

STATS=$(fl admin storage analyze)
TOTAL_SIZE=$(echo "$STATS" | grep "Total Size" | awk '{print $3}')

if [ "$TOTAL_SIZE" -gt 100 ]; then
  echo "Storage alert: ${TOTAL_SIZE}GB used!"
  # Send alert via email/Slack/etc
fi
```

---

## Exit Codes

- `0`: Success
- `1`: Error (check stderr for details)

---

## Troubleshooting

### "Error: no token found"

Run `fl login` first to authenticate.

### "Error: unauthorized (status 401)"

Your token may have expired. Login again:

```bash
fl logout
fl login --token new-token
```

### "Error: failed to connect"

Check server URL in `~/.filelocker.json` or specify with `--host`:

```bash
fl login --host https://correct-url.com
```

### Custom Server URL

For production deployments:

```bash
# First-time setup
fl login --host https://files.company.com --token your-token

# Subsequent commands use saved URL
fl ls
fl upload file.pdf
```

---

## API Coverage

The CLI provides complete coverage of the File Locker REST API:

| API Category | Coverage | Commands |
|--------------|----------|----------|
| Authentication | ✅ 100% | login, logout, me |
| Files | ✅ 100% | ls, upload, download, rm, search, export, update |
| Tokens | ✅ 100% | tokens list/create/revoke |
| User | ✅ 100% | password |
| Announcements | ✅ 100% | announcements, announcements dismiss |
| Admin Stats | ✅ 100% | admin stats |
| Admin Users | ✅ 100% | admin users (all operations) |
| Admin Settings | ✅ 100% | admin settings |
| Admin Files | ✅ 100% | admin files |
| Admin Storage | ✅ 100% | admin storage analyze/cleanup |
| Admin Logs | ✅ 100% | admin logs |
| Admin Announcements | ✅ 100% | admin announcements |

**Total: 32 API endpoints, 100% coverage**

---

## Integration with Other Tools

### curl Alternative

```bash
# Instead of complex curl commands:
curl -H "Authorization: Bearer token" http://localhost:9010/api/v1/files

# Use simple CLI:
fl ls
```

### jq for JSON Processing

```bash
# Get file count
fl ls --json | jq 'length'

# Extract specific fields
fl ls --json | jq '.[] | {name: .file_name, size: .size}'
```

### Watch for Changes

```bash
# Monitor file count
watch -n 30 'fl ls | wc -l'
```

---

## Environment Variables

```bash
# Set token via environment
export FILELOCKER_TOKEN="fl_abc123..."
fl login --token "$FILELOCKER_TOKEN"

# Set host via environment
export FILELOCKER_HOST="https://files.example.com"
fl login --host "$FILELOCKER_HOST"
```

---

## Best Practices

1. **Use PATs for Automation**: Create dedicated tokens for scripts/CI/CD
2. **Set Expiration for Temporary Access**: Use `--expire` for time-limited access
3. **Tag Files Consistently**: Use standardized tags for easier searching
4. **Regular Backups**: Use `fl export` for periodic backups
5. **Monitor Storage**: Regular `admin storage analyze` for admins
6. **Audit Regularly**: Review `admin logs` for security monitoring

---

## Support

- **Documentation**: See `/docs` in web UI or Swagger at `/swagger/`
- **API Reference**: [backend/docs/openapi.yaml](../backend/docs/openapi.yaml)
- **Issues**: Check logs with `fl admin logs` (admin only)
