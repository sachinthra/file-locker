import { Router } from 'preact-router';
import { useState, useEffect } from 'preact/hooks';
import Header from './components/Header';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import Settings from './pages/Settings';
import { getToken } from './utils/auth';
import { initTheme } from './utils/theme';

export function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    const token = getToken();
    setIsAuthenticated(!!token);
    initTheme();
  }, []);

  return (
    <div class="app">
      <Header isAuthenticated={isAuthenticated} setIsAuthenticated={setIsAuthenticated} />
      <main class="container">
        <Router>
          <Login path="/" setIsAuthenticated={setIsAuthenticated} />
          <Login path="/login" setIsAuthenticated={setIsAuthenticated} />
          <Register path="/register" />
          <Dashboard path="/dashboard" isAuthenticated={isAuthenticated} setIsAuthenticated={setIsAuthenticated} />
          <Settings path="/settings" isAuthenticated={isAuthenticated} />
        </Router>
      </main>
    </div>
  );
}
