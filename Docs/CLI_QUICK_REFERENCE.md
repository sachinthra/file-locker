# File Locker CLI - Quick Reference Card

## Installation
```bash
cd backend && go build -o fl cmd/cli/main.go
sudo mv fl /usr/local/bin/  # Optional: install globally
```

## Authentication
```bash
fl login --token fl_abc123...        # Login with PAT
fl login -u user -p pass             # Login with credentials
fl logout                            # Logout
fl me                                # Show current user
```

## Files
```bash
fl ls                                # List files
fl ls --json                         # List (JSON output)
fl upload file.pdf                   # Upload file
fl upload file.pdf --tags t1,t2      # Upload with tags
fl download file-id                  # Download file
fl download file-id -o myfile.pdf    # Download with name
fl rm file-id                        # Delete file
fl search "query"                    # Search files
fl export -o backup.zip              # Export all files
fl update file-id --tags new,tags    # Update tags
fl update file-id --name newname.pdf # Rename file
```

## Personal Access Tokens
```bash
fl tokens list                       # List PATs
fl tokens create "Token Name"        # Create PAT
fl tokens revoke token-id            # Revoke PAT
```

## User
```bash
fl password --old old --new new      # Change password
```

## Announcements
```bash
fl announcements                     # List announcements
fl announcements dismiss id          # Dismiss announcement
```

## Admin - System
```bash
fl admin stats                       # System statistics
fl admin settings                    # View settings
fl admin settings key value          # Update setting
```

## Admin - Users
```bash
fl admin users                       # List users
fl admin users --status pending      # Filter by status
fl admin users approve id            # Approve user
fl admin users reject id             # Reject user
fl admin users delete id             # Delete user
fl admin users id status active      # Update status
fl admin users id role admin         # Update role
fl admin users id reset-password     # Reset password
fl admin users id logout             # Force logout
```

## Admin - Files
```bash
fl admin files                       # List all files
fl admin files delete id             # Delete any file
```

## Admin - Storage
```bash
fl admin storage analyze             # Analyze storage
fl admin storage cleanup             # Cleanup orphaned files
```

## Admin - Logs
```bash
fl admin logs                        # View all logs
fl admin logs --action login         # Filter by action
fl admin logs --user_id id           # Filter by user
```

## Admin - Announcements
```bash
fl admin announcements               # List announcements
fl admin announcements create        # Create announcement
  --title "Title" 
  --message "Message" 
  --severity info
fl admin announcements delete id     # Delete announcement
```

## Scripting Examples

### Daily Backup
```bash
#!/bin/bash
fl login --token "$FILELOCKER_TOKEN"
fl export -o "backup-$(date +%Y%m%d).zip"
```

### CI/CD Upload
```bash
#!/bin/bash
fl login --token "$CI_TOKEN"
for file in dist/*.zip; do
  fl upload "$file" --tags ci,build
done
```

### Storage Monitoring
```bash
#!/bin/bash
fl login --token "$ADMIN_TOKEN"
if fl admin storage analyze | grep -q "Orphaned Files:  [1-9]"; then
  fl admin storage cleanup
fi
```

## Exit Codes
- `0`: Success
- `1`: Error

## Environment Variables
```bash
export FILELOCKER_TOKEN="fl_abc123..."
export FILELOCKER_HOST="https://files.example.com"
```

## Help
```bash
fl              # Show all commands
fl --help       # Show all commands
```

---

**Full Documentation:** [Docs/CLI_GUIDE.md](./CLI_GUIDE.md)
