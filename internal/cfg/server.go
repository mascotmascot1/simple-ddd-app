package cfg

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	// -------------ENVS----------------------------
	envHTTPHost = "HTTP_HOST"
	envHTTPPort = "HTTP_PORT"

	envHTTPShutdownTimeout = "HTTP_SHUTDOWN_TIMEOUT" // UNNECESSARY — in seconds
	envHTTPIdleTimeout     = "HTTP_IDLE_TIMEOUT"     // UNNECESSARY — in seconds
	envHTTPWriteTimeout    = "HTTP_WRITE_TIMEOUT"    // UNNECESSARY — in seconds
	envHTTPReadTimeout     = "HTTP_READ_TIMEOUT"     // UNNECESSARY — in seconds
	envHTTPRequestTimeout  = "HTTP_REQUEST_TIMEOUT"  // UNNECESSARY — in seconds

	// ------------default-values------------------
	defaultHTTPShutdownTimeout = 10 * time.Second
	defaultHTTPReadTimeout     = 5 * time.Second
	defaultHTTPWriteTimeout    = 15 * time.Second
	defaultHTTPIdleTimeout     = 15 * time.Second
	defaultHTTPRequestTimeout  = 10 * time.Second
)

// в мнч лучше или ед.ч
type Timeouts struct {
	Shutdown time.Duration
	Read     time.Duration
	Write    time.Duration
	Idle     time.Duration
	Request  time.Duration
}

func newTimeouts(logger *log.Logger) Timeouts {
	return Timeouts{
		Shutdown: parseEnvDuration(envHTTPShutdownTimeout, defaultHTTPShutdownTimeout, logger),
		Read:     parseEnvDuration(envHTTPReadTimeout, defaultHTTPReadTimeout, logger),
		Write:    parseEnvDuration(envHTTPWriteTimeout, defaultHTTPWriteTimeout, logger),
		Idle:     parseEnvDuration(envHTTPIdleTimeout, defaultHTTPIdleTimeout, logger),
		Request:  parseEnvDuration(envHTTPRequestTimeout, defaultHTTPRequestTimeout, logger),
	}
}

type Server struct {
	Host     string
	Port     string
	Timeouts Timeouts
}

func newServer(logger *log.Logger) (Server, error) {
	eHost := os.Getenv(envHTTPHost)
	if eHost == "" {
		return Server{}, fmt.Errorf("empty http host env")
	}

	ePort := os.Getenv(envHTTPPort)
	if ePort == "" {
		return Server{}, fmt.Errorf("empty http port env")
	}

	return Server{
		Host:     eHost,
		Port:     ePort,
		Timeouts: newTimeouts(logger),
	}, nil
}
