package cfg

import (
	"log"
	"os"
	"strconv"
	"time"
)

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
	postgres, err := newPostgres(logger)
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

func parseEnvDuration(envKey string, defaultVal time.Duration, logger *log.Logger) time.Duration {
	valStr := os.Getenv(envKey)
	if valStr == "" {
		return defaultVal
	}

	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		logger.Printf("[WARNING] invalid %q, default (%.1f sec) applied: %v\n",
			envKey, defaultVal.Seconds(), err)
		return defaultVal
	}
	return time.Duration(val) * time.Second
}

func parseEnvInt32(envKey string, defaultVal int32, logger *log.Logger) int32 {
	valStr := os.Getenv(envKey)
	if valStr == "" {
		return defaultVal
	}

	// Размерность 32, тогда парсер сам выдаст ошибку, если кто-то в env положит число, превышающее лимит `int32`.
	val, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		logger.Printf("[WARNING] invalid %q, default (%d conns) applied: %v\n",
			envKey, defaultVal, err)
		return defaultVal
	}
	return int32(val)
}
