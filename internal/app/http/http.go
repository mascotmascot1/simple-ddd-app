package http

import (
	"context"
	"ddd/internal/cfg"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	defaultReadTimeout  = time.Second * 5
	defaultWriteTimeout = time.Second * 15
	defaultIdleTimeout  = time.Second * 15
)

func NewRouter(logger *log.Logger) chi.Router {
	r := chi.NewRouter()

	// Архитектурно правильный подход: middleware.Timeout выставляется на требуемое
	// бизнес-время (например, 10 секунд), а WriteTimeout в http.Server увеличивается с запасом
	// (например, 15 или 20 секунд).
	//
	// Этот запас гарантирует, что если мидлварь отменит контекст на 10-й секунде, у транспортного
	// слоя останется время сформировать и протолкнуть JSON со статусом 504 в сокет до того, как
	// сработает жесткий сетевой обрыв от WriteTimeout
	r.Use(middleware.Timeout(time.Second * 10))

	r.Use(WithLogging(logger))
	return r
}

func NewServer(conf *cfg.Server, r chi.Router, logger *log.Logger) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		ErrorLog:     logger,
		Handler:      r,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}
}

func Run(srv *http.Server, conf *cfg.Server) error {
	// Буферизированный канал для получения результата остановки сервера
	shutdownError := make(chan error, 1)

	// Горутина для перехвата системных сигналов
	go func() {
		quit := make(chan os.Signal, 1)

		// Подписываемся на мягкие сигналы остановки (Ctrl+C, Docker stop)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit // Блокируемся и ждем сигнала от ОС

		// Выделяем таймаут на завершение обработки текущих запросов клиентов
		ctx, cancel := context.WithTimeout(context.Background(), conf.ShutdownTimeout)
		defer cancel() // Очищаем системный таймер контекста

		// Инициируем мягкую остановку и отправляем результат (ошибку или nil) в канал
		shutdownError <- srv.Shutdown(ctx)
	}()

	// Запускаем сервер (функция блокирует текущий поток)
	err := srv.ListenAndServe()
	// ErrServerClosed означает штатную остановку через вызов Shutdown() выше
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("launch server: %w", err)
	}

	// Ждем завершения процесса Shutdown() и читаем результат из горутины
	err = <-shutdownError
	if err != nil {
		// Ошибка здесь чаще всего означает, что сработал таймаут контекста
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
