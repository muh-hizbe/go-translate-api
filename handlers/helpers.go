package handlers

import (
	"encoding/json"
	"net/http"
)

// jsonOK writes a 200 JSON response.
func jsonOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, data)
}

// jsonError writes an error JSON response.
func jsonError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(data)
}

// decodeJSON reads and decodes the request body into dst.
// Returns false and writes a 400 response if decoding fails.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		jsonError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}
