import { useState, useEffect } from 'preact/hooks';
import { route } from 'preact-router';
import { getUser, getToken } from '../utils/auth';
import api from '../utils/api';
import Toast from '../components/Toast';
import ConfirmDialog from '../components/ConfirmDialog';

export default function AdminUserDetails({ isAuthenticated, userId }) {
  const [userDetails, setUserDetails] = useState(null);
  const [userFiles, setUserFiles] = useState([]);
  const [userActivity, setUserActivity] = useState([]);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState(null);
  const [confirmDialog, setConfirmDialog] = useState(null);
  const currentUser = getUser();

  const showToast = (message, type = 'info') => {
    setToast({ message, type });
  };

  const closeToast = () => {
    setToast(null);
  };

  useEffect(() => {
    if (!isAuthenticated || !getToken() || currentUser?.role !== 'admin') {
      route('/admin', true);
      return;
    }
    if (!userId) {
      route('/admin', true);
      return;
    }
    loadUserData();
  }, [isAuthenticated, userId]);

  const loadUserData = async () => {
    setLoading(true);
    try {
      // Load all users to find the specific user
      const usersResponse = await api.get('/admin/users');
      const users = usersResponse.data?.users || [];
      const foundUser = users.find(u => u.id === userId);
      
      if (!foundUser) {
        showToast('User not found', 'error');
        setTimeout(() => route('/admin', true), 2000);
        return;
      }

      setUserDetails(foundUser);

      // Load user's files
      try {
        const filesResponse = await api.get('/admin/files');
        const allFiles = filesResponse.data?.files || [];
        setUserFiles(allFiles.filter(f => f.user_id === userId));
      } catch (err) {
        console.error('Failed to load user files:', err);
      }

      // Load user's activity from audit logs
      try {
        const logsResponse = await api.get('/admin/logs?limit=1000');
        const allLogs = logsResponse.data?.logs || [];
        setUserActivity(allLogs.filter(log => 
          log.actor_id === userId || log.target_id === userId
        ).slice(0, 50));
      } catch (err) {
        console.error('Failed to load user activity:', err);
      }
    } catch (err) {
      showToast('Failed to load user data', 'error');
      console.error('Load user data error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSuspend = () => {
    const currentUserId = currentUser?.user_id || currentUser?.id;
    if (String(userId) === String(currentUserId)) {
      showToast('You cannot suspend your own account', 'error');
      return;
    }
    const action = userDetails.is_active !== false ? 'suspend' : 'activate';
    setConfirmDialog({
      type: 'suspend',
      newStatus: !userDetails.is_active,
      title: userDetails.is_active !== false ? '🚫 Suspend User?' : '✅ Activate User?',
      message: userDetails.is_active !== false
        ? `Suspend user "${userDetails.username}"? They will be immediately logged out and unable to access their account.`
        : `Activate user "${userDetails.username}"? They will be able to log in and access their account again.`,
      confirmText: userDetails.is_active !== false ? 'Yes, Suspend' : 'Yes, Activate',
      confirmStyle: userDetails.is_active !== false ? 'danger' : 'primary'
    });
  };

  const handleChangeRole = () => {
    const currentUserId = currentUser?.user_id || currentUser?.id;
    if (String(userId) === String(currentUserId)) {
      showToast('You cannot change your own role', 'error');
      return;
    }
    const newRole = userDetails.role === 'admin' ? 'user' : 'admin';
    setConfirmDialog({
      type: 'role',
      newRole,
      title: newRole === 'admin' ? '🛡️ Promote to Admin?' : '👤 Demote to User?',
      message: newRole === 'admin'
        ? `Promote "${userDetails.username}" to administrator? They will have full access to admin features.`
        : `Demote "${userDetails.username}" to regular user? They will lose access to admin features.`,
      confirmText: newRole === 'admin' ? 'Yes, Promote' : 'Yes, Demote',
      confirmStyle: 'primary'
    });
  };

  const handleResetPassword = () => {
    const currentUserId = currentUser?.user_id || currentUser?.id;
    if (String(userId) === String(currentUserId)) {
      showToast('You cannot reset your own password from here. Use Settings instead.', 'error');
      return;
    }
    setConfirmDialog({
      type: 'reset',
      title: '🔑 Reset Password?',
      message: `Reset password for "${userDetails.username}"? They will be logged out and need to contact you for the new password.`,
      confirmText: 'Yes, Reset Password',
      confirmStyle: 'warning'
    });
  };

  const handleForceLogout = () => {    if (String(userId) === String(currentUser?.id)) {
      showToast('You cannot force logout yourself. Use the normal logout instead.', 'error');
      return;
    }    setConfirmDialog({
      type: 'logout',
      title: '🚪 Force Logout?',
      message: `Force logout user "${userDetails.username}"? All their active sessions will be terminated.`,
      confirmText: 'Yes, Force Logout',
      confirmStyle: 'warning'
    });
  };

  const handleDeleteUser = () => {
    const currentUserId = currentUser?.user_id || currentUser?.id;
    if (String(userId) === String(currentUserId)) {
      showToast('You cannot delete your own account', 'error');
      return;
    }
    setConfirmDialog({
      type: 'delete',
      title: '⚠️ Delete User?',
      message: `Are you sure you want to delete user "${userDetails.username}"? This will permanently delete their account and all their files. This action cannot be undone.`,
      confirmText: 'Yes, Delete User',
      confirmStyle: 'danger'
    });
  };

  const confirmAction = async () => {
    const dialog = confirmDialog;
    setConfirmDialog(null);

    try {
      switch (dialog.type) {
        case 'suspend':
          await api.patch(`/admin/users/${userId}/status`, { is_active: dialog.newStatus });
          showToast(`User ${dialog.newStatus ? 'activated' : 'suspended'} successfully`, 'success');
          loadUserData();
          break;
        case 'role':
          await api.patch(`/admin/users/${userId}/role`, { role: dialog.newRole });
          showToast(`User ${dialog.newRole === 'admin' ? 'promoted to admin' : 'demoted to user'}`, 'success');
          loadUserData();
          break;
        case 'reset':
          const resetResponse = await api.post(`/admin/users/${userId}/reset-password`);
          const newPassword = resetResponse.data?.new_password || 'N/A';
          showToast(`Password reset! New password: ${newPassword}`, 'success');
          break;
        case 'logout':
          await api.post(`/admin/users/${userId}/logout`);
          showToast('User logged out from all sessions', 'success');
          break;
        case 'delete':
          await api.delete(`/admin/users/${userId}`);
          showToast('User deleted successfully', 'success');
          setTimeout(() => route('/admin', true), 1500);
          break;
      }
    } catch (err) {
      showToast(err.response?.data?.error || `Failed to ${dialog.type} user`, 'error');
      console.error(`${dialog.type} user error:`, err);
    }
  };

  const cancelAction = () => {
    setConfirmDialog(null);
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  if (loading) {
    return (
      <div class="loading" style="min-height: 100vh; display: flex; align-items: center; justify-content: center;">
        <div>
          <div class="spinner"></div>
          <p>Loading user details...</p>
        </div>
      </div>
    );
  }

  if (!userDetails) {
    return (
      <div class="admin-container" style="padding: 2rem;">
        <div class="alert alert-error">User not found</div>
      </div>
    );
  }

  // Ensure proper string comparison for UUIDs
  // Backend returns user_id, not id
  const currentUserId = currentUser?.user_id || currentUser?.id;
  const isCurrentUser = String(userId) === String(currentUserId);
  const isAdmin = userDetails.role === 'admin';
  
  // Debug log to verify comparison
  console.log('[AdminUserDetails] userId:', userId, 'type:', typeof userId);
  console.log('[AdminUserDetails] currentUser.user_id:', currentUser?.user_id, 'type:', typeof currentUser?.user_id);
  console.log('[AdminUserDetails] isCurrentUser:', isCurrentUser);

  return (
    <div class="admin-container" style="padding: 2rem; max-width: 1400px; margin: 0 auto;">
      {toast && <Toast message={toast.message} type={toast.type} onClose={closeToast} />}

      {/* Header */}
      <div style="margin-bottom: 2rem;">
        <button 
          class="btn btn-secondary btn-sm"
          onClick={() => route('/admin')}
          style="margin-bottom: 0.5rem;"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
            <polyline points="15 18 9 12 15 6"></polyline>
          </svg>
          Back to Admin
        </button>
        <h1 style="margin: 0; font-size: 2rem; color: var(--text-color);">
          👤 User Details
        </h1>
      </div>

      {/* Self-Account Notice */}
      {isCurrentUser && (
        <div class="alert alert-info" style="margin-bottom: 2rem; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; border: none;">
          <strong>👤 Viewing Your Own Account</strong>
          <p style="margin: 0.5rem 0 0 0; opacity: 0.9;">You cannot perform admin actions on your own account for security reasons.</p>
        </div>
      )}

      {/* User Info Card */}
      <div class="card" style="margin-bottom: 2rem;">
        <div style="display: flex; justify-content: space-between; align-items: flex-start; gap: 2rem; flex-wrap: wrap;">
          <div style="display: flex; align-items: start; gap: 1.5rem;">
            <div style="width: 80px; height: 80px; border-radius: 50%; background: linear-gradient(135deg, var(--primary-color) 0%, #764ba2 100%); display: flex; align-items: center; justify-content: center; color: white; font-size: 2rem; font-weight: 700;">
              {userDetails.username.charAt(0).toUpperCase()}
            </div>
            <div>
              <div style="display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.5rem;">
                <h2 style="margin: 0; font-size: 1.75rem;">{userDetails.username}</h2>
                {userDetails.is_active === false && (
                  <span style="display: inline-block; padding: 0.25rem 0.75rem; background: #ef4444; color: white; border-radius: 9999px; font-size: 0.8rem; font-weight: 600;">
                    SUSPENDED
                  </span>
                )}
              </div>
              <p style="margin: 0 0 0.5rem 0; color: var(--text-secondary);">{userDetails.email}</p>
              <span style={{
                display: 'inline-block',
                padding: '0.25rem 0.75rem',
                borderRadius: 'var(--radius-xl)',
                fontSize: '0.9rem',
                fontWeight: '600',
                background: isAdmin ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' : 'var(--bg-secondary)',
                color: isAdmin ? 'white' : 'var(--text-secondary)'
              }}>
                {isAdmin ? '🛡️ Administrator' : '👤 User'}
              </span>
            </div>
          </div>

          <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(120px, 1fr)); gap: 1.5rem;">
            <div>
              <div style="font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 0.25rem;">Files</div>
              <div style="font-size: 1.5rem; font-weight: 700; color: var(--text-color);">{userDetails.file_count}</div>
            </div>
            <div>
              <div style="font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 0.25rem;">Storage</div>
              <div style="font-size: 1.5rem; font-weight: 700; color: var(--text-color);">{formatBytes(userDetails.total_storage)}</div>
            </div>
            <div>
              <div style="font-size: 0.85rem; color: var(--text-secondary); margin-bottom: 0.25rem;">Joined</div>
              <div style="font-size: 1rem; font-weight: 600; color: var(--text-color);">{formatDate(userDetails.created_at)}</div>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        {!isCurrentUser && (
          <div style="margin-top: 2rem; padding-top: 2rem; border-top: 1px solid var(--border-color); display: flex; gap: 0.75rem; flex-wrap: wrap;">
            <button 
              class="btn btn-secondary"
              onClick={handleSuspend}
            >
              {userDetails.is_active === false ? (
                <>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
                    <polyline points="20 6 9 17 4 12"></polyline>
                  </svg>
                  Activate Account
                </>
              ) : (
                <>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="15" y1="9" x2="9" y2="15"></line>
                    <line x1="9" y1="9" x2="15" y2="15"></line>
                  </svg>
                  Suspend Account
                </>
              )}
            </button>

            {!isAdmin && (
              <button 
                class="btn btn-primary"
                onClick={handleChangeRole}
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
                  <path d="M12 2L2 7l10 5 10-5-10-5z"></path>
                  <polyline points="2 17 12 22 22 17"></polyline>
                  <polyline points="2 12 12 17 22 12"></polyline>
                </svg>
                Promote to Admin
              </button>
            )}

            <button 
              class="btn btn-secondary"
              onClick={handleResetPassword}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
              </svg>
              Reset Password
            </button>

            <button 
              class="btn btn-secondary"
              onClick={handleForceLogout}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                <polyline points="16 17 21 12 16 7"></polyline>
                <line x1="21" y1="12" x2="9" y2="12"></line>
              </svg>
              Force Logout
            </button>

            {!isAdmin && (
              <button 
                class="btn btn-danger"
                onClick={handleDeleteUser}
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
                  <polyline points="3 6 5 6 21 6"></polyline>
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                </svg>
                Delete User
              </button>
            )}
          </div>
        )}
      </div>

      {/* User's Files */}
      <div class="card" style="margin-bottom: 2rem;">
        <h3 style="margin: 0 0 1.5rem 0; font-size: 1.25rem;">📁 Files ({userFiles.length})</h3>
        
        {userFiles.length === 0 ? (
          <div style="text-align: center; padding: 2rem 1rem; color: var(--text-muted);">
            <p>No files uploaded yet</p>
          </div>
        ) : (
          <div style="overflow-x: auto;">
            <table style="width: 100%; border-collapse: collapse; font-size: 0.9rem;">
              <thead>
                <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Filename</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Type</th>
                  <th style="padding: 0.75rem; text-align: right; font-weight: 600; color: var(--text-secondary);">Size</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Uploaded</th>
                </tr>
              </thead>
              <tbody>
                {userFiles.map(file => (
                  <tr key={file.id} style="border-bottom: 1px solid var(--border-color);">
                    <td style="padding: 0.75rem; font-weight: 500; word-break: break-word; max-width: 300px;">
                      {file.filename}
                    </td>
                    <td style="padding: 0.75rem;">
                      <span style="display: inline-block; padding: 0.25rem 0.5rem; background: var(--bg-secondary); border-radius: var(--radius-sm); font-size: 0.8rem; color: var(--text-secondary);">
                        {file.content_type}
                      </span>
                    </td>
                    <td style="padding: 0.75rem; text-align: right; color: var(--text-secondary);">
                      {formatBytes(file.size)}
                    </td>
                    <td style="padding: 0.75rem; color: var(--text-secondary);">
                      {formatDate(file.created_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* User Activity */}
      <div class="card">
        <h3 style="margin: 0 0 1.5rem 0; font-size: 1.25rem;">📋 Recent Activity ({userActivity.length})</h3>
        
        {userActivity.length === 0 ? (
          <div style="text-align: center; padding: 2rem 1rem; color: var(--text-muted);">
            <p>No activity recorded</p>
          </div>
        ) : (
          <div style="overflow-x: auto;">
            <table style="width: 100%; border-collapse: collapse; font-size: 0.85rem;">
              <thead>
                <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Time</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Action</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Target</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Role</th>
                </tr>
              </thead>
              <tbody>
                {userActivity.map(log => (
                  <tr key={log.id} style="border-bottom: 1px solid var(--border-color);">
                    <td style="padding: 0.75rem; white-space: nowrap; color: var(--text-secondary);">
                      {formatDate(log.created_at)}
                    </td>
                    <td style="padding: 0.75rem;">
                      <span style={{
                        display: 'inline-block',
                        padding: '0.25rem 0.5rem',
                        background: 'var(--bg-secondary)',
                        borderRadius: 'var(--radius-sm)',
                        fontSize: '0.8rem',
                        fontWeight: '500'
                      }}>
                        {log.action}
                      </span>
                    </td>
                    <td style="padding: 0.75rem; color: var(--text-secondary);">
                      {log.target_type && `${log.target_type}`}
                    </td>
                    <td style="padding: 0.75rem;">
                      <span style="font-size: 0.8rem; color: var(--text-muted);">
                        {log.actor_id === userId ? 'Actor' : 'Target'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Confirmation Dialog */}
      <ConfirmDialog
        isOpen={confirmDialog !== null}
        title={confirmDialog?.title || ''}
        message={confirmDialog?.message || ''}
        confirmText={confirmDialog?.confirmText || 'Confirm'}
        cancelText="Cancel"
        confirmStyle={confirmDialog?.confirmStyle || 'primary'}
        onConfirm={confirmAction}
        onCancel={cancelAction}
      />
    </div>
  );
}
