package ports // Пакет `ports` — это интерфейс ядра системы, контракт взаимодействия между ядром и инфраструктурой

import (
	"context"
	"ddd/internal/orders/domain"
)

// Исходящее API (Outbound ports): Это твои интерфейсы в пакете ports (OrderRepository, ProductChecker).
// Это контракты, которые само ядро выставляет внешнему миру и говорит:
// "Чтобы я могло работать, инфраструктура должна реализовать для меня вот это API".

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error

	// FindByID retrieves an Order by its unique identifier.
	// If the order does not exist in the data store, it returns (nil, nil).
	// It returns a non-nil error strictly in cases of technical or infrastructure failures.
	FindByID(ctx context.Context, id string) (*domain.Order, error)
	Delete(ctx context.Context, order *domain.Order) error
}

type ProductChecker interface {
	Exists(ctx context.Context, productID string) (bool, error)
}

type CustomerChecker interface {
	Exists(ctx context.Context, customerID string) (bool, error)
}
