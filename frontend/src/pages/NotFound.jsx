import { route } from "preact-router";
import "./NotFound.css";

const NotFound = () => {
  return (
    <div class="not-found-container">
      <div class="not-found-content">
        <h1 class="error-code">404</h1>
        <h2 class="error-title">Page Not Found</h2>
        <p class="error-description">
          The page you're looking for doesn't exist or has been moved.
        </p>
        <div class="error-actions">
          <button onClick={() => route("/dashboard")} class="btn btn-primary">
            Go to Dashboard
          </button>
          <button onClick={() => route("/login")} class="btn btn-secondary">
            Go to Login
          </button>
        </div>
      </div>
    </div>
  );
};

export default NotFound;
