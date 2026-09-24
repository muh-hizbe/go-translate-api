package handlers

import (
	"net/http"

	gt "github.com/muh-hizbe/google-translate-api-go"
)

// -----------------------------------------------------------------
// POST /speak
// -----------------------------------------------------------------
// Body:   { "text": "Hello", "options": { "to": "en" } }
// 200:    { "audio": "<base64 mp3>" }
// The text must be ≤ 200 characters.

type speakRequest struct {
	Text    string           `json:"text"`
	Options translateOptions `json:"options"`
}

type speakResponse struct {
	Audio *string `json:"audio"`
}

func Speak(defaultTLD string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req speakRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.Text == "" {
			jsonError(w, "text is required", http.StatusBadRequest)
			return
		}

		audio, err := gt.SpeakOne(req.Text, req.Options.toGT(defaultTLD))
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadGateway)
			return
		}
		jsonOK(w, speakResponse{Audio: audio})
	}
}

// -----------------------------------------------------------------
// POST /speak/batch
// -----------------------------------------------------------------
// Body:
//
//	{
//	  "queries": [
//	    { "text": "Hello", "to": "en" },
//	    { "text": "World", "to": "id" }
//	  ],
//	  "options": { "to": "en" }
//	}
//
// 200: array of { "audio": "<base64>" } objects; null entries allowed when
// rejectOnPartialFail=false.

type speakBatchRequest struct {
	Queries []batchQueryItem `json:"queries"`
	Options translateOptions `json:"options"`
}

func SpeakBatch(defaultTLD string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req speakBatchRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if len(req.Queries) == 0 {
			jsonError(w, "queries must not be empty", http.StatusBadRequest)
			return
		}

		queries := make([]gt.Query, len(req.Queries))
		for i, q := range req.Queries {
			queries[i] = gt.Query{
				Text:    q.Text,
				To:      q.To,
				ForceTo: q.ForceTo,
			}
		}

		opts := req.Options.toGT(defaultTLD)
		results, err := gt.SpeakBatch(queries, opts)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadGateway)
			return
		}

		out := make([]*speakResponse, len(results))
		for i, audio := range results {
			if audio == nil {
				out[i] = nil
				continue
			}
			out[i] = &speakResponse{Audio: audio}
		}
		jsonOK(w, out)
	}
}
