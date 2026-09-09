package rest

import (
	"ddd/pkg/jsonx"
	"errors"

	"github.com/google/uuid"
)

// Представленные структуры являются эталонной реализацией паттерна DTO (Data Transfer Object).
// Их единственная архитектурная ответственность — определить схему сериализации и
// десериализации данных для HTTP-протокола. Они изолируют ядро приложения, не позволяя
// доменным моделям (таким как domain.Order) напрямую конвертироваться в JSON.
type errorResponse struct {
	Error string `json:"error"`
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id"`
}

func (r *createOrderRequest) Validate() error {
	if err := uuid.Validate(r.CustomerID); err != nil {
		return errors.New("missing or invalid customer_id")
	}
	return nil
}

type createOrderResponse struct {
	OrderID string `json:"order_id"`
}

// Вся перечисленная информация из агрегата должна уходить в ответ. Вывод абсолютно верен:
// в текущем состоянии ядра отсутствуют сугубо технические или инфраструктурные поля.
//
// Каждый атрибут (id, customerID, items, totalPrice, productID, quantity, price с валютами)
// является чистыми бизнес-данными. Внешнему потребителю необходим полный набор этих полей для
// идентификации заказа, понимания его стоимости и состава позиций.
type orderItemDTO struct {
	ProductID string `json:"product_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Quantity  int64  `json:"quantity"`
}

type getOrderResponse struct {
	OrderID     string         `json:"order_id"`
	CustomerID  string         `json:"customer_id"`
	Currency    string         `json:"currency"`
	TotalAmount int64          `json:"total_amount"`
	Items       []orderItemDTO `json:"items"`
}

type addItemToOrderRequest struct {
	ProductID string             `json:"product_id"`
	Amount    jsonx.Field[int64] `json:"amount"`
	Currency  string             `json:"currency"`
	Quantity  int64              `json:"quantity"`
}

func (r *addItemToOrderRequest) Validate() error {
	if err := uuid.Validate(r.ProductID); err != nil {
		return errors.New("missing or invalid product_id")
	}
	if !r.Amount.Valid || r.Amount.Value < 0 {
		return errors.New("missing or invalid amount")
	}
	if r.Quantity <= 0 {
		return errors.New("missing or invalid quantity")
	}
	if r.Currency == "" {
		return errors.New("missing or invalid currency")
	}
	return nil
}
