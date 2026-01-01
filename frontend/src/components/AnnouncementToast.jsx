import { h } from "preact";
import { useState, useEffect } from "preact/hooks";

export default function AnnouncementToast({ announcement, onDismiss }) {
  const [isVisible, setIsVisible] = useState(false);
  const [isExiting, setIsExiting] = useState(false);

  useEffect(() => {
    // Slide in after mount
    setTimeout(() => setIsVisible(true), 100);

    // Auto-dismiss after 8 seconds for info, 12 seconds for warning, never for critical
    if (announcement.type !== "critical") {
      const duration = announcement.type === "warning" ? 12000 : 8000;
      const timer = setTimeout(() => {
        handleDismiss();
      }, duration);
      return () => clearTimeout(timer);
    }
  }, [announcement.type]);

  const handleDismiss = () => {
    setIsExiting(true);
    setTimeout(() => {
      onDismiss(announcement.id);
    }, 300);
  };

  const typeColors = {
    info: { bg: "#3b82f6", text: "#ffffff" },
    warning: { bg: "#f59e0b", text: "#ffffff" },
    critical: { bg: "#ef4444", text: "#ffffff" },
  };

  const colors = typeColors[announcement.type] || typeColors.info;

  return (
    <div
      style={`
        position: fixed;
        top: 80px;
        right: ${isVisible && !isExiting ? "20px" : "-400px"};
        max-width: 400px;
        background: ${colors.bg};
        color: ${colors.text};
        padding: 1rem 1.25rem;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        z-index: 1000;
        transition: right 0.3s ease-out;
      `}
    >
      <div style="display: flex; justify-content: space-between; align-items: start; gap: 1rem;">
        <div style="flex: 1;">
          <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.25rem;">
            {announcement.type === "info" && (
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="12" y1="16" x2="12" y2="12"></line>
                <line x1="12" y1="8" x2="12.01" y2="8"></line>
              </svg>
            )}
            {announcement.type === "warning" && (
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                <line x1="12" y1="9" x2="12" y2="13"></line>
                <line x1="12" y1="17" x2="12.01" y2="17"></line>
              </svg>
            )}
            {announcement.type === "critical" && (
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="15" y1="9" x2="9" y2="15"></line>
                <line x1="9" y1="9" x2="15" y2="15"></line>
              </svg>
            )}
            <strong style="font-size: 1rem; font-weight: 600;">
              {announcement.title}
            </strong>
          </div>
          <p style="margin: 0.25rem 0 0 1.75rem; font-size: 0.9rem; opacity: 0.95;">
            {announcement.message}
          </p>
        </div>
        <button
          onClick={handleDismiss}
          style="
            background: transparent;
            border: none;
            cursor: pointer;
            color: currentColor;
            padding: 0.25rem;
            display: flex;
            align-items: center;
            opacity: 0.8;
            transition: opacity 0.2s;
          "
          onMouseEnter={(e) => (e.target.style.opacity = "1")}
          onMouseLeave={(e) => (e.target.style.opacity = "0.8")}
          title="Dismiss"
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>
    </div>
  );
}
