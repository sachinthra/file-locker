import { useState, useEffect } from 'preact/hooks';
import { route } from 'preact-router';
import { getUser, getToken } from '../utils/auth';
import api from '../utils/api';
import Toast from '../components/Toast';

export default function Admin({ isAuthenticated }) {
  const [stats, setStats] = useState({
    total_users: 0,
    total_files: 0,
    total_storage_bytes: 0,
    active_users_24h: 0
  });
  const [users, setUsers] = useState([]);
  const [recentLogs, setRecentLogs] = useState([]);
  const [recentFiles, setRecentFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [toast, setToast] = useState(null);
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

      // Load recent audit logs
      try {
        const logsResponse = await api.get('/admin/logs?limit=50');
        setRecentLogs(logsResponse.data?.logs || []);
      } catch (err) {
        console.error('Failed to load audit logs:', err);
      }

      // Load recent files
      try {
        const filesResponse = await api.get('/admin/files');
        setRecentFiles(filesResponse.data?.files?.slice(0, 10) || []);
      } catch (err) {
        console.error('Failed to load files:', err);
      }
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
        <div 
          class="card" 
          style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; box-shadow: var(--shadow-lg); cursor: pointer; transition: transform 0.2s;"
          onClick={() => document.getElementById('users-section')?.scrollIntoView({ behavior: 'smooth' })}
          onMouseOver={(e) => e.currentTarget.style.transform = 'translateY(-4px)'}
          onMouseOut={(e) => e.currentTarget.style.transform = 'translateY(0)'}
        >
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

        <div 
          class="card" 
          style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; box-shadow: var(--shadow-lg); cursor: pointer; transition: transform 0.2s;"
          onClick={() => route('/admin/files')}
          onMouseOver={(e) => e.currentTarget.style.transform = 'translateY(-4px)'}
          onMouseOut={(e) => e.currentTarget.style.transform = 'translateY(0)'}
        >
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

        <div 
          class="card" 
          style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); color: white; box-shadow: var(--shadow-lg); cursor: pointer; transition: transform 0.2s;"
          onClick={() => route('/admin/files')}
          onMouseOver={(e) => e.currentTarget.style.transform = 'translateY(-4px)'}
          onMouseOut={(e) => e.currentTarget.style.transform = 'translateY(0)'}
        >
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

        <div 
          class="card" 
          style="background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); color: white; box-shadow: var(--shadow-lg); cursor: pointer; transition: transform 0.2s;"
          onClick={() => document.getElementById('users-section')?.scrollIntoView({ behavior: 'smooth' })}
          onMouseOver={(e) => e.currentTarget.style.transform = 'translateY(-4px)'}
          onMouseOut={(e) => e.currentTarget.style.transform = 'translateY(0)'}
        >
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
      <div id="users-section" class="card">
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
                  <tr 
                    key={u.id} 
                    style={{
                      borderBottom: '1px solid var(--border-color)',
                      transition: 'background-color 0.2s',
                      opacity: u.is_active === false ? '0.5' : '1',
                      background: u.is_active === false ? 'var(--bg-secondary)' : 'transparent'
                    }}
                  >
                    <td style="padding: 1rem;">
                      <div style="display: flex; align-items: center; gap: 0.5rem;">
                        <div style="width: 32px; height: 32px; border-radius: 50%; background: var(--primary-color); display: flex; align-items: center; justify-content: center; color: white; font-weight: 600;">
                          {u.username.charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <div style="display: flex; align-items: center; gap: 0.5rem;">
                            <span style="font-weight: 500;">{u.username}</span>
                            {u.is_active === false && (
                              <span style="display: inline-block; padding: 0.125rem 0.5rem; background: #ef4444; color: white; border-radius: 9999px; font-size: 0.7rem; font-weight: 600;">
                                SUSPENDED
                              </span>
                            )}
                          </div>
                        </div>
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
                      {u.id !== user?.id ? (
                        <button 
                          class="btn btn-primary btn-sm"
                          onClick={() => route(`/admin/users/${u.id}`)}
                          title="View user details"
                        >
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.25rem;">
                            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                            <circle cx="12" cy="7" r="4"></circle>
                          </svg>
                          Manage
                        </button>
                      ) : (
                        <span style="color: var(--text-muted); font-size: 0.85rem;">You</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Recent Audit Logs */}
      <div class="card" style="margin-top: 2rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
          <h2 style="margin: 0; font-size: 1.5rem;">📋 Recent Audit Logs</h2>
          <button 
            class="btn btn-secondary btn-sm"
            onClick={() => route('/admin/audit-logs')}
          >
            View All Logs
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-left: 0.5rem;">
              <polyline points="9 18 15 12 9 6"></polyline>
            </svg>
          </button>
        </div>

        {recentLogs.length === 0 ? (
          <div style="text-align: center; padding: 2rem 1rem; color: var(--text-muted);">
            <p>No audit logs yet</p>
          </div>
        ) : (
          <div style="overflow-x: auto;">
            <table style="width: 100%; border-collapse: collapse; font-size: 0.9rem;">
              <thead>
                <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Action</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Target</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Time</th>
                </tr>
              </thead>
              <tbody>
                {recentLogs.slice(0, 10).map(log => (
                  <tr key={log.id} style="border-bottom: 1px solid var(--border-color);">
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
                      {log.target_type} {log.target_id ? `(${log.target_id.substring(0, 8)}...)` : ''}
                    </td>
                    <td style="padding: 0.75rem; color: var(--text-secondary);">{formatDate(log.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Recent Files */}
      <div class="card" style="margin-top: 2rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
          <h2 style="margin: 0; font-size: 1.5rem;">📁 Recent Files</h2>
          <button 
            class="btn btn-secondary btn-sm"
            onClick={() => route('/admin/files')}
          >
            View All Files
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-left: 0.5rem;">
              <polyline points="9 18 15 12 9 6"></polyline>
            </svg>
          </button>
        </div>

        {recentFiles.length === 0 ? (
          <div style="text-align: center; padding: 2rem 1rem; color: var(--text-muted);">
            <p>No files yet</p>
          </div>
        ) : (
          <div style="overflow-x: auto;">
            <table style="width: 100%; border-collapse: collapse; font-size: 0.9rem;">
              <thead>
                <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Filename</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Owner</th>
                  <th style="padding: 0.75rem; text-align: right; font-weight: 600; color: var(--text-secondary);">Size</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Uploaded</th>
                </tr>
              </thead>
              <tbody>
                {recentFiles.map(file => (
                  <tr key={file.id} style="border-bottom: 1px solid var(--border-color);">
                    <td style="padding: 0.75rem; font-weight: 500;">{file.filename}</td>
                    <td style="padding: 0.75rem; color: var(--text-secondary);">
                      {file.username?.Valid ? file.username.String : 'Unknown'}
                    </td>
                    <td style="padding: 0.75rem; text-align: right; color: var(--text-secondary);">{formatBytes(file.size)}</td>
                    <td style="padding: 0.75rem; color: var(--text-secondary);">{formatDate(file.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
