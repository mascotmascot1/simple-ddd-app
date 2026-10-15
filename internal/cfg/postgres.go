package cfg

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	// -------------ENVS----------------------------
	envPostgresDialTimeout = "POSTGRES_DIAL_TIMEOUT" // UNNECESSARY — in seconds

	// ------------dsn---------------
	envPostgresHost     = "POSTGRES_HOST"
	envPostgresPort     = "POSTGRES_PORT"
	envPostgresUser     = "POSTGRES_USER"
	envPostgresPassword = "POSTGRES_PASSWORD"
	envPostgresDatabase = "POSTGRES_DATABASE"
	envPostgresSSLMode  = "POSTGRES_SSLMODE"

	// ------------Pool---------------
	envPostgresMaxConns = "POSTGRES_MAX_CONNS" // UNNECESSARY
	envPostgresMinConns = "POSTGRES_MIN_CONNS" // UNNECESSARY

	// ------------default-values--------------------
	defaultPostgresDialTimeout       = 10 * time.Second
	defaultPostgresMaxConns    int32 = 20
	defaultPostgresMinConns    int32 = 3
)

type dsn struct {
	host     string
	port     string
	user     string
	password string
	database string
	sslMode  string
}

func newDSN() (dsn, error) {
	eHost := os.Getenv(envPostgresHost)
	if eHost == "" {
		return dsn{}, fmt.Errorf("empty postgres host env")
	}

	ePort := os.Getenv(envPostgresPort)
	if ePort == "" {
		return dsn{}, fmt.Errorf("empty postgres port env")
	}

	eUser := os.Getenv(envPostgresUser)
	if eUser == "" {
		return dsn{}, fmt.Errorf("empty postgres user env")
	}

	ePassword := os.Getenv(envPostgresPassword)
	if ePassword == "" {
		return dsn{}, fmt.Errorf("empty postgres password env")
	}

	eDatabase := os.Getenv(envPostgresDatabase)
	if eDatabase == "" {
		return dsn{}, fmt.Errorf("empty postgres database env")
	}

	eSSLMode := os.Getenv(envPostgresSSLMode)
	if eSSLMode == "" {
		return dsn{}, fmt.Errorf("empty postgres ssl mode env")
	}

	return dsn{
		host:     eHost,
		port:     ePort,
		user:     eUser,
		password: ePassword,
		database: eDatabase,
		sslMode:  eSSLMode,
	}, nil
}

func (d dsn) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.user, d.password, d.host, d.port, d.database, d.sslMode)
}

type Pool struct {
	MinConns int32
	MaxConns int32
}

func newPool(logger *log.Logger) Pool {
	return Pool{
		MinConns: parseEnvInt32(envPostgresMinConns, defaultPostgresMinConns, logger),
		MaxConns: parseEnvInt32(envPostgresMaxConns, defaultPostgresMaxConns, logger),
	}
}

type Postgres struct {
	dsn
	Pool        Pool
	DialTimeout time.Duration
}

func newPostgres(logger *log.Logger) (Postgres, error) {
	dsn, err := newDSN()
	if err != nil {
		return Postgres{}, fmt.Errorf("create dsn: %w", err)
	}
	return Postgres{
		dsn:         dsn,
		Pool:        newPool(logger),
		DialTimeout: parseEnvDuration(envPostgresDialTimeout, defaultPostgresDialTimeout, logger),
	}, nil
}
