package config

import (
	"flag"
	"fmt"
	"os"
)

type ServerConfig struct {
	Port               string
	MonitoringPort     string
	MaxConnections     int
	LogLevel           string
	MessageHistorySize int
}

func Parse() ServerConfig {
	flag.Usage = printUsage

	var (
		port               string
		monitoringPort     string
		logLevel           string
		maxConnections     int
		messageHistorySize int
	)

	flag.StringVar(&port, "port", ":8080", "Port to listen on")
	flag.StringVar(&monitoringPort, "monitoring-port", ":9090", "HTTP monitoring port")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (error, warn, info)")
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

func printUsage() {
	fmt.Fprintln(os.Stderr, "TCP Chat Server — production-ready TCP chat with HTTP monitoring.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  server [flags]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Flags:")
	flag.PrintDefaults()
}
