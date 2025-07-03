package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// WebHandler handles web interface requests
type WebHandler struct {
	templates *template.Template
}

// NewWebHandler creates a new web handler
func NewWebHandler() *WebHandler {
	// Parse templates
	templatePath := filepath.Join("web", "templates", "*.html")
	templates, err := template.ParseGlob(templatePath)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse templates, serving without templates")
	}

	return &WebHandler{
		templates: templates,
	}
}

// Index serves the main application page
func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	if h.templates != nil {
		err := h.templates.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			log.Error().Err(err).Msg("Failed to execute template")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback HTML if templates fail to load
		fallbackHTML := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>BlogWriter</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0;
            color: white;
        }
        .container {
            text-align: center;
            background: rgba(255, 255, 255, 0.1);
            padding: 3rem;
            border-radius: 16px;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
        h1 { margin-bottom: 1rem; }
        p { margin-bottom: 2rem; opacity: 0.8; }
        .api-links { margin-top: 2rem; }
        .api-links a {
            color: #f093fb;
            text-decoration: none;
            margin: 0 1rem;
            padding: 0.5rem 1rem;
            border: 1px solid rgba(240, 147, 251, 0.3);
            border-radius: 8px;
            transition: all 0.2s ease;
        }
        .api-links a:hover {
            background: rgba(240, 147, 251, 0.1);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 BlogWriter API</h1>
        <p>Your production-ready blog API is running successfully!</p>
        <div class="api-links">
            <a href="/api/health">Health Check</a>
            <a href="/api/posts">View Posts</a>
            <a href="/api/users">View Users</a>
        </div>
        <p style="margin-top: 2rem; font-size: 0.875rem;">
            Templates not loaded. Place your templates in the web/templates directory.
        </p>
    </div>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fallbackHTML))
	}
}
