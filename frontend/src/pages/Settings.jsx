import { useState, useEffect } from "preact/hooks";
import { route } from "preact-router";
import { getUser, getToken } from "../utils/auth";
import api from "../utils/api";
import Toast from "../components/Toast";
import ConfirmDialog from "../components/ConfirmDialog";

export default function Settings({ isAuthenticated, addNotification }) {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [toast, setToast] = useState(null);
  const [showShortcuts, setShowShortcuts] = useState(true);
  const user = getUser();

  const showToast = (message, type = "info") => {
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
      route("/login", true);
    }
  }, [isAuthenticated]);

  // Auto-hide keyboard shortcuts hint after 5 seconds
  useEffect(() => {
    const timer = setTimeout(() => {
      setShowShortcuts(false);
    }, 5000);

    return () => clearTimeout(timer);
  }, []);

  const handlePasswordChange = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    // Validation
    if (!currentPassword || !newPassword || !confirmPassword) {
      setError("All fields are required");
      return;
    }

    if (newPassword !== confirmPassword) {
      setError("New passwords do not match");
      return;
    }

    if (newPassword.length < 6) {
      setError("New password must be at least 6 characters");
      return;
    }

    if (newPassword === currentPassword) {
      setError("New password must be different from current password");
      return;
    }

    setLoading(true);

    try {
      const response = await api.patch("/user/password", {
        current_password: currentPassword,
        new_password: newPassword,
      });

      setSuccess(response.data.message || "Password changed successfully");
      showToast("Password changed successfully!", "success");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err) {
      const errorMsg = err.response?.data?.error || "Failed to change password";
      setError(errorMsg);
      showToast(errorMsg, "error");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div class="settings-container">
      {toast && (
        <Toast message={toast.message} type={toast.type} onClose={closeToast} />
      )}

      {/* Keyboard shortcuts hint */}
      {showShortcuts && (
        <div style="position: fixed; bottom: 10px; left: 10px; background: var(--card-bg); padding: 0.5rem 1rem; border-radius: 4px; font-size: 0.75rem; color: #666; border: 1px solid var(--border-color); z-index: 10; display: flex; align-items: center; gap: 1rem;">
          <div>
            <strong>Shortcuts:</strong>
            <span style="margin-left: 0.5rem;">
              <kbd>ESC</kbd> Close/Cancel
            </span>
            <span style="margin-left: 0.5rem;">
              <kbd>⌘/Ctrl+Enter</kbd> Submit
            </span>
            <span style="margin-left: 0.5rem;">
              <kbd>⌘/Ctrl+D</kbd> Dashboard
            </span>
            <span style="margin-left: 0.5rem;">
              <kbd>⌘/Ctrl+L</kbd> Logout
            </span>
          </div>
          <button
            onClick={() => setShowShortcuts(false)}
            class="btn-icon"
            style="padding: 0.25rem; font-size: 0.75rem;"
            title="Close"
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
      )}

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
            <input
              type="text"
              class="form-input"
              disabled
              value={user?.username || "N/A"}
            />
            <small>Username cannot be changed</small>
          </div>
          <div class="settings-item">
            <label>Email</label>
            <input
              type="email"
              class="form-input"
              disabled
              value={user?.email || "user@example.com"}
            />
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
              {loading ? "Changing Password..." : "Change Password"}
            </button>
          </form>
        </div>

        {/* Appearance Settings */}
        <div class="card settings-section">
          <h3>Appearance</h3>
          <div class="settings-item">
            <label>Theme</label>
            <p>
              Use the theme toggle in the header to switch between light and
              dark modes
            </p>
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
            <p>
              Personal Access Tokens for CLI and script usage are managed below
              in Developer Settings.
            </p>
          </div>
          <div class="settings-item">
            <label>Export Data</label>
            <button class="btn btn-secondary" disabled>
              Download All Files (Coming Soon)
            </button>
            <small>Export all your encrypted files</small>
          </div>
        </div>

        {/* Keyboard Shortcuts */}
        <div class="card settings-section">
          <h3>Keyboard Shortcuts</h3>
          <p style="margin-bottom: 1rem; color: #666;">
            Use these shortcuts to navigate faster
          </p>

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
              <span>Open dashboard</span>
              <div>
                <kbd>⌘</kbd> / <kbd>Ctrl</kbd> + <kbd>D</kbd>
              </div>
            </div>

            <div style="display: flex; justify-content: space-between; align-items: center; padding: 0.5rem; background: var(--bg-color); border-radius: 4px;">
              <span>Logout</span>
              <div>
                <kbd>⌘</kbd> / <kbd>Ctrl</kbd> + <kbd>L</kbd>
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
            <strong>Tip:</strong> Shortcuts are context-aware and won't
            interfere when typing in input fields.
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
  const [copyButtonText, setCopyButtonText] = useState("📋 Copy Token");
  const [revokeConfirm, setRevokeConfirm] = useState(null); // { id, name }

  const loadTokens = async () => {
    try {
      const res = await api.get("/auth/tokens");
      setTokens(res.data.tokens || []);
    } catch (e) {
      addNotification && addNotification("Failed to load tokens", "error");
    }
  };

  useEffect(() => {
    loadTokens();
  }, []);

  // Keyboard shortcuts for modal
  useEffect(() => {
    if (!showCreateModal) return;

    const handleKeyDown = (e) => {
      // ESC to close
      if (e.key === "Escape") {
        setShowCreateModal(false);
        setNewTokenPlain(null);
        setCopyButtonText("📋 Copy Token");
      }
      // Cmd/Ctrl + C to copy
      if ((e.metaKey || e.ctrlKey) && e.key === "c" && newTokenPlain) {
        e.preventDefault();
        handleCopyToken();
      }
      // Cmd/Ctrl + Enter to close after copying
      if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
        if (copyButtonText.includes("Copied")) {
          setShowCreateModal(false);
          setNewTokenPlain(null);
          setCopyButtonText("📋 Copy Token");
        } else {
          handleCopyToken();
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [showCreateModal, newTokenPlain, copyButtonText]);

  const handleCopyToken = () => {
    if (newTokenPlain) {
      navigator.clipboard.writeText(newTokenPlain);
      setCopyButtonText("✅ Copied!");
      addNotification &&
        addNotification("Token copied to clipboard!", "success");
      setTimeout(() => {
        setCopyButtonText("📋 Copy Token");
      }, 2000);
    }
  };

  const handleCreate = async () => {
    if (!newTokenName.trim()) {
      addNotification && addNotification("Please enter a token name", "error");
      return;
    }
    setIsCreating(true);
    try {
      const res = await api.post("/auth/tokens", {
        name: newTokenName.trim(),
        expires_in_days: Number(expiresDays),
      });
      setNewTokenPlain(res.data.token);
      setShowCreateModal(true);
      setNewTokenName("");
      setExpiresDays(0);
      await loadTokens();
    } catch (e) {
      addNotification && addNotification("Failed to create token", "error");
    } finally {
      setIsCreating(false);
    }
  };

  const handleRevoke = async (id, name) => {
    setRevokeConfirm({ id, name });
  };

  const confirmRevoke = async () => {
    if (!revokeConfirm) return;

    try {
      await api.delete(`/auth/tokens/${revokeConfirm.id}`);
      await loadTokens();
      addNotification &&
        addNotification("Token revoked successfully", "success");
    } catch (e) {
      addNotification && addNotification("Failed to revoke token", "error");
    } finally {
      setRevokeConfirm(null);
    }
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return "Never";
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = date - now;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffMs < 0) return "Expired";
    if (diffDays === 0) return "Today";
    if (diffDays === 1) return "Tomorrow";
    if (diffDays < 7) return `In ${diffDays} days`;
    return date.toLocaleDateString();
  };

  const getTokenStatus = (expiresAt) => {
    if (!expiresAt) return { text: "Active", color: "var(--success-color)" };
    const now = new Date();
    const expires = new Date(expiresAt);
    const diffMs = expires - now;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffMs < 0) return { text: "Expired", color: "var(--error-color)" };
    if (diffDays <= 7)
      return { text: "Expiring Soon", color: "var(--warning-color)" };
    return { text: "Active", color: "var(--success-color)" };
  };

  return (
    <div style={{ maxWidth: "900px" }}>
      {/* Info Banner */}
      <div
        style={{
          background:
            "linear-gradient(135deg, var(--primary-color) 0%, var(--primary-light) 100%)",
          padding: "1.5rem",
          borderRadius: "var(--radius-lg)",
          marginBottom: "2rem",
          color: "white",
          boxShadow: "var(--shadow-lg)",
        }}
      >
        <h3
          style={{
            margin: "0 0 0.5rem 0",
            fontSize: "1.25rem",
            fontWeight: "600",
          }}
        >
          🔑 Personal Access Tokens
        </h3>
        <p
          style={{
            margin: 0,
            opacity: 0.95,
            lineHeight: 1.5,
            fontSize: "0.95rem",
          }}
        >
          Generate tokens for CLI access and automation. Treat tokens like
          passwords - they provide full account access.
        </p>
      </div>

      {/* Create Token Card */}
      <div
        style={{
          background: "var(--card-bg)",
          border: "1px solid var(--border-color)",
          borderRadius: "var(--radius-lg)",
          padding: "1.5rem",
          marginBottom: "2rem",
          boxShadow: "var(--shadow-sm)",
        }}
      >
        <h4
          style={{
            margin: "0 0 1rem 0",
            fontSize: "1.1rem",
            color: "var(--text-color)",
          }}
        >
          Generate New Token
        </h4>

        <div style={{ display: "flex", flexDirection: "column", gap: "1rem" }}>
          <div>
            <label
              style={{
                display: "block",
                marginBottom: "0.5rem",
                fontSize: "0.9rem",
                color: "var(--text-secondary)",
                fontWeight: "500",
              }}
            >
              Token Name <span style={{ color: "var(--error-color)" }}>*</span>
            </label>
            <input
              type="text"
              placeholder="e.g., MacBook CLI, Production Server, CI/CD Pipeline"
              value={newTokenName}
              onInput={(e) => setNewTokenName(e.target.value)}
              class="form-input"
            />
          </div>

          <div>
            <label
              style={{
                display: "block",
                marginBottom: "0.5rem",
                fontSize: "0.9rem",
                color: "var(--text-secondary)",
                fontWeight: "500",
              }}
            >
              Expiration
            </label>
            <div
              style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}
            >
              <input
                type="number"
                placeholder="0"
                value={expiresDays}
                onInput={(e) => setExpiresDays(e.target.value)}
                class="form-input"
                style={{ width: "120px" }}
              />
              <span style={{ color: "var(--text-muted)", fontSize: "0.9rem" }}>
                days (0 = never expires)
              </span>
            </div>
          </div>

          <button
            class="btn btn-primary"
            onClick={handleCreate}
            disabled={isCreating}
          >
            {isCreating ? "Generating..." : "✨ Generate Token"}
          </button>
        </div>
      </div>

      {/* Tokens List */}
      <div
        style={{
          background: "var(--card-bg)",
          border: "1px solid var(--border-color)",
          borderRadius: "var(--radius-lg)",
          padding: "1.5rem",
          boxShadow: "var(--shadow-sm)",
        }}
      >
        <h4
          style={{
            margin: "0 0 1rem 0",
            fontSize: "1.1rem",
            color: "var(--text-color)",
          }}
        >
          Your Tokens ({tokens.length})
        </h4>

        {tokens.length === 0 ? (
          <div
            style={{
              textAlign: "center",
              padding: "3rem 1rem",
              color: "var(--text-muted)",
            }}
          >
            <div style={{ fontSize: "3rem", marginBottom: "1rem" }}>🔐</div>
            <p style={{ margin: 0 }}>
              No tokens yet. Create your first token to get started.
            </p>
          </div>
        ) : (
          <div
            style={{ display: "flex", flexDirection: "column", gap: "1rem" }}
          >
            {tokens.map((t) => {
              const status = getTokenStatus(t.expires_at);
              return (
                <div
                  key={t.id}
                  style={{
                    background: "var(--bg-secondary)",
                    border: "1px solid var(--border-color)",
                    borderRadius: "var(--radius-md)",
                    padding: "1.25rem",
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "flex-start",
                    gap: "1rem",
                    transition: "all 0.2s ease",
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <div
                      style={{
                        display: "flex",
                        alignItems: "center",
                        gap: "0.75rem",
                        marginBottom: "0.75rem",
                      }}
                    >
                      <h5
                        style={{
                          margin: 0,
                          fontSize: "1.05rem",
                          fontWeight: "600",
                          color: "var(--text-color)",
                        }}
                      >
                        {t.name}
                      </h5>
                      <span
                        style={{
                          background: `${status.color}20`,
                          color: status.color,
                          padding: "0.25rem 0.75rem",
                          borderRadius: "var(--radius-xl)",
                          fontSize: "0.8rem",
                          fontWeight: "600",
                        }}
                      >
                        {status.text}
                      </span>
                    </div>

                    <div
                      style={{
                        display: "grid",
                        gridTemplateColumns:
                          "repeat(auto-fit, minmax(200px, 1fr))",
                        gap: "0.75rem",
                        fontSize: "0.85rem",
                        color: "var(--text-secondary)",
                      }}
                    >
                      <div>
                        <span style={{ color: "var(--text-muted)" }}>
                          Created:
                        </span>{" "}
                        {new Date(t.created_at).toLocaleString("en-US", {
                          year: "numeric",
                          month: "short",
                          day: "numeric",
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </div>
                      <div>
                        <span style={{ color: "var(--text-muted)" }}>
                          Last Used:
                        </span>{" "}
                        {t.last_used_at
                          ? new Date(t.last_used_at).toLocaleString("en-US", {
                              year: "numeric",
                              month: "short",
                              day: "numeric",
                              hour: "2-digit",
                              minute: "2-digit",
                            })
                          : "Never"}
                      </div>
                      <div>
                        <span style={{ color: "var(--text-muted)" }}>
                          Expires:
                        </span>{" "}
                        {formatDate(t.expires_at)}
                      </div>
                    </div>

                    {!t.last_used_at && (
                      <div
                        style={{
                          marginTop: "0.75rem",
                          padding: "0.5rem 0.75rem",
                          background: `${status.color}15`,
                          border: `1px solid ${status.color}40`,
                          borderRadius: "var(--radius-sm)",
                          fontSize: "0.85rem",
                          color: "var(--warning-color)",
                        }}
                      >
                        ⚠️ This token has never been used
                      </div>
                    )}
                  </div>

                  <div
                    style={{
                      display: "flex",
                      gap: "0.5rem",
                      flexShrink: 0,
                    }}
                  >
                    <button
                      class="btn btn-danger btn-sm"
                      onClick={() => handleRevoke(t.id, t.name)}
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
          onClick={() => {
            setShowCreateModal(false);
            setNewTokenPlain(null);
            setCopyButtonText("📋 Copy Token");
          }}
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: "rgba(0, 0, 0, 0.85)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 1000,
            backdropFilter: "blur(4px)",
            animation: "fadeIn 0.2s ease",
          }}
        >
          <div
            class="modal-content"
            onClick={(e) => e.stopPropagation()}
            style={{
              background: "var(--card-bg)",
              borderRadius: "var(--radius-xl)",
              maxWidth: "600px",
              width: "90%",
              border: "1px solid var(--border-color)",
              boxShadow: "var(--shadow-xl)",
              animation: "slideUp 0.3s ease",
            }}
          >
            <div
              class="modal-header"
              style={{
                padding: "1.5rem",
                borderBottom: "1px solid var(--border-color)",
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
              }}
            >
              <div>
                <h3
                  style={{
                    margin: "0 0 0.25rem 0",
                    fontSize: "1.5rem",
                    color: "var(--text-color)",
                  }}
                >
                  ✅ Token Created Successfully
                </h3>
                <p
                  style={{
                    margin: 0,
                    color: "var(--text-secondary)",
                    fontSize: "0.9rem",
                  }}
                >
                  Copy your token now - it won't be shown again
                </p>
              </div>
              <button
                class="btn-icon"
                onClick={() => {
                  setShowCreateModal(false);
                  setNewTokenPlain(null);
                  setCopyButtonText("📋 Copy Token");
                }}
                style={{
                  background: "transparent",
                  border: "none",
                  color: "var(--text-muted)",
                  fontSize: "1.5rem",
                  cursor: "pointer",
                  padding: "0.5rem",
                  lineHeight: 1,
                  borderRadius: "var(--radius-md)",
                  transition: "all 0.2s ease",
                }}
              >
                ✕
              </button>
            </div>

            <div style={{ padding: "1.5rem" }}>
              {/* Security Warning */}
              <div
                style={{
                  background: `${getComputedStyle(document.documentElement).getPropertyValue("--error-color")}15`,
                  border: `1px solid ${getComputedStyle(document.documentElement).getPropertyValue("--error-color")}40`,
                  borderRadius: "var(--radius-md)",
                  padding: "1rem",
                  marginBottom: "1.5rem",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    alignItems: "flex-start",
                    gap: "0.75rem",
                    color: "var(--error-color)",
                    fontSize: "0.9rem",
                    lineHeight: 1.6,
                  }}
                >
                  <span style={{ fontSize: "1.25rem" }}>🔒</span>
                  <div>
                    <strong>Important Security Notice:</strong>
                    <ul
                      style={{ margin: "0.5rem 0 0 0", paddingLeft: "1.25rem" }}
                    >
                      <li>This token provides full access to your account</li>
                      <li>Store it securely (e.g., password manager)</li>
                      <li>Never commit it to version control</li>
                      <li>You won't be able to see it again</li>
                    </ul>
                  </div>
                </div>
              </div>

              {/* Token Display */}
              <div style={{ marginBottom: "1.5rem" }}>
                <label
                  style={{
                    display: "block",
                    marginBottom: "0.5rem",
                    fontSize: "0.9rem",
                    color: "var(--text-secondary)",
                    fontWeight: "600",
                  }}
                >
                  Your Personal Access Token:
                </label>
                <div style={{ position: "relative" }}>
                  <pre
                    style={{
                      background: "var(--token-surface)",
                      padding: "1rem",
                      borderRadius: "var(--radius-md)",
                      color: "var(--success-color)",
                      fontSize: "0.95rem",
                      fontFamily: "monospace",
                      margin: 0,
                      wordBreak: "break-all",
                      whiteSpace: "pre-wrap",
                      border: "2px solid var(--success-color)",
                      userSelect: "all",
                    }}
                  >
                    {newTokenPlain}
                  </pre>
                </div>
              </div>

              {/* CLI Usage Example */}
              <div
                style={{
                  background: "var(--bg-secondary)",
                  border: "1px solid var(--border-color)",
                  borderRadius: "var(--radius-md)",
                  padding: "1rem",
                  marginBottom: "1.5rem",
                }}
              >
                <div
                  style={{
                    fontSize: "0.85rem",
                    color: "var(--text-muted)",
                    marginBottom: "0.5rem",
                    fontWeight: "600",
                  }}
                >
                  💻 CLI Usage:
                </div>
                <pre
                  style={{
                    margin: 0,
                    fontSize: "0.85rem",
                    color: "var(--text-secondary)",
                    fontFamily: "monospace",
                    whiteSpace: "pre-wrap",
                    wordBreak: "break-all",
                  }}
                >
                  fl login --token {newTokenPlain}
                </pre>
              </div>

              {/* Keyboard Shortcuts Info */}
              <div
                style={{
                  background: "var(--bg-secondary)",
                  border: "1px solid var(--border-color)",
                  borderRadius: "var(--radius-md)",
                  padding: "1rem",
                  marginBottom: "1.5rem",
                  fontSize: "0.85rem",
                  color: "var(--text-secondary)",
                }}
              >
                <div style={{ fontWeight: "600", marginBottom: "0.5rem" }}>
                  ⌨️ Keyboard Shortcuts:
                </div>
                <div style={{ display: "grid", gap: "0.25rem" }}>
                  <div>
                    <kbd>Cmd/Ctrl</kbd> + <kbd>C</kbd> - Copy token
                  </div>
                  <div>
                    <kbd>Cmd/Ctrl</kbd> + <kbd>Enter</kbd> - Copy and close
                  </div>
                  <div>
                    <kbd>ESC</kbd> - Close modal
                  </div>
                </div>
              </div>

              {/* Action Buttons */}
              <div
                style={{
                  display: "flex",
                  gap: "0.75rem",
                  justifyContent: "flex-end",
                }}
              >
                <button class="btn btn-primary" onClick={handleCopyToken}>
                  {copyButtonText}
                </button>
                <button
                  class="btn btn-secondary"
                  onClick={() => {
                    setShowCreateModal(false);
                    setNewTokenPlain(null);
                    setCopyButtonText("📋 Copy Token");
                  }}
                >
                  I've Saved It
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Revoke Confirmation Dialog */}
      <ConfirmDialog
        isOpen={revokeConfirm !== null}
        title="⚠️ Revoke Token?"
        message={
          revokeConfirm
            ? `Are you sure you want to revoke the token "${revokeConfirm.name}"? This action cannot be undone and any applications using this token will immediately lose access.`
            : ""
        }
        confirmText="Yes, Revoke Token"
        cancelText="Cancel"
        confirmStyle="danger"
        onConfirm={confirmRevoke}
        onCancel={() => setRevokeConfirm(null)}
      />
    </div>
  );
}
