# Frontend Implementation Summary

## Overview

The File Locker frontend is a modern single-page application (SPA) built with Preact and Vite, providing a secure and user-friendly interface for encrypted file management.

## Technology Stack

- **Framework**: Preact 10.28.1 (lightweight React alternative)
- **Build Tool**: Vite 5.0.10 (fast development and optimized builds)
- **Routing**: preact-router 4.1.2 (client-side routing)
- **HTTP Client**: axios 1.13.2 (API communication)
- **Styling**: Custom CSS with CSS variables

## Project Structure

```
frontend/
├── src/
│   ├── components/          # Reusable UI components
│   │   ├── Header.jsx       # Navigation header with auth state
│   │   ├── FileList.jsx     # File listing with actions
│   │   └── FileUpload.jsx   # Drag-and-drop upload component
│   ├── pages/               # Route pages
│   │   ├── Login.jsx        # User login
│   │   ├── Register.jsx     # User registration
│   │   └── Dashboard.jsx    # Main file management interface
│   ├── utils/               # Utility modules
│   │   ├── api.js           # API client with axios
│   │   └── auth.js          # Authentication helpers
│   ├── app.jsx              # Main app with routing
│   ├── main.jsx             # Entry point
│   └── style.css            # Global styles
├── index.html               # HTML template
├── vite.config.js           # Vite configuration
├── package.json             # Dependencies and scripts
└── README.md                # Frontend documentation
```

## Components

### Header Component
**File**: `src/components/Header.jsx`

- Displays navigation and authentication state
- Shows username when logged in
- Provides login/register or dashboard/logout buttons
- Handles logout with token cleanup and redirect

**Props**:
- `isAuthenticated`: boolean - Current auth state
- `setIsAuthenticated`: function - Auth state setter

### FileList Component
**File**: `src/components/FileList.jsx`

- Displays list of user's files
- Shows file metadata (name, size, upload date, expiration, tags)
- Provides action buttons (download, stream, delete)
- Handles video file detection for streaming
- Empty state when no files exist

**Props**:
- `files`: array - List of file objects
- `onDelete`: function - Delete handler

**Features**:
- File size formatting (Bytes, KB, MB, GB)
- Date formatting with locale support
- Video file icon differentiation
- Tag display with styling
- Confirmation dialog for deletion

### FileUpload Component
**File**: `src/components/FileUpload.jsx`

- Drag-and-drop file upload interface
- Manual file selection fallback
- Upload progress tracking with percentage
- Optional tags input (comma-separated)
- Optional expiration time (hours)
- File size display

**Props**:
- `onUploadComplete`: function - Callback after successful upload

**Features**:
- Active state for drag-over
- Progress bar with percentage
- File preview before upload
- Form validation
- Error handling with user feedback

## Pages

### Login Page
**File**: `src/pages/Login.jsx`
**Route**: `/login` (default route)

- Username and password inputs
- Form validation
- Loading state during authentication
- Error display
- Token storage on success
- Redirect to dashboard after login
- Link to registration page

### Register Page
**File**: `src/pages/Register.jsx`
**Route**: `/register`

- Username, email, and password inputs
- Password length validation (min 8 characters)
- Form validation
- Loading state during registration
- Error display
- Redirect to login after successful registration
- Link to login page

### Dashboard Page
**File**: `src/pages/Dashboard.jsx`
**Route**: `/dashboard`

- Protected route (requires authentication)
- Welcome message with username
- File upload component
- Search functionality
- File list display
- Auto-refresh on upload/delete
- Error handling

**Features**:
- Search files by name or tags
- Clear search to return to full list
- Loading spinner during operations
- Empty state handling

## Utilities

### API Client
**File**: `src/utils/api.js`

Centralized API communication with axios:

**Base Configuration**:
- Base URL: `http://localhost:9010/api/v1`
- Automatic Bearer token injection via interceptor
- Error handling at interceptor level

