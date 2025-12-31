import { useEffect } from "preact/hooks";

const Toast = ({ message, type = "info", onClose, duration = 3000 }) => {
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        onClose();
      }, duration);
      return () => clearTimeout(timer);
    }
  }, [duration, onClose]);

  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.key === "Escape") {
        onClose();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const colors = {
    success: {
      bg: "#10b981",
      icon: "✓",
    },
    error: {
      bg: "#ef4444",
      icon: "✕",
    },
    warning: {
      bg: "#f59e0b",
      icon: "⚠",
    },
    info: {
      bg: "#3b82f6",
      icon: "ℹ",
    },
  };

  const color = colors[type] || colors.info;

  return (
    <div
      class="toast"
      style={{
        position: "fixed",
        top: "20px",
        right: "20px",
        background: color.bg,
        color: "white",
        padding: "1rem 1.5rem",
        borderRadius: "8px",
        boxShadow: "0 4px 6px rgba(0, 0, 0, 0.1)",
        display: "flex",
        alignItems: "center",
        gap: "0.75rem",
        minWidth: "300px",
        maxWidth: "500px",
        zIndex: 9999,
        animation: "slideInRight 0.3s ease-out",
      }}
    >
      <span style={{ fontSize: "1.25rem", fontWeight: "bold" }}>
        {color.icon}
      </span>
      <span style={{ flex: 1 }}>{message}</span>
      <button
        onClick={onClose}
        style={{
          background: "rgba(255, 255, 255, 0.2)",
          border: "none",
          color: "white",
          width: "24px",
          height: "24px",
          borderRadius: "50%",
          cursor: "pointer",
          fontSize: "16px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          padding: 0,
        }}
      >
        ×
      </button>
    </div>
  );
};

export default Toast;
