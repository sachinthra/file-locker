import { Router } from 'preact-router';
import { useState, useEffect } from 'preact/hooks';
import Header from './components/Header';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import { getToken } from './utils/auth';

export function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    const token = getToken();
    setIsAuthenticated(!!token);
  }, []);

  return (
    <div class="app">
      <Header isAuthenticated={isAuthenticated} setIsAuthenticated={setIsAuthenticated} />
      <main class="container">
        <Router>
          <Login path="/" setIsAuthenticated={setIsAuthenticated} />
          <Login path="/login" setIsAuthenticated={setIsAuthenticated} />
          <Register path="/register" />
          <Dashboard path="/dashboard" isAuthenticated={isAuthenticated} />
        </Router>
      </main>
    </div>
  );
}
