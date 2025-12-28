import { useState, useEffect } from 'preact/hooks';
import { route } from 'preact-router';
import { getUser } from '../utils/auth';
import { listFiles, searchFiles, deleteFile, uploadFile } from '../utils/api';
import FileList from '../components/FileList';
import FileUpload from '../components/FileUpload';

export default function Dashboard({ isAuthenticated }) {
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState('');
  const user = getUser();

  useEffect(() => {
    if (!isAuthenticated) {
      route('/login');
      return;
    }
    loadFiles();
  }, [isAuthenticated]);

  const loadFiles = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await listFiles();
      setFiles(data.files || []);
    } catch (err) {
      setError('Failed to load files');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async (e) => {
    e.preventDefault();
    if (!searchQuery.trim()) {
      loadFiles();
      return;
    }

    setLoading(true);
    setError('');
    try {
      const data = await searchFiles(searchQuery);
      setFiles(data.files || []);
    } catch (err) {
      setError('Search failed');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleUploadComplete = () => {
    loadFiles();
  };

  const handleDelete = async (fileId) => {
    if (!confirm('Are you sure you want to delete this file?')) {
      return;
    }

    try {
      await deleteFile(fileId);
      setFiles(files.filter(f => f.id !== fileId));
    } catch (err) {
      setError('Failed to delete file');
      console.error(err);
    }
  };

  return (
    <div class="dashboard">
      <div class="dashboard-header">
        <h1>Welcome, {user?.username}!</h1>
        <p>Manage your encrypted files</p>
      </div>

      {error && <div class="alert alert-error">{error}</div>}

      <FileUpload onUploadComplete={handleUploadComplete} />

      <div class="card" style="margin-top: 2rem">
        <form onSubmit={handleSearch} class="search-form">
          <input
            type="text"
            class="form-input"
            placeholder="Search files by name or tags..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <button type="submit" class="btn btn-primary">Search</button>
          {searchQuery && (
            <button
              type="button"
              class="btn btn-secondary"
              onClick={() => {
                setSearchQuery('');
                loadFiles();
              }}
            >
              Clear
            </button>
          )}
        </form>
      </div>

      {loading ? (
        <div class="loading">
          <div class="spinner"></div>
          <p>Loading files...</p>
        </div>
      ) : (
        <FileList files={files} onDelete={handleDelete} />
      )}
    </div>
  );
}
