import { Router } from "preact-router";
import { useState, useEffect } from "preact/hooks";
import Header from "./components/Header";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";
import Settings from "./pages/Settings";
import Admin from "./pages/Admin";
import AdminAuditLogs from "./pages/AdminAuditLogs";
import AdminFiles from "./pages/AdminFiles";
import AdminUserDetails from "./pages/AdminUserDetails";
import ApiDocs from "./pages/ApiDocs";
import NotFound from "./pages/NotFound";
import { getToken } from "./utils/auth";
import { initTheme } from "./utils/theme";

export function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [notifications, setNotifications] = useState([]);

  useEffect(() => {
    const token = getToken();
    setIsAuthenticated(!!token);
    initTheme();
  }, []);

  const addNotification = (message, type = "info") => {
    const notification = {
      id: Date.now() + Math.random(),
      message,
      type,
      timestamp: Date.now(),
      read: false,
    };
    setNotifications((prev) => [notification, ...prev]);
  };

  const clearNotification = (id) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  };

  const clearAllNotifications = () => {
    setNotifications([]);
  };

  return (
    <div class="app">
      <Header
        isAuthenticated={isAuthenticated}
        setIsAuthenticated={setIsAuthenticated}
        notifications={notifications}
        onClearAllNotifications={clearAllNotifications}
        onClearNotification={clearNotification}
      />
      <main class="container">
        <Router>
          <Login path="/" setIsAuthenticated={setIsAuthenticated} />
          <Login path="/login" setIsAuthenticated={setIsAuthenticated} />
          <Register path="/register" />
          <Dashboard
            path="/dashboard"
            isAuthenticated={isAuthenticated}
            setIsAuthenticated={setIsAuthenticated}
            addNotification={addNotification}
          />
          <Settings
            path="/settings"
            isAuthenticated={isAuthenticated}
            addNotification={addNotification}
          />
          <Admin path="/admin" isAuthenticated={isAuthenticated} />
          <AdminAuditLogs
            path="/admin/audit-logs"
            isAuthenticated={isAuthenticated}
          />
          <AdminFiles path="/admin/files" isAuthenticated={isAuthenticated} />
          <AdminUserDetails
            path="/admin/users/:userId"
            isAuthenticated={isAuthenticated}
          />
          <ApiDocs path="/docs" />
          <NotFound default />
        </Router>
      </main>
    </div>
  );
}
