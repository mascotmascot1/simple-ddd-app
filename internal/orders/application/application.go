package application

import (
	"context"
	"ddd/internal/orders/domain"
	"ddd/internal/orders/ports"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	// Поскольку валидация существования происходит на уровне `application` при помощи интерфейсов
	// из пакета `ports`, эти ошибки логичнее всего определять прямо в пакете `application` или `ports`.
	ErrCustomerNotFound = errors.New("customer not found")
	ErrProductNotFound  = errors.New("product not found")
	ErrOrderNotFound    = errors.New("order not found")
)

type NotFoundError struct {
	err error
}

func (e NotFoundError) Error() string {
	return e.err.Error()
}

func (e NotFoundError) Unwrap() error {
	return e.err
}

// Входящее API (Inbound ports): Это публичные методы твоего OrderService.
// Это контракты, по которым внешний мир (HTTP-хендлеры, gRPC-адаптеры) имеет право вызывать твою бизнес-логику.
type OrderService struct {
	repo            ports.OrderRepository
	customerChecker ports.CustomerChecker
	productChecker  ports.ProductChecker
}

func NewOrderService(repo ports.OrderRepository,
	customerChecker ports.CustomerChecker,
	productChecker ports.ProductChecker,
) *OrderService {
	return &OrderService{
		repo:            repo,
		customerChecker: customerChecker,
		productChecker:  productChecker,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, customerID string) (*domain.Order, error) {
	customerExists, err := s.customerChecker.Exists(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("check customer existence: %w", err)
	}
	if !customerExists {
		return nil, NotFoundError{err: ErrCustomerNotFound}
	}

	orderID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid: %w", err)
	}

	order, err := domain.NewOrder(orderID.String(), customerID)
	if err != nil {
		return nil, fmt.Errorf("create domain order: %w", err)
	}
	if err := s.repo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("save new order: %w", err)
	}
	// Здесь можно публиковать Domain Event: OrderCreated
	return order, nil
}

func (s *OrderService) AddItemToOrder(ctx context.Context, orderID, productID, currency string, amount, quantity int64) error {
	productExists, err := s.productChecker.Exists(ctx, productID)
	if err != nil {
		return fmt.Errorf("check product existence: %w", err)
	}
	if !productExists {
		return NotFoundError{err: ErrProductNotFound}
	}

	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("find order by id: %w", err)
	}
	if order == nil {
		return NotFoundError{err: ErrOrderNotFound}
	}

	price, err := domain.NewMoney(amount, currency)
	if err != nil {
		return fmt.Errorf("create money: %w", err)
	}

	err = order.AddItem(productID, price, quantity)
	if err != nil {
		return fmt.Errorf("add item to order: %w", err)
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	// Здесь можно публиковать Domain Event: OrderItemAdded
	return nil
}

// Публикация событий силами инфраструктурного слоя используется в распределенных системах
// для решения проблемы Dual Write (двойной записи). Если Application Service сначала успешно
// удалит заказ в базе, а на моменте публикации события в Kafka произойдет сетевой сбой,
// система потеряет консистентность.
//
// Для обхода этого сбоя применяется паттерн Transactional Outbox в связке с транзакционным
// декоратором. Декоратор — это структурный паттерн (обертка), который перехватывает вызов
// Application Service. Он открывает транзакцию в PostgreSQL, выполняет бизнес-логику, а затем
// репозиторий в рамках одной транзакции атомарно удаляет заказ и записывает сериализованное
// событие в соседнюю таблицу outbox. Чтением из этой таблицы и отправкой в брокер занимается
// отдельный фоновый процесс (инфраструктура).
//
// В текущей монолитной реализации внедрение Outbox избыточно. Вариант с публикацией
// непосредственно из DeleteOrder после успешного s.repo.Delete полностью отвечает
// требованиям системы.
func (s *OrderService) DeleteOrder(ctx context.Context, orderID string) error {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("find order by id: %w", err)
	}
	if order == nil {
		return NotFoundError{err: ErrOrderNotFound}
	}

	order.Delete()
	if err := s.repo.Delete(ctx, order); err != nil {
		return fmt.Errorf("delete order: %w", err)
	}
	// Здесь можно публиковать Domain Event: OrderDeleted
	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("find order by id: %w", err)
	}
	if order == nil {
		return nil, NotFoundError{err: ErrOrderNotFound}
	}
	// Для идемпотентных операций чтения, не изменяющих состояние системы,
	// доменные события не генерируются.
	return order, nil
}
