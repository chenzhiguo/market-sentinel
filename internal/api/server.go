package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

type Server struct {
	cfg    *config.Config
	store  *storage.Storage
	router *chi.Mux
	http   *http.Server
}

func NewServer(cfg *config.Config, store *storage.Storage) *Server {
	s := &Server{
		cfg:   cfg,
		store: store,
	}
	s.setupRouter()
	return s
}

func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware)

	// Health check (no auth)
	r.Get("/api/v1/health", s.handleHealth)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)

		// News
		r.Get("/api/v1/news", s.handleListNews)
		r.Get("/api/v1/news/{id}", s.handleGetNews)

		// Analysis
		r.Get("/api/v1/analysis", s.handleListAnalysis)
		r.Get("/api/v1/analysis/{id}", s.handleGetAnalysis)

		// Reports
		r.Get("/api/v1/reports", s.handleListReports)
		r.Get("/api/v1/reports/latest", s.handleGetLatestReport)
		r.Get("/api/v1/reports/{id}", s.handleGetReport)

		// Stocks
		r.Get("/api/v1/stocks/{symbol}/sentiment", s.handleGetStockSentiment)

		// Alerts
		r.Get("/api/v1/alerts", s.handleListAlerts)

		// Manual scan trigger
		r.Post("/api/v1/scan", s.handleTriggerScan)
	})

	s.router = r
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.http = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
	}

	log.Printf("Starting API server on %s", addr)
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
