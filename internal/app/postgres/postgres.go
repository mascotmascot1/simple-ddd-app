package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

/*
DSN (Data Source Name) — это стандартизированная строка подключения,
которая содержит всю информацию для драйвера БД.

Формат для PostgreSQL выглядит так:
postgres://username:password@host:port/database_name?sslmode=disable.
Параметры для формирования этой строки обязательно выносятся в структуру конфигурации
(например, config.Database). В функцию NewPostgresPool передается строка,
собранная из конфига на этапе инициализации в main.go.
*/

// NewPool настраивает pgxpool и возвращает готовое соединение.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	// настройка конфига, pgxpool.NewWithConfig, пинг базы
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	var success bool
	defer func() {
		if !success {
			db.Close()
		}
	}()

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	success = true

	return db, nil
}
