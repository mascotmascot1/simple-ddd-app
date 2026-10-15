package postgres

import (
	"context"
	"ddd/internal/cfg"
	"fmt"
	"time"

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
func NewPool(ctx context.Context, poolInfo *cfg.Pool, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	// Явно задаем лимиты пула.
	// Если запросов больше, чем MaxConns, пул выстроит горутины в очередь, а не пойдет открывать новые сокеты.
	// Это гарантирует, что мы не превысим жесткий лимит max_connections на самом сервере PostgreSQL,
	// при превышении которого БД отбивает новые подключения с ошибкой FATAL.
	cfg.MaxConns = poolInfo.MaxConns
	cfg.MinConns = poolInfo.MinConns

	// MaxConnLifetime принудительно закрывает старые, но полностью рабочие соединения.
	// В PostgreSQL каждый коннект — это отдельный процесс ОС. Со временем он обрастает кэшами
	// и распухает по RAM. Ротация коннектов заставляет БД убивать старые процессы и освобождать память.
	cfg.MaxConnLifetime = time.Hour

	// HealthCheckPeriod пингует простаивающие соединения в фоне.
	// Если сетевое оборудование (балансировщик, фаервол, роутер) "тихо" дропнуло TCP-сессию
	// по таймауту неактивности (удалило стейт из таблицы соединений) или мигнула сеть,
	// пул выкинет мертвый сокет до того, как он достанется клиентскому запросу (защита от 500-х ошибок).
	cfg.HealthCheckPeriod = time.Minute

	db, err := pgxpool.NewWithConfig(ctx, cfg)
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
