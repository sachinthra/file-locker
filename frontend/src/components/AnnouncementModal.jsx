import { h } from "preact";
import { useState, useEffect } from "preact/hooks";

export default function AnnouncementModal({ announcement, onClose }) {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    setTimeout(() => setIsVisible(true), 50);
  }, []);

  const handleClose = () => {
    setIsVisible(false);
    setTimeout(() => {
      onClose(announcement.id);
    }, 300);
  };

  const typeColors = {
    info: {
      bg: "#dbeafe",
      border: "#3b82f6",
      text: "#1e40af",
      icon: "#3b82f6",
    },
    warning: {
      bg: "#fef3c7",
      border: "#f59e0b",
      text: "#92400e",
      icon: "#f59e0b",
    },
    critical: {
      bg: "#fee2e2",
      border: "#ef4444",
      text: "#991b1b",
      icon: "#ef4444",
    },
  };

  const colors = typeColors[announcement.type] || typeColors.info;

  return (
    <div
      style={`
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, ${isVisible ? "0.5" : "0"});
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 2000;
        transition: background 0.3s ease-out;
      `}
      onClick={handleClose}
    >
      <div
        style={`
          background: ${colors.bg};
          border: 3px solid ${colors.border};
          border-radius: 12px;
          padding: 2rem;
          max-width: 500px;
          width: 90%;
          transform: scale(${isVisible ? "1" : "0.9"});
          opacity: ${isVisible ? "1" : "0"};
          transition: all 0.3s ease-out;
        `}
        onClick={(e) => e.stopPropagation()}
      >
        <div style="display: flex; align-items: start; gap: 1rem; margin-bottom: 1rem;">
          <div
            style={`
              flex-shrink: 0;
              width: 48px;
              height: 48px;
              border-radius: 50%;
              background: ${colors.icon};
              display: flex;
              align-items: center;
              justify-content: center;
              color: white;
            `}
          >
            {announcement.type === "info" && (
              <svg
                width="28"
                height="28"
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
                width="28"
                height="28"
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
                width="28"
                height="28"
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
          </div>
          <div style="flex: 1;">
            <h3
              style={`
                margin: 0 0 0.5rem 0;
                color: ${colors.text};
                font-size: 1.25rem;
                font-weight: 600;
              `}
            >
              {announcement.title}
            </h3>
            <p
              style={`
                margin: 0;
                color: ${colors.text};
                font-size: 1rem;
                line-height: 1.5;
              `}
            >
              {announcement.message}
            </p>
          </div>
        </div>
        <div style="display: flex; justify-content: flex-end; gap: 0.5rem; margin-top: 1.5rem;">
          <button
            onClick={handleClose}
            class="btn btn-primary"
            style={`background: ${colors.border}; border-color: ${colors.border};`}
          >
            Got it
          </button>
        </div>
      </div>
    </div>
  );
}
