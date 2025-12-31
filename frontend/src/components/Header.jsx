import { useState, useEffect } from "preact/hooks";
import { route } from "preact-router";
import { removeToken, getUser, getToken } from "../utils/auth";
import { logout } from "../utils/api";
import { getTheme, toggleTheme } from "../utils/theme";
import NotificationCenter from "./NotificationCenter";
import api from "../utils/api";

export default function Header({
  isAuthenticated,
  setIsAuthenticated,
  notifications = [],
  onClearAllNotifications,
  onClearNotification,
}) {
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [theme, setTheme] = useState("light");
  const [pendingCount, setPendingCount] = useState(0);
  const user = isAuthenticated ? getUser() : null;

  useEffect(() => {
    setTheme(getTheme());

    // Fetch pending users count for admin
    if (isAuthenticated && getToken() && user?.role === "admin") {
      fetchPendingCount();
      // Poll every 30 seconds
      const interval = setInterval(fetchPendingCount, 30000);
      return () => clearInterval(interval);
    } else {
      setPendingCount(0);
    }
  }, [isAuthenticated, user?.role]);

  const fetchPendingCount = async () => {
    if (!isAuthenticated || !getToken() || user?.role !== "admin") return;
    try {
      const response = await api.get("/admin/users/pending");
      setPendingCount(response.data?.count || 0);
    } catch (err) {
      // Silently fail for 401 errors (user not admin or not logged in)
      if (err.response?.status !== 401) {
        console.error("Failed to fetch pending users count:", err);
      }
    }
  };

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e) => {
      // Ignore if typing in input/textarea
      if (e.target.tagName === "INPUT" || e.target.tagName === "TEXTAREA")
        return;

      // Ctrl/Cmd + D - Go to Dashboard
      if ((e.metaKey || e.ctrlKey) && e.key === "d") {
        e.preventDefault();
        if (isAuthenticated) {
          route("/dashboard");
        }
      }

      // Ctrl/Cmd + L - Logout
      if ((e.metaKey || e.ctrlKey) && e.key === "l") {
        e.preventDefault();
        if (isAuthenticated) {
          handleLogout();
        }
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isAuthenticated]);

  useEffect(() => {
    const handleClickOutside = (e) => {
      if (showUserMenu && !e.target.closest(".user-menu")) {
        setShowUserMenu(false);
      }
    };

    document.addEventListener("click", handleClickOutside);
    return () => document.removeEventListener("click", handleClickOutside);
  }, [showUserMenu]);

  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error("Logout error:", error);
    } finally {
      removeToken();
      setIsAuthenticated(false);
      route("/login");
    }
  };

  const handleThemeToggle = () => {
    const newTheme = toggleTheme();
    setTheme(newTheme);
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
              <NotificationCenter
                notifications={notifications}
                onClearAll={onClearAllNotifications}
                onClear={onClearNotification}
              />

              <button
                onClick={handleThemeToggle}
                class="btn-icon theme-toggle"
                title="Toggle theme"
              >
                {theme === "light" ? (
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                  >
                    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
                  </svg>
                ) : (
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                  >
                    <circle cx="12" cy="12" r="5"></circle>
                    <line x1="12" y1="1" x2="12" y2="3"></line>
                    <line x1="12" y1="21" x2="12" y2="23"></line>
                    <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
                    <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
                    <line x1="1" y1="12" x2="3" y2="12"></line>
                    <line x1="21" y1="12" x2="23" y2="12"></line>
                    <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
                    <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
                  </svg>
                )}
              </button>

              <div class="user-menu">
                <button
                  class="user-menu-trigger"
                  onClick={(e) => {
                    e.stopPropagation();
                    setShowUserMenu(!showUserMenu);
                  }}
                >
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                  >
                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                    <circle cx="12" cy="7" r="4"></circle>
                  </svg>
                  <span>{user?.username || "User"}</span>
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                  >
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>

                {showUserMenu && (
                  <div class="user-menu-dropdown">
                    <a href="/dashboard" class="user-menu-item">
                      <svg
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                      >
                        <rect x="3" y="3" width="7" height="7"></rect>
                        <rect x="14" y="3" width="7" height="7"></rect>
                        <rect x="14" y="14" width="7" height="7"></rect>
                        <rect x="3" y="14" width="7" height="7"></rect>
                      </svg>
                      <span style="flex: 1">Dashboard</span>
                      <span class="menu-shortcut">
                        <kbd>⌘</kbd> / <kbd>Ctrl</kbd> + <kbd>D</kbd>
                      </span>
                    </a>
                    <a href="/settings" class="user-menu-item">
                      <svg
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                      >
                        <circle cx="12" cy="12" r="3"></circle>
                        <path d="M12 1v6m0 6v6m7.071-13.071l-4.243 4.243m-5.656 5.656l-4.243 4.243m16.97.001l-4.243-4.243m-5.656-5.656L1.929 1.929"></path>
                      </svg>
                      <span style="flex: 1">Settings</span>
                      <span class="menu-shortcut">
                        <kbd>⌘</kbd> / <kbd>Ctrl</kbd> + <kbd>S</kbd>
                      </span>
                    </a>
                    {user?.role === "admin" && (
                      <a
                        href="/admin"
                        class="user-menu-item"
                        style="background: linear-gradient(135deg, rgba(102, 126, 234, 0.1), rgba(118, 75, 162, 0.1));"
                      >
                        <svg
                          width="18"
                          height="18"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          style="color: var(--primary-color);"
                        >
                          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
                        </svg>
                        <span style="flex: 1; color: var(--primary-color); font-weight: 600;">
                          Admin Panel
                        </span>
                        <div style="display: flex; align-items: center; gap: 0.5rem;">
                          {pendingCount > 0 && (
                            <span style="background: #ef4444; color: white; padding: 0.125rem 0.5rem; border-radius: 9999px; font-size: 0.7rem; font-weight: 700; animation: pulse 2s infinite;">
                              {pendingCount}
                            </span>
                          )}
                          <span style="background: var(--primary-color); color: white; padding: 0.125rem 0.5rem; border-radius: var(--radius-xl); font-size: 0.7rem; font-weight: 600;">
                            ADMIN
                          </span>
                        </div>
                      </a>
                    )}
                    <div class="user-menu-divider"></div>
                    <button onClick={handleLogout} class="user-menu-item">
                      <svg
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                      >
                        <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                        <polyline points="16 17 21 12 16 7"></polyline>
                        <line x1="21" y1="12" x2="9" y2="12"></line>
                      </svg>
                      <span style="flex: 1">Logout</span>
                      <span class="menu-shortcut">
                        <kbd>⌘</kbd> / <kbd>Ctrl</kbd> + <kbd>L</kbd>
                      </span>
                    </button>
                  </div>
                )}
              </div>
            </>
          ) : (
            <>
              <button
                onClick={handleThemeToggle}
                class="btn-icon theme-toggle"
                title="Toggle theme"
              >
                {theme === "light" ? (
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                  >
                    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
                  </svg>
                ) : (
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                  >
                    <circle cx="12" cy="12" r="5"></circle>
                    <line x1="12" y1="1" x2="12" y2="3"></line>
                    <line x1="12" y1="21" x2="12" y2="23"></line>
                    <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
                    <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
                    <line x1="1" y1="12" x2="3" y2="12"></line>
                    <line x1="21" y1="12" x2="23" y2="12"></line>
                    <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
                    <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
                  </svg>
                )}
              </button>
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
