# File Locker - Frontend

Modern web interface for the File Locker secure file storage system built with Preact and Vite.

## Features

- 🔐 User authentication (login/register)
- 📤 File upload with drag-and-drop support
- 📊 Upload progress tracking
- 📁 File listing and search
- 🏷️ Tag-based organization
- ⏰ File expiration settings
- 📥 Secure file downloads
- 🎥 Video streaming support
- 🔒 AES-256-GCM encryption (handled by backend)

## Tech Stack

- **Framework**: Preact 10.28.1
- **Build Tool**: Vite 5.0.10
- **Router**: preact-router 4.1.2
- **HTTP Client**: axios 1.13.2
- **Styling**: Custom CSS with CSS variables

## Prerequisites

- Node.js 16+ and npm/yarn
- Backend server running on `http://localhost:9010`

## Installation

```bash
# Install dependencies
npm install
```

## Development

```bash
# Start development server
npm run dev
```

The app will be available at `http://localhost:5173`

## Build for Production

```bash
# Create production build
npm run build

# Preview production build
npm run preview
```

## Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── Header.jsx          # Navigation header
│   │   ├── FileList.jsx        # File list display
│   │   └── FileUpload.jsx      # Upload component with drag-and-drop
│   ├── pages/
│   │   ├── Login.jsx           # Login page
│   │   ├── Register.jsx        # Registration page
│   │   └── Dashboard.jsx       # Main dashboard
│   ├── utils/
│   │   ├── api.js              # API client with axios
│   │   └── auth.js             # Authentication utilities
│   ├── app.jsx                 # Main app component with routing
│   ├── main.jsx                # Entry point
│   └── style.css               # Global styles
├── index.html
├── vite.config.js
└── package.json
```

## API Integration

The frontend communicates with the backend REST API at `http://localhost:9010/api/v1`:

### Authentication Endpoints
- `POST /auth/login` - User login
- `POST /auth/register` - User registration
- `POST /auth/logout` - User logout

### File Endpoints
- `POST /upload` - Upload file with metadata
- `GET /files` - List user's files
- `GET /files/search?q=query` - Search files
- `DELETE /files?id=fileId` - Delete file
- `GET /download/:id` - Download file
- `GET /stream/:id` - Stream video file

## Authentication

The app uses JWT tokens stored in localStorage:
- Tokens are automatically attached to API requests via axios interceptors
- Login sets the token and redirects to dashboard
- Logout clears the token and redirects to login
- Protected routes check authentication state

## File Upload

The upload component supports:
- Drag-and-drop file selection
- Progress tracking with percentage display
- Optional tags (comma-separated)
- Optional expiration time (in hours)
- File size display

## File Operations

### Download
Files are downloaded with the original filename through authenticated endpoints.

### Streaming
Video files can be streamed directly in the browser with Range request support for seeking.

### Delete
Files can be deleted with confirmation prompt. Deletion removes both metadata (Redis) and storage (MinIO).

## Styling

The app uses CSS variables for theming:
```css
--primary-color: #567e25   /* Main brand color */
--primary-dark: #3d5a1a    /* Hover states */
--bg-color: #f5f5f5        /* Page background */
--card-bg: #ffffff         /* Card backgrounds */
--text-color: #222         /* Text color */
--border-color: #ddd       /* Borders */
```

## Environment Configuration

Update `vite.config.js` to change the backend API URL:

```javascript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:9010',
      changeOrigin: true,
    },
  },
},
```

Or update the API base URL in `src/utils/api.js`:

```javascript
const API_BASE_URL = 'http://localhost:9010/api/v1';
```

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

## Troubleshooting

### Backend Connection Issues
- Ensure the backend is running on port 9010
- Check CORS settings in backend configuration
- Verify network connectivity

### Authentication Problems
- Clear localStorage and try logging in again
- Check token expiration in backend logs
- Verify JWT secret matches between frontend and backend

### Upload Failures
- Check file size limits (default: 100MB)
- Verify MinIO is running and accessible
- Check backend logs for encryption errors

## Development Tips

1. **Hot Reload**: Vite provides instant hot module replacement
2. **Debugging**: Use browser DevTools to inspect API calls
3. **State Management**: Simple useState hooks for local state
4. **Error Handling**: All API calls include try-catch with user feedback

## License

Same as main project license.
