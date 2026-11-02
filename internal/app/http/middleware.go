package http

import (
	"context"
	"log"
	"net/http"
	"time"
)

// WithLogging возвращает мидлварь, которая логирует каждый входящий запрос.
// Логируются метод запроса, URI и IP-адрес отправителя.
// Примечание: если перед приложением стоит балансировщик нагрузки с терминацией TCP,
// и реальный IP не прокидывается (например, через ConnContext (L4 LB) или X-Forwarded-For (L7 LB)),
// то r.RemoteAddr будет содержать IP-адрес балансировщика, а не реального клиента.
func WithLogging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("Request: %s %s from %s", r.Method, r.RequestURI, r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout добавляет таймаут к контексту запроса.
// Важно: эта мидлварь сама не пишет HTTP-ответ при истечении времени.
// Ответственность за проверку и возврат ошибки 503/504 лежит на хендлерах.
func Timeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
