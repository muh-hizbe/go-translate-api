package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-translate-api/handlers"
)

func main() {
	cfg := loadConfig()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("starting go-translate-api on :%s", cfg.Port)

	rl := newRateLimiter(cfg.RateLimit, cfg.RateBurst)

	mux := http.NewServeMux()
	registerRoutes(mux, cfg)

	// Apply global middleware: recovery → CORS → rate-limit → logging
	handler := chain(
		mux,
		loggingMiddleware,
		rateLimitMiddleware(rl),
		corsMiddleware(cfg.CORSAllowOrigin),
		recoveryMiddleware,
	)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: cfg.RequestTimeout + 5*time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	log.Printf("server ready — press Ctrl+C to stop")
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}

func registerRoutes(mux *http.ServeMux, cfg Config) {
	tld := cfg.DefaultTLD

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Translation endpoints
	mux.HandleFunc("POST /translate", handlers.Translate(tld))
	mux.HandleFunc("POST /translate/batch", handlers.TranslateBatch(tld))
	mux.HandleFunc("POST /translate/map", handlers.TranslateMap(tld))

	// TTS endpoints
	mux.HandleFunc("POST /speak", handlers.Speak(tld))
	mux.HandleFunc("POST /speak/batch", handlers.SpeakBatch(tld))

	// Language utility endpoints
	mux.HandleFunc("GET /languages", handlers.Languages())
	mux.HandleFunc("GET /languages/", handlers.LanguageByCode())
}
