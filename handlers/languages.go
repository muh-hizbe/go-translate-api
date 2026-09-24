package handlers

import (
	"net/http"
	"sort"
	"strings"

	gt "github.com/muh-hizbe/google-translate-api-go"
)

// languageEntry is a single item in the languages list response.
type languageEntry struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// -----------------------------------------------------------------
// GET /languages
// -----------------------------------------------------------------
// Returns the full list of supported languages sorted by code.
// Optional query param: ?q=<search> — filters by code or name (case-insensitive).
//
// 200: [ { "code": "en", "name": "English" }, ... ]

func Languages() http.HandlerFunc {
	// Pre-build the sorted list once at handler creation time.
	all := buildLanguageList()

	return func(w http.ResponseWriter, r *http.Request) {
		q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		if q == "" {
			jsonOK(w, all)
			return
		}

		filtered := make([]languageEntry, 0)
		for _, entry := range all {
			if strings.Contains(strings.ToLower(entry.Code), q) ||
				strings.Contains(strings.ToLower(entry.Name), q) {
				filtered = append(filtered, entry)
			}
		}
		jsonOK(w, filtered)
	}
}

// -----------------------------------------------------------------
// GET /languages/{code}
// -----------------------------------------------------------------
// Returns the language entry for the given ISO code or name.
// 200: { "code": "en", "name": "English" }
// 404: { "error": "language not found" }

func LanguageByCode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the {code} path segment manually (compatible with net/http
		// without a router dependency).
		// Path is expected to be /languages/{code}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			jsonError(w, "missing language code", http.StatusBadRequest)
			return
		}
		raw := parts[len(parts)-1]

		code, ok := gt.GetCode(raw)
		if !ok {
			jsonError(w, "language not found", http.StatusNotFound)
			return
		}
		jsonOK(w, languageEntry{Code: code, Name: gt.Languages[code]})
	}
}

// buildLanguageList returns a stable sorted slice of all supported languages.
func buildLanguageList() []languageEntry {
	list := make([]languageEntry, 0, len(gt.Languages))
	for code, name := range gt.Languages {
		list = append(list, languageEntry{Code: code, Name: name})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Code < list[j].Code
	})
	return list
}
