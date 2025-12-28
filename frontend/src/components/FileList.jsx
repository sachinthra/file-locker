import { useState } from 'preact/hooks';
import { getDownloadUrl, getStreamUrl } from '../utils/api';

export default function FileList({ files, onDelete }) {
  const [streamingFile, setStreamingFile] = useState(null);
  
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

  const handleDownload = (fileId, filename) => {
    const url = getDownloadUrl(fileId);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  };

  const handleStream = (fileId, filename) => {
    setStreamingFile({ fileId, filename });
  };

  const closePlayer = () => {
    setStreamingFile(null);
  };

  return (
    <>
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
            <video 
              controls 
              autoplay
              style="width: 100%; max-height: 70vh; background: #000;"
              src={getStreamUrl(streamingFile.fileId)}
            >
              Your browser does not support video playback.
            </video>
          </div>
        </div>
      )}
      
    <div class="card" style="margin-top: 2rem">
      <h3>Your Files ({files.length})</h3>
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
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                  <polyline points="8 9 12 13 16 9"></polyline>
                  <line x1="12" y1="3" x2="12" y2="13"></line>
                </svg>
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
    </div>
    </>
  );
}