**Auth Methods**:
- `login(username, password)` → POST /auth/login
- `register(username, password, email)` → POST /auth/register
- `logout()` → POST /auth/logout

**File Methods**:
- `uploadFile(file, tags, expireAfter, onProgress)` → POST /upload
- `listFiles()` → GET /files
- `searchFiles(query)` → GET /files/search?q=...
- `deleteFile(fileId)` → DELETE /files?id=...
- `getDownloadUrl(fileId)` → Returns authenticated download URL
- `getStreamUrl(fileId)` → Returns authenticated stream URL

### Authentication Helper
**File**: `src/utils/auth.js`

LocalStorage-based token management:

**Methods**:
- `saveToken(token)` - Store JWT token
- `getToken()` - Retrieve JWT token
- `removeToken()` - Clear token and user data
- `saveUser(user)` - Store user object
- `getUser()` - Retrieve user object

**Storage Keys**:
- `token` - JWT authentication token
- `user` - User object (username, email, user_id)

## Routing

The app uses preact-router for client-side routing:

```jsx
<Router>
  <Login path="/" setIsAuthenticated={setIsAuthenticated} />
  <Login path="/login" setIsAuthenticated={setIsAuthenticated} />
  <Register path="/register" />
  <Dashboard path="/dashboard" isAuthenticated={isAuthenticated} />
</Router>
```

**Routes**:
- `/` or `/login` - Login page (default)
- `/register` - Registration page
- `/dashboard` - Main dashboard (protected)

## Authentication Flow

1. **Login**:
   - User enters credentials
   - API call to `/auth/login`
   - Token saved to localStorage
   - User object saved to localStorage
   - `setIsAuthenticated(true)` called
   - Redirect to `/dashboard`

2. **Protected Routes**:
   - Dashboard checks `isAuthenticated` prop
   - Redirects to `/login` if not authenticated
   - Token automatically attached to API requests via interceptor

3. **Logout**:
   - API call to `/auth/logout`
   - Token and user data removed from localStorage
   - `setIsAuthenticated(false)` called
   - Redirect to `/login`

## Styling

Custom CSS with CSS variables for theming:

**CSS Variables**:
```css
--primary-color: #567e25      /* Main brand color (green) */
--primary-dark: #3d5a1a       /* Hover states */
--bg-color: #f5f5f5           /* Page background */
--card-bg: #ffffff            /* Card backgrounds */
--text-color: #222            /* Text color */
--border-color: #ddd          /* Borders */
--error-color: #dc3545        /* Error messages */
--success-color: #28a745      /* Success messages */
--warning-color: #ffc107      /* Warnings */
```

**Responsive Design**:
- Breakpoint at 768px for mobile/tablet
- Flexible grid layouts
- Mobile-friendly forms and buttons

## API Integration

### Request Flow
1. User action triggers API call
2. Axios interceptor adds Bearer token from localStorage
3. Request sent to backend
4. Response handled (success or error)
5. UI updated with result or error message

### Error Handling
- Network errors caught and displayed to user
- 401 errors can trigger re-authentication
- Validation errors shown in forms
- Generic fallback messages for unknown errors

## File Operations

### Upload
- FormData with file, tags, and expiration
- Progress tracking with `onUploadProgress` callback
- Success triggers file list refresh

### Download
- Authenticated URL with token query parameter
- Browser handles download with original filename
- Link element programmatically clicked

### Stream (Video)
- Authenticated URL with token query parameter
- Opens in new tab for browser native player
- Backend supports Range requests for seeking

### Delete
- Confirmation dialog before deletion
- API call removes from Redis and MinIO
- File removed from local state immediately on success

## Security

- **Authentication**: JWT tokens in localStorage
- **Token Expiry**: Backend handles token validation
- **Protected Routes**: Dashboard redirects if not authenticated
- **HTTPS**: Recommended for production (handled by reverse proxy)
- **CORS**: Configured in backend for allowed origins

## Development

