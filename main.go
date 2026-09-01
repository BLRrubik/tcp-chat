package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"tcp-chat/internal/server"
	"time"
)

func main() {
	cfg := parseCommandLineArgs()
	printStartupBanner(cfg)

	logger := server.SetupLogging(cfg.LogLevel)
	h := server.NewHub(logger, cfg.MessageHistorySize, cfg.MaxConnections)

	go h.Run()
	go startHTTPMonitoring(h, cfg.MonitoringPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := setupSignalHandling()

	go func() {
		<-sigChan
		cancel()
	}()

	go func() {
		<-ctx.Done()
		logger.Println("INFO Shutdown signal received")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := h.Shutdown(shutdownCtx); err != nil {
			logger.Printf("ERROR shutdown: %v", err)
		}

		os.Exit(0)
	}()

	server.StartEchoServer(cfg.Port, h)
}

func setupSignalHandling() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	return sigChan
}

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

func startHTTPMonitoring(hub *server.Hub, port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealthEndpoint(hub))
	mux.HandleFunc("/stats", handleStatsEndpoint(hub))

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}

func handleHealthEndpoint(hub *server.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := hub.GetStats()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(healthResponse{
			Status:            "healthy",
			ActiveConnections: stats.ActiveConnections,
			UptimeSeconds:     stats.UptimeSeconds,
		})
	}
}

func handleStatsEndpoint(hub *server.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := hub.GetStats()

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

type ServerConfig struct {
	Port               string
	MonitoringPort     string
	MaxConnections     int
	LogLevel           string
	MessageHistorySize int
}

func parseCommandLineArgs() ServerConfig {
	var (
		port               string
		monitoringPort     string
		logLevel           string
		maxConnections     int
		messageHistorySize int
	)

	flag.StringVar(&port, "port", ":8080", "Port to listen on")
	flag.StringVar(&monitoringPort, "monitoring-port", ":9090", "HTTP monitoring port")
	flag.StringVar(&logLevel, "log-level", "info", "Log level")
	flag.IntVar(&maxConnections, "max-connections", 10, "Maximum number of concurrent connections")
	flag.IntVar(&messageHistorySize, "message-history-size", 100, "Message history size")

	flag.Parse()

	return ServerConfig{
		Port:               port,
		MonitoringPort:     monitoringPort,
		MaxConnections:     maxConnections,
		LogLevel:           logLevel,
		MessageHistorySize: messageHistorySize,
	}
}

const banner = `
╔══════════════════════════════════════╗
║         TCP Chat Server              ║
╚══════════════════════════════════════╝
Port:            %s
Monitoring Port: %s
Max Connections: %d
Log Level:       %s
Connect using: telnet <ip> %s
`

func printStartupBanner(config ServerConfig) {
	port := strings.TrimLeft(config.Port, ":")
	monitoringPort := strings.TrimLeft(config.MonitoringPort, ":")
	fmt.Printf(banner, port, monitoringPort, config.MaxConnections, config.LogLevel, port)
}
