import { useState } from 'preact/hooks';
import { uploadFile } from '../utils/api';

export default function FileUpload({ onUploadComplete }) {
  const [file, setFile] = useState(null);
  const [tags, setTags] = useState('');
  const [expiresIn, setExpiresIn] = useState('');
  const [description, setDescription] = useState('');
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState('');
  const [dragActive, setDragActive] = useState(false);

  const handleDrag = (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      setFile(e.dataTransfer.files[0]);
    }
  };

  const handleFileChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0]);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!file) {
      setError('Please select a file');
      return;
    }

    setUploading(true);
    setError('');
    setProgress(0);

    try {
      const tagArray = tags.split(',').map(t => t.trim()).filter(t => t);
      await uploadFile(file, tagArray, expiresIn, description, (progressEvent) => {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        setProgress(percentCompleted);
      });

      // Reset form
      setFile(null);
      setTags('');
      setExpiresIn('');
      setDescription('');
      setProgress(0);
      
      if (onUploadComplete) {
        onUploadComplete();
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Upload failed');
      console.error(err);
    } finally {
      setUploading(false);
    }
  };

  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  return (
    <div class="card">
      <h3>Upload File</h3>
      {error && <div class="alert alert-error">{error}</div>}
      
      <form onSubmit={handleSubmit}>
        <div
          class={`upload-zone ${dragActive ? 'active' : ''}`}
          onDragEnter={handleDrag}
          onDragLeave={handleDrag}
          onDragOver={handleDrag}
          onDrop={handleDrop}
        >
          <input
            type="file"
            id="file-input"
            onChange={handleFileChange}
            style="display: none"
            disabled={uploading}
          />
          
          {file ? (
            <div class="upload-zone-selected">
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path>
                <polyline points="13 2 13 9 20 9"></polyline>
              </svg>
              <p><strong>{file.name}</strong></p>
              <p>{formatFileSize(file.size)}</p>
              <button
                type="button"
                class="btn btn-secondary"
                onClick={() => setFile(null)}
                disabled={uploading}
              >
                Remove
              </button>
            </div>
          ) : (
            <label for="file-input" style="cursor: pointer; width: 100%; text-align: center">
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" style="margin: 0 auto">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                <polyline points="17 8 12 3 7 8"></polyline>
                <line x1="12" y1="3" x2="12" y2="15"></line>
              </svg>
              <p style="margin-top: 1rem"><strong>Drop files here or click to browse</strong></p>
              <p style="color: #666">Any file type supported</p>
            </label>
          )}
        </div>

        {uploading && (
          <div class="progress-bar">
            <div class="progress-bar-fill" style={`width: ${progress}%`}></div>
            <span class="progress-bar-text">{progress}%</span>
          </div>
        )}

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Tags (comma-separated)</label>
            <input
              type="text"
              class="form-input"
              placeholder="e.g., document, work, important"
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              disabled={uploading}
            />
          </div>

          <div class="form-group">
            <label class="form-label">Expires In (hours)</label>
            <input
              type="number"
              class="form-input"
              placeholder="Leave empty for no expiration"
              value={expiresIn}
              onChange={(e) => setExpiresIn(e.target.value)}
              min="1"
              disabled={uploading}
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">Description (optional)</label>
          <textarea
            class="form-input"
            placeholder="Add a description for this file..."
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={uploading}
            rows="3"
          />
        </div>

        <button type="submit" class="btn btn-primary" style="width: 100%" disabled={uploading || !file}>
          {uploading ? `Uploading ${progress}%...` : 'Upload File'}
        </button>
      </form>
    </div>
  );
}
