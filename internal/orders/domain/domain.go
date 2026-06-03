package domain

// Разобрать как бы это выглядело в MVC - ==**Классическая многослойная архитектура (Layered)==:**

import (
	"errors"
)

const averageCountOfItems = 10

var (
	ErrNegativeAmount = errors.New("amount mustn't be negative")
	ErrEmptyCurrency  = errors.New("currency mustn't be empty")

	ErrEmptyProductID      = errors.New("product's ID mustn't be empty")
	ErrNonPositiveQuantity = errors.New("quantity must be positive")

	ErrEmptyOrderID    = errors.New("order's ID mustn't be empty")
	ErrEmptyCustomerID = errors.New("customer's ID mustn't be empty")

	ErrMixedCurrency = errors.New("cannot mix the currencies in an order")
)

// В контексте чистой архитектуры и DDD этот подход классифицируется как
// Typed Domain Errors (Типизированные ошибки домена).
type ValidateError struct {
	err error
}

func (e ValidateError) Error() string {
	return e.err.Error()
}

func (e ValidateError) Unwrap() error {
	return e.err
}

// Value Object: Money представляет сумму денег (неизменяемый)
type Money struct {
	amount   int64
	currency string
}

func (m Money) Amount() int64 {
	return m.amount
}

func (m Money) Currency() string {
	return m.currency
}

// Паттерн Always Valid Domain Model — возврат error из конструктора value object и других сущностей
func NewMoney(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, ValidateError{err: ErrNegativeAmount}
	}
	if currency == "" {
		return Money{}, ValidateError{err: ErrEmptyCurrency}
	}
	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

type orderItem struct {
	productID string
	price     Money
	quantity  int64
}

func (i orderItem) ProductID() string {
	return i.productID
}

func (i orderItem) Price() Money {
	return i.price
}

func (i orderItem) Quantity() int64 {
	return i.quantity
}

// Фабричный метод для создания нового заказа
func newOrderItem(productID string, price Money, quantity int64) (orderItem, error) {
	if productID == "" {
		return orderItem{}, ValidateError{err: ErrEmptyProductID}
	}
	if quantity <= 0 {
		return orderItem{}, ValidateError{err: ErrNonPositiveQuantity}
	}
	return orderItem{
		productID: productID,
		price:     price,
		quantity:  quantity,
	}, nil
}

// Entity: OrderItem представляет позицию в заказе
type Order struct {
	id         string
	customerID string
	items      []orderItem
	totalPrice Money
}

func NewOrder(orderID, customerID string) (*Order, error) {
	if orderID == "" {
		return nil, ValidateError{err: ErrEmptyOrderID}
	}
	if customerID == "" {
		return nil, ValidateError{err: ErrEmptyCustomerID}
	}
	return &Order{
		id:         orderID,
		customerID: customerID,
		items:      make([]orderItem, 0, averageCountOfItems),
		// totalPrice можно не указывать если нет бизнес правила о том, например, что все новые заказы по умолчнию с USD валютой
	}, nil
	// Здесь можно генерировать Domain Event: OrderCreated
}

func (o *Order) ID() string {
	return o.id
}

func (o *Order) CustomerID() string {
	return o.customerID
}

func (o *Order) TotalPrice() Money {
	return o.totalPrice
}

func (o *Order) Items() []orderItem {
	items := make([]orderItem, len(o.items))
	copy(items, o.items)
	return items
}

// Метод на агрегате для добавления позиции (инкапсулирует логику)
func (o *Order) AddItem(productID string, price Money, quantity int64) error {
	if o.totalPrice.amount != 0 && o.totalPrice.currency != price.currency {
		return ValidateError{err: ErrMixedCurrency}
	}

	item, err := newOrderItem(productID, price, quantity)
	if err != nil {
		return err
	}
	o.items = append(o.items, item)

	total := o.recalculateTotal()
	o.totalPrice, err = NewMoney(total, price.currency)
	if err != nil {
		return err
	}
	// Здесь можно генерировать Domain Event: OrderItemAdded
	return nil
}

func (o *Order) Delete() {
	// Если, согласно бизнес-правилам, удалить заказ можно только при определенных
	// условиях - эта логика проверяется здесь И тогда добавляется error в
	// качестве возвращаемого значения

	// Здесь можно генерировать Domain Event: OrderDeleted
}

func (o *Order) recalculateTotal() int64 {
	var total int64
	for _, item := range o.items {
		total += item.price.amount * item.quantity
	}
	return total
}

/*
↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
Такие объекты называются DTO (Data Transfer Object) или Data Model. В строгом DDD их принято
называть Snapshot (снимок) или State. Это структуры без методов и бизнес-логики, предназначенные
исключительно для переноса плоских данных между слоем инфраструктуры и ядром.
↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
*/

type MoneyState struct {
	Amount   int64
	Currency string
}

type OrderItemState struct {
	ProductID string
	Price     MoneyState
	Quantity  int64
}

// Специальный конструктор (для загрузки из БД) / Экспортируемый маппер (Mapper). Паттерн Reconstitution.
func RestoreOrder(id, customerID string, totalPrice MoneyState, itemsStates []OrderItemState) *Order {
	items := make([]orderItem, 0, len(itemsStates))
	for _, item := range itemsStates {
		items = append(items, orderItem{
			productID: item.ProductID,
			price:     Money{item.Price.Amount, item.Price.Currency},
			quantity:  item.Quantity,
		})
	}

	return &Order{
		id:         id,
		customerID: customerID,
		totalPrice: Money{totalPrice.Amount, totalPrice.Currency},
		items:      items,
	}
}

// Логика паттерна Always Valid Domain Model применена верно.
// Твое рассуждение насчет поля price абсолютно точное: так как структура Money создается
// исключительно через функцию NewMoney, где уже зашита валидация, остальные функции могут
// полностью доверять переданному объекту. В этом и заключается главная польза Value Object в доменной модели.
