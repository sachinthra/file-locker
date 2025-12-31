import { useState, useEffect } from 'preact/hooks';
import { route } from 'preact-router';
import { login, getMe } from '../utils/api';
import { saveToken, saveUser, getToken } from '../utils/auth';
import api from '../utils/api';

export default function Login({ setIsAuthenticated }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [announcements, setAnnouncements] = useState([]);
  const [showAnnouncementModal, setShowAnnouncementModal] = useState(false);

  useEffect(() => {
    // If already logged in, redirect to dashboard
    if (getToken()) {
      route('/dashboard', true);
    }
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await login(username, password);
      const { token } = response.data;
      
      saveToken(token);
      
      // Fetch user info including role
      const userResponse = await getMe();
      saveUser(userResponse.data);
      
      setIsAuthenticated(true);
      
      // Check for announcements after successful login
      try {
        const announcementResponse = await api.get('/announcements');
        const undismissedAnnouncements = announcementResponse.data?.announcements || [];
        
        if (undismissedAnnouncements.length > 0) {
          setAnnouncements(undismissedAnnouncements);
          setShowAnnouncementModal(true);
        } else {
          route('/dashboard');
        }
      } catch (err) {
        console.error('Failed to load announcements:', err);
        route('/dashboard');
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleDismissAnnouncement = async (announcementId) => {
    try {
      await api.post(`/announcements/${announcementId}/dismiss`);
      setAnnouncements(announcements.filter(a => a.id !== announcementId));
      
      // If all announcements dismissed, redirect to dashboard
      if (announcements.length === 1) {
        setShowAnnouncementModal(false);
        route('/dashboard');
      }
    } catch (err) {
      console.error('Failed to dismiss announcement:', err);
    }
  };

  const handleContinueToDashboard = () => {
    setShowAnnouncementModal(false);
    route('/dashboard');
  };

  const typeColors = {
    info: { bg: '#dbeafe', border: '#3b82f6', text: '#1e40af' },
    warning: { bg: '#fef3c7', border: '#f59e0b', text: '#92400e' },
    critical: { bg: '#fee2e2', border: '#ef4444', text: '#991b1b' }
  };

  return (
    <>
      <div class="form">
        <h2>Login to File Locker</h2>
        {error && <div class="alert alert-error">{error}</div>}
        
        <form onSubmit={handleSubmit}>
          <div class="form-group">
            <label class="form-label">Username</label>
            <input
              type="text"
              class="form-input"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
          </div>

          <div class="form-group">
            <label class="form-label">Password</label>
            <input
              type="password"
              class="form-input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>

          <button type="submit" class="btn btn-primary" style="width: 100%" disabled={loading}>
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <p style="text-align: center; margin-top: 1rem">
          Don't have an account? <a href="/register">Register here</a>
        </p>
      </div>

      {/* Announcements Modal */}
      {showAnnouncementModal && announcements.length > 0 && (
        <div 
          style="
            position: fixed; 
            top: 0; 
            left: 0; 
            right: 0; 
            bottom: 0; 
            background: rgba(0, 0, 0, 0.8); 
            backdrop-filter: blur(4px);
            display: flex; 
            align-items: center; 
            justify-content: center; 
            z-index: 1000;
            animation: fadeIn 0.3s ease-out;
          "
          onClick={(e) => {
            if (e.target === e.currentTarget) {
              handleContinueToDashboard();
            }
          }}
        >
          <div 
            style="
              background: var(--bg-primary); 
              border-radius: 12px; 
              max-width: 600px; 
              width: 90%; 
              max-height: 80vh; 
              overflow-y: auto;
              box-shadow: 0 20px 60px rgba(0,0,0,0.4);
              animation: slideUp 0.3s ease-out;
            "
          >
            <div style="padding: 1.5rem; border-bottom: 1px solid var(--border-color);">
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <h2 style="margin: 0; font-size: 1.5rem;">System Announcements</h2>
                <button 
                  onClick={handleContinueToDashboard}
                  style="
                    background: transparent; 
                    border: none; 
                    cursor: pointer; 
                    padding: 0.5rem;
                    color: var(--text-secondary);
                    transition: color 0.2s;
                  "
                  onMouseEnter={(e) => e.target.style.color = 'var(--text-primary)'}
                  onMouseLeave={(e) => e.target.style.color = 'var(--text-secondary)'}
                  title="Close"
                >
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                </button>
              </div>
              <p style="margin: 0.5rem 0 0 0; color: var(--text-secondary); font-size: 0.9rem;">
                Please review the following important announcements
              </p>
            </div>

            <div style="padding: 1.5rem; display: flex; flex-direction: column; gap: 1rem;">
              {announcements.map(announcement => {
                const colors = typeColors[announcement.type] || typeColors.info;
                
                return (
                  <div 
                    key={announcement.id}
                    style={`
                      border: 2px solid ${colors.border}; 
                      border-radius: 8px; 
                      padding: 1.25rem; 
                      background: ${colors.bg};
                    `}
                  >
                    <div style="margin-bottom: 1rem;">
                      <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
                        {announcement.type === 'info' && (
                          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke={colors.border} stroke-width="2">
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="12" y1="16" x2="12" y2="12"></line>
                            <line x1="12" y1="8" x2="12.01" y2="8"></line>
                          </svg>
                        )}
                        {announcement.type === 'warning' && (
                          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke={colors.border} stroke-width="2">
                            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                            <line x1="12" y1="9" x2="12" y2="13"></line>
                            <line x1="12" y1="17" x2="12.01" y2="17"></line>
                          </svg>
                        )}
                        {announcement.type === 'critical' && (
                          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke={colors.border} stroke-width="2">
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="15" y1="9" x2="9" y2="15"></line>
                            <line x1="9" y1="9" x2="15" y2="15"></line>
                          </svg>
                        )}
                        <strong style={`color: ${colors.text}; font-size: 1.1rem;`}>{announcement.title}</strong>
                        <span 
                          style={`
                            padding: 0.25rem 0.5rem; 
                            border-radius: 4px; 
                            font-size: 0.7rem; 
                            font-weight: 600; 
                            background: ${colors.border}; 
                            color: white; 
                            text-transform: uppercase;
                          `}
                        >
                          {announcement.type}
                        </span>
                      </div>
                      <p style={`margin: 0; color: ${colors.text}; line-height: 1.6;`}>
                        {announcement.message}
                      </p>
                    </div>
                    <button 
                      class="btn btn-secondary btn-sm"
                      onClick={() => handleDismissAnnouncement(announcement.id)}
                      style="width: 100%;"
                    >
                      Dismiss
                    </button>
                  </div>
                );
              })}
            </div>

            <div style="padding: 1.5rem; border-top: 1px solid var(--border-color);">
              <button 
                class="btn btn-primary"
                onClick={handleContinueToDashboard}
                style="width: 100%;"
              >
                Continue to Dashboard
              </button>
            </div>
          </div>
        </div>
      )}

      <style>{`
        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
        @keyframes slideUp {
          from {
            opacity: 0;
            transform: translateY(20px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </>  );
}