package cfg

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	// -------------ENVS----------------------------
	envHTTPHost            = "HTTP_HOST"
	envHTTPPort            = "HTTP_PORT"
	envHTTPShutdownTimeout = "HTTP_SHUTDOWN_TIMEOUT" // UNNECESSARY — in seconds

	// ------------default-values------------------
	defaultHTTPShutdownTimeout time.Duration = 10 // seconds
)

type Server struct {
	Host            string
	Port            string
	ShutdownTimeout time.Duration
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

	stTimeout := defaultHTTPShutdownTimeout * time.Second
	eSTimeout := os.Getenv(envHTTPShutdownTimeout)
	if eSTimeout != "" {
		st, err := strconv.ParseInt(eSTimeout, 10, 64)
		if err != nil {
			logger.Printf("[WARNING] invalid %q, default (%d sec) for http server was applied: %v\n",
				envHTTPShutdownTimeout, defaultHTTPShutdownTimeout, err)
		} else {
			stTimeout = time.Duration(st) * time.Second
		}
	}

	return Server{
		Host:            eHost,
		Port:            ePort,
		ShutdownTimeout: stTimeout,
	}, nil
}
