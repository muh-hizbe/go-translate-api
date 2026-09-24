package main

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration, populated from environment variables.
type Config struct {
	// Port the HTTP server listens on. Env: PORT (default: "8080")
	Port string

	// RequestTimeout is the max duration for an outbound Google Translate call.
	// Env: REQUEST_TIMEOUT_MS in milliseconds (default: 10000)
	RequestTimeout time.Duration

	// RateLimit is the maximum number of requests per second accepted from a
	// single IP address. Env: RATE_LIMIT (default: 10)
	RateLimit int

	// RateBurst is the burst capacity on top of the per-second rate.
	// Env: RATE_BURST (default: 20)
	RateBurst int

	// CORSAllowOrigin sets the Access-Control-Allow-Origin header.
	// Env: CORS_ALLOW_ORIGIN (default: "*")
	CORSAllowOrigin string

	// DefaultTLD is the Google Translate domain suffix used when the request
	// body does not specify one. Env: DEFAULT_TLD (default: "com")
	DefaultTLD string
}

// loadConfig reads Config from environment variables, falling back to defaults.
func loadConfig() Config {
	c := Config{
		Port:            envStr("PORT", "8080"),
		RequestTimeout:  time.Duration(envInt("REQUEST_TIMEOUT_MS", 10000)) * time.Millisecond,
		RateLimit:       envInt("RATE_LIMIT", 10),
		RateBurst:       envInt("RATE_BURST", 20),
		CORSAllowOrigin: envStr("CORS_ALLOW_ORIGIN", "*"),
		DefaultTLD:      envStr("DEFAULT_TLD", "com"),
	}
	return c
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
