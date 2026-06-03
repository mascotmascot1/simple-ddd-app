package main

import (
	"context"
	"ddd/internal/app/http"
	"ddd/internal/app/postgres"
	"ddd/internal/cfg"
	"ddd/internal/orders"
	"log"
	"os"
	"time"
)

func main() {
	logger := log.New(os.Stdout, "[DDD-PROJECT] ", log.LstdFlags)

	conf, err := cfg.New(logger)
	if err != nil {
		logger.Fatalf("init config: %v\n", err)
	}

	ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Second) // тут тоже надо в константу вынести
	db, err := postgres.NewPostgresPool(ctxDB, conf.Postgres.DSN())

	// Конкретно в этом месте, в main.go, если ты не вызовешь cancelDB(), таймер просто дотикает
	// свои 10 секунд, закроет канал, отвалится, и сборщик мусора его заберет. Глобальной утечки
	// памяти на всё время работы сервера (как я сказал ранее) тут не будет. В масштабах старта
	// main это микроскопическая фигня.
	cancelDB()

	if err != nil {
		logger.Fatalf("init postgres pool: %v\n", err)
	}
	defer db.Close()

	mux := http.NewRouter(logger)

	orders.InitModule(db, mux, &conf.Orders, logger)

	srv := http.NewServer(&conf.Server, mux, logger)

	if err := http.Run(srv, &conf.Server); err != nil {
		logger.Printf("http server stopped with error: %v\n", err)
	} else {
		logger.Println("http server stopped successfully")
	}
}
