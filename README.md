# go-translate-api

REST API server built in Go that wraps [google-translate-api-go](https://muh-hizbe/google-translate-api-go).  
No Google API key required — uses the same internal endpoint as translate.google.com.

---

## Running

```bash
go run .
# Server starts on :8080 by default
```

Or build a binary:

```bash
go build -o go-translate-api .
./go-translate-api
```

---

## Configuration (environment variables)

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `REQUEST_TIMEOUT_MS` | `10000` | Max ms for outbound Google Translate calls |
| `RATE_LIMIT` | `10` | Max requests/second per IP |
| `RATE_BURST` | `20` | Burst capacity on top of rate limit |
| `CORS_ALLOW_ORIGIN` | `*` | `Access-Control-Allow-Origin` header value |
| `DEFAULT_TLD` | `com` | Google Translate domain suffix (e.g. `co.id`) |

Example:

```bash
PORT=3000 DEFAULT_TLD=co.id go run .
```

---

## Endpoints

### `GET /health`

```json
{ "status": "ok" }
```

---

### `POST /translate`

Translate a single text string.

**Request body:**

```json
{
  "text": "Hello, world!",
  "options": {
    "from": "auto",
    "to": "id",
    "tld": "com"
  }
}
```

**Response `200`:**

```json
{
  "text": "Halo Dunia!",
  "pronunciation": null,
  "from": {
    "language": { "iso": "en", "didYouMean": false },
    "text":     { "autoCorrected": false, "value": "", "didYouMean": false }
  }
}
```

---

### `POST /translate/batch`

Translate multiple texts in one request.

**Request body:**

```json
{
  "queries": [
    { "text": "Hello" },
    { "text": "Good morning", "to": "ja" }
  ],
  "options": { "to": "id" }
}
```

**Response `200`:** Array of translation objects in the same order.  
Failed items are `null` when `rejectOnPartialFail` is `false`.

---

### `POST /translate/map`

Translate a keyed map of texts.

**Request body:**

```json
{
  "queries": {
    "greeting": { "text": "Hello" },
    "farewell":  { "text": "Goodbye", "to": "ja" }
  },
  "options": { "to": "id" }
}
```

**Response `200`:** Object with matching keys.

```json
{
  "greeting": { "text": "Halo", ... },
  "farewell":  { "text": "さようなら", ... }
}
```

---

### `POST /speak`

Get a Base64-encoded MP3 of the text spoken aloud.  
Input must be ≤ 200 characters.

**Request body:**

```json
{
  "text": "Hello",
  "options": { "to": "en" }
}
```

**Response `200`:**

```json
{ "audio": "<base64-encoded mp3>" }
```

Decode and play the audio:

```bash
echo "<base64>" | base64 -d > hello.mp3
```

---

### `POST /speak/batch`

Get TTS audio for multiple texts.

**Request body:**

```json
{
  "queries": [
    { "text": "Hello", "to": "en" },
    { "text": "Halo",  "to": "id" }
  ],
  "options": {}
}
```

**Response `200`:** Array of `{ "audio": "..." }` objects.

---

### `GET /languages`

Returns all supported languages sorted by code.

Optional query param `?q=<search>` filters by code or name (case-insensitive).

```
GET /languages
GET /languages?q=indo
```

**Response `200`:**

```json
[
  { "code": "id", "name": "Indonesian" },
  ...
]
```

---

### `GET /languages/{code}`

Look up a single language by its ISO code or English name.

```
GET /languages/id
GET /languages/Indonesian
```

**Response `200`:**

```json
{ "code": "id", "name": "Indonesian" }
```

**Response `404`:**

```json
{ "error": "language not found" }
```

---

## Options object (all endpoints)

| Field | Type | Default | Description |
|---|---|---|---|
| `from` | string | `"auto"` | Source language code or name |
| `to` | string | `"en"` | Target language code or name |
| `forceFrom` | bool | `false` | Skip source language validation |
| `forceTo` | bool | `false` | Skip target language validation |
| `autoCorrect` | bool | `false` | Use Google's auto-corrected source text |
| `tld` | string | server default | Google domain suffix (overrides `DEFAULT_TLD`) |
| `forceBatch` | bool | `true` | Always use the batch endpoint |
| `fallbackBatch` | bool | `true` | Fall back to batch if single endpoint fails |
| `rejectOnPartialFail` | bool | `true` | Fail entire request if any batch item fails |

---

## Error responses

All error responses follow the same shape:

```json
{ "error": "<message>" }
```

| Status | Meaning |
|---|---|
| `400` | Invalid JSON body or missing required field |
| `429` | Rate limit exceeded |
| `502` | Google Translate returned an error |
| `500` | Unexpected server error |

---

## Project structure

```
go-translate-api/
├── main.go         — server setup, routing, graceful shutdown
├── config.go       — env-based configuration
├── middleware.go   — logging, recovery, CORS, per-IP rate limiter
├── response.go     — shared JSON helpers (main package level)
├── go.mod
└── handlers/
    ├── helpers.go      — JSON encode/decode utilities
    ├── translate.go    — /translate, /translate/batch, /translate/map
    ├── speak.go        — /speak, /speak/batch
    └── languages.go    — /languages, /languages/{code}
```
