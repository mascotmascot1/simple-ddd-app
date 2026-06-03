package cfg

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	envHTTPHost            = "HTTP_HOST"
	envHTTPPort            = "HTTP_PORT"
	envHTTPShutdownTimeout = "HTTP_SHUTDOWN_TIMEOUT" // UNNECESSARY — in seconds

	envOrdersHTTPMaxUploadSize = "ORDERS_HTTP_MAXUPLOADSIZE" // UNNECESSARY

	envPostgresHost     = "POSTGRES_HOST"
	envPostgresPort     = "POSTGRES_PORT"
	envPostgresUser     = "POSTGRES_USER"
	envPostgresPassword = "POSTGRES_PASSWORD"
	envPostgresDatabase = "POSTGRES_DATABASE"
	envPostgresSSLMode  = "POSTGRES_SSLMODE"
)

const (
	defaultHTTPShutdownTimeout     time.Duration = 10   // seconds
	defaultOrdersHTTPMaxUploadSize int64         = 1000 // KB
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

	timeout := defaultHTTPShutdownTimeout * time.Second
	eSTimeout := os.Getenv(envHTTPShutdownTimeout)
	if eSTimeout != "" {
		st, err := strconv.ParseInt(eSTimeout, 10, 64)
		if err != nil {
			logger.Printf("[WARNING] %q (%d sec) for http server was applied: %v\n",
				envHTTPShutdownTimeout, defaultHTTPShutdownTimeout, err)
		} else {
			timeout = time.Duration(st) * time.Second
		}
	}

	return Server{
		Host:            eHost,
		Port:            ePort,
		ShutdownTimeout: timeout,
	}, nil
}

type Postgres struct {
	host     string
	port     string
	user     string
	password string
	database string
	sslMode  string
}

func newPostgres() (Postgres, error) {
	eHost := os.Getenv(envPostgresHost)
	if eHost == "" {
		return Postgres{}, fmt.Errorf("empty postgres host env")
	}

	ePort := os.Getenv(envPostgresPort)
	if ePort == "" {
		return Postgres{}, fmt.Errorf("empty postgres port env")
	}

	eUser := os.Getenv(envPostgresUser)
	if eUser == "" {
		return Postgres{}, fmt.Errorf("empty postgres user env")
	}

	ePassword := os.Getenv(envPostgresPassword)
	if ePassword == "" {
		return Postgres{}, fmt.Errorf("empty postgres password env")
	}

	eDatabase := os.Getenv(envPostgresDatabase)
	if eDatabase == "" {
		return Postgres{}, fmt.Errorf("empty postgres database env")
	}

	eSSLMode := os.Getenv(envPostgresSSLMode)
	if eSSLMode == "" {
		return Postgres{}, fmt.Errorf("empty postgres ssl mode env")
	}

	return Postgres{
		host:     eHost,
		port:     ePort,
		user:     eUser,
		password: ePassword,
		database: eDatabase,
		sslMode:  eSSLMode,
	}, nil
}

func (p Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.user, p.password, p.host, p.port, p.database, p.sslMode)
}

type Orders struct {
	MaxUploadSize int64
}

func newOrders(logger *log.Logger) Orders {
	uploadSize := defaultOrdersHTTPMaxUploadSize << 10

	eMaxUploadSize := os.Getenv(envOrdersHTTPMaxUploadSize)
	if eMaxUploadSize != "" {
		mus, err := strconv.ParseInt(eMaxUploadSize, 10, 64)
		if err != nil {
			logger.Printf("[WARNING] standard %q (%d KB) for http server was applied: %v\n",
				envOrdersHTTPMaxUploadSize, defaultOrdersHTTPMaxUploadSize, err)
		} else {
			uploadSize = mus << 10
		}
	}

	return Orders{
		MaxUploadSize: uploadSize,
	}
}

type Config struct {
	Server   Server
	Postgres Postgres
	Orders   Orders
}

func New(logger *log.Logger) (*Config, error) {
	server, err := newServer(logger)
	if err != nil {
		return nil, err
	}
	postgres, err := newPostgres()
	if err != nil {
		return nil, err
	}
	orders := newOrders(logger)

	return &Config{
		Server:   server,
		Postgres: postgres,
		Orders:   orders,
	}, nil
}
