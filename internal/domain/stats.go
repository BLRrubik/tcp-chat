package domain

type ServerStats struct {
	ActiveConnections      int
	TotalMessagesProcessed int64
	UptimeSeconds          int64
	ErrorCount             int
}
