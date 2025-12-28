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
            <p>Personal Access Tokens for CLI and script usage are managed below in Developer Settings.</p>
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
      {/* Developer Settings - Personal Access Tokens */}
      <div class="card settings-section" style="margin-top: 1rem;">
        <h3>Developer Settings</h3>
        <p>Personal Access Tokens for CLI and script usage.</p>
        <TokenManager addNotification={addNotification} />
      </div>
    </div>
  );
}

// TokenManager component
function TokenManager({ addNotification }) {
  const [tokens, setTokens] = useState([]);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newTokenPlain, setNewTokenPlain] = useState(null);
  const [newTokenName, setNewTokenName] = useState("");
  const [expiresDays, setExpiresDays] = useState(0);
  const [isCreating, setIsCreating] = useState(false);

  const loadTokens = async () => {
    try {
      const res = await api.get('/auth/tokens');
      setTokens(res.data.tokens || []);
    } catch (e) {
      addNotification && addNotification('Failed to load tokens', 'error');
    }
  };

  useEffect(() => { loadTokens(); }, []);

  const handleCreate = async () => {
    if (!newTokenName.trim()) {
      addNotification && addNotification('Please enter a token name', 'error');
      return;
    }
    setIsCreating(true);
    try {
      const res = await api.post('/auth/tokens', { 
        name: newTokenName.trim(), 
        expires_in_days: Number(expiresDays) 
      });
      setNewTokenPlain(res.data.token);
      setShowCreateModal(true);
      setNewTokenName(''); 
      setExpiresDays(0);
      await loadTokens();
    } catch (e) {
      addNotification && addNotification('Failed to create token', 'error');
    } finally {
      setIsCreating(false);
    }
  };

  const handleRevoke = async (id, name) => {
    if (!confirm(`Are you sure you want to revoke the token "${name}"? This action cannot be undone.`)) {
      return;
    }
    try {
      await api.delete(`/auth/tokens/${id}`);
      await loadTokens();
      addNotification && addNotification('Token revoked successfully', 'success');
    } catch (e) {
      addNotification && addNotification('Failed to revoke token', 'error');
    }
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return 'Never';
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = date - now;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    
    if (diffMs < 0) return 'Expired';
    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Tomorrow';
    if (diffDays < 7) return `In ${diffDays} days`;
    return date.toLocaleDateString();
  };

  const getTokenStatus = (expiresAt) => {
    if (!expiresAt) return { text: 'Active', color: '#10b981' };
    const now = new Date();
    const expires = new Date(expiresAt);
    const diffMs = expires - now;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    
    if (diffMs < 0) return { text: 'Expired', color: '#ef4444' };
    if (diffDays <= 7) return { text: 'Expiring Soon', color: '#f59e0b' };
    return { text: 'Active', color: '#10b981' };
  };

  return (
    <div style={{ maxWidth: '900px' }}>
      {/* Info Banner */}
      <div style={{
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: '1.5rem',
        borderRadius: '12px',
        marginBottom: '2rem',
        color: 'white'
      }}>
        <h3 style={{ margin: '0 0 0.5rem 0', fontSize: '1.25rem' }}>🔑 Personal Access Tokens</h3>
        <p style={{ margin: 0, opacity: 0.95, lineHeight: 1.5 }}>
          Generate tokens for CLI access and automation. Treat tokens like passwords - they provide full account access.
        </p>
      </div>

      {/* Create Token Card */}
      <div style={{
        background: '#1a1a2e',
        border: '1px solid #2a2a3e',
        borderRadius: '12px',
        padding: '1.5rem',
        marginBottom: '2rem'
      }}>
        <h4 style={{ margin: '0 0 1rem 0', fontSize: '1.1rem' }}>Generate New Token</h4>
        
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontSize: '0.9rem', color: '#aaa' }}>
              Token Name <span style={{ color: '#ef4444' }}>*</span>
            </label>
            <input 
              type="text" 
              placeholder="e.g., MacBook CLI, Production Server, CI/CD Pipeline" 
              value={newTokenName} 
              onInput={e => setNewTokenName(e.target.value)}
              style={{
                width: '100%',
                padding: '0.75rem',
                background: '#0f0f1e',
                border: '1px solid #2a2a3e',
                borderRadius: '8px',
                color: 'white',
                fontSize: '0.95rem'
              }}
            />
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontSize: '0.9rem', color: '#aaa' }}>
              Expiration
            </label>
            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
              <input 
                type="number" 
                placeholder="0" 
                value={expiresDays} 
                onInput={e => setExpiresDays(e.target.value)}
                style={{
                  width: '120px',
                  padding: '0.75rem',
                  background: '#0f0f1e',
                  border: '1px solid #2a2a3e',
                  borderRadius: '8px',
                  color: 'white',
                  fontSize: '0.95rem'
                }}
              />
              <span style={{ color: '#aaa', fontSize: '0.9rem' }}>
                days (0 = never expires)
              </span>
            </div>
          </div>

          <button 
            class="btn btn-primary" 
            onClick={handleCreate}
            disabled={isCreating}
            style={{
              padding: '0.75rem 1.5rem',
              fontSize: '1rem',
              fontWeight: '600',
              opacity: isCreating ? 0.6 : 1
            }}
          >
            {isCreating ? 'Generating...' : '✨ Generate Token'}
          </button>
        </div>
      </div>

      {/* Tokens List */}
      <div style={{
        background: '#1a1a2e',
        border: '1px solid #2a2a3e',
        borderRadius: '12px',
        padding: '1.5rem'
      }}>
        <h4 style={{ margin: '0 0 1rem 0', fontSize: '1.1rem' }}>
          Your Tokens ({tokens.length})
        </h4>

        {tokens.length === 0 ? (
          <div style={{
            textAlign: 'center',
            padding: '3rem 1rem',
            color: '#666'
          }}>
            <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>🔐</div>
            <p style={{ margin: 0 }}>No tokens yet. Create your first token to get started.</p>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {tokens.map(t => {
              const status = getTokenStatus(t.expires_at);
              return (
                <div 
                  key={t.id}
                  style={{
                    background: '#0f0f1e',
                    border: '1px solid #2a2a3e',
                    borderRadius: '10px',
                    padding: '1.25rem',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-start',
                    gap: '1rem'
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <div style={{ 
                      display: 'flex', 
                      alignItems: 'center', 
                      gap: '0.75rem',
                      marginBottom: '0.75rem'
                    }}>
                      <h5 style={{ 
                        margin: 0, 
                        fontSize: '1.05rem',
                        fontWeight: '600'
                      }}>
                        {t.name}
                      </h5>
                      <span style={{
                        background: status.color + '20',
                        color: status.color,
                        padding: '0.25rem 0.75rem',
                        borderRadius: '12px',
                        fontSize: '0.8rem',
                        fontWeight: '600'
                      }}>
                        {status.text}
                      </span>
                    </div>

                    <div style={{ 
                      display: 'grid', 
                      gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
                      gap: '0.75rem',
                      fontSize: '0.85rem',
                      color: '#aaa'
                    }}>
                      <div>
                        <span style={{ color: '#666' }}>Created:</span>{' '}
                        {new Date(t.created_at).toLocaleString('en-US', {
                          year: 'numeric',
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit'
                        })}
                      </div>
                      <div>
                        <span style={{ color: '#666' }}>Last Used:</span>{' '}
                        {t.last_used_at 
                          ? new Date(t.last_used_at).toLocaleString('en-US', {
                              year: 'numeric',
                              month: 'short',
                              day: 'numeric',
                              hour: '2-digit',
                              minute: '2-digit'
                            })
                          : 'Never'}
                      </div>
                      <div>
                        <span style={{ color: '#666' }}>Expires:</span>{' '}
                        {formatDate(t.expires_at)}
                      </div>
                    </div>

                    {!t.last_used_at && (
                      <div style={{
                        marginTop: '0.75rem',
                        padding: '0.5rem 0.75rem',
                        background: '#f59e0b20',
                        border: '1px solid #f59e0b40',
                        borderRadius: '6px',
                        fontSize: '0.85rem',
                        color: '#f59e0b'
                      }}>
                        ⚠️ This token has never been used
                      </div>
                    )}
                  </div>

                  <div style={{ 
                    display: 'flex', 
                    gap: '0.5rem',
                    flexShrink: 0 
                  }}>
                    <button 
                      class="btn btn-danger btn-sm" 
                      onClick={() => handleRevoke(t.id, t.name)}
                      style={{
                        padding: '0.5rem 1rem',
                        fontSize: '0.9rem'
                      }}
                    >
                      🗑️ Revoke
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Create Token Success Modal */}
      {showCreateModal && (
        <div 
          class="modal-overlay" 
          onClick={() => { setShowCreateModal(false); setNewTokenPlain(null); }}
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: 'rgba(0, 0, 0, 0.85)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
            backdropFilter: 'blur(4px)'
          }}
        >
          <div 
            class="modal-content" 
            onClick={e => e.stopPropagation()}
            style={{
              background: '#1a1a2e',
              borderRadius: '16px',
              maxWidth: '600px',
              width: '90%',
              border: '1px solid #2a2a3e',
              boxShadow: '0 20px 60px rgba(0, 0, 0, 0.5)'
            }}
          >
            <div 
              class="modal-header"
              style={{
                padding: '1.5rem',
                borderBottom: '1px solid #2a2a3e',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between'
              }}
            >
              <div>
                <h3 style={{ margin: '0 0 0.25rem 0', fontSize: '1.5rem' }}>
                  ✅ Token Created Successfully
                </h3>
                <p style={{ margin: 0, color: '#aaa', fontSize: '0.9rem' }}>
                  Copy your token now - it won't be shown again
                </p>
              </div>
              <button 
                class="btn-icon" 
                onClick={() => { setShowCreateModal(false); setNewTokenPlain(null); }}
                style={{
                  background: 'transparent',
                  border: 'none',
                  color: '#aaa',
                  fontSize: '1.5rem',
                  cursor: 'pointer',
                  padding: '0.5rem',
                  lineHeight: 1
                }}
              >
                ✕
              </button>
            </div>

            <div style={{ padding: '1.5rem' }}>
              {/* Security Warning */}
              <div style={{
                background: '#ef444420',
                border: '1px solid #ef444440',
                borderRadius: '8px',
                padding: '1rem',
                marginBottom: '1.5rem'
              }}>
                <div style={{ 
                  display: 'flex', 
                  alignItems: 'flex-start', 
                  gap: '0.75rem',
                  color: '#ef4444',
                  fontSize: '0.9rem',
                  lineHeight: 1.6
                }}>
                  <span style={{ fontSize: '1.25rem' }}>🔒</span>
                  <div>
                    <strong>Important Security Notice:</strong>
                    <ul style={{ margin: '0.5rem 0 0 0', paddingLeft: '1.25rem' }}>
                      <li>This token provides full access to your account</li>
                      <li>Store it securely (e.g., password manager)</li>
                      <li>Never commit it to version control</li>
                      <li>You won't be able to see it again</li>
                    </ul>
                  </div>
                </div>
              </div>

              {/* Token Display */}
              <div style={{ marginBottom: '1.5rem' }}>
                <label style={{ 
                  display: 'block', 
                  marginBottom: '0.5rem',
                  fontSize: '0.9rem',
                  color: '#aaa',
                  fontWeight: '600'
                }}>
                  Your Personal Access Token:
                </label>
                <div style={{
                  position: 'relative'
                }}>
                  <pre style={{
                    background: '#0f0f1e',
                    padding: '1rem',
                    borderRadius: '8px',
                    color: '#10b981',
                    fontSize: '0.95rem',
                    fontFamily: 'monospace',
                    margin: 0,
                    wordBreak: 'break-all',
                    whiteSpace: 'pre-wrap',
                    border: '2px solid #10b981',
                    userSelect: 'all'
                  }}>
                    {newTokenPlain}
                  </pre>
                </div>
              </div>

              {/* CLI Usage Example */}
              <div style={{
                background: '#0f0f1e',
                border: '1px solid #2a2a3e',
                borderRadius: '8px',
                padding: '1rem',
                marginBottom: '1.5rem'
              }}>
                <div style={{ 
                  fontSize: '0.85rem', 
                  color: '#666',
                  marginBottom: '0.5rem',
                  fontWeight: '600'
                }}>
                  💻 CLI Usage:
                </div>
                <pre style={{
                  margin: 0,
                  fontSize: '0.85rem',
                  color: '#aaa',
                  fontFamily: 'monospace',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-all'
                }}>
                  fl login --token {newTokenPlain}
                </pre>
              </div>

              {/* Action Buttons */}
              <div style={{ 
                display: 'flex', 
                gap: '0.75rem',
                justifyContent: 'flex-end'
              }}>
                <button 
                  class="btn btn-primary" 
                  onClick={() => { 
                    navigator.clipboard.writeText(newTokenPlain || ''); 
                    addNotification && addNotification('✅ Token copied to clipboard!', 'success'); 
                  }}
                  style={{
                    padding: '0.75rem 1.5rem',
                    fontSize: '1rem',
                    fontWeight: '600'
                  }}
                >
                  📋 Copy Token
                </button>
                <button 
                  class="btn" 
                  onClick={() => { setShowCreateModal(false); setNewTokenPlain(null); }}
                  style={{
                    padding: '0.75rem 1.5rem',
                    fontSize: '1rem'
                  }}
                >
                  I've Saved It
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
