package handlers

import (
	"net/http"

	gt "github.com/muh-hizbe/google-translate-api-go"
)

// -----------------------------------------------------------------
// Shared option mapping
// -----------------------------------------------------------------

// translateOptions is the JSON-visible subset of gt.Options.
// Pointer booleans let callers explicitly send false.
type translateOptions struct {
	From                string `json:"from"`
	To                  string `json:"to"`
	ForceFrom           bool   `json:"forceFrom"`
	ForceTo             bool   `json:"forceTo"`
	AutoCorrect         bool   `json:"autoCorrect"`
	TLD                 string `json:"tld"`
	ForceBatch          *bool  `json:"forceBatch"`
	FallbackBatch       *bool  `json:"fallbackBatch"`
	RejectOnPartialFail *bool  `json:"rejectOnPartialFail"`
}

func (o translateOptions) toGT(defaultTLD string) gt.Options {
	opts := gt.DefaultOptions()
	if o.From != "" {
		opts.From = o.From
	}
	if o.To != "" {
		opts.To = o.To
	}
	if o.TLD != "" {
		opts.TLD = o.TLD
	} else {
		opts.TLD = defaultTLD
	}
	opts.ForceFrom = o.ForceFrom
	opts.ForceTo = o.ForceTo
	opts.AutoCorrect = o.AutoCorrect
	if o.ForceBatch != nil {
		opts.ForceBatch = *o.ForceBatch
	}
	if o.FallbackBatch != nil {
		opts.FallbackBatch = *o.FallbackBatch
	}
	if o.RejectOnPartialFail != nil {
		opts.RejectOnPartialFail = *o.RejectOnPartialFail
	}
	return opts
}

// -----------------------------------------------------------------
// Shared response shapes
// -----------------------------------------------------------------

type translationResponse struct {
	Text          string              `json:"text"`
	Pronunciation *string             `json:"pronunciation,omitempty"`
	From          translationFromInfo `json:"from"`
}

type translationFromInfo struct {
	Language translationLanguageInfo `json:"language"`
	Text     translationTextInfo     `json:"text"`
}

type translationLanguageInfo struct {
	DidYouMean bool   `json:"didYouMean"`
	ISO        string `json:"iso"`
}

type translationTextInfo struct {
	AutoCorrected bool   `json:"autoCorrected"`
	Value         string `json:"value"`
	DidYouMean    bool   `json:"didYouMean"`
}

func resultToResponse(r *gt.TranslationResult) translationResponse {
	if r == nil {
		return translationResponse{}
	}
	return translationResponse{
		Text:          r.Text,
		Pronunciation: r.Pronunciation,
		From: translationFromInfo{
			Language: translationLanguageInfo{
				DidYouMean: r.From.Language.DidYouMean,
				ISO:        r.From.Language.ISO,
			},
			Text: translationTextInfo{
				AutoCorrected: r.From.Text.AutoCorrected,
				Value:         r.From.Text.Value,
				DidYouMean:    r.From.Text.DidYouMean,
			},
		},
	}
}

// -----------------------------------------------------------------
// POST /translate
// -----------------------------------------------------------------
// Body:   { "text": "Hello", "options": { "to": "id" } }
// 200:    { "text": "Halo", "from": { "language": { "iso": "en", ... }, "text": {} } }

type translateRequest struct {
	Text    string           `json:"text"`
	Options translateOptions `json:"options"`
}

func Translate(defaultTLD string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req translateRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.Text == "" {
			jsonError(w, "text is required", http.StatusBadRequest)
			return
		}

		result, err := gt.Translate(req.Text, req.Options.toGT(defaultTLD))
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadGateway)
			return
		}
		jsonOK(w, resultToResponse(result))
	}
}

// -----------------------------------------------------------------
// POST /translate/batch
// -----------------------------------------------------------------
// Body:
//
//	{
//	  "queries": [
//	    { "text": "Hello", "to": "id" },
//	    { "text": "World" }
//	  ],
//	  "options": { "to": "id" }
//	}
//
// 200: array of translation objects; null entries when rejectOnPartialFail=false.

type batchQueryItem struct {
	Text        string `json:"text"`
	From        string `json:"from"`
	To          string `json:"to"`
	ForceFrom   bool   `json:"forceFrom"`
	ForceTo     bool   `json:"forceTo"`
	AutoCorrect bool   `json:"autoCorrect"`
}

type batchTranslateRequest struct {
	Queries []batchQueryItem `json:"queries"`
	Options translateOptions `json:"options"`
}

func TranslateBatch(defaultTLD string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req batchTranslateRequest
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
				Text:        q.Text,
				From:        q.From,
				To:          q.To,
				ForceFrom:   q.ForceFrom,
				ForceTo:     q.ForceTo,
				AutoCorrect: q.AutoCorrect,
			}
		}

		results, err := gt.TranslateBatch(queries, req.Options.toGT(defaultTLD))
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadGateway)
			return
		}

		out := make([]*translationResponse, len(results))
		for i, res := range results {
			if res == nil {
				out[i] = nil
				continue
			}
			v := resultToResponse(res)
			out[i] = &v
		}
		jsonOK(w, out)
	}
}

// -----------------------------------------------------------------
// POST /translate/map
// -----------------------------------------------------------------
// Body:
//
//	{
//	  "queries": {
//	    "greeting": { "text": "Hello" },
//	    "farewell": { "text": "Goodbye", "to": "ja" }
//	  },
//	  "options": { "to": "id" }
//	}
//
// 200: object with matching keys.

type mapTranslateRequest struct {
	Queries map[string]batchQueryItem `json:"queries"`
	Options translateOptions          `json:"options"`
}

func TranslateMap(defaultTLD string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req mapTranslateRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if len(req.Queries) == 0 {
			jsonError(w, "queries must not be empty", http.StatusBadRequest)
			return
		}

		queries := make(map[string]gt.Query, len(req.Queries))
		for k, q := range req.Queries {
			queries[k] = gt.Query{
				Text:        q.Text,
				From:        q.From,
				To:          q.To,
				ForceFrom:   q.ForceFrom,
				ForceTo:     q.ForceTo,
				AutoCorrect: q.AutoCorrect,
			}
		}

		results, err := gt.TranslateMap(queries, req.Options.toGT(defaultTLD))
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadGateway)
			return
		}

		out := make(map[string]*translationResponse, len(results))
		for k, res := range results {
			if res == nil {
				out[k] = nil
				continue
			}
			v := resultToResponse(res)
			out[k] = &v
		}
		jsonOK(w, out)
	}
}
