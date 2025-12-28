import { useState, useEffect } from 'preact/hooks';
import { route } from 'preact-router';
import { getUser, getToken } from '../utils/auth';
import api from '../utils/api';
import Toast from '../components/Toast';
import ConfirmDialog from '../components/ConfirmDialog';

export default function Admin({ isAuthenticated }) {
  const [stats, setStats] = useState({
    total_users: 0,
    total_files: 0,
    total_storage_bytes: 0,
    active_users_24h: 0
  });
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [toast, setToast] = useState(null);
  const [deleteConfirm, setDeleteConfirm] = useState(null);
  const user = getUser();

  const showToast = (message, type = 'info') => {
    setToast({ message, type });
  };

  const closeToast = () => {
    setToast(null);
  };

  useEffect(() => {
    // Check if user is authenticated
    if (!isAuthenticated || !getToken()) {
      route('/login', true);
      return;
    }

    // Check if user is admin
    if (user?.role !== 'admin') {
      showToast('Admin access required', 'error');
      route('/dashboard', true);
      return;
    }

    loadData();
  }, [isAuthenticated]);

  const loadData = async () => {
    setLoading(true);
    setError('');
    try {
      // Load stats
      const statsResponse = await api.get('/admin/stats');
      setStats(statsResponse.data);

      // Load users
      const usersResponse = await api.get('/admin/users');
      setUsers(usersResponse.data?.users || []);
    } catch (err) {
      if (err.response?.status === 403) {
        setError('Admin access required');
        showToast('Admin access required', 'error');
        setTimeout(() => route('/dashboard', true), 2000);
      } else {
        setError('Failed to load admin data');
        showToast('Failed to load admin data', 'error');
      }
      console.error('Load admin data error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteUser = (userId, username) => {
    setDeleteConfirm({ id: userId, username });
  };

  const confirmDelete = async () => {
    const userId = deleteConfirm.id;
    const username = deleteConfirm.username;
    setDeleteConfirm(null);

    try {
      await api.delete(`/admin/users/${userId}`);
      showToast(`User ${username} deleted successfully`, 'success');
      // Reload data
      loadData();
    } catch (err) {
      showToast(err.response?.data?.error || 'Failed to delete user', 'error');
      console.error('Delete user error:', err);
    }
  };

  const cancelDelete = () => {
    setDeleteConfirm(null);
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
    return date.toLocaleDateString('en-US', {
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
          <p>Loading admin panel...</p>
        </div>
      </div>
    );
  }

  if (error && users.length === 0) {
    return (
      <div class="admin-container" style="padding: 2rem;">
        <div class="alert alert-error">{error}</div>
      </div>
    );
  }

  return (
    <div class="admin-container" style="padding: 2rem; max-width: 1400px; margin: 0 auto;">
      {toast && <Toast message={toast.message} type={toast.type} onClose={closeToast} />}

      {/* Header */}
      <div style="margin-bottom: 2rem;">
        <h1 style="margin: 0 0 0.5rem 0; font-size: 2rem; color: var(--text-color);">
          🛡️ Admin Dashboard
        </h1>
        <p style="margin: 0; color: var(--text-secondary);">
          Manage users and monitor system statistics
        </p>
      </div>

      {/* Stats Cards */}
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem; margin-bottom: 2rem;">
        <div class="card" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; box-shadow: var(--shadow-lg);">
          <div style="display: flex; justify-content: space-between; align-items: flex-start;">
            <div>
              <h3 style="margin: 0 0 0.5rem 0; font-size: 2.5rem; font-weight: 700;">
                {stats.total_users}
              </h3>
              <p style="margin: 0; opacity: 0.9; font-size: 1.1rem;">Total Users</p>
            </div>
            <div style="font-size: 2.5rem; opacity: 0.7;">👥</div>
          </div>
        </div>

        <div class="card" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; box-shadow: var(--shadow-lg);">
          <div style="display: flex; justify-content: space-between; align-items: flex-start;">
            <div>
              <h3 style="margin: 0 0 0.5rem 0; font-size: 2.5rem; font-weight: 700;">
                {stats.total_files}
              </h3>
              <p style="margin: 0; opacity: 0.9; font-size: 1.1rem;">Total Files</p>
            </div>
            <div style="font-size: 2.5rem; opacity: 0.7;">📁</div>
          </div>
        </div>

        <div class="card" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); color: white; box-shadow: var(--shadow-lg);">
          <div style="display: flex; justify-content: space-between; align-items: flex-start;">
            <div>
              <h3 style="margin: 0 0 0.5rem 0; font-size: 2.5rem; font-weight: 700;">
                {formatBytes(stats.total_storage_bytes)}
              </h3>
              <p style="margin: 0; opacity: 0.9; font-size: 1.1rem;">Total Storage</p>
            </div>
            <div style="font-size: 2.5rem; opacity: 0.7;">💾</div>
          </div>
        </div>

        <div class="card" style="background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); color: white; box-shadow: var(--shadow-lg);">
          <div style="display: flex; justify-content: space-between; align-items: flex-start;">
            <div>
              <h3 style="margin: 0 0 0.5rem 0; font-size: 2.5rem; font-weight: 700;">
                {stats.active_users_24h}
              </h3>
              <p style="margin: 0; opacity: 0.9; font-size: 1.1rem;">Active Today</p>
            </div>
            <div style="font-size: 2.5rem; opacity: 0.7;">⚡</div>
          </div>
        </div>
      </div>

      {/* Users Table */}
      <div class="card">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
          <h2 style="margin: 0; font-size: 1.5rem;">User Management</h2>
          <button 
            class="btn btn-secondary btn-sm" 
            onClick={loadData}
            title="Refresh data"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
              <polyline points="23 4 23 10 17 10"></polyline>
              <polyline points="1 20 1 14 7 14"></polyline>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
            </svg>
            Refresh
          </button>
        </div>

        {users.length === 0 ? (
          <div style="text-align: center; padding: 3rem 1rem; color: var(--text-muted);">
            <div style="font-size: 3rem; margin-bottom: 1rem;">👤</div>
            <p>No users found</p>
          </div>
        ) : (
          <div style="overflow-x: auto;">
            <table class="admin-table" style="width: 100%; border-collapse: collapse;">
              <thead>
                <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Username</th>
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Email</th>
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Role</th>
                  <th style="padding: 1rem; text-align: right; font-weight: 600; color: var(--text-secondary);">Files</th>
                  <th style="padding: 1rem; text-align: right; font-weight: 600; color: var(--text-secondary);">Storage</th>
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Joined</th>
                  <th style="padding: 1rem; text-align: center; font-weight: 600; color: var(--text-secondary);">Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map(u => (
                  <tr key={u.id} style="border-bottom: 1px solid var(--border-color); transition: background-color 0.2s;">
                    <td style="padding: 1rem;">
                      <div style="display: flex; align-items: center; gap: 0.5rem;">
                        <div style="width: 32px; height: 32px; border-radius: 50%; background: var(--primary-color); display: flex; align-items: center; justify-content: center; color: white; font-weight: 600;">
                          {u.username.charAt(0).toUpperCase()}
                        </div>
                        <span style="font-weight: 500;">{u.username}</span>
                      </div>
                    </td>
                    <td style="padding: 1rem; color: var(--text-secondary);">{u.email}</td>
                    <td style="padding: 1rem;">
                      <span style={{
                        display: 'inline-block',
                        padding: '0.25rem 0.75rem',
                        borderRadius: 'var(--radius-xl)',
                        fontSize: '0.85rem',
                        fontWeight: '600',
                        background: u.role === 'admin' ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' : 'var(--bg-secondary)',
                        color: u.role === 'admin' ? 'white' : 'var(--text-secondary)'
                      }}>
                        {u.role === 'admin' ? '🛡️ Admin' : 'User'}
                      </span>
                    </td>
                    <td style="padding: 1rem; text-align: right; font-weight: 500;">{u.file_count}</td>
                    <td style="padding: 1rem; text-align: right; font-weight: 500;">{formatBytes(u.total_storage)}</td>
                    <td style="padding: 1rem; color: var(--text-secondary); font-size: 0.9rem;">{formatDate(u.created_at)}</td>
                    <td style="padding: 1rem; text-align: center;">
                      {u.id !== user?.id && u.role !== 'admin' ? (
                        <button 
                          class="btn btn-danger btn-sm"
                          onClick={() => handleDeleteUser(u.id, u.username)}
                          title="Delete user"
                        >
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.25rem;">
                            <polyline points="3 6 5 6 21 6"></polyline>
                            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                            <line x1="10" y1="11" x2="10" y2="17"></line>
                            <line x1="14" y1="11" x2="14" y2="17"></line>
                          </svg>
                          Delete
                        </button>
                      ) : (
                        <span style="color: var(--text-muted); font-size: 0.85rem;">
                          {u.id === user?.id ? 'You' : 'Protected'}
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={deleteConfirm !== null}
        title="⚠️ Delete User?"
        message={deleteConfirm ? `Are you sure you want to delete user "${deleteConfirm.username}"? This will permanently delete their account and all their files. This action cannot be undone.` : ''}
        confirmText="Yes, Delete User"
        cancelText="Cancel"
        confirmStyle="danger"
        onConfirm={confirmDelete}
        onCancel={cancelDelete}
      />
    </div>
  );
}
