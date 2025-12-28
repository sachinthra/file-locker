import { useState, useEffect } from 'preact/hooks';
import { route } from 'preact-router';
import { login } from '../utils/api';
import { saveToken, saveUser, getToken } from '../utils/auth';

export default function Login({ setIsAuthenticated }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // If already logged in, redirect to dashboard
    if (getToken()) {
      route('/dashboard', true);
    }
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await login(username, password);
      const { token, user_id, email } = response.data;
      
      saveToken(token);
      saveUser({ user_id, email, username });
      setIsAuthenticated(true);
      route('/dashboard');
    } catch (err) {
      setError(err.response?.data?.error || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div class="form">
      <h2>Login to File Locker</h2>
      {error && <div class="alert alert-error">{error}</div>}
      
      <form onSubmit={handleSubmit}>
        <div class="form-group">
          <label class="form-label">Username</label>
          <input
            type="text"
            class="form-input"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
        </div>

        <div class="form-group">
          <label class="form-label">Password</label>
          <input
            type="password"
            class="form-input"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>

        <button type="submit" class="btn btn-primary" style="width: 100%" disabled={loading}>
          {loading ? 'Logging in...' : 'Login'}
        </button>
      </form>

      <p style="text-align: center; margin-top: 1rem">
        Don't have an account? <a href="/register">Register here</a>
      </p>
    </div>
  );
}
