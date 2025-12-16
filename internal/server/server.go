package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/mdgen/internal/config"
	"github.com/example/mdgen/internal/note"
)

type Server struct {
	cfg    config.Config
	logger *slog.Logger
	mux    *http.ServeMux
}

func New(cfg config.Config, logger *slog.Logger) *Server {
	s := &Server{cfg: cfg, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.wrap(s.health))
	mux.HandleFunc("/version", s.wrap(s.version))
	mux.HandleFunc("/api/note", s.wrap(s.authMiddleware(s.createNote)))
	s.mux = mux
	return s
}

func (s *Server) wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := randomID()
		w.Header().Set("X-Request-ID", reqID)

		next(w, r)

		s.logger.Info("request",
			"request_id", reqID,
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) version(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{"version": "0.1.0"})
}

func (s *Server) createNote(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxBodySize)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "READ_ERROR", "failed to read body", nil)
		return
	}

	var req note.NoteRequest
	if err := json.Unmarshal(data, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json", nil)
		return
	}

	if err := validateNoteRequest(req); err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), map[string]string{"field": "title"})
		return
	}

	rendered := note.RenderMarkdown(req, time.Now())
	saved, err := note.WriteFile(s.cfg.OutputDir, req.Folder, rendered)
	if err != nil {
		status := http.StatusInternalServerError
		code := "WRITE_ERROR"
		if errors.Is(err, note.ErrInvalidFolder) {
			status = http.StatusBadRequest
			code = "VALIDATION_ERROR"
		}
		respondError(w, status, code, err.Error(), nil)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"filename": saved.FileName,
		"path":     saved.FullPath,
		"status":   "saved",
		"took_ms":  time.Since(start).Milliseconds(),
	})
}

func validateNoteRequest(req note.NoteRequest) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(req.Title) > 200 {
		return fmt.Errorf("title too long")
	}
	return nil
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	if s.cfg.APIKey == "" {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != s.cfg.APIKey {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid api key", nil)
			return
		}
		next(w, r)
	}
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, code, message string, details any) {
	respondJSON(w, status, map[string]any{
		"ok": false,
		"error": map[string]any{
			"code":    code,
			"message": message,
			"details": details,
		},
	})
}

func randomID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err == nil {
		return hex.EncodeToString(b)
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
