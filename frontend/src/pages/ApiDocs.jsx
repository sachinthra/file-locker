import { useEffect } from "preact/hooks";
import "./ApiDocs.css";

const ApiDocs = () => {
  useEffect(() => {
    // Load Swagger UI CSS and JS
    const link = document.createElement("link");
    link.rel = "stylesheet";
    link.href = "https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css";
    document.head.appendChild(link);

    const script = document.createElement("script");
    script.src =
      "https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js";
    script.onload = () => {
      window.SwaggerUIBundle({
        url: "/api/v1/docs/openapi.yaml",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [
          window.SwaggerUIBundle.presets.apis,
          window.SwaggerUIBundle.SwaggerUIStandalonePreset,
        ],
        layout: "BaseLayout",
        defaultModelsExpandDepth: 1,
        defaultModelExpandDepth: 1,
        docExpansion: "list",
        filter: true,
        tryItOutEnabled: true,
      });
    };
    document.body.appendChild(script);

    return () => {
      document.head.removeChild(link);
      document.body.removeChild(script);
    };
  }, []);

  return (
    <div class="api-docs-container">
      <div class="api-docs-header">
        <h1>🔒 File Locker API Documentation</h1>
        <p>Interactive API documentation with try-it-out functionality</p>
      </div>
      <div id="swagger-ui"></div>
    </div>
  );
};

export default ApiDocs;
