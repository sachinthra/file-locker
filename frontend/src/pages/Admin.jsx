import { useState, useEffect } from "preact/hooks";
import { route } from "preact-router";
import { getUser, getToken } from "../utils/auth";
import api from "../utils/api";
import Toast from "../components/Toast";

export default function Admin({ isAuthenticated }) {
  const [stats, setStats] = useState({
    total_users: 0,
    total_files: 0,
    total_storage_bytes: 0,
    active_users_24h: 0,
  });
  const [users, setUsers] = useState([]);
  const [pendingUsers, setPendingUsers] = useState([]);
  const [announcements, setAnnouncements] = useState([]);
  const [recentLogs, setRecentLogs] = useState([]);
  const [recentFiles, setRecentFiles] = useState([]);
  const [settings, setSettings] = useState({});
  const [showAnnouncementForm, setShowAnnouncementForm] = useState(false);
  const [newAnnouncement, setNewAnnouncement] = useState({
    title: "",
    message: "",
    type: "info",
    target_type: "all",
  });
  const [storageAnalysis, setStorageAnalysis] = useState(null);
  const [analyzingStorage, setAnalyzingStorage] = useState(false);
  const [cleaningStorage, setCleaningStorage] = useState(false);
  const [selectedOrphans, setSelectedOrphans] = useState([]);
  const [selectedGhosts, setSelectedGhosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [toast, setToast] = useState(null);
  const user = getUser();

  const showToast = (message, type = "info") => {
    setToast({ message, type });
  };

  const closeToast = () => {
    setToast(null);
  };

  useEffect(() => {
    // Check if user is authenticated
    if (!isAuthenticated || !getToken()) {
      route("/login", true);
      return;
    }

    // Check if user is admin
    if (user?.role !== "admin") {
      showToast("Admin access required", "error");
      route("/dashboard", true);
      return;
    }

    loadData();
  }, [isAuthenticated]);

  const loadData = async () => {
    setLoading(true);
    setError("");
    try {
      // Load stats
      const statsResponse = await api.get("/admin/stats");
      setStats(statsResponse.data);

      // Load users
      const usersResponse = await api.get("/admin/users");
      setUsers(usersResponse.data?.users || []);

      // Load pending users
      try {
        const pendingResponse = await api.get("/admin/users/pending");
        setPendingUsers(pendingResponse.data?.pending_users || []);
      } catch (err) {
        console.error("Failed to load pending users:", err);
      }

      // Load settings
      try {
        const settingsResponse = await api.get("/admin/settings");
        setSettings(settingsResponse.data?.settings || {});
      } catch (err) {
        console.error("Failed to load settings:", err);
      }

      // Load announcements
      try {
        const announcementsResponse = await api.get("/admin/announcements");
        setAnnouncements(announcementsResponse.data?.announcements || []);
      } catch (err) {
        console.error("Failed to load announcements:", err);
      }

      // Load recent audit logs
      try {
        const logsResponse = await api.get("/admin/logs?limit=50");
        setRecentLogs(logsResponse.data?.logs || []);
      } catch (err) {
        console.error("Failed to load audit logs:", err);
      }

      // Load recent files
      try {
        const filesResponse = await api.get("/admin/files");
        setRecentFiles(filesResponse.data?.files?.slice(0, 10) || []);
      } catch (err) {
        console.error("Failed to load files:", err);
      }
    } catch (err) {
      if (err.response?.status === 403) {
        setError("Admin access required");
        showToast("Admin access required", "error");
        setTimeout(() => route("/dashboard", true), 2000);
      } else {
        setError("Failed to load admin data");
        showToast("Failed to load admin data", "error");
      }
      console.error("Load admin data error:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleApproveUser = async (userId, username) => {
    try {
      await api.post(`/admin/users/${userId}/approve`);
      showToast(`User ${username} approved successfully`, "success");
      loadData(); // Reload to update pending list
    } catch (err) {
      showToast(err.response?.data?.error || "Failed to approve user", "error");
      console.error("Approve user error:", err);
    }
  };

  const handleRejectUser = async (userId, username) => {
    try {
      await api.post(`/admin/users/${userId}/reject`);
      showToast(`User ${username} rejected`, "success");
      loadData(); // Reload to update pending list
    } catch (err) {
      showToast(err.response?.data?.error || "Failed to reject user", "error");
      console.error("Reject user error:", err);
    }
  };

  const handleToggleAutoApprove = async () => {
    const currentValue = settings.registration_auto_approve?.value || "false";
    const newValue = currentValue === "true" ? "false" : "true";

    try {
      await api.patch("/admin/settings", {
        key: "registration_auto_approve",
        value: newValue,
      });
      showToast(
        `Auto-approve ${newValue === "true" ? "enabled" : "disabled"}`,
        "success",
      );
      loadData(); // Reload settings
    } catch (err) {
      showToast(
        err.response?.data?.error || "Failed to update setting",
        "error",
      );
      console.error("Update setting error:", err);
    }
  };

  const handleCreateAnnouncement = async (e) => {
    e.preventDefault();
    if (!newAnnouncement.title.trim() || !newAnnouncement.message.trim()) {
      showToast("Title and message are required", "error");
      return;
    }

    try {
      await api.post("/admin/announcements", newAnnouncement);
      showToast("Announcement created successfully", "success");
      setNewAnnouncement({
        title: "",
        message: "",
        type: "info",
        target_type: "all",
      });
      setShowAnnouncementForm(false);
      loadData(); // Reload announcements
    } catch (err) {
      showToast(
        err.response?.data?.error || "Failed to create announcement",
        "error",
      );
      console.error("Create announcement error:", err);
    }
  };

  const handleDeleteAnnouncement = async (id) => {
    if (!confirm("Are you sure you want to delete this announcement?")) return;

    try {
      await api.delete(`/admin/announcements/${id}`);
      showToast("Announcement deleted successfully", "success");
      loadData(); // Reload announcements
    } catch (err) {
      showToast(
        err.response?.data?.error || "Failed to delete announcement",
        "error",
      );
      console.error("Delete announcement error:", err);
    }
  };

  const handleAnalyzeStorage = async () => {
    setAnalyzingStorage(true);
    try {
      const response = await api.get("/admin/storage/analyze");
      setStorageAnalysis(response.data);
      showToast("Storage analysis completed", "success");
    } catch (err) {
      showToast(
        err.response?.data?.error || "Failed to analyze storage",
        "error",
      );
      console.error("Analyze storage error:", err);
    } finally {
      setAnalyzingStorage(false);
    }
  };

  const handleCleanupStorage = async () => {
    if (selectedOrphans.length === 0 && selectedGhosts.length === 0) {
      showToast("Please select items to cleanup", "warning");
      return;
    }

    const confirmMsg = `Are you sure you want to cleanup ${selectedOrphans.length} orphaned file(s) and ${selectedGhosts.length} ghost record(s)? This action cannot be undone.`;
    if (!confirm(confirmMsg)) return;

    setCleaningStorage(true);
    try {
      const response = await api.post("/admin/storage/cleanup", {
        orphaned_paths: selectedOrphans,
        ghost_file_ids: selectedGhosts,
      });

      const { orphaned_deleted, orphaned_failed, ghost_deleted, ghost_failed } =
        response.data;

      showToast(
        `Cleanup completed: ${orphaned_deleted} files deleted, ${ghost_deleted} records removed. ${orphaned_failed + ghost_failed} failed.`,
        orphaned_failed + ghost_failed > 0 ? "warning" : "success",
      );

      // Reset and re-analyze
      setSelectedOrphans([]);
      setSelectedGhosts([]);
      await handleAnalyzeStorage();
    } catch (err) {
      showToast(
        err.response?.data?.error || "Failed to cleanup storage",
        "error",
      );
      console.error("Cleanup storage error:", err);
    } finally {
      setCleaningStorage(false);
    }
  };

  const toggleOrphanSelection = (path) => {
    setSelectedOrphans((prev) =>
      prev.includes(path) ? prev.filter((p) => p !== path) : [...prev, path],
    );
  };

  const toggleGhostSelection = (fileId) => {
    setSelectedGhosts((prev) =>
      prev.includes(fileId)
        ? prev.filter((id) => id !== fileId)
        : [...prev, fileId],
    );
  };

  const selectAllOrphans = () => {
    if (!storageAnalysis) return;
    setSelectedOrphans(storageAnalysis.orphaned_files.map((f) => f.path));
  };

  const selectAllGhosts = () => {
    if (!storageAnalysis) return;
    setSelectedGhosts(storageAnalysis.ghost_files.map((f) => f.file_id));
  };

  const deselectAllOrphans = () => setSelectedOrphans([]);
  const deselectAllGhosts = () => setSelectedGhosts([]);

  const formatBytes = (bytes) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  };

  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  if (loading) {
    return (
      <div
        class="loading"
        style="min-height: 100vh; display: flex; align-items: center; justify-content: center;"
      >
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
    <div
      class="admin-container"
      style="padding: 2rem; max-width: 1400px; margin: 0 auto;"
    >
      {toast && (
        <Toast message={toast.message} type={toast.type} onClose={closeToast} />
      )}

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
          onClick={() =>
            document
              .getElementById("users-section")
              ?.scrollIntoView({ behavior: "smooth" })
          }
          onMouseOver={(e) =>
            (e.currentTarget.style.transform = "translateY(-4px)")
          }
          onMouseOut={(e) =>
            (e.currentTarget.style.transform = "translateY(0)")
          }
        >
          <div style="display: flex; justify-content: space-between; align-items: flex-start;">
            <div>
              <h3 style="margin: 0 0 0.5rem 0; font-size: 2.5rem; font-weight: 700;">
                {stats.total_users}
              </h3>
              <p style="margin: 0; opacity: 0.9; font-size: 1.1rem;">
                Total Users
              </p>
            </div>
            <div style="font-size: 2.5rem; opacity: 0.7;">👥</div>
          </div>
        </div>

        <div
          class="card"
          style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; box-shadow: var(--shadow-lg); cursor: pointer; transition: transform 0.2s;"
          onClick={() => route("/admin/files")}
          onMouseOver={(e) =>
            (e.currentTarget.style.transform = "translateY(-4px)")
          }
          onMouseOut={(e) =>
            (e.currentTarget.style.transform = "translateY(0)")
          }
        >
          <div style="display: flex; justify-content: space-between; align-items: flex-start;">
            <div>
              <h3 style="margin: 0 0 0.5rem 0; font-size: 2.5rem; font-weight: 700;">
                {stats.total_files}
              </h3>
              <p style="margin: 0; opacity: 0.9; font-size: 1.1rem;">
                Total Files
              </p>
            </div>
            <div style="font-size: 2.5rem; opacity: 0.7;">📁</div>
          </div>
        </div>

        <div
          class="card"
          style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); color: white; box-shadow: var(--shadow-lg); cursor: pointer; transition: transform 0.2s;"
          onClick={() => route("/admin/files")}
          onMouseOver={(e) =>
            (e.currentTarget.style.transform = "translateY(-4px)")
          }
          onMouseOut={(e) =>
            (e.currentTarget.style.transform = "translateY(0)")
          }
        >
          <div style="display: flex; justify-content: space-between; align-items: flex-start;">
            <div>
              <h3 style="margin: 0 0 0.5rem 0; font-size: 2.5rem; font-weight: 700;">
                {formatBytes(stats.total_storage_bytes)}
              </h3>
              <p style="margin: 0; opacity: 0.9; font-size: 1.1rem;">
                Total Storage
              </p>
            </div>
            <div style="font-size: 2.5rem; opacity: 0.7;">💾</div>
          </div>
        </div>

        <div
          class="card"
          style="background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); color: white; box-shadow: var(--shadow-lg); cursor: pointer; transition: transform 0.2s;"
          onClick={() =>
            document
              .getElementById("users-section")
              ?.scrollIntoView({ behavior: "smooth" })
          }
          onMouseOver={(e) =>
            (e.currentTarget.style.transform = "translateY(-4px)")
          }
          onMouseOut={(e) =>
            (e.currentTarget.style.transform = "translateY(0)")
          }
        >
          <div style="display: flex; justify-content: space-between; align-items: flex-start;">
            <div>
              <h3 style="margin: 0 0 0.5rem 0; font-size: 2.5rem; font-weight: 700;">
                {stats.active_users_24h}
              </h3>
              <p style="margin: 0; opacity: 0.9; font-size: 1.1rem;">
                Active Today
              </p>
            </div>
            <div style="font-size: 2.5rem; opacity: 0.7;">⚡</div>
          </div>
        </div>
      </div>

      {/* Pending Users Section */}
      {pendingUsers.length > 0 && (
        <div
          class="card"
          style="margin-bottom: 2rem; border-left: 4px solid #ef4444;"
        >
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
            <div style="display: flex; align-items: center; gap: 0.75rem;">
              <h2 style="margin: 0; font-size: 1.5rem;">
                ⏳ Pending User Approvals
              </h2>
              <span style="display: inline-flex; align-items: center; justify-content: center; background: #ef4444; color: white; border-radius: 9999px; padding: 0.25rem 0.75rem; font-size: 0.875rem; font-weight: 700;">
                {pendingUsers.length}
              </span>
            </div>
            <div style="display: flex; align-items: center; gap: 1rem;">
              <label style="display: flex; align-items: center; gap: 0.5rem; font-size: 0.9rem; color: var(--text-secondary); cursor: pointer;">
                <input
                  type="checkbox"
                  checked={settings.registration_auto_approve?.value === "true"}
                  onChange={handleToggleAutoApprove}
                  style="width: 18px; height: 18px; cursor: pointer;"
                />
                <span>Auto-approve new users</span>
              </label>
            </div>
          </div>

          <div style="overflow-x: auto;">
            <table style="width: 100%; border-collapse: collapse;">
              <thead>
                <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Username
                  </th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Email
                  </th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Registered
                  </th>
                  <th style="padding: 0.75rem; text-align: center; font-weight: 600; color: var(--text-secondary);">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {pendingUsers.map((u) => (
                  <tr
                    key={u.id}
                    style="border-bottom: 1px solid var(--border-color);"
                  >
                    <td style="padding: 0.75rem;">
                      <div style="display: flex; align-items: center; gap: 0.5rem;">
                        <div style="width: 32px; height: 32px; border-radius: 50%; background: #fbbf24; display: flex; align-items: center; justify-content: center; color: white; font-weight: 600;">
                          {u.username.charAt(0).toUpperCase()}
                        </div>
                        <span style="font-weight: 500;">{u.username}</span>
                      </div>
                    </td>
                    <td style="padding: 0.75rem; color: var(--text-secondary);">
                      {u.email}
                    </td>
                    <td style="padding: 0.75rem; color: var(--text-secondary); font-size: 0.9rem;">
                      {formatDate(u.created_at)}
                    </td>
                    <td style="padding: 0.75rem; text-align: center;">
                      <div style="display: flex; gap: 0.5rem; justify-content: center;">
                        <button
                          class="btn btn-primary btn-sm"
                          onClick={() => handleApproveUser(u.id, u.username)}
                          title="Approve user"
                        >
                          <svg
                            width="14"
                            height="14"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            style="margin-right: 0.25rem;"
                          >
                            <polyline points="20 6 9 17 4 12"></polyline>
                          </svg>
                          Approve
                        </button>
                        <button
                          class="btn btn-danger btn-sm"
                          onClick={() => handleRejectUser(u.id, u.username)}
                          title="Reject user"
                        >
                          <svg
                            width="14"
                            height="14"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            style="margin-right: 0.25rem;"
                          >
                            <line x1="18" y1="6" x2="6" y2="18"></line>
                            <line x1="6" y1="6" x2="18" y2="18"></line>
                          </svg>
                          Reject
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Announcements Section */}
      <div id="announcements-section" class="card">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
          <h2 style="margin: 0; font-size: 1.5rem;">System Announcements</h2>
          <button
            class="btn btn-primary btn-sm"
            onClick={() => setShowAnnouncementForm(!showAnnouncementForm)}
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              style="margin-right: 0.5rem;"
            >
              <line x1="12" y1="5" x2="12" y2="19"></line>
              <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg>
            New Announcement
          </button>
        </div>

        {showAnnouncementForm && (
          <div style="background: var(--bg-secondary); border-radius: 8px; padding: 1.5rem; margin-bottom: 1.5rem;">
            <form onSubmit={handleCreateAnnouncement}>
              <div style="margin-bottom: 1rem;">
                <label style="display: block; margin-bottom: 0.5rem; font-weight: 500;">
                  Title
                </label>
                <input
                  type="text"
                  class="input"
                  value={newAnnouncement.title}
                  onChange={(e) =>
                    setNewAnnouncement({
                      ...newAnnouncement,
                      title: e.target.value,
                    })
                  }
                  placeholder="Announcement title"
                  required
                />
              </div>

              <div style="margin-bottom: 1rem;">
                <label style="display: block; margin-bottom: 0.5rem; font-weight: 500;">
                  Message
                </label>
                <textarea
                  class="input"
                  value={newAnnouncement.message}
                  onChange={(e) =>
                    setNewAnnouncement({
                      ...newAnnouncement,
                      message: e.target.value,
                    })
                  }
                  placeholder="Announcement message"
                  rows="4"
                  required
                  style="resize: vertical;"
                />
              </div>

              <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-bottom: 1rem;">
                <div>
                  <label style="display: block; margin-bottom: 0.5rem; font-weight: 500;">
                    Type
                  </label>
                  <select
                    class="input"
                    value={newAnnouncement.type}
                    onChange={(e) =>
                      setNewAnnouncement({
                        ...newAnnouncement,
                        type: e.target.value,
                      })
                    }
                  >
                    <option value="info">Info</option>
                    <option value="warning">Warning</option>
                    <option value="critical">Critical</option>
                  </select>
                </div>
                <div>
                  <label style="display: block; margin-bottom: 0.5rem; font-weight: 500;">
                    Target
                  </label>
                  <select
                    class="input"
                    value={newAnnouncement.target_type}
                    onChange={(e) =>
                      setNewAnnouncement({
                        ...newAnnouncement,
                        target_type: e.target.value,
                      })
                    }
                  >
                    <option value="all">All Users</option>
                    <option value="specific_users">
                      Specific Users (Future)
                    </option>
                  </select>
                </div>
              </div>

              <div style="display: flex; gap: 0.5rem; justify-content: flex-end;">
                <button
                  type="button"
                  class="btn btn-secondary btn-sm"
                  onClick={() => setShowAnnouncementForm(false)}
                >
                  Cancel
                </button>
                <button type="submit" class="btn btn-primary btn-sm">
                  Create Announcement
                </button>
              </div>
            </form>
          </div>
        )}

        <div style="display: flex; flex-direction: column; gap: 1rem;">
          {announcements.length === 0 ? (
            <div style="text-align: center; padding: 2rem; color: var(--text-secondary);">
              No announcements yet
            </div>
          ) : (
            announcements.map((announcement) => {
              const typeColors = {
                info: { bg: "#dbeafe", border: "#3b82f6", text: "#1e40af" },
                warning: { bg: "#fef3c7", border: "#f59e0b", text: "#92400e" },
                critical: { bg: "#fee2e2", border: "#ef4444", text: "#991b1b" },
              };
              const colors = typeColors[announcement.type] || typeColors.info;

              return (
                <div
                  key={announcement.id}
                  style={`border: 2px solid ${colors.border}; border-radius: 8px; padding: 1rem; background: ${colors.bg};`}
                >
                  <div style="display: flex; justify-content: space-between; align-items: start; margin-bottom: 0.5rem;">
                    <div style="flex: 1;">
                      <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.25rem;">
                        <h3
                          style={`margin: 0; font-size: 1.1rem; color: ${colors.text};`}
                        >
                          {announcement.title}
                        </h3>
                        <span
                          style={`padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: 600; background: ${colors.border}; color: white; text-transform: uppercase;`}
                        >
                          {announcement.type}
                        </span>
                      </div>
                      <p style={`margin: 0.5rem 0 0 0; color: ${colors.text};`}>
                        {announcement.message}
                      </p>
                      <p style="margin: 0.5rem 0 0 0; font-size: 0.85rem; color: var(--text-secondary);">
                        Created: {formatDate(announcement.created_at)}
                        {announcement.expires_at &&
                          ` • Expires: ${formatDate(announcement.expires_at)}`}
                      </p>
                    </div>
                    <button
                      class="btn btn-secondary btn-sm"
                      onClick={() => handleDeleteAnnouncement(announcement.id)}
                      title="Delete announcement"
                      style="flex-shrink: 0;"
                    >
                      <svg
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                      >
                        <polyline points="3 6 5 6 21 6"></polyline>
                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                      </svg>
                    </button>
                  </div>
                </div>
              );
            })
          )}
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
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              style="margin-right: 0.5rem;"
            >
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
            <table
              class="admin-table"
              style="width: 100%; border-collapse: collapse;"
            >
              <thead>
                <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Username
                  </th>
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Email
                  </th>
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Role
                  </th>
                  <th style="padding: 1rem; text-align: right; font-weight: 600; color: var(--text-secondary);">
                    Files
                  </th>
                  <th style="padding: 1rem; text-align: right; font-weight: 600; color: var(--text-secondary);">
                    Storage
                  </th>
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Joined
                  </th>
                  <th style="padding: 1rem; text-align: center; font-weight: 600; color: var(--text-secondary);">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr
                    key={u.id}
                    style={{
                      borderBottom: "1px solid var(--border-color)",
                      transition: "background-color 0.2s",
                      opacity: u.is_active === false ? "0.5" : "1",
                      background:
                        u.is_active === false
                          ? "var(--bg-secondary)"
                          : "transparent",
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
                    <td style="padding: 1rem; color: var(--text-secondary);">
                      {u.email}
                    </td>
                    <td style="padding: 1rem;">
                      <span
                        style={{
                          display: "inline-block",
                          padding: "0.25rem 0.75rem",
                          borderRadius: "var(--radius-xl)",
                          fontSize: "0.85rem",
                          fontWeight: "600",
                          background:
                            u.role === "admin"
                              ? "linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
                              : "var(--bg-secondary)",
                          color:
                            u.role === "admin"
                              ? "white"
                              : "var(--text-secondary)",
                        }}
                      >
                        {u.role === "admin" ? "🛡️ Admin" : "User"}
                      </span>
                    </td>
                    <td style="padding: 1rem; text-align: right; font-weight: 500;">
                      {u.file_count}
                    </td>
                    <td style="padding: 1rem; text-align: right; font-weight: 500;">
                      {formatBytes(u.total_storage)}
                    </td>
                    <td style="padding: 1rem; color: var(--text-secondary); font-size: 0.9rem;">
                      {formatDate(u.created_at)}
                    </td>
                    <td style="padding: 1rem; text-align: center;">
                      {u.id !== (user?.user_id || user?.id) ? (
                        <button
                          class="btn btn-primary btn-sm"
                          onClick={() => route(`/admin/users/${u.id}`)}
                          title="View user details"
                        >
                          <svg
                            width="14"
                            height="14"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            style="margin-right: 0.25rem;"
                          >
                            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                            <circle cx="12" cy="7" r="4"></circle>
                          </svg>
                          Manage
                        </button>
                      ) : (
                        <span style="color: var(--text-muted); font-size: 0.85rem;">
                          You
                        </span>
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
            onClick={() => route("/admin/audit-logs")}
          >
            View All Logs
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              style="margin-left: 0.5rem;"
            >
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
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Action
                  </th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Target
                  </th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Time
                  </th>
                </tr>
              </thead>
              <tbody>
                {recentLogs.slice(0, 10).map((log) => {
                  // Safely extract nullable DB fields
                  const targetType =
                    typeof log.target_type === "object" &&
                    log.target_type !== null
                      ? log.target_type.Valid
                        ? log.target_type.String
                        : ""
                      : log.target_type || "";
                  const targetId =
                    typeof log.target_id === "object" && log.target_id !== null
                      ? log.target_id.Valid
                        ? log.target_id.String
                        : null
                      : log.target_id;

                  return (
                    <tr
                      key={log.id}
                      style="border-bottom: 1px solid var(--border-color);"
                    >
                      <td style="padding: 0.75rem;">
                        <span
                          style={{
                            display: "inline-block",
                            padding: "0.25rem 0.5rem",
                            background: "var(--bg-secondary)",
                            borderRadius: "var(--radius-sm)",
                            fontSize: "0.8rem",
                            fontWeight: "500",
                          }}
                        >
                          {log.action}
                        </span>
                      </td>
                      <td style="padding: 0.75rem; color: var(--text-secondary);">
                        {targetType}{" "}
                        {targetId
                          ? `(${String(targetId).substring(0, 8)}...)`
                          : ""}
                      </td>
                      <td style="padding: 0.75rem; color: var(--text-secondary);">
                        {formatDate(log.created_at)}
                      </td>
                    </tr>
                  );
                })}
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
            onClick={() => route("/admin/files")}
          >
            View All Files
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              style="margin-left: 0.5rem;"
            >
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
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Filename
                  </th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Owner
                  </th>
                  <th style="padding: 0.75rem; text-align: right; font-weight: 600; color: var(--text-secondary);">
                    Size
                  </th>
                  <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Uploaded
                  </th>
                </tr>
              </thead>
              <tbody>
                {recentFiles.map((file) => {
                  // Safely extract username, handling nullable DB fields
                  const username =
                    typeof file.username === "object" && file.username !== null
                      ? file.username.Valid
                        ? file.username.String
                        : "Unknown"
                      : file.username || "Unknown";

                  return (
                    <tr
                      key={file.id}
                      style="border-bottom: 1px solid var(--border-color);"
                    >
                      <td style="padding: 0.75rem; font-weight: 500;">
                        {file.filename}
                      </td>
                      <td style="padding: 0.75rem; color: var(--text-secondary);">
                        {username}
                      </td>
                      <td style="padding: 0.75rem; text-align: right; color: var(--text-secondary);">
                        {formatBytes(file.size)}
                      </td>
                      <td style="padding: 0.75rem; color: var(--text-secondary);">
                        {formatDate(file.created_at)}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Storage Cleanup Section */}
      <div id="storage-section" class="card">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
          <div>
            <h2 style="margin: 0; font-size: 1.5rem;">Storage Cleanup</h2>
            <p style="margin: 0.5rem 0 0 0; color: var(--text-secondary); font-size: 0.9rem;">
              Identify and cleanup orphaned files and ghost records
            </p>
          </div>
          <button
            class="btn btn-primary btn-sm"
            onClick={handleAnalyzeStorage}
            disabled={analyzingStorage}
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              style="margin-right: 0.5rem;"
            >
              <path d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
            </svg>
            {analyzingStorage ? "Analyzing..." : "Analyze Storage"}
          </button>
        </div>

        {storageAnalysis && (
          <div>
            {/* Summary Cards */}
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 1.5rem;">
              <div style="background: var(--bg-secondary); border-radius: 8px; padding: 1rem;">
                <div style="color: var(--text-secondary); font-size: 0.85rem; margin-bottom: 0.25rem;">
                  Orphaned Files
                </div>
                <div style="font-size: 1.5rem; font-weight: 600; color: #f59e0b;">
                  {storageAnalysis.orphaned_files?.length || 0}
                </div>
                <div style="color: var(--text-secondary); font-size: 0.8rem; margin-top: 0.25rem;">
                  {formatBytes(
                    storageAnalysis.orphaned_files?.reduce(
                      (sum, f) => sum + f.size,
                      0,
                    ) || 0,
                  )}
                </div>
              </div>

              <div style="background: var(--bg-secondary); border-radius: 8px; padding: 1rem;">
                <div style="color: var(--text-secondary); font-size: 0.85rem; margin-bottom: 0.25rem;">
                  Ghost Records
                </div>
                <div style="font-size: 1.5rem; font-weight: 600; color: #ef4444;">
                  {storageAnalysis.ghost_files?.length || 0}
                </div>
                <div style="color: var(--text-secondary); font-size: 0.8rem; margin-top: 0.25rem;">
                  Database entries without files
                </div>
              </div>

              <div style="background: var(--bg-secondary); border-radius: 8px; padding: 1rem;">
                <div style="color: var(--text-secondary); font-size: 0.85rem; margin-bottom: 0.25rem;">
                  Selected for Cleanup
                </div>
                <div style="font-size: 1.5rem; font-weight: 600; color: #3b82f6;">
                  {selectedOrphans.length + selectedGhosts.length}
                </div>
                <div style="color: var(--text-secondary); font-size: 0.8rem; margin-top: 0.25rem;">
                  {selectedOrphans.length} files + {selectedGhosts.length}{" "}
                  records
                </div>
              </div>
            </div>

            {/* Orphaned Files Table */}
            {storageAnalysis.orphaned_files?.length > 0 && (
              <div style="margin-bottom: 1.5rem;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem;">
                  <h3 style="margin: 0; font-size: 1.1rem; color: #f59e0b;">
                    Orphaned Files (in storage, not in database)
                  </h3>
                  <div style="display: flex; gap: 0.5rem;">
                    <button
                      class="btn btn-secondary btn-sm"
                      onClick={selectAllOrphans}
                      disabled={
                        selectedOrphans.length ===
                        storageAnalysis.orphaned_files.length
                      }
                    >
                      Select All
                    </button>
                    <button
                      class="btn btn-secondary btn-sm"
                      onClick={deselectAllOrphans}
                      disabled={selectedOrphans.length === 0}
                    >
                      Deselect All
                    </button>
                  </div>
                </div>

                <div style="overflow-x: auto; border: 1px solid var(--border-color); border-radius: 8px;">
                  <table style="width: 100%; border-collapse: collapse; font-size: 0.9rem;">
                    <thead>
                      <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                        <th style="padding: 0.75rem; text-align: center; width: 40px;">
                          <input
                            type="checkbox"
                            checked={
                              selectedOrphans.length ===
                              storageAnalysis.orphaned_files.length
                            }
                            onChange={(e) =>
                              e.target.checked
                                ? selectAllOrphans()
                                : deselectAllOrphans()
                            }
                            style="width: 16px; height: 16px; cursor: pointer;"
                          />
                        </th>
                        <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                          Path
                        </th>
                        <th style="padding: 0.75rem; text-align: right; font-weight: 600; color: var(--text-secondary);">
                          Size
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {storageAnalysis.orphaned_files.map((file) => (
                        <tr
                          key={file.path}
                          style={`border-bottom: 1px solid var(--border-color); ${selectedOrphans.includes(file.path) ? "background: rgba(251, 146, 60, 0.1);" : ""}`}
                        >
                          <td style="padding: 0.75rem; text-align: center;">
                            <input
                              type="checkbox"
                              checked={selectedOrphans.includes(file.path)}
                              onChange={() => toggleOrphanSelection(file.path)}
                              style="width: 16px; height: 16px; cursor: pointer;"
                            />
                          </td>
                          <td style="padding: 0.75rem; font-family: monospace; font-size: 0.85rem; word-break: break-all;">
                            {file.path}
                          </td>
                          <td style="padding: 0.75rem; text-align: right; color: var(--text-secondary);">
                            {formatBytes(file.size)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {/* Ghost Records Table */}
            {storageAnalysis.ghost_files?.length > 0 && (
              <div style="margin-bottom: 1.5rem;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem;">
                  <h3 style="margin: 0; font-size: 1.1rem; color: #ef4444;">
                    Ghost Records (in database, not in storage)
                  </h3>
                  <div style="display: flex; gap: 0.5rem;">
                    <button
                      class="btn btn-secondary btn-sm"
                      onClick={selectAllGhosts}
                      disabled={
                        selectedGhosts.length ===
                        storageAnalysis.ghost_files.length
                      }
                    >
                      Select All
                    </button>
                    <button
                      class="btn btn-secondary btn-sm"
                      onClick={deselectAllGhosts}
                      disabled={selectedGhosts.length === 0}
                    >
                      Deselect All
                    </button>
                  </div>
                </div>

                <div style="overflow-x: auto; border: 1px solid var(--border-color); border-radius: 8px;">
                  <table style="width: 100%; border-collapse: collapse; font-size: 0.9rem;">
                    <thead>
                      <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                        <th style="padding: 0.75rem; text-align: center; width: 40px;">
                          <input
                            type="checkbox"
                            checked={
                              selectedGhosts.length ===
                              storageAnalysis.ghost_files.length
                            }
                            onChange={(e) =>
                              e.target.checked
                                ? selectAllGhosts()
                                : deselectAllGhosts()
                            }
                            style="width: 16px; height: 16px; cursor: pointer;"
                          />
                        </th>
                        <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                          File ID
                        </th>
                        <th style="padding: 0.75rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                          Path
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {storageAnalysis.ghost_files.map((file) => (
                        <tr
                          key={file.file_id}
                          style={`border-bottom: 1px solid var(--border-color); ${selectedGhosts.includes(file.file_id) ? "background: rgba(239, 68, 68, 0.1);" : ""}`}
                        >
                          <td style="padding: 0.75rem; text-align: center;">
                            <input
                              type="checkbox"
                              checked={selectedGhosts.includes(file.file_id)}
                              onChange={() =>
                                toggleGhostSelection(file.file_id)
                              }
                              style="width: 16px; height: 16px; cursor: pointer;"
                            />
                          </td>
                          <td style="padding: 0.75rem; font-family: monospace; font-size: 0.85rem;">
                            {file.file_id}
                          </td>
                          <td style="padding: 0.75rem; font-family: monospace; font-size: 0.85rem; word-break: break-all;">
                            {file.path}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {/* Cleanup Actions */}
            {(storageAnalysis.orphaned_files?.length > 0 ||
              storageAnalysis.ghost_files?.length > 0) && (
              <div style="display: flex; justify-content: flex-end; gap: 0.5rem;">
                <button
                  class="btn btn-danger"
                  onClick={handleCleanupStorage}
                  disabled={
                    cleaningStorage ||
                    (selectedOrphans.length === 0 &&
                      selectedGhosts.length === 0)
                  }
                >
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    style="margin-right: 0.5rem;"
                  >
                    <polyline points="3 6 5 6 21 6"></polyline>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                  </svg>
                  {cleaningStorage
                    ? "Cleaning..."
                    : `Cleanup Selected (${selectedOrphans.length + selectedGhosts.length})`}
                </button>
              </div>
            )}

            {/* No Issues Found */}
            {storageAnalysis.orphaned_files?.length === 0 &&
              storageAnalysis.ghost_files?.length === 0 && (
                <div style="text-align: center; padding: 3rem 1rem; color: var(--text-secondary);">
                  <svg
                    width="64"
                    height="64"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    style="margin: 0 auto 1rem; color: #10b981;"
                  >
                    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                    <polyline points="22 4 12 14.01 9 11.01"></polyline>
                  </svg>
                  <h3 style="margin: 0 0 0.5rem 0; color: #10b981;">
                    Storage is Clean!
                  </h3>
                  <p style="margin: 0; font-size: 0.9rem;">
                    No orphaned files or ghost records found.
                  </p>
                </div>
              )}
          </div>
        )}

        {!storageAnalysis && !analyzingStorage && (
          <div style="text-align: center; padding: 3rem 1rem; color: var(--text-secondary);">
            <svg
              width="64"
              height="64"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              style="margin: 0 auto 1rem;"
            >
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
              <polyline points="7 10 12 15 17 10"></polyline>
              <line x1="12" y1="15" x2="12" y2="3"></line>
            </svg>
            <h3 style="margin: 0 0 0.5rem 0;">Storage Analysis</h3>
            <p style="margin: 0 0 1.5rem 0; font-size: 0.9rem;">
              Click "Analyze Storage" to scan for orphaned files and ghost
              records
            </p>
            <div style="background: var(--bg-secondary); border-radius: 8px; padding: 1rem; text-align: left; max-width: 500px; margin: 0 auto;">
              <p style="margin: 0 0 0.5rem 0; font-weight: 600; font-size: 0.9rem;">
                What does this do?
              </p>
              <ul style="margin: 0; padding-left: 1.5rem; font-size: 0.85rem; line-height: 1.6;">
                <li>
                  <strong>Orphaned Files:</strong> Files in storage without
                  database records (safe to delete)
                </li>
                <li>
                  <strong>Ghost Records:</strong> Database entries without
                  actual files (should be removed)
                </li>
                <li>
                  <strong>Manual Review:</strong> Select items before cleanup to
                  prevent accidental deletion
                </li>
              </ul>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
