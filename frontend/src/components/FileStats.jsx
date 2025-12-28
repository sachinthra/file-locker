export default function FileStats({ files }) {
  const totalFiles = files?.length || 0;
  
  const totalSize = files?.reduce((acc, file) => acc + (file.size || 0), 0) || 0;
  
  const videoFiles = files?.filter(file => {
    const videoExtensions = ['.mp4', '.webm', '.ogg', '.mov', '.avi', '.mkv'];
    return videoExtensions.some(ext => file.file_name?.toLowerCase().endsWith(ext));
  }).length || 0;
  
  const expiringFiles = files?.filter(file => file.expires_at).length || 0;
  
  const formatSize = (bytes) => {
    if (!bytes) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  return (
    <div class="stats-container">
      {/* <h3>Overview</h3> */}
      
      <div class="stat-card">
        <div class="stat-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path>
            <polyline points="13 2 13 9 20 9"></polyline>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">Total Files</div>
          <div class="stat-value">{totalFiles}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">Total Storage</div>
          <div class="stat-value">{formatSize(totalSize)}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <polygon points="5 3 19 12 5 21 5 3"></polygon>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">Video Files</div>
          <div class="stat-value">{videoFiles}</div>
        </div>
      </div>

      {expiringFiles > 0 && (
        <div class="stat-card stat-warning">
          <div class="stat-icon">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <circle cx="12" cy="12" r="10"></circle>
              <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
          </div>
          <div class="stat-content">
            <div class="stat-label">Expiring Soon</div>
            <div class="stat-value">{expiringFiles}</div>
          </div>
        </div>
      )}

      <div class="quick-actions">
        <h4>Quick Tips</h4>
        <ul>
          <li>🔐 All files are encrypted with AES-256</li>
          <li>🎥 Video files can be streamed</li>
          <li>🏷️ Use tags to organize files</li>
          <li>⏰ Set expiration for temporary files</li>
        </ul>
      </div>
    </div>
  );
}
