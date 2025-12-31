import { useState, useEffect } from "preact/hooks";
import { route } from "preact-router";
import { register } from "../utils/api";
import { getToken } from "../utils/auth";

export default function Register() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // If already logged in, redirect to dashboard
    if (getToken()) {
      route("/dashboard", true);
    }
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      setLoading(false);
      return;
    }

    try {
      await register(username, password, email);
      route("/login");
    } catch (err) {
      setError(
        err.response?.data?.error || "Registration failed. Please try again.",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div class="form">
      <h2>Create Account</h2>
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
            minLength={3}
          />
        </div>

        <div class="form-group">
          <label class="form-label">Email</label>
          <input
            type="email"
            class="form-input"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
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
            minLength={8}
          />
          <small style="color: #666">Must be at least 8 characters</small>
        </div>

        <button
          type="submit"
          class="btn btn-primary"
          style="width: 100%"
          disabled={loading}
        >
          {loading ? "Creating account..." : "Register"}
        </button>
      </form>

      <p style="text-align: center; margin-top: 1rem">
        Already have an account? <a href="/login">Login here</a>
      </p>
    </div>
  );
}
