export default function ConfirmDialog({ 
  isOpen, 
  title = 'Confirm Action',
  message, 
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  confirmStyle = 'danger', // 'danger', 'primary', 'secondary'
  onConfirm, 
  onCancel 
}) {
  if (!isOpen) return null;

  const confirmButtonClass = `btn btn-${confirmStyle}`;

  return (
    <div class="modal-overlay" onClick={onCancel}>
      <div class="modal-content" onClick={(e) => e.stopPropagation()} style="max-width: 400px;">
        <div class="modal-header">
          <h3>{title}</h3>
          <button class="btn btn-icon" onClick={onCancel} title="Close">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
        <div style="padding: 1.5rem;">
          <p style="margin-bottom: 1.5rem;">{message}</p>
          <div style="display: flex; gap: 0.5rem; justify-content: flex-end;">
            <button class="btn btn-secondary" onClick={onCancel}>
              {cancelText}
            </button>
            <button class={confirmButtonClass} onClick={onConfirm}>
              {confirmText}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
