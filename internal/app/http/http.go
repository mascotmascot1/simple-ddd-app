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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(conf *cfg.Server, logger *log.Logger) chi.Router {
	r := chi.NewRouter()

	// Архитектурно правильный подход: middleware.Timeout выставляется на требуемое
	// бизнес-время (например, 10 секунд), а WriteTimeout в http.Server увеличивается с запасом
	// (например, 15 или 20 секунд).
	//
	// Этот запас гарантирует, что если мидлварь отменит контекст на 10-й секунде, у транспортного
	// слоя останется время сформировать и протолкнуть JSON со статусом 504 в сокет до того, как
	// сработает жесткий сетевой обрыв от WriteTimeout
	r.Use(middleware.Timeout(conf.Timeouts.Request))

	r.Use(WithLogging(logger))
	return r
}

func NewServer(conf *cfg.Server, r chi.Router, logger *log.Logger) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		ErrorLog:     logger,
		Handler:      r,
		ReadTimeout:  conf.Timeouts.Read,
		WriteTimeout: conf.Timeouts.Write,
		IdleTimeout:  conf.Timeouts.Idle,
	}
}

func Run(srv *http.Server, conf *cfg.Server) error {
	var (
		// Небуферизованный канал отмены. Закрываем через close(), что моментально
		// разблокирует select и не заблокирует main, даже если горутина уже мертва.
		cancelCh = make(chan struct{})
		// Буфер = 1 обязателен! Защищает от утечки горутины в race condition,
		// когда OS сигнал прилетел ровно в момент падения ListenAndServe, и main
		// завершился раньше, чем горутина успела записать результат Shutdown.
		shutdownError = make(chan error, 1)
	)
	// Горутина для перехвата системных сигналов
	go func() {
		// Буфер = 1 обязателен! Пакет signal пишет в канал через неблокирующий send.
		// Если буфера не будет, а горутина еще не дошла до select, сигнал будет потерян.
		quit := make(chan os.Signal, 1)

		// Подписываемся на мягкие сигналы остановки (Ctrl+C, Docker stop)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		select {
		case <-quit: // Блокируемся и ждем сигнала от ОС
		case <-cancelCh:
			// Если ListenAndServe упал на старте, отписываем канал от пакета signal.
			// Иначе ссылка на quit останется в глобальной мапе пакета, и GC не сможет его собрать.
			signal.Stop(quit)
			return
		}
		// Выделяем таймаут на завершение обработки текущих запросов клиентов
		ctx, cancel := context.WithTimeout(context.Background(), conf.Timeouts.Shutdown)
		defer cancel() // в данном случае (когда родитель `Background` и горутина сразу умирает) вызов `defer cancel()` — это **исключительно идиоматика и defensive programming**.

		// Инициируем мягкую остановку и отправляем результат (ошибку или nil) в канал
		shutdownError <- srv.Shutdown(ctx)
	}()

	// Запускаем сервер (функция блокирует текущий поток)
	err := srv.ListenAndServe()
	// ErrServerClosed означает штатную остановку через вызов Shutdown() выше
	if !errors.Is(err, http.ErrServerClosed) {
		close(cancelCh) // Сигнализируем горутине, что сервер упал и ждать сигналов не нужно
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
