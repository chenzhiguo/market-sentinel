package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Response structures
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func writeSuccessWithMeta(w http.ResponseWriter, data interface{}, total, limit, offset int) {
	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// Query parameter helpers
func queryInt(r *http.Request, key string, defaultVal int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func queryTime(r *http.Request, key string) time.Time {
	if v := r.URL.Query().Get(key); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
	}
	return time.Time{}
}

// Handlers
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, map[string]interface{}{
		"status":  "healthy",
		"version": "0.1.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleListNews(w http.ResponseWriter, r *http.Request) {
	since := queryTime(r, "since")
	until := queryTime(r, "until")
	source := r.URL.Query().Get("source")
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)

	if limit > 200 {
		limit = 200
	}

	items, total, err := s.store.ListNews(since, until, source, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	writeSuccessWithMeta(w, items, total, limit, offset)
}

func (s *Server) handleGetNews(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := s.store.GetNews(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if item == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "News item not found")
		return
	}
	writeSuccess(w, item)
}

func (s *Server) handleListAnalysis(w http.ResponseWriter, r *http.Request) {
	since := queryTime(r, "since")
	until := queryTime(r, "until")
	impact := r.URL.Query().Get("impact")
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)

	if limit > 200 {
		limit = 200
	}

	items, total, err := s.store.ListAnalysis(since, until, impact, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	writeSuccessWithMeta(w, items, total, limit, offset)
}

func (s *Server) handleGetAnalysis(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := s.store.GetAnalysis(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if item == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Analysis not found")
		return
	}
	writeSuccess(w, item)
}

func (s *Server) handleListReports(w http.ResponseWriter, r *http.Request) {
	reportType := r.URL.Query().Get("type")
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)

	if limit > 200 {
		limit = 200
	}

	items, total, err := s.store.ListReports(reportType, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	writeSuccessWithMeta(w, items, total, limit, offset)
}

func (s *Server) handleGetLatestReport(w http.ResponseWriter, r *http.Request) {
	report, err := s.store.GetLatestReport()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if report == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "No reports available")
		return
	}
	writeSuccess(w, report)
}

func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	report, err := s.store.GetReport(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if report == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Report not found")
		return
	}
	writeSuccess(w, report)
}

func (s *Server) handleGetStockSentiment(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	hours := queryInt(r, "hours", 24)

	sentiment, err := s.store.GetStockSentiment(symbol, hours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if sentiment == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "No sentiment data for this stock")
		return
	}
	writeSuccess(w, sentiment)
}

func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	limit := queryInt(r, "limit", 50)
	offset := queryInt(r, "offset", 0)

	if limit > 200 {
		limit = 200
	}

	items, total, err := s.store.ListAlerts(level, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	writeSuccessWithMeta(w, items, total, limit, offset)
}

func (s *Server) handleTriggerScan(w http.ResponseWriter, r *http.Request) {
	// TODO: Trigger actual scan
	writeSuccess(w, map[string]interface{}{
		"status":  "triggered",
		"message": "Scan has been triggered",
	})
}
