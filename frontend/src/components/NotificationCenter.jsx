import { useState, useEffect } from "preact/hooks";
import api from "../utils/api";
import { getToken } from "../utils/auth";

export default function NotificationCenter({
  notifications,
  onClearAll,
  onClear,
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [announcements, setAnnouncements] = useState([]);

  useEffect(() => {
    const handleClickOutside = (e) => {
      if (isOpen && !e.target.closest(".notification-center")) {
        setIsOpen(false);
      }
    };

    document.addEventListener("click", handleClickOutside);
    return () => document.removeEventListener("click", handleClickOutside);
  }, [isOpen]);

  useEffect(() => {
    if (getToken()) {
      loadAnnouncements();
    }
  }, []);

  const loadAnnouncements = async () => {
    try {
      const response = await api.get("/announcements");
      const allAnnouncements = response.data?.announcements || [];
      const dismissed = getDismissedAnnouncements();
      setAnnouncements(
        allAnnouncements.filter((a) => !dismissed.includes(a.id)),
      );
    } catch (err) {
      if (err.response?.status !== 401) {
        console.error("Failed to load announcements:", err);
      }
    }
  };

  const getDismissedAnnouncements = () => {
    try {
      const dismissed = localStorage.getItem("dismissedAnnouncements");
      return dismissed ? JSON.parse(dismissed) : [];
    } catch {
      return [];
    }
  };

  const markAsDismissed = (announcementId) => {
    try {
      const dismissed = getDismissedAnnouncements();
      if (!dismissed.includes(announcementId)) {
        dismissed.push(announcementId);
        localStorage.setItem(
          "dismissedAnnouncements",
          JSON.stringify(dismissed),
        );
      }
    } catch (err) {
      console.error("Failed to save dismissed announcement:", err);
    }
  };

  const handleDismissAnnouncement = async (announcementId) => {
    try {
      await api.post(`/announcements/${announcementId}/dismiss`);
      markAsDismissed(announcementId);
      setAnnouncements(announcements.filter((a) => a.id !== announcementId));
    } catch (err) {
      console.error("Failed to dismiss announcement:", err);
    }
  };

  const unreadCount =
    notifications.filter((n) => !n.read).length + announcements.length;

  const getIcon = (type) => {
    switch (type) {
      case "success":
        return (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            style="color: #10b981;"
          >
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
            <polyline points="22 4 12 14.01 9 11.01"></polyline>
          </svg>
        );
      case "error":
        return (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            style="color: #ef4444;"
          >
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="15" y1="9" x2="9" y2="15"></line>
            <line x1="9" y1="9" x2="15" y2="15"></line>
          </svg>
        );
      case "warning":
        return (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            style="color: #f59e0b;"
          >
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
            <line x1="12" y1="9" x2="12" y2="13"></line>
            <line x1="12" y1="17" x2="12.01" y2="17"></line>
          </svg>
        );
      default:
        return (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            style="color: #3b82f6;"
          >
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="16" x2="12" y2="12"></line>
            <line x1="12" y1="8" x2="12.01" y2="8"></line>
          </svg>
        );
    }
  };

  const formatTime = (timestamp) => {
    const now = new Date();
    const time = new Date(timestamp);
    const diff = Math.floor((now - time) / 1000); // seconds

    if (diff < 60) return "Just now";
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    return time.toLocaleDateString();
  };

  return (
    <div class="notification-center">
      <button
        class="btn-icon notification-trigger"
        onClick={(e) => {
          e.stopPropagation();
          setIsOpen(!isOpen);
        }}
        title="Notifications"
      >
        <svg
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
        >
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
          <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
        </svg>
      </button>
      {unreadCount > 0 && (
        <span class="notification-badge">
          {unreadCount > 9 ? "9+" : unreadCount}
        </span>
      )}

      {isOpen && (
        <div class="notification-dropdown">
          <div class="notification-header">
            <h3>Notifications</h3>
            {(notifications.length > 0 || announcements.length > 0) && (
              <button class="btn-link" onClick={onClearAll}>
                Clear all
              </button>
            )}
          </div>

          <div class="notification-list">
            {notifications.length === 0 && announcements.length === 0 ? (
              <div class="notification-empty">
                <svg
                  width="48"
                  height="48"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  style="opacity: 0.3; margin: 0 auto;"
                >
                  <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
                  <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
                </svg>
                <p>No notifications</p>
              </div>
            ) : (
              <>
                {announcements.map((announcement) => {
                  const typeColors = {
                    info: "#3b82f6",
                    warning: "#f59e0b",
                    critical: "#ef4444",
                  };
                  return (
                    <div
                      key={announcement.id}
                      class="notification-item unread"
                      style={`border-left: 3px solid ${typeColors[announcement.type] || typeColors.info};`}
                    >
                      <div class="notification-icon">
                        {getIcon(announcement.type)}
                      </div>
                      <div class="notification-content">
                        <p
                          class="notification-message"
                          style="font-weight: 600;"
                        >
                          {announcement.title}
                        </p>
                        <p
                          class="notification-message"
                          style="font-size: 0.85rem; margin-top: 0.25rem;"
                        >
                          {announcement.message}
                        </p>
                        <span class="notification-time">
                          {formatTime(announcement.created_at)}
                        </span>
                      </div>
                      <button
                        class="notification-close"
                        onClick={() =>
                          handleDismissAnnouncement(announcement.id)
                        }
                        title="Dismiss"
                      >
                        <svg
                          width="14"
                          height="14"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                        >
                          <line x1="18" y1="6" x2="6" y2="18"></line>
                          <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                      </button>
                    </div>
                  );
                })}
                {notifications.map((notification) => (
                  <div
                    key={notification.id}
                    class={`notification-item ${!notification.read ? "unread" : ""}`}
                  >
                    <div class="notification-icon">
                      {getIcon(notification.type)}
                    </div>
                    <div class="notification-content">
                      <p class="notification-message">{notification.message}</p>
                      <span class="notification-time">
                        {formatTime(notification.timestamp)}
                      </span>
                    </div>
                    <button
                      class="notification-close"
                      onClick={() => onClear(notification.id)}
                      title="Dismiss"
                    >
                      <svg
                        width="14"
                        height="14"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                      >
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                      </svg>
                    </button>
                  </div>
                ))}
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
