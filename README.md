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
  "original": "Hello, world!",
  "text": "Halo Dunia!",
  "pronunciation": null,
  "from": {
    "language": { "iso": "en", "didYouMean": false },
    "text":     { "autoCorrected": false, "value": "", "didYouMean": false }
  }
}
```

**Failure responses:**

| Status | Condition | Error message |
|---|---|---|
| `400` | Body bukan JSON valid | `invalid request body: ...` |
| `400` | Field `text` kosong atau tidak ada | `text is required` |
| `400` | Field tidak dikenal di body (unknown field) | `invalid request body: json: unknown field "..."` |
| `502` | Kode bahasa `from` atau `to` tidak dikenali dan `forceFrom`/`forceTo` tidak di-set | `from language "xx" is not supported; ...` |
| `502` | Google Translate mengembalikan status HTTP non-200 | `batchTranslate: server returned 429 Too Many Requests ...` |
| `502` | Koneksi ke Google Translate gagal (network error, timeout) | `batchTranslate: http: ...` |
| `429` | Terlalu banyak request dari IP yang sama | `rate limit exceeded` |
| `500` | Panic tak terduga di dalam handler | `internal server error` |

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

```json
[
  {
    "original": "Hello",
    "text": "Halo",
    "from": {
      "language": { "iso": "en", "didYouMean": false },
      "text": { "autoCorrected": false, "value": "", "didYouMean": false }
    }
  },
  {
    "original": "Good morning",
    "text": "おはようございます",
    "from": {
      "language": { "iso": "en", "didYouMean": false },
      "text": { "autoCorrected": false, "value": "", "didYouMean": false }
    }
  }
]
```

> **`from.text.value`** hanya terisi ketika Google mendeteksi typo pada teks sumber. Kosong `""` adalah kondisi normal. Lihat bagian [Translation response fields](#translation-response-fields) untuk penjelasan lengkap.

**Failure responses:**

| Status | Condition | Error message |
|---|---|---|
| `400` | Body bukan JSON valid | `invalid request body: ...` |
| `400` | Array `queries` kosong atau tidak ada | `queries must not be empty` |
| `400` | Field tidak dikenal di body | `invalid request body: json: unknown field "..."` |
| `502` | Kode bahasa tidak dikenali di salah satu query (global atau per-item) | `from/to language "xx" is not supported; ...` |
| `502` | Salah satu item ditolak Google dan `rejectOnPartialFail: true` (default) | `batchTranslate: partial failure at index N — item was rejected by the server ...` |
| `502` | Google Translate mengembalikan status HTTP non-200 | `batchTranslate: server returned 429 Too Many Requests ...` |
| `502` | Koneksi ke Google Translate gagal | `batchTranslate: http: ...` |
| `429` | Rate limit terlampaui | `rate limit exceeded` |
| `500` | Panic tak terduga | `internal server error` |

> **Partial failure:** Jika `rejectOnPartialFail` di-set `false`, response tetap `200` tetapi item yang gagal menjadi `null` di dalam array hasil. Tidak ada error HTTP yang dikembalikan.

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
  "greeting": {
    "original": "Hello",
    "text": "Halo",
    "from": {
      "language": { "iso": "en", "didYouMean": false },
      "text": { "autoCorrected": false, "value": "", "didYouMean": false }
    }
  },
  "farewell": {
    "original": "Goodbye",
    "text": "さようなら",
    "from": {
      "language": { "iso": "en", "didYouMean": false },
      "text": { "autoCorrected": false, "value": "", "didYouMean": false }
    }
  }
}
```

> **Catatan:** Urutan key di response tidak dijamin sama dengan urutan di request karena Go `map` tidak memiliki urutan. Gunakan key name untuk mengambil hasilnya, bukan posisi.

> **`from.text.value`** hanya terisi ketika Google mendeteksi typo pada teks sumber. Kosong `""` adalah kondisi normal. Lihat bagian [Translation response fields](#translation-response-fields) untuk penjelasan lengkap.

**Failure responses:**

| Status | Condition | Error message |
|---|---|---|
| `400` | Body bukan JSON valid | `invalid request body: ...` |
| `400` | Object `queries` kosong atau tidak ada | `queries must not be empty` |
| `400` | Field tidak dikenal di body | `invalid request body: json: unknown field "..."` |
| `502` | Kode bahasa tidak dikenali di salah satu entry | `from/to language "xx" is not supported; ...` |
| `502` | Salah satu entry ditolak Google dan `rejectOnPartialFail: true` (default) | `batchTranslate: partial failure at index N — item was rejected by the server ...` |
| `502` | Google Translate mengembalikan status HTTP non-200 | `batchTranslate: server returned 429 Too Many Requests ...` |
| `502` | Koneksi ke Google Translate gagal | `batchTranslate: http: ...` |
| `429` | Rate limit terlampaui | `rate limit exceeded` |
| `500` | Panic tak terduga | `internal server error` |

> **Partial failure:** Sama seperti `/translate/batch` — jika `rejectOnPartialFail: false`, response tetap `200` dengan value `null` untuk key yang gagal.

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

## Translation response fields

Setiap objek hasil terjemahan dari `/translate`, `/translate/batch`, dan `/translate/map` memiliki field berikut:

| Field | Type | Deskripsi |
|---|---|---|
| `original` | string | Teks asli yang dikirim dalam request, selalu terisi |
| `text` | string | Hasil terjemahan |
| `pronunciation` | string \| null | Panduan pelafalan / romanisasi (jika tersedia, selain itu tidak muncul) |
| `from.language.iso` | string | Kode ISO bahasa sumber yang terdeteksi oleh Google |
| `from.language.didYouMean` | bool | `true` jika bahasa yang terdeteksi berbeda dari nilai `from` yang dikirim |
| `from.text.autoCorrected` | bool | `true` jika Google **diam-diam mengoreksi** typo dan menerjemahkan versi yang sudah diperbaiki (hanya terjadi bila `autoCorrect: true` dikirim) |
| `from.text.didYouMean` | bool | `true` jika Google **mendeteksi typo** namun tidak mengoreksinya secara otomatis — koreksi hanya disarankan |
| `from.text.value` | string | Teks sumber beserta saran koreksi dari Google. **Kosong `""` jika tidak ada typo.** Jika ada typo, kata yang disarankan untuk diganti dibungkus dengan `[...]`, contoh: `"I [have] a [dream]"`. Terisi ketika `autoCorrected: true` **atau** `didYouMean: true` |

### Ilustrasi perilaku `from.text`

**Tidak ada typo** — `value` selalu kosong:
```json
{
  "original": "Hello world",
  "text": "Halo dunia",
  "from": {
    "language": { "iso": "en", "didYouMean": false },
    "text": { "autoCorrected": false, "value": "", "didYouMean": false }
  }
}
```

**Ada typo, `autoCorrect: false` (default)** — Google menyarankan koreksi tapi terjemahan tetap dari teks asli yang salah:
```json
{
  "original": "I havv a dreeem",
  "text": "Aku punya mimpi",
  "from": {
    "language": { "iso": "en", "didYouMean": false },
    "text": { "autoCorrected": false, "value": "I [have] a [dream]", "didYouMean": true }
  }
}
```

**Ada typo, `autoCorrect: true`** — Google mengoreksi teks sebelum menerjemahkan:
```json
{
  "original": "I havv a dreeem",
  "text": "Saya punya mimpi",
  "from": {
    "language": { "iso": "en", "didYouMean": false },
    "text": { "autoCorrected": true, "value": "I [have] a [dream]", "didYouMean": false }
  }
}
```

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
