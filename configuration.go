package main

import "time"

const (
	version                = "1.5.7"
	agentName              = "muffet"
	concurrency            = 4096
	tcpTimeout             = time.Minute
	defaultBufferSize      = 4096
	defaultMaxConnections  = 512
	defaultMaxRedirections = 64
	defaultHTTPTimeout     = 10 * time.Second
)
