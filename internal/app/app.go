package app

import (
	"context"
	"fmt"
	"github.com/blrrubik/tcp-chat-server/internal/config"
	"github.com/blrrubik/tcp-chat-server/internal/hub"
	"github.com/blrrubik/tcp-chat-server/internal/logging"
	"github.com/blrrubik/tcp-chat-server/internal/server"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

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

func Run() {
	cfg := config.Parse()
	printStartupBanner(cfg)

	logger := logging.SetupLogging(cfg.LogLevel)
	h := hub.NewHub(logger, cfg.MessageHistorySize, cfg.MaxConnections)

	go h.Run()
	go server.StartHTTPMonitoring(h, cfg.MonitoringPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := setupSignalHandling()

	go func() {
		<-sigChan
		cancel()
	}()

	go func() {
		<-ctx.Done()
		logger.Info("shutdown signal received")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := h.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown error", "error", err)
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

func printStartupBanner(cfg config.ServerConfig) {
	port := strings.TrimLeft(cfg.Port, ":")
	monitoringPort := strings.TrimLeft(cfg.MonitoringPort, ":")
	fmt.Printf(banner, port, monitoringPort, cfg.MaxConnections, cfg.LogLevel, port)
}
