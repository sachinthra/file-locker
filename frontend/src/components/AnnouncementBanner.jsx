import { h } from 'preact';
import { useState, useEffect } from 'preact/hooks';
import api from '../utils/api';
import { getToken } from '../utils/auth';

export default function AnnouncementBanner() {
  const [announcements, setAnnouncements] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (getToken()) {
      loadAnnouncements();
    } else {
      setLoading(false);
    }
  }, []);

  const loadAnnouncements = async () => {
    if (!getToken()) {
      setLoading(false);
      return;
    }
    try {
      const response = await api.get('/announcements');
      setAnnouncements(response.data?.announcements || []);
    } catch (err) {
      // Silently fail for 401 errors (not authenticated)
      if (err.response?.status !== 401) {
        console.error('Failed to load announcements:', err);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleDismiss = async (announcementId) => {
    try {
      await api.post(`/announcements/${announcementId}/dismiss`);
      setAnnouncements(announcements.filter(a => a.id !== announcementId));
    } catch (err) {
      console.error('Failed to dismiss announcement:', err);
    }
  };

  if (loading || announcements.length === 0) {
    return null;
  }

  const typeColors = {
    info: { bg: '#dbeafe', border: '#3b82f6', text: '#1e40af' },
    warning: { bg: '#fef3c7', border: '#f59e0b', text: '#92400e' },
    critical: { bg: '#fee2e2', border: '#ef4444', text: '#991b1b' }
  };

  return (
    <div style="position: relative; z-index: 100;">
      {announcements.map(announcement => {
        const colors = typeColors[announcement.type] || typeColors.info;
        
        return (
          <div 
            key={announcement.id}
            style={`
              background: ${colors.bg}; 
              border-left: 4px solid ${colors.border}; 
              padding: 1rem 1.5rem; 
              margin-bottom: 0.5rem;
              border-radius: 4px;
              box-shadow: 0 2px 4px rgba(0,0,0,0.1);
              animation: slideDown 0.3s ease-out;
            `}
          >
            <div style="display: flex; justify-content: space-between; align-items: start; gap: 1rem;">
              <div style="flex: 1;">
                <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.25rem;">
                  {announcement.type === 'info' && (
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={colors.border} stroke-width="2">
                      <circle cx="12" cy="12" r="10"></circle>
                      <line x1="12" y1="16" x2="12" y2="12"></line>
                      <line x1="12" y1="8" x2="12.01" y2="8"></line>
                    </svg>
                  )}
                  {announcement.type === 'warning' && (
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={colors.border} stroke-width="2">
                      <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                      <line x1="12" y1="9" x2="12" y2="13"></line>
                      <line x1="12" y1="17" x2="12.01" y2="17"></line>
                    </svg>
                  )}
                  {announcement.type === 'critical' && (
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={colors.border} stroke-width="2">
                      <circle cx="12" cy="12" r="10"></circle>
                      <line x1="15" y1="9" x2="9" y2="15"></line>
                      <line x1="9" y1="9" x2="15" y2="15"></line>
                    </svg>
                  )}
                  <strong style={`color: ${colors.text}; font-size: 1rem;`}>{announcement.title}</strong>
                  <span 
                    style={`
                      padding: 0.125rem 0.5rem; 
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
                <p style={`margin: 0.25rem 0 0 1.75rem; color: ${colors.text}; font-size: 0.95rem;`}>
                  {announcement.message}
                </p>
              </div>
              <button 
                onClick={() => handleDismiss(announcement.id)}
                style={`
                  background: transparent; 
                  border: none; 
                  cursor: pointer; 
                  color: ${colors.text}; 
                  padding: 0.25rem;
                  display: flex;
                  align-items: center;
                  opacity: 0.7;
                  transition: opacity 0.2s;
                `}
                onMouseEnter={(e) => e.target.style.opacity = '1'}
                onMouseLeave={(e) => e.target.style.opacity = '0.7'}
                title="Dismiss"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>
          </div>
        );
      })}
      <style>{`
        @keyframes slideDown {
          from {
            opacity: 0;
            transform: translateY(-10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </div>
  );
}
