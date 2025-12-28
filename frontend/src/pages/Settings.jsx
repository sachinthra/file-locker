import { useState } from 'preact/hooks';

export default function Settings() {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [message, setMessage] = useState('');

  const handlePasswordChange = (e) => {
    e.preventDefault();
    setMessage('⚠️ Password change feature is not yet implemented');
    // TODO: Implement password change API call
  };

  return (
    <div class="settings-container">
      <div class="settings-header">
        <h1>Settings</h1>
        <p>Manage your account preferences and security</p>
      </div>

      <div class="settings-grid">
        {/* Account Settings */}
        <div class="card settings-section">
          <h3>Account Information</h3>
          <div class="settings-item">
            <label>Username</label>
            <input type="text" class="form-input" disabled value="Current user" />
            <small>Username cannot be changed</small>
          </div>
          <div class="settings-item">
            <label>Email</label>
            <input type="email" class="form-input" disabled value="user@example.com" />
            <small>Contact admin to change email</small>
          </div>
        </div>

        {/* Security Settings */}
        <div class="card settings-section">
          <h3>Security</h3>
          {message && <div class="alert alert-warning">{message}</div>}
          <form onSubmit={handlePasswordChange}>
            <div class="settings-item">
              <label>Current Password</label>
              <input
                type="password"
                class="form-input"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder="Enter current password"
              />
            </div>
            <div class="settings-item">
              <label>New Password</label>
              <input
                type="password"
                class="form-input"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="Enter new password"
              />
            </div>
            <div class="settings-item">
              <label>Confirm New Password</label>
              <input
                type="password"
                class="form-input"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Confirm new password"
              />
            </div>
            <button type="submit" class="btn btn-primary" disabled>
              Change Password (Coming Soon)
            </button>
          </form>
        </div>

        {/* Storage Settings */}
        <div class="card settings-section">
          <h3>Storage</h3>
          <div class="settings-item">
            <label>Default File Expiration</label>
            <select class="form-input" disabled>
              <option>Never</option>
              <option>24 hours</option>
              <option>7 days</option>
              <option>30 days</option>
            </select>
            <small>Coming soon</small>
          </div>
          <div class="settings-item">
            <label>Auto-delete After Download</label>
            <input type="checkbox" disabled />
            <small>Coming soon</small>
          </div>
        </div>

        {/* Appearance Settings */}
        <div class="card settings-section">
          <h3>Appearance</h3>
          <div class="settings-item">
            <label>Theme</label>
            <p>Use the theme toggle in the header to switch between light and dark modes</p>
          </div>
        </div>

        {/* Notifications */}
        <div class="card settings-section">
          <h3>Notifications</h3>
          <div class="settings-item">
            <label>Email Notifications</label>
            <input type="checkbox" disabled />
            <small>Notify when files expire (Coming soon)</small>
          </div>
          <div class="settings-item">
            <label>Upload Completion</label>
            <input type="checkbox" disabled />
            <small>Desktop notifications (Coming soon)</small>
          </div>
        </div>

        {/* Advanced */}
        <div class="card settings-section">
          <h3>Advanced</h3>
          <div class="settings-item">
            <label>API Access</label>
            <button class="btn btn-secondary" disabled>Generate API Key (Coming Soon)</button>
            <small>Use API keys for programmatic access</small>
          </div>
          <div class="settings-item">
            <label>Export Data</label>
            <button class="btn btn-secondary" disabled>Download All Files (Coming Soon)</button>
            <small>Export all your encrypted files</small>
          </div>
        </div>
      </div>
    </div>
  );
}
