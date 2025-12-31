import { useState, useEffect } from "preact/hooks";
import { route } from "preact-router";
import { getUser, getToken } from "../utils/auth";
import api from "../utils/api";
import Toast from "../components/Toast";
import ConfirmDialog from "../components/ConfirmDialog";

export default function AdminFiles({ isAuthenticated }) {
  const [files, setFiles] = useState([]);
  const [filteredFiles, setFilteredFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState(null);
  const [deleteConfirm, setDeleteConfirm] = useState(null);
  const [search, setSearch] = useState("");
  const user = getUser();

  const showToast = (message, type = "info") => {
    setToast({ message, type });
  };

  const closeToast = () => {
    setToast(null);
  };

  useEffect(() => {
    if (!isAuthenticated || !getToken() || user?.role !== "admin") {
      route("/admin", true);
      return;
    }
    loadFiles();
  }, [isAuthenticated]);

  useEffect(() => {
    if (search) {
      const searchLower = search.toLowerCase();
      setFilteredFiles(
        files.filter(
          (file) =>
            file.filename.toLowerCase().includes(searchLower) ||
            file.username?.String?.toLowerCase().includes(searchLower) ||
            file.content_type.toLowerCase().includes(searchLower),
        ),
      );
    } else {
      setFilteredFiles(files);
    }
  }, [files, search]);

  const loadFiles = async () => {
    setLoading(true);
    try {
      const response = await api.get("/admin/files");
      setFiles(response.data?.files || []);
    } catch (err) {
      showToast("Failed to load files", "error");
      console.error("Load files error:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteFile = (fileId, filename) => {
    setDeleteConfirm({ id: fileId, filename });
  };

  const confirmDelete = async () => {
    const fileId = deleteConfirm.id;
    const filename = deleteConfirm.filename;
    setDeleteConfirm(null);

    try {
      await api.delete(`/admin/files/${fileId}`);
      showToast(`File "${filename}" deleted successfully`, "success");
      loadFiles();
    } catch (err) {
      showToast(err.response?.data?.error || "Failed to delete file", "error");
      console.error("Delete file error:", err);
    }
  };

  const cancelDelete = () => {
    setDeleteConfirm(null);
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
  };

  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getTotalSize = () => {
    return filteredFiles.reduce((sum, file) => sum + file.size, 0);
  };

  if (loading) {
    return (
      <div
        class="loading"
        style="min-height: 100vh; display: flex; align-items: center; justify-content: center;"
      >
        <div>
          <div class="spinner"></div>
          <p>Loading files...</p>
        </div>
      </div>
    );
  }

  return (
    <div
      class="admin-container"
      style="padding: 2rem; max-width: 1600px; margin: 0 auto;"
    >
      {toast && (
        <Toast message={toast.message} type={toast.type} onClose={closeToast} />
      )}

      {/* Header */}
      <div style="margin-bottom: 2rem;">
        <button
          class="btn btn-secondary btn-sm"
          onClick={() => route("/admin")}
          style="margin-bottom: 0.5rem;"
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            style="margin-right: 0.5rem;"
          >
            <polyline points="15 18 9 12 15 6"></polyline>
          </svg>
          Back to Admin
        </button>
        <h1 style="margin: 0; font-size: 2rem; color: var(--text-color);">
          📁 Global Files Management
        </h1>
        <p style="margin: 0.5rem 0 0 0; color: var(--text-secondary);">
          {filteredFiles.length} files • {formatBytes(getTotalSize())} total
        </p>
      </div>

      {/* Search */}
      <div class="card" style="margin-bottom: 2rem;">
        <input
          type="text"
          class="input"
          placeholder="Search files by name, owner, or type..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {/* Files Table */}
      <div class="card">
        {filteredFiles.length === 0 ? (
          <div style="text-align: center; padding: 3rem 1rem; color: var(--text-muted);">
            <div style="font-size: 3rem; margin-bottom: 1rem;">📁</div>
            <p>
              {files.length === 0
                ? "No files in system"
                : "No files match your search"}
            </p>
          </div>
        ) : (
          <div style="overflow-x: auto;">
            <table style="width: 100%; border-collapse: collapse;">
              <thead>
                <tr style="border-bottom: 2px solid var(--border-color); background: var(--bg-secondary);">
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Filename
                  </th>
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Owner
                  </th>
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Type
                  </th>
                  <th style="padding: 1rem; text-align: right; font-weight: 600; color: var(--text-secondary);">
                    Size
                  </th>
                  <th style="padding: 1rem; text-align: left; font-weight: 600; color: var(--text-secondary);">
                    Uploaded
                  </th>
                  <th style="padding: 1rem; text-align: center; font-weight: 600; color: var(--text-secondary);">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {filteredFiles.map((file) => (
                  <tr
                    key={file.id}
                    style="border-bottom: 1px solid var(--border-color);"
                  >
                    <td style="padding: 1rem;">
                      <div style="display: flex; align-items: center; gap: 0.5rem;">
                        <svg
                          width="20"
                          height="20"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          style="color: var(--primary-color);"
                        >
                          <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path>
                          <polyline points="13 2 13 9 20 9"></polyline>
                        </svg>
                        <div>
                          <div style="font-weight: 500; word-break: break-word; max-width: 300px;">
                            {file.filename}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td style="padding: 1rem; color: var(--text-secondary);">
                      {file.username?.Valid ? (
                        <div style="display: flex; align-items: center; gap: 0.5rem;">
                          <div style="width: 24px; height: 24px; border-radius: 50%; background: var(--primary-color); display: flex; align-items: center; justify-content: center; color: white; font-size: 0.7rem; font-weight: 600;">
                            {file.username.String.charAt(0).toUpperCase()}
                          </div>
                          <span>{file.username.String}</span>
                        </div>
                      ) : (
                        <span style="color: var(--text-muted);">Unknown</span>
                      )}
                    </td>
                    <td style="padding: 1rem;">
                      <span style="display: inline-block; padding: 0.25rem 0.5rem; background: var(--bg-secondary); border-radius: var(--radius-sm); font-size: 0.8rem; color: var(--text-secondary);">
                        {file.content_type}
                      </span>
                    </td>
                    <td style="padding: 1rem; text-align: right; font-weight: 500;">
                      {formatBytes(file.size)}
                    </td>
                    <td style="padding: 1rem; color: var(--text-secondary);">
                      {formatDate(file.created_at)}
                    </td>
                    <td style="padding: 1rem; text-align: center;">
                      <button
                        class="btn btn-danger btn-sm"
                        onClick={() => handleDeleteFile(file.id, file.filename)}
                        title="Delete file"
                      >
                        <svg
                          width="14"
                          height="14"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          style="margin-right: 0.25rem;"
                        >
                          <polyline points="3 6 5 6 21 6"></polyline>
                          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                        </svg>
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={deleteConfirm !== null}
        title="⚠️ Delete File?"
        message={
          deleteConfirm
            ? `Are you sure you want to delete "${deleteConfirm.filename}"? This action cannot be undone.`
            : ""
        }
        confirmText="Yes, Delete File"
        cancelText="Cancel"
        confirmStyle="danger"
        onConfirm={confirmDelete}
        onCancel={cancelDelete}
      />
    </div>
  );
}
