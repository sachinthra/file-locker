import { useState, useEffect } from 'preact/hooks';
import { route } from 'preact-router';
import { getUser, getToken } from '../utils/auth';
import api from '../utils/api';
import Toast from '../components/Toast';

export default function Settings({ isAuthenticated, addNotification }) {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [toast, setToast] = useState(null);
  const user = getUser();

  const showToast = (message, type = 'info') => {
    setToast({ message, type });
    if (addNotification) {
      addNotification(message, type);
    }
  };

  const closeToast = () => {
    setToast(null);
  };

  useEffect(() => {
    if (!isAuthenticated || !getToken()) {
      route('/login', true);
    }
  }, [isAuthenticated]);

  const handlePasswordChange = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    // Validation
    if (!currentPassword || !newPassword || !confirmPassword) {
      setError('All fields are required');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('New passwords do not match');
      return;
    }

    if (newPassword.length < 6) {
      setError('New password must be at least 6 characters');
      return;
    }

    if (newPassword === currentPassword) {
      setError('New password must be different from current password');
      return;
    }

    setLoading(true);

    try {
      const response = await api.patch('/user/password', {
        current_password: currentPassword,
        new_password: newPassword,
      });

      setSuccess(response.data.message || 'Password changed successfully');
      showToast('Password changed successfully!', 'success');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err) {
      const errorMsg = err.response?.data?.error || 'Failed to change password';
      setError(errorMsg);
      showToast(errorMsg, 'error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div class="settings-container">
      {toast && <Toast message={toast.message} type={toast.type} onClose={closeToast} />}
        <div class="dashboard-header">
            <h1>Welcome, {user?.username}!</h1>
            <p>Manage your encrypted files securely</p>
      </div>

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
            <input type="text" class="form-input" disabled value={user?.username || 'N/A'} />
            <small>Username cannot be changed</small>
          </div>
          <div class="settings-item">
            <label>Email</label>
            <input type="email" class="form-input" disabled value={user?.email || "user@example.com"} />
            <small>Contact admin to change email</small>
          </div>
        </div>

        {/* Security Settings */}
        <div class="card settings-section">
          <h3>Change Password</h3>
          
          {error && (
            <div class="alert alert-error" style="margin-bottom: 1rem;">
              {error}
            </div>
          )}

          {success && (
            <div class="alert alert-success" style="margin-bottom: 1rem;">
              {success}
            </div>
          )}
          
          <form onSubmit={handlePasswordChange}>
            <div class="settings-item">
              <label>Current Password</label>
              <input
                type="password"
                class="form-input"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder="Enter current password"
                disabled={loading}
                required
              />
            </div>
            <div class="settings-item">
              <label>New Password</label>
              <input
                type="password"
                class="form-input"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="Enter new password (min 6 characters)"
                disabled={loading}
                required
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
                disabled={loading}
                required
              />
            </div>
            <button type="submit" class="btn btn-primary" disabled={loading}>
              {loading ? 'Changing Password...' : 'Change Password'}
            </button>
          </form>
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

        {/* Keyboard Shortcuts */}
        <div class="card settings-section">
          <h3>Keyboard Shortcuts</h3>
          <p style="margin-bottom: 1rem; color: #666;">Use these shortcuts to navigate faster</p>
          
          <div style="display: grid; gap: 0.75rem;">
            <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.5rem; background: var(--bg-color); border-radius: 4px;">
              <span>Focus search box</span>
              <kbd>/</kbd>
            </div>
            
            <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.5rem; background: var(--bg-color); border-radius: 4px;">
              <span>Clear search / Close dialogs</span>
              <kbd>ESC</kbd>
            </div>
            
            <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.5rem; background: var(--bg-color); border-radius: 4px;">
              <span>Export all files</span>
              <div>
                <kbd>⌘</kbd> / <kbd>Ctrl</kbd> + <kbd>E</kbd>
              </div>
            </div>
            
            <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.5rem; background: var(--bg-color); border-radius: 4px;">
              <span>Open settings</span>
              <div>
                <kbd>⌘</kbd> / <kbd>Ctrl</kbd> + <kbd>S</kbd>
              </div>
            </div>
            
            <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.5rem; background: var(--bg-color); border-radius: 4px;">
              <span>Confirm action in dialogs</span>
              <div>
                <kbd>⌘</kbd> / <kbd>Ctrl</kbd> + <kbd>Enter</kbd>
              </div>
            </div>
          </div>
          
          <small style="display: block; margin-top: 1rem; color: #999;">
            <strong>Tip:</strong> Shortcuts are context-aware and won't interfere when typing in input fields.
          </small>
        </div>
      </div>
    </div>
  );
}
