package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

const addr = ":8085"

type server struct {
	httpServer *http.Server
	logger     *slog.Logger
	service    *Service
	// Add any necessary fields for your server
}

// NewServer initializes a new server instance
func NewServer(logger *slog.Logger, service *Service) *server {
	return &server{
		service: service,
		logger:  logger,
	}
}

func (s *server) Start() {
	s.logger.Info("Starting server", slog.String("address", addr))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.healthCheck)
	mux.HandleFunc("POST /receipts", s.PrintReceiptHandler)

	s.httpServer = &http.Server{Addr: addr, Handler: mux}

	go func() {
		err := s.httpServer.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				s.logger.Info("Server closed gracefully")
				return
			}
			s.logger.Error("Failed to start server", slog.String("error", err.Error()))
		}
	}()

}

func (s *server) Shutdown() {
	s.logger.Info("Shutting down server")

	if err := s.httpServer.Close(); err != nil {
		s.logger.Error("Error shutting down server", slog.String("error", err.Error()))
	}
}

func (s *server) healthCheck(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Health check endpoint hit", slog.String("method", r.Method), slog.String("path", r.URL.Path))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *server) PrintReceiptHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Print receipt endpoint hit", slog.String("method", r.Method), slog.String("path", r.URL.Path))

	receipt := Receipt{}
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		s.logger.Error("Failed to decode receipt", slog.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call the service to handle the receipt printing
	printJob, err := s.service.PrintReceipt(receipt)
	if err != nil {
		s.logger.Error("Failed to print receipt", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(printJob); err != nil {
		s.logger.Error("Failed to encode print job response", slog.String("error", err.Error()))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
