import { useState, useEffect } from 'preact/hooks';
import { route } from 'preact-router';
import { getUser, getToken, removeToken } from '../utils/auth';
import { listFiles, searchFiles, deleteFile, exportAllFiles } from '../utils/api';
import FileList from '../components/FileList';
import FileUpload from '../components/FileUpload';
import FileStats from '../components/FileStats';
import Toast from '../components/Toast';
import ConfirmDialog from '../components/ConfirmDialog';

export default function Dashboard({ isAuthenticated, setIsAuthenticated, addNotification }) {
  const [allFiles, setAllFiles] = useState([]);
  const [displayedFiles, setDisplayedFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [exportProgress, setExportProgress] = useState(0);
  const [toast, setToast] = useState(null);
  const [deleteConfirm, setDeleteConfirm] = useState(null);
  const [showShortcuts, setShowShortcuts] = useState(true);
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
    // Check if token exists on page load/reload
    const token = getToken();
    if (!token) {
      // No token found, redirect to login
      removeToken();
      if (setIsAuthenticated) setIsAuthenticated(false);
      route('/login', true);
      return;
    }

    if (!isAuthenticated) {
      route('/login', true);
      return;
    }
    loadFiles();
  }, [isAuthenticated]);

  // Global keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e) => {
      // Ignore if user is typing in an input/textarea or if modal is open
      const isInputFocused = ['INPUT', 'TEXTAREA'].includes(document.activeElement.tagName);
      const isModalOpen = deleteConfirm !== null;
      
      // ESC to clear search (when input is focused)
      if (e.key === 'Escape' && isInputFocused && isSearching) {
        e.preventDefault();
        handleClearSearch();
        document.activeElement.blur();
        return;
      }

      // Don't process other shortcuts if modal is open or typing
      if (isModalOpen || isInputFocused) return;

      // Ctrl/Cmd + E for Export All
      if ((e.metaKey || e.ctrlKey) && e.key === 'e') {
        e.preventDefault();
        handleExportAll();
      }
      // Ctrl/Cmd + S for Settings
      else if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault();
        route('/settings');
      }
      // Forward slash (/) to focus search
      else if (e.key === '/' && !isInputFocused) {
        e.preventDefault();
        document.querySelector('.search-form input')?.focus();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isSearching, deleteConfirm]);

  // Auto-hide keyboard shortcuts hint after 5 seconds
  useEffect(() => {
    const timer = setTimeout(() => {
      setShowShortcuts(false);
    }, 5000);

    return () => clearTimeout(timer);
  }, []);

  const loadFiles = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await listFiles();
      console.log('API Response:', response);
      console.log('Response data:', response.data);
      const files = response.data?.files || [];
      console.log('Parsed files:', files);
      setAllFiles(files);
      setDisplayedFiles(files);
      setIsSearching(false);
      setSearchQuery('');
    } catch (err) {
      // Check if error is due to authentication
      if (err.response?.status === 401) {
        setError('Session expired. Please login again.');
        removeToken();
        if (setIsAuthenticated) setIsAuthenticated(false);
        setTimeout(() => route('/login', true), 2000);
        return;
      }
      setError('Failed to load files');
      console.error('Load files error:', err);
      console.error('Error response:', err.response);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async (e) => {
    e.preventDefault();
    if (!searchQuery.trim()) {
      // If search is empty, show all files
      setDisplayedFiles(allFiles);
      setIsSearching(false);
      return;
    }

    setLoading(true);
    setError('');
    setIsSearching(true);
    try {
      const response = await searchFiles(searchQuery);
      setDisplayedFiles(response.data?.files || []);
    } catch (err) {
      setError('Search failed');
      console.error('Search error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleClearSearch = () => {
    setSearchQuery('');
    setDisplayedFiles(allFiles);
    setIsSearching(false);
  };

  const handleUploadComplete = () => {
    loadFiles();
    showToast('File uploaded successfully!', 'success');
  };

  const handleDelete = (fileId) => {
    setDeleteConfirm(fileId);
  };

  const confirmDelete = async () => {
    const fileId = deleteConfirm;
    setDeleteConfirm(null);

    try {
      await deleteFile(fileId);
      const updatedFiles = allFiles.filter(f => f.file_id !== fileId);
      setAllFiles(updatedFiles);
      setDisplayedFiles(updatedFiles.filter(f => 
        !isSearching || displayedFiles.some(df => df.file_id === f.file_id)
      ));
      showToast('File deleted successfully', 'success');
    } catch (err) {
      showToast('Failed to delete file', 'error');
      console.error(err);
    }
  };

  const cancelDelete = () => {
    setDeleteConfirm(null);
  };

  const handleExportAll = async () => {
    if (allFiles.length === 0) {
      showToast('No files to export', 'warning');
      return;
    }

    setExporting(true);
    setExportProgress(0);
    try {
      const response = await exportAllFiles((progressEvent) => {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        setExportProgress(percentCompleted);
      });

      // Create download link for ZIP
      const url = window.URL.createObjectURL(response.data);
      const a = document.createElement('a');
      a.href = url;
      a.download = `filelocker-export-${Date.now()}.zip`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
      showToast('Files exported successfully!', 'success');
    } catch (err) {
      console.error('Export failed:', err);
      showToast(err.response?.data?.error || 'Failed to export files', 'error');
    } finally {
      setExporting(false);
      setExportProgress(0);
    }
  };

  return (
    <div class="dashboard">
      {toast && <Toast message={toast.message} type={toast.type} onClose={closeToast} />}

      {error && <div class="alert alert-error">{error}</div>}

      {/* Keyboard shortcuts hint */}
      {showShortcuts && (
        <div style="position: fixed; bottom: 10px; left: 10px; background: var(--card-bg); padding: 0.5rem 1rem; border-radius: 4px; font-size: 0.75rem; color: #666; border: 1px solid var(--border-color); z-index: 10; display: flex; align-items: center; gap: 1rem;">
          <div>
            <strong>Shortcuts:</strong> 
            <span style="margin-left: 0.5rem;"><kbd>/</kbd> Search</span>
            <span style="margin-left: 0.5rem;"><kbd>ESC</kbd> Clear/Close</span>
            <span style="margin-left: 0.5rem;"><kbd>⌘/Ctrl+E</kbd> Export</span>
            <span style="margin-left: 0.5rem;"><kbd>⌘/Ctrl+S</kbd> Settings</span>
          </div>
          <button 
            onClick={() => setShowShortcuts(false)} 
            class="btn-icon" 
            style="padding: 0.25rem; font-size: 0.75rem;"
            title="Close"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
      )}

      <div class="dashboard-grid">
        {/* Column 1: Statistics */}
        <div class="dashboard-col stats-col">
          <FileStats files={allFiles} />
        </div>

        {/* Column 2: Upload */}
        <div class="dashboard-col upload-col">
          <FileUpload onUploadComplete={handleUploadComplete} />
        </div>

        {/* Column 3: File List & Search */}
        <div class="dashboard-col files-col">
          <div class="card">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
              <h2 style="margin: 0;">Your Files ({allFiles.length})</h2>
              <button 
                onClick={handleExportAll} 
                class="btn btn-primary"
                style="padding: 0.5rem 1rem; display: flex; align-items: center; white-space: nowrap;"
                disabled={exporting || allFiles.length === 0}
                title="Export all files as ZIP"
              >
                {exporting ? (
                  <span>Exporting... {exportProgress}%</span>
                ) : (
                  <>
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
                      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                      <polyline points="7 10 12 15 17 10"></polyline>
                      <line x1="12" y1="15" x2="12" y2="3"></line>
                    </svg>
                    Export All
                  </>
                )}
              </button>
            </div>
            <form onSubmit={handleSearch} class="search-form">
              <input
                type="text"
                class="form-input"
                placeholder="Search files by name or tags..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
              <button type="submit" class="btn btn-primary">Search</button>
              {isSearching && (
                <button
                  type="button"
                  class="btn btn-secondary"
                  onClick={handleClearSearch}
                >
                  Clear
                </button>
              )}
            </form>
            {isSearching && (
              <div class="search-status">
                Showing {displayedFiles.length} of {allFiles.length} files
              </div>
            )}
          </div>

          {loading ? (
            <div class="loading">
              <div class="spinner"></div>
              <p>Loading files...</p>
            </div>
          ) : (
            <FileList 
              files={displayedFiles} 
              onDelete={handleDelete}
              onUpdate={loadFiles}
            />
          )}
        </div>
      </div>

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={deleteConfirm !== null}
        title="Confirm Delete"
        message="Are you sure you want to delete this file? This action cannot be undone."
        confirmText="Delete"
        cancelText="Cancel"
        confirmStyle="danger"
        onConfirm={confirmDelete}
        onCancel={cancelDelete}
      />

      {toast && <Toast message={toast.message} type={toast.type} onClose={closeToast} />}
    </div>
  );
}
