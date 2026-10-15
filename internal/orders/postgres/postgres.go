package postgres

// TODO: Внедрить Optimistic Locking при переходе к микросервисной архитектуре.
// При масштабировании добавить поле version в доменную модель Order для защиты от Lost Update
// в условиях распределенных транзакций.
// SQL-запрос должен быть расширен: DELETE FROM orders WHERE id = $1 AND version = $2.
// Если pgxpool.Exec возвращает 0 затронутых строк (RowsAffected() == 0),
// транслировать это в доменную ошибку ErrConcurrencyConflict для инициации retry-политики на стороне клиента.

import (
	"context"
	"ddd/internal/orders/domain"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxArgumentCount    = 8
	defaultItemCapacity = 10

	noItemsMarker = "NO_ITEMS_MARKER"
)

type OrderRepo struct {
	db *pgxpool.Pool
}

func NewOrderRepo(db *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{
		db: db,
	}
}

func (p *OrderRepo) Save(ctx context.Context, order *domain.Order) error {
	// Открываем транзакцию
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Отложенный rollback гарантирует, что если функция завершится ошибкой
	// до вызова Commit, транзакция будет отменена.
	// При успешном Commit этот Rollback просто ничего не сделает.
	defer tx.Rollback(ctx) // nolint:errcheck

	upsetOrderQuery := `
		INSERT INTO orders (order_id, customer_id, total_amount, currency)
		VALUES (@order_id, @customer_id, @total_amount, @currency)
		ON CONFLICT (order_id) DO UPDATE
		SET total_amount = EXCLUDED.total_amount,
			currency = EXCLUDED.currency;
	`
	args := make(pgx.NamedArgs, maxArgumentCount)
	args["order_id"] = order.ID()
	args["customer_id"] = order.CustomerID()
	args["total_amount"] = order.TotalPrice().Amount()
	args["currency"] = order.TotalPrice().Currency()

	_, err = tx.Exec(ctx, upsetOrderQuery, args)
	if err != nil {
		return fmt.Errorf("upsert order record: %w", err)
	}

	deleteItemsQuery := `DELETE FROM order_items WHERE order_id = @order_id;`
	_, err = tx.Exec(ctx, deleteItemsQuery, args)
	if err != nil {
		return fmt.Errorf("delete old order items: %w", err)
	}

	insertItemQuery := `
	INSERT INTO order_items (order_id, product_id, amount, currency, quantity)
	VALUES (@order_id, @product_id, @amount, @currency, @quantity);
	`
	for _, item := range order.Items() {
		args["product_id"] = item.ProductID()
		args["amount"] = item.Price().Amount()
		args["currency"] = item.Price().Currency()
		args["quantity"] = item.Quantity()

		_, err := tx.Exec(ctx, insertItemQuery, args)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (p *OrderRepo) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	// TODO - вариант с json_agg
	selectQuery := `
	SELECT 
		o.customer_id, 
		o.total_amount, 
		o.currency,
		COALESCE(i.product_id::TEXT, @no_items_marker), 
		COALESCE(i.amount, 0), 
		COALESCE(i.currency, ''), 
		COALESCE(i.quantity, 0)
	FROM orders o
	LEFT JOIN order_items i ON o.order_id = i.order_id
	WHERE o.order_id = @order_id;
	`
	args := pgx.NamedArgs{"order_id": id, "no_items_marker": noItemsMarker}

	rows, err := p.db.Query(ctx, selectQuery, args)
	if err != nil {
		return nil, fmt.Errorf("query order with items: %w", err)
	}
	defer rows.Close()

	var (
		customerID  string
		totalPrice  domain.MoneyState
		itemsStates = make([]domain.OrderItemState, 0, defaultItemCapacity)
		orderFound  bool
	)

	for rows.Next() {
		var iState domain.OrderItemState
		orderFound = true

		err := rows.Scan(&customerID, &totalPrice.Amount, &totalPrice.Currency, &iState.ProductID,
			&iState.Price.Amount, &iState.Price.Currency, &iState.Quantity)
		if err != nil {
			return nil, fmt.Errorf("scan joined order row: %w", err)
		}

		// Используем инфраструктурный паттерн Sentinel Value (значение-маркер).
		// Если в заказе еще нет позиций, LEFT JOIN в SQL-запросе сгенерирует строку,
		// где вместо реального идентификатора товара вернется маркер noItemsMarker.
		// Мы отсекаем этот артефакт БД на границе слоя, предотвращая попадание
		// невалидных фиктивных данных в доменную модель агрегата.
		if iState.ProductID == noItemsMarker {
			continue
		}
		itemsStates = append(itemsStates, iState)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if !orderFound {
		return nil, nil // CONTRACT from the ports package.
	}

	return domain.RestoreOrder(id, customerID, totalPrice, itemsStates), nil
}

func (p *OrderRepo) Delete(ctx context.Context, order *domain.Order) error {
	// Запрос удаляет только родительский агрегат.
	// Очистка дочерних позиций делегирована СУБД через ограничение on delete cascade.
	// Использование составных запросов (cte) или ручного управления транзакциями не требуется.
	deleteQuery := `DELETE FROM orders WHERE order_id = @order_id;`
	args := pgx.NamedArgs{"order_id": order.ID()}

	if _, err := p.db.Exec(ctx, deleteQuery, args); err != nil {
		return fmt.Errorf("exec delete query: %w", err)
	}
	return nil
}

/*
##### Ошибка из `defer tx.Rollback(ctx)` — это технический побочный продукт.
##### *- Если всё хорошо, откат выдаст ошибку, потому что откатывать нечего.*
##### *- Если всё плохо, откат выдаст ошибку, потому что он сломается по той же самой причине,
по которой сломался основной запрос.*

Драйвер `pgx` устроен так, что если `Rollback` не может выполниться, он просто удаляет
проблемное соединение на стороне Go. База данных всё чистит на своей стороне.
Именно поэтому мы ставим `//nolint:errcheck` и идем дальше со спокойной душой.
*/
