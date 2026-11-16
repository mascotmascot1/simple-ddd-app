package main

import (
	"context"
	"log"
	"os"

	"ddd/internal/app/http"
	"ddd/internal/app/postgres"
	"ddd/internal/cfg"
	"ddd/internal/orders"
)

func main() {
	logger := log.New(os.Stdout, "[DDD-PROJECT] ", log.LstdFlags)

	conf, err := cfg.New(logger)
	if err != nil {
		logger.Fatalf("init config: %v\n", err)
	}

	ctxDB, cancelDB := context.WithTimeout(context.Background(), conf.Postgres.DialTimeout)
	db, err := postgres.NewPool(ctxDB, &conf.Postgres.Pool, conf.Postgres.DSN())

	// Конкретно в этом месте, в main.go, если ты не вызовешь cancelDB(), таймер просто дотикает
	// свои 10 секунд, закроет канал, отвалится, и сборщик мусора его заберет. Глобальной утечки
	// памяти на всё время работы сервера (как я сказал ранее) тут не будет. В масштабах старта
	// main это микроскопическая фигня.
	cancelDB()

	if err != nil {
		logger.Fatalf("init postgres pool: %v\n", err)
	}
	defer db.Close()

	mux := http.NewRouter(&conf.Server, logger)

	orders.InitModule(db, mux, &conf.Orders, logger)

	srv := http.NewServer(&conf.Server, mux, logger)

	if err := http.Run(srv, &conf.Server); err != nil {
		// Если Run вернул ошибку, значит Shutdown не уложился в отведенное время
		// (например, зависло TCP-соединение к отвалившейся БД и мы
		// не дождались ответа или зависшая горутина без контекста).
		// Вызов logger.Fatalf делает os.Exit(1).
		// Это критически важно: жесткий выход предотвратит зависание процесса
		// на defer db.Close(), делегируя очистку TCP-соединений операционной системе.
		logger.Fatalf("http server stopped with error: %v\n", err)
	}
	logger.Println("http server stopped successfully")
}
