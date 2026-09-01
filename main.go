package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcp-chat/internal/server"
	"time"
)

func main() {
	logger := SetupLogging("info")
	h := server.NewHub(logger)

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

	server.StartEchoServer(":8010", h)
}

func setupSignalHandling() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	return sigChan
}

func SetupLogging(level string) *log.Logger {
	return log.New(os.Stdout, "[TCP-CHAT] ", log.Ldate|log.Ltime)
}
