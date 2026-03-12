// Package server provides the HTTP server that backs the workload-capacity UI.
package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/KimHansenCubris/gh-godo/internal/excel"
)

//go:embed static
var staticFiles embed.FS

// Start initialises routes and listens on addr.
func Start(addr string) error {
	mux := http.NewServeMux()

	// Serve embedded static files (HTML/CSS/JS).
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("static fs: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// API endpoints.
	mux.HandleFunc("/api/rows", handleRows)
	mux.HandleFunc("/api/rows/", handleRowByID)

	return http.ListenAndServe(addr, mux)
}

// filePath extracts and validates the ?file= query parameter.
func filePath(r *http.Request) (string, error) {
	p := r.URL.Query().Get("file")
	if p == "" {
		return "", fmt.Errorf("missing 'file' query parameter")
	}
	return filepath.Clean(p), nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// handleRows handles GET /api/rows and POST /api/rows.
func handleRows(w http.ResponseWriter, r *http.Request) {
	path, err := filePath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := excel.ReadRows(path)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if rows == nil {
			rows = []excel.Row{}
		}
		writeJSON(w, http.StatusOK, rows)

	case http.MethodPost:
		var row excel.Row
		if err := json.NewDecoder(r.Body).Decode(&row); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if err := excel.AppendRow(path, row); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleRowByID handles PUT /api/rows/{id} and DELETE /api/rows/{id}.
func handleRowByID(w http.ResponseWriter, r *http.Request) {
	path, err := filePath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		// Fallback for Go versions without PathValue: strip prefix.
		idStr = r.URL.Path[len("/api/rows/"):]
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid row id")
		return
	}

	switch r.Method {
	case http.MethodPut:
		var row excel.Row
		if err := json.NewDecoder(r.Body).Decode(&row); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		row.ID = id
		if err := excel.UpdateRow(path, row); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	case http.MethodDelete:
		if err := excel.DeleteRow(path, id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
