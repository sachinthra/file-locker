import { useState, useEffect } from 'preact/hooks';
import { route } from 'preact-router';
import { getUser, getToken, removeToken } from '../utils/auth';
import { listFiles, searchFiles, deleteFile, exportAllFiles } from '../utils/api';
import FileList from '../components/FileList';
import FileUpload from '../components/FileUpload';
import FileStats from '../components/FileStats';
import Toast from '../components/Toast';

export default function Dashboard({ isAuthenticated, setIsAuthenticated }) {
  const [allFiles, setAllFiles] = useState([]);
  const [displayedFiles, setDisplayedFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [exportProgress, setExportProgress] = useState(0);
  const [toast, setToast] = useState(null);
  const user = getUser();

  const showToast = (message, type = 'info') => {
    setToast({ message, type });
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

  const handleDelete = async (fileId) => {
    if (!confirm('Are you sure you want to delete this file? This action cannot be undone.')) {
      return;
    }

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

      <div class="dashboard-header">
        <div>
          <h1>Welcome, {user?.username}!</h1>
          <p>Manage your encrypted files securely</p>
        </div>
        <div>
          <button 
            onClick={handleExportAll} 
            class="btn btn-primary"
            disabled={exporting || allFiles.length === 0}
            title="Export all files as ZIP"
          >
            {exporting ? (
              <>
                <span>Exporting... {exportProgress}%</span>
              </>
            ) : (
              <>
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin-right: 0.5rem;">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                  <polyline points="7 10 12 15 17 10"></polyline>
                  <line x1="12" y1="15" x2="12" y2="3"></line>
                </svg>
                Export All ({allFiles.length})
              </>
            )}
          </button>
        </div>
      </div>

      {error && <div class="alert alert-error">{error}</div>}

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
    </div>
  );
}