### Running Locally
```bash
npm install         # Install dependencies
npm run dev         # Start dev server at localhost:5173
```

### Building for Production
```bash
npm run build       # Creates dist/ folder
npm run preview     # Preview production build locally
```

### Vite Configuration
**File**: `vite.config.js`

- Preact plugin for JSX transformation
- Proxy configuration for backend API at `/api`
- Development server on port 5173

## Future Enhancements

### Planned Features
- [ ] Bulk file upload
- [ ] Folder organization
- [ ] Share links with expiration
- [ ] File preview (images, PDFs)
- [ ] Download multiple files as ZIP
- [ ] Advanced search with filters
- [ ] User profile management
- [ ] Dark mode toggle
- [ ] Upload queue management
- [ ] Thumbnail generation for images/videos

### Technical Improvements
- [ ] Add unit tests with Vitest
- [ ] E2E tests with Playwright
- [ ] State management with signals
- [ ] Service worker for offline support
- [ ] Progressive Web App (PWA) features
- [ ] Internationalization (i18n)
- [ ] Accessibility improvements (ARIA)
- [ ] Performance monitoring

## Browser Compatibility

**Minimum Requirements**:
- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

**Features Used**:
- ES6+ JavaScript
- Fetch API / XMLHttpRequest (via axios)
- LocalStorage
- FormData
- CSS Grid and Flexbox

## Performance

**Optimizations**:
- Vite's optimized build with code splitting
- Preact's small bundle size (~3KB)
- CSS in single file (no CSS-in-JS overhead)
- Lazy loading not needed due to small app size
- No heavy dependencies

**Bundle Size** (approximate):
- Vendor: ~40KB (preact, router, axios)
- App: ~15KB (components and pages)
- CSS: ~8KB
- **Total**: ~63KB minified + gzipped

## Deployment

### Static Hosting
Build the app and serve the `dist/` folder:

```bash
npm run build
```

Deploy to:
- Netlify
- Vercel
- GitHub Pages
- AWS S3 + CloudFront
- Nginx/Apache static hosting

### Environment Variables
Update API base URL for different environments:

**Development**: `http://localhost:9010/api/v1`
**Production**: Update in `src/utils/api.js` or use environment variables

### Reverse Proxy
For production, serve frontend and backend through same domain:

```nginx
location /api/ {
    proxy_pass http://backend:9010;
}

location / {
    root /var/www/frontend/dist;
    try_files $uri $uri/ /index.html;
}
```

## Testing

### Manual Testing Checklist
- [ ] Register new account
- [ ] Login with credentials
- [ ] Upload file with drag-and-drop
- [ ] Upload file with click-to-browse
- [ ] Upload file with tags
- [ ] Upload file with expiration
- [ ] View file list
- [ ] Search files by name
- [ ] Search files by tag
- [ ] Download file
- [ ] Stream video file
- [ ] Delete file
- [ ] Logout
- [ ] Login again and verify files persist

### Test Accounts
Create test accounts with various scenarios:
- User with no files
- User with many files
- User with expired files
- User with video files

## Troubleshooting

### Frontend won't start
- Check Node.js version (16+)
- Delete `node_modules` and reinstall
- Check port 5173 is available

### Can't connect to backend
- Verify backend is running on port 9010
- Check Vite proxy configuration
- Check CORS settings in backend

### Authentication issues
- Clear localStorage and try again
- Check token in DevTools → Application → Local Storage
- Verify JWT secret matches backend

### Upload failures
- Check file size limits
- Verify MinIO is running
- Check backend logs for encryption errors

### Styling issues
- Hard refresh (Cmd+Shift+R / Ctrl+Shift+R)
- Clear browser cache
- Check CSS file is loaded in DevTools

## Resources

- [Preact Documentation](https://preactjs.com/)
- [Vite Documentation](https://vitejs.dev/)
- [Axios Documentation](https://axios-http.com/)
- [Backend API Documentation](../backend/docs/openapi.yaml)

## License

Same as main project (MIT License)
