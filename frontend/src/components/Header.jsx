import { route } from 'preact-router';
import { removeToken } from '../utils/auth';
import { logout } from '../utils/api';

export default function Header({ isAuthenticated, setIsAuthenticated }) {
  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      removeToken();
      setIsAuthenticated(false);
      route('/login');
    }
  };

  return (
    <header class="header">
      <div class="header-content">
        <a href="/" class="logo">
          🔐 File Locker
        </a>
        <nav class="nav">
          {isAuthenticated ? (
            <>
              <a href="/dashboard" class="btn btn-primary btn-sm">
                Dashboard
              </a>
              <button onClick={handleLogout} class="btn btn-secondary btn-sm">
                Logout
              </button>
            </>
          ) : (
            <>
              <a href="/login" class="btn btn-primary btn-sm">
                Login
              </a>
              <a href="/register" class="btn btn-secondary btn-sm">
                Register
              </a>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
