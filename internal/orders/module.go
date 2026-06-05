package orders

import (
	"ddd/internal/cfg"
	"ddd/internal/orders/application"
	"ddd/internal/orders/postgres"
	"ddd/internal/orders/rest"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InitModule собирает все слои поддомена и подключает их к глобальному роутеру.
func InitModule(db *pgxpool.Pool, router chi.Router, conf *cfg.Orders, logger *log.Logger) {
	// Создаем дочерний логгер, наследуя вывод и флаги, но расширяя префикс.
	ordersLogger := log.New(logger.Writer(), logger.Prefix()+"[ORDERS] ", logger.Flags())

	// 1. Инфраструктура (Outbound)
	orderRepo := postgres.NewOrderRepo(db)
	customerChecker := postgres.NewCustomerChecker(db)
	productChecker := postgres.NewProductChecker(db)

	// 2. Бизнес-логика (Application).
	// Компилятор автоматически проверяет, что репозитории соответствуют интерфейсам из пакета ports.
	orderService := application.NewOrderService(orderRepo, customerChecker, productChecker)

	// 3. Транспорт (Inbound).
	// Компилятор автоматически проверяет, что orderService соответствует интерфейсу api.OrderService.
	handlers := rest.NewOrderHandlers(orderService, ordersLogger, conf.MaxUploadSize)

	// 4. Регистрация маршрутов
	handlers.RegisterRoutes(router)
}

/*
Строгая последовательность сборки:
«инфраструктура (outbound-адаптеры) -> бизнес-логика (сервисы) -> транспорт (inbound-адаптеры) ->
 регистрация маршрутов»
— это каноничная реализация Composition Root в гексагональной архитектуре.
*/
