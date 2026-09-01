package server

import (
	"encoding/json"
	"github.com/blrrubik/tcp-chat-server/internal/hub"
	"log"
	"net/http"
)

type healthResponse struct {
	Status            string `json:"status"`
	ActiveConnections int    `json:"active_connections"`
	UptimeSeconds     int64  `json:"uptime_seconds"`
}

type statsResponse struct {
	ActiveConnections      int   `json:"active_connections"`
	TotalMessagesProcessed int64 `json:"total_messages_processed"`
	UptimeSeconds          int64 `json:"uptime_seconds"`
	ErrorCount             int   `json:"error_count"`
}

func StartHTTPMonitoring(h *hub.Hub, port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealthEndpoint(h))
	mux.HandleFunc("/stats", handleStatsEndpoint(h))

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}

func handleHealthEndpoint(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := h.GetStats()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(healthResponse{
			Status:            "healthy",
			ActiveConnections: stats.ActiveConnections,
			UptimeSeconds:     stats.UptimeSeconds,
		})
	}
}

func handleStatsEndpoint(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := h.GetStats()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(statsResponse{
			ActiveConnections:      stats.ActiveConnections,
			TotalMessagesProcessed: stats.TotalMessagesProcessed,
			UptimeSeconds:          stats.UptimeSeconds,
			ErrorCount:             stats.ErrorCount,
		})
	}
}
