package api

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var openAPISpecJSON []byte

const scalarHTML = `<!doctype html>
<html lang="pt-BR">
  <head>
    <title>Tesouro Litúrgico API - Documentação Interativa</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="description" content="Documentação interativa e playground da API Tesouro Litúrgico." />
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>✝️</text></svg>">
    <style>
      body {
        margin: 0;
        padding: 0;
        background-color: #0f172a;
      }
    </style>
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/openapi.json"
      data-configuration='{
        "theme": "purple",
        "darkMode": true,
        "showSidebar": true,
        "searchHotKey": "k",
        "hideDownloadButton": false
      }'
    ></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`

// HandleDocs renders the modern Scalar interactive API documentation.
func (h *Handler) HandleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(scalarHTML))
}

// HandleOpenAPISpec serves the OpenAPI 3.1 JSON specification.
func (h *Handler) HandleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpecJSON)
}

// HandleSwaggerRedirect redirects legacy /swagger requests to modern /docs.
func (h *Handler) HandleSwaggerRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
}
