package main

import (
	"encoding/json"
	"net/http"
)

// jsonError writes an error JSON response at the main package level
// (used only by middleware before requests reach handlers).
func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(map[string]string{"error": message})
}
