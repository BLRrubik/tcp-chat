package main

import (
	"context"
	"flag"
	"fmt"
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

type ServerConfig struct {
	Port               string
	MaxConnections     int
	LogLevel           string
	MessageHistorySize int
}

func parseCommandLineArgs() ServerConfig {
	port := flag.String("port", ":8080", "Port to listen on")
	logLevel := flag.String("log-level", "info", "Log level")
	maxConnections := flag.Int("max-connections", 10, "Maximum number of concurrent connections")
	messageHistorySize := flag.Int("message-history-size", 100, "Message history size")

	flag.Parse()

	return ServerConfig{
		Port:               *port,
		MaxConnections:     *maxConnections,
		LogLevel:           *logLevel,
		MessageHistorySize: *messageHistorySize,
	}
}

const banner = `
╔══════════════════════════════════════╗
║         TCP Chat Server              ║
╚══════════════════════════════════════╝
Port:            %s
Max Connections: %d
Log Level:       %s
Connect using: telnet <ip> %s
`

func printStartupBanner(config ServerConfig) {
	port := strings.TrimLeft(config.Port, ":")
	fmt.Printf(banner, port, config.MaxConnections, config.LogLevel, port)
}
