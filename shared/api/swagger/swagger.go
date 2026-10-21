// Package swagger serves the generated OpenAPI specs together with Swagger UI.
package swagger

import (
	"embed"
	"net/http"
)

//go:embed index.html spec
var files embed.FS

// Handler serves Swagger UI at /swagger-ui.html and the spec at /swagger.json.
// specPath is relative to the spec dir, e.g. "inventory/v1/inventory.swagger.json".
func Handler(specPath string) http.Handler {
	mux := http.NewServeMux()

	serve := func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			http.ServeFileFS(w, r, files, name)
		}
	}

	mux.HandleFunc("/swagger-ui.html", serve("index.html"))
	mux.HandleFunc("/swagger.json", serve("spec/"+specPath))

	return mux
}
