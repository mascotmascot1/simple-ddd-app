package cfg

import (
	"log"
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
