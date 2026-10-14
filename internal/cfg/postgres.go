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
	envPostgresDialTimeout = "POSTGRES_DIAL_TIMEOUT" // UNNECESSARY

	// ------------dsn-info---------------
	envPostgresHost     = "POSTGRES_HOST"
	envPostgresPort     = "POSTGRES_PORT"
	envPostgresUser     = "POSTGRES_USER"
	envPostgresPassword = "POSTGRES_PASSWORD"
	envPostgresDatabase = "POSTGRES_DATABASE"
	envPostgresSSLMode  = "POSTGRES_SSLMODE"

	// ------------pool-info---------------
	envPostgresMaxConns = "POSTGRES_MAX_CONNS" // UNNECESSARY
	envPostgresMinConns = "POSTGRES_MIN_CONNS" // UNNECESSARY

	// ------------default-values--------------------
	defaultPostgresDialTimeout time.Duration = 10 // seconds
	defaultPostgresMaxConns    int32         = 20
	defaultPostgresMinConns    int32         = 4
)

type dsn struct {
	host     string
	port     string
	user     string
	password string
	database string
	sslMode  string
}

func newDSNInfo() (dsn, error) {
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

// хз оставить  poolInfo и dsnInfo или убрать 'Info'
type Pool struct {
	MinConns int32
	MaxConns int32
}

func newPoolInfo(logger *log.Logger) Pool {
	var (
		minConns  = defaultPostgresMinConns
		eMinConns = os.Getenv(envPostgresMinConns)
	)
	if eMinConns != "" {

		mnc, err := strconv.ParseInt(eMinConns, 10, 32) // третьим аргументом размерность 32, тогда парсер сам выдаст ошибку, если кто-то в env запихнет число, превышающее лимит `int32`.
		if err != nil {
			logger.Printf("[WARNING] invalid %q, default (%d conns) for db pool was applied: %v\n",
				envPostgresMinConns, defaultPostgresMinConns, err)
		} else {
			minConns = int32(mnc)
		}
	}

	var (
		maxConns  = defaultPostgresMaxConns
		eMaxConns = os.Getenv(envPostgresMaxConns)
	)
	if eMaxConns != "" {
		mnc, err := strconv.ParseInt(eMaxConns, 10, 32)
		if err != nil {
			logger.Printf("[WARNING] invalid %q, default (%d conns) for db pool was applied: %v\n",
				envPostgresMaxConns, defaultPostgresMaxConns, err)
		} else {
			maxConns = int32(mnc)
		}
	}

	return Pool{
		MinConns: minConns,
		MaxConns: maxConns,
	}
}

type Postgres struct {
	dsn
	Pool        Pool
	DialTimeout time.Duration
}

func newPostgres(logger *log.Logger) (Postgres, error) {
	var (
		dlTimeout = defaultPostgresDialTimeout * time.Second
		eDTimeout = os.Getenv(envPostgresDialTimeout)
	)
	if eDTimeout != "" {
		dt, err := strconv.ParseInt(eDTimeout, 10, 64)
		if err != nil {
			logger.Printf("[WARNING] invalid %q, default (%d sec) for dial timeout was applied: %v\n",
				envPostgresDialTimeout, defaultPostgresDialTimeout, err)
		} else {
			dlTimeout = time.Duration(dt) * time.Second
		}
	}

	dsnInfo, err := newDSNInfo()
	if err != nil {
		return Postgres{}, fmt.Errorf("create dsn info: %w", err)
	}
	poolinfo := newPoolInfo(logger)

	return Postgres{
		dsn:         dsnInfo,
		Pool:        poolinfo,
		DialTimeout: dlTimeout,
	}, nil
}
