import { useState, useEffect } from 'preact/hooks';
import { route } from 'preact-router';
import { getUser, getToken } from '../utils/auth';
import api from '../utils/api';
import Toast from '../components/Toast';

export default function AdminAuditLogs({ isAuthenticated }) {
  const [logs, setLogs] = useState([]);
  const [filteredLogs, setFilteredLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState(null);
  const [filters, setFilters] = useState({
    action: '',
    targetType: '',
    dateFrom: '',
    dateTo: '',
    search: ''
  });
  const user = getUser();

  const showToast = (message, type = 'info') => {
    setToast({ message, type });
  };

  const closeToast = () => {
    setToast(null);
  };

  useEffect(() => {
    if (!isAuthenticated || !getToken() || user?.role !== 'admin') {
      route('/admin', true);
      return;
    }
    loadLogs();
  }, [isAuthenticated]);

  useEffect(() => {
    applyFilters();
  }, [logs, filters]);

  const loadLogs = async () => {
    setLoading(true);
    try {
      const response = await api.get('/admin/logs?limit=1000');
      setLogs(response.data?.logs || []);
    } catch (err) {
      showToast('Failed to load audit logs', 'error');
      console.error('Load logs error:', err);
    } finally {
      setLoading(false);
    }
  };

  const applyFilters = () => {
    let filtered = [...logs];

    if (filters.action) {
      filtered = filtered.filter(log => log.action === filters.action);
    }

    if (filters.targetType) {
      filtered = filtered.filter(log => log.target_type === filters.targetType);
    }

    if (filters.dateFrom) {
      const fromDate = new Date(filters.dateFrom);
      filtered = filtered.filter(log => new Date(log.created_at) >= fromDate);
    }

    if (filters.dateTo) {
      const toDate = new Date(filters.dateTo);
      toDate.setHours(23, 59, 59);
      filtered = filtered.filter(log => new Date(log.created_at) <= toDate);
    }

    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      filtered = filtered.filter(log => 
        log.action.toLowerCase().includes(searchLower) ||
        log.target_type?.toLowerCase().includes(searchLower) ||
        log.target_id?.toLowerCase().includes(searchLower) ||
        JSON.stringify(log.metadata || {}).toLowerCase().includes(searchLower)
      );
    }

    setFilteredLogs(filtered);
  };

  const downloadLogs = () => {
    const logContent = filteredLogs.map(log => {
      const timestamp = new Date(log.created_at).toISOString();
      const metadata = log.metadata ? JSON.stringify(log.metadata) : '';
      return `[${timestamp}] ${log.action} - Target: ${log.target_type || 'N/A'} (${log.target_id || 'N/A'}) - Actor: ${log.actor_id} - IP: ${log.ip_address || 'N/A'} - Metadata: ${metadata}`;
    }).join('\n');

    const blob = new Blob([logContent], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `audit-logs-${new Date().toISOString().split('T')[0]}.log`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    showToast('Logs downloaded successfully', 'success');
  };

  const clearFilters = () => {
    setFilters({
      action: '',
      targetType: '',
      dateFrom: '',
      dateTo: '',
      search: ''
    });
  };

  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  const getUniqueActions = () => {
    return [...new Set(logs.map(log => log.action))].sort();
  };

  const getUniqueTargetTypes = () => {
    return [...new Set(logs.map(log => log.target_type).filter(Boolean))].sort();
  };

  if (loading) {
    return (
      <div class="loading" style="min-height: 100vh; display: flex; align-items: center; justify-content: center;">
        <div>
          <div class="spinner"></div>
          <p>Loading audit logs...</p>
        </div>
      </div>
    );
  }

  return (
    <div class="admin-container" style="padding: 2rem; max-width: 1600px; margin: 0 auto;">
      {toast && <Toast message={toast.message} type={toast.type} onClose={closeToast} />}

      {/* Header */}
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
        <div>
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
            📋 Audit Logs
          </h1>
          <p style="margin: 0.5rem 0 0 0; color: var(--text-secondary);">
            Showing {filteredLogs.length} of {logs.length} logs
          </p>
        </div>
        <button 
          class="btn btn-primary"
          onClick={downloadLogs}
          disabled={filteredLogs.length === 0}
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
            <polyline points="7 10 12 15 17 10"></polyline>
            <line x1="12" y1="15" x2="12" y2="3"></line>
          </svg>
          Download Logs (.log)
        </button>
      </div>

      {/* Filters */}
      <div class="card" style="margin-bottom: 2rem;">
        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem;">
          <div>
            <label style="display: block; margin-bottom: 0.5rem; font-weight: 500; color: var(--text-secondary); font-size: 0.9rem;">
              Action
            </label>
            <select
              class="input"
              value={filters.action}
              onChange={(e) => setFilters({ ...filters, action: e.target.value })}
            >
              <option value="">All Actions</option>
              {getUniqueActions().map(action => (
                <option key={action} value={action}>{action}</option>
              ))}
            </select>
          </div>

          <div>
            <label style="display: block; margin-bottom: 0.5rem; font-weight: 500; color: var(--text-secondary); font-size: 0.9rem;">
              Target Type
            </label>
            <select
              class="input"
              value={filters.targetType}
              onChange={(e) => setFilters({ ...filters, targetType: e.target.value })}
            >
              <option value="">All Types</option>
              {getUniqueTargetTypes().map(type => (
                <option key={type} value={type}>{type}</option>
              ))}
            </select>
          </div>

          <div>
            <label style="display: block; margin-bottom: 0.5rem; font-weight: 500; color: var(--text-secondary); font-size: 0.9rem;">
              From Date
            </label>
            <input
              type="date"
              class="input"
              value={filters.dateFrom}
              onChange={(e) => setFilters({ ...filters, dateFrom: e.target.value })}
            />
          </div>

          <div>
            <label style="display: block; margin-bottom: 0.5rem; font-weight: 500; color: var(--text-secondary); font-size: 0.9rem;">
              To Date
            </label>
            <input
              type="date"
              class="input"
              value={filters.dateTo}
              onChange={(e) => setFilters({ ...filters, dateTo: e.target.value })}
            />
          </div>

          <div style="grid-column: 1 / -1;">
            <label style="display: block; margin-bottom: 0.5rem; font-weight: 500; color: var(--text-secondary); font-size: 0.9rem;">
              Search
            </label>
            <input
              type="text"
              class="input"
              placeholder="Search logs..."
              value={filters.search}
              onChange={(e) => setFilters({ ...filters, search: e.target.value })}
            />
          </div>
        </div>

        <div style="margin-top: 1rem; display: flex; justify-content: flex-end;">
          <button 
            class="btn btn-secondary btn-sm"
            onClick={clearFilters}
          >
            Clear Filters
          </button>
        </div>
      </div>

      {/* Logs Table */}
      <div class="card">
        {filteredLogs.length === 0 ? (
          <div style="text-align: center; padding: 3rem 1rem; color: var(--text-muted);">
            <div style="font-size: 3rem; margin-bottom: 1rem;">📋</div>
            <p>{logs.length === 0 ? 'No audit logs yet' : 'No logs match your filters'}</p>
          </div>
        ) : (
          <div style="overflow-x: auto;">
            <table style="width: 100%; border-collapse: collapse; font-size: 0.85rem;">
              <thead>
                <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Timestamp</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Action</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Target</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Actor ID</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">IP Address</th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">Metadata</th>
                </tr>
              </thead>
              <tbody>
                {filteredLogs.map(log => (
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
                      {log.target_type && (
                        <div>
                          <div style="font-weight: 500;">{log.target_type}</div>
                          {log.target_id && (
                            <div style="font-size: 0.75rem; color: var(--text-muted); font-family: monospace;">
                              {log.target_id.substring(0, 16)}...
                            </div>
                          )}
                        </div>
                      )}
                    </td>
                    <td style="padding: 0.75rem; font-family: monospace; font-size: 0.75rem; color: var(--text-muted);">
                      {log.actor_id.substring(0, 16)}...
                    </td>
                    <td style="padding: 0.75rem; color: var(--text-secondary);">
                      {log.ip_address || 'N/A'}
                    </td>
                    <td style="padding: 0.75rem;">
                      {log.metadata && (
                        <details style="cursor: pointer;">
                          <summary style="color: var(--primary-color); font-size: 0.85rem;">View</summary>
                          <pre style="margin-top: 0.5rem; padding: 0.5rem; background: var(--bg-secondary); border-radius: var(--radius-sm); font-size: 0.75rem; overflow-x: auto;">
                            {JSON.stringify(log.metadata, null, 2)}
                          </pre>
                        </details>
                      )}
                    </td>
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
