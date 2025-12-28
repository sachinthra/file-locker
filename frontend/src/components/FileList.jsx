import { useState } from 'preact/hooks';
import { downloadFile, getStreamUrl, updateFile } from '../utils/api';

export default function FileList({ files, onDelete, onUpdate }) {
  const [streamingFile, setStreamingFile] = useState(null);
  const [downloadingFile, setDownloadingFile] = useState(null);
  const [downloadProgress, setDownloadProgress] = useState(0);
  const [streamLoading, setStreamLoading] = useState(false);
  const [editingFile, setEditingFile] = useState(null);
  const [editDescription, setEditDescription] = useState('');
  const [editTags, setEditTags] = useState('');
  const [updating, setUpdating] = useState(false);
  
  if (!files || files.length === 0) {
    return (
      <div class="card" style="margin-top: 2rem; text-align: center; padding: 3rem">
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin: 0 auto; opacity: 0.5">
          <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path>
          <polyline points="13 2 13 9 20 9"></polyline>
        </svg>
        <h3 style="margin-top: 1rem; color: #666">No files yet</h3>
        <p style="color: #999">Upload your first encrypted file to get started</p>
      </div>
    );
  }

  const formatFileSize = (bytes) => {
    if (!bytes) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'Never';
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
  };

  const isVideoFile = (filename) => {
    const videoExtensions = ['.mp4', '.webm', '.ogg', '.mov', '.avi', '.mkv'];
    return videoExtensions.some(ext => filename.toLowerCase().endsWith(ext));
  };

  const handleDownload = async (fileId, filename) => {
    try {
      setDownloadingFile(fileId);
      setDownloadProgress(0);

      const response = await downloadFile(fileId, (progressEvent) => {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        setDownloadProgress(percentCompleted);
      });

      // Create download link
      const url = window.URL.createObjectURL(response.data);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Download failed:', err);
      alert('Failed to download file');
    } finally {
      setDownloadingFile(null);
      setDownloadProgress(0);
    }
  };

  const handleStream = (fileId, filename) => {
    setStreamLoading(true);
    setStreamingFile({ fileId, filename });
  };

  const closePlayer = () => {
    setStreamingFile(null);
    setStreamLoading(false);
  };

  const handleEdit = (file) => {
    setEditingFile(file);
    setEditDescription(file.description || '');
    setEditTags(file.tags ? file.tags.join(', ') : '');
  };

  const handleUpdateSubmit = async (e) => {
    e.preventDefault();
    if (!editingFile) return;

    setUpdating(true);
    try {
      const tagsArray = editTags
        .split(',')
        .map(tag => tag.trim())
        .filter(tag => tag.length > 0);

      await updateFile(editingFile.file_id, {
        description: editDescription,
        tags: tagsArray
      });

      setEditingFile(null);
      if (onUpdate) {
        onUpdate();
      }
    } catch (err) {
      console.error('Update failed:', err);
      alert(err.response?.data?.error || 'Failed to update file');
    } finally {
      setUpdating(false);
    }
  };

  const closeEditModal = () => {
    setEditingFile(null);
    setEditDescription('');
    setEditTags('');
  };

  return (
    <>
      {editingFile && (
        <div class="modal-overlay" onClick={closeEditModal}>
          <div class="modal-content" onClick={(e) => e.stopPropagation()} style="max-width: 500px">
            <div class="modal-header">
              <h3>Edit File</h3>
              <button class="btn btn-icon" onClick={closeEditModal} title="Close">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>
            <form onSubmit={handleUpdateSubmit} style="padding: 1.5rem">
              <div class="form-group">
                <label class="form-label">Description</label>
                <textarea
                  class="form-input"
                  value={editDescription}
                  onChange={(e) => setEditDescription(e.target.value)}
                  disabled={updating}
                  rows="4"
                  placeholder="Add a description..."
                />
              </div>
              <div class="form-group">
                <label class="form-label">Tags</label>
                <input
                  type="text"
                  class="form-input"
                  value={editTags}
                  onChange={(e) => setEditTags(e.target.value)}
                  disabled={updating}
                  placeholder="Enter tags separated by commas (e.g., work, important, 2024)"
                />
                <small style="color: #666; font-size: 0.875rem; margin-top: 0.25rem; display: block;">
                  Separate multiple tags with commas
                </small>
              </div>
              <div style="display: flex; gap: 0.5rem; justify-content: flex-end">
                <button type="button" class="btn btn-secondary" onClick={closeEditModal} disabled={updating}>
                  Cancel
                </button>
                <button type="submit" class="btn btn-primary" disabled={updating}>
                  {updating ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {streamingFile && (
        <div class="modal-overlay" onClick={closePlayer}>
          <div class="modal-content" onClick={(e) => e.stopPropagation()}>
            <div class="modal-header">
              <h3>{streamingFile.filename}</h3>
              <button class="btn btn-icon" onClick={closePlayer} title="Close">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>
            {streamLoading && (
              <div class="loading-overlay">
                <div class="spinner-large"></div>
              </div>
            )}
            <video 
              controls 
              autoplay
              style="width: 100%; max-height: 70vh; background: #000;"
              src={getStreamUrl(streamingFile.fileId)}
              onLoadedData={() => setStreamLoading(false)}
              onError={() => setStreamLoading(false)}
            >
              Your browser does not support video playback.
            </video>
          </div>
        </div>
      )}
      
      {downloadingFile && (
        <div class="progress-bar" style="margin-bottom: 1rem;">
          <div class="progress-bar-fill" style={`width: ${downloadProgress}%`}></div>
          <span class="progress-bar-text">Downloading... {downloadProgress}%</span>
        </div>
      )}
      
      <div class="file-list">
        {files.map(file => (
          <div key={file.file_id} class="file-item">
            <div class="file-icon">
              {isVideoFile(file.file_name) ? (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <polygon points="5 3 19 12 5 21 5 3"></polygon>
                </svg>
              ) : (
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path>
                  <polyline points="13 2 13 9 20 9"></polyline>
                </svg>
              )}
            </div>
            
            <div class="file-details">
              <div class="file-name">{file.file_name}</div>
              {file.description && (
                <div class="file-description" style="color: #666; font-size: 0.9rem; margin-top: 0.25rem">
                  {file.description}
                </div>
              )}
              <div class="file-meta">
                <span>{formatFileSize(file.size)}</span>
                <span>•</span>
                <span>Uploaded: {formatDate(file.created_at)}</span>
                {file.expires_at && (
                  <>
                    <span>•</span>
                    <span style="color: #f59e0b">Expires: {formatDate(file.expires_at)}</span>
                  </>
                )}
              </div>
              {file.tags && file.tags.length > 0 && (
                <div class="file-tags">
                  {file.tags.map(tag => (
                    <span key={tag} class="file-tag">{tag}</span>
                  ))}
                </div>
              )}
            </div>

            <div class="file-actions">
              <button
                class="btn btn-icon"
                onClick={() => handleEdit(file)}
                title="Edit"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                  <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                </svg>
              </button>

              {isVideoFile(file.file_name) && (
                <button
                  class="btn btn-icon"
                  onClick={() => handleStream(file.file_id, file.file_name)}
                  title="Stream video"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                    <polygon points="5 3 19 12 5 21 5 3"></polygon>
                  </svg>
                </button>
              )}
              
              <button
                class="btn btn-icon"
                onClick={() => handleDownload(file.file_id, file.file_name)}
                title="Download"
                disabled={downloadingFile === file.file_id}
              >
                {downloadingFile === file.file_id ? (
                  <div class="spinner"></div>
                ) : (
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                    <polyline points="8 9 12 13 16 9"></polyline>
                    <line x1="12" y1="3" x2="12" y2="13"></line>
                  </svg>
                )}
              </button>
              
              <button
                class="btn btn-icon btn-danger"
                onClick={() => onDelete(file.file_id)}
                title="Delete"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <polyline points="3 6 5 6 21 6"></polyline>
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                </svg>
              </button>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
