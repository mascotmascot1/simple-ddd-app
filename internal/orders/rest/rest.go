package rest

import (
	"context"
	"ddd/internal/orders/application"
	"ddd/internal/orders/domain"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// В рамках чистой архитектуры и DDD один эндпоинт (хендлер) транспортного слоя проецируется
// строго на один метод (use case) слоя application.

// Термин Inbound-контракт абсолютно корректен. В терминологии гексагональной архитектуры
// (Ports and Adapters) этот интерфейс представляет собой Inbound Port
// (первичный или управляющий порт).
//
// Совпадение названия интерфейса в пакете api и реализующей его структуры в пакете
// application является технической нормой.
type OrderService interface {
	CreateOrder(ctx context.Context, customerID string) (*domain.Order, error)
	AddItemToOrder(ctx context.Context, orderID, productID, currency string, amount, quantity int64) error
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)
	DeleteOrder(ctx context.Context, orderID string) error
}

type OrderHandlers struct {
	service       OrderService
	maxUploadSize int64
	logger        *log.Logger
}

func NewOrderHandlers(service OrderService, logger *log.Logger, maxUploadSize int64) *OrderHandlers {
	return &OrderHandlers{
		service:       service,
		logger:        logger,
		maxUploadSize: maxUploadSize,
	}
}

func (h *OrderHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/orders", func(r chi.Router) {
		// Мидлварь применяется только к эндпоинтам этого поддомена
		r.Use(h.withSizeLimit)

		r.Post("/", h.createOrder)
		r.Get("/{orderID}", h.getOrder)
		r.Post("/{orderID}/items", h.addItemToOrder)
		r.Delete("/{orderID}", h.deleteOrder)
	})
}

// withSizeLimit returns a middleware that limits the size of each incoming request
// to the specified value in limits. It's intended to be used to prevent abuse and
// protect server from running out of memory.
func (h *OrderHandlers) withSizeLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadSize)
		next.ServeHTTP(w, r)
	})
}

func (h *OrderHandlers) createOrder(w http.ResponseWriter, r *http.Request) {
	caller := "createOrder"

	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Printf("%s: read request body: %v\n", caller, err)
		h.writeJSON(w, errorResponse{Error: "internal server error"}, http.StatusInternalServerError, caller)
		return
	}

	var req createOrderRequest
	if err := json.Unmarshal(data, &req); err != nil {
		h.logger.Printf("%s: json unmarshalling: %v\n", caller, err)
		h.writeJSON(w, errorResponse{Error: "bad request"}, http.StatusBadRequest, caller)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Printf("%s: validation failed: %v\n", caller, err)
		h.writeJSON(w, errorResponse{Error: "bad request"}, http.StatusBadRequest, caller)
		return
	}

	order, err := h.service.CreateOrder(r.Context(), req.CustomerID)
	if err != nil {
		h.logger.Printf("%s: service.CreateOrder: %v\n", caller, err)
		h.handleCoreError(w, caller, err)
		return
	}

	h.writeJSON(w, createOrderResponse{OrderID: order.ID()}, http.StatusCreated, caller)
}

func (h *OrderHandlers) getOrder(w http.ResponseWriter, r *http.Request) {
	caller := "getOrder"

	// Извлечение параметра из URI.
	// Ключ должен точно совпадать с названием в r.Get("/{orderID}", ...)
	orderID := chi.URLParam(r, "orderID")
	if err := uuid.Validate(orderID); err != nil {
		h.logger.Printf("%s: invalid order_id: %v\n", caller, err)
		h.writeJSON(w, errorResponse{Error: "bad request"}, http.StatusBadRequest, caller)
		return
	}

	order, err := h.service.GetOrder(r.Context(), orderID)
	if err != nil {
		h.logger.Printf("%s: service.GetOrder: %v\n", caller, err)
		h.handleCoreError(w, caller, err)
		return
	}

	orderItems := order.Items()
	itemsDTO := make([]orderItemDTO, 0, len(orderItems))
	for _, item := range orderItems {
		itemsDTO = append(itemsDTO, orderItemDTO{
			ProductID: item.ProductID(),
			Amount:    item.Price().Amount(),
			Currency:  item.Price().Currency(),
			Quantity:  item.Quantity(),
		})
	}
	orderDTO := getOrderResponse{
		OrderID:     order.ID(),
		CustomerID:  order.CustomerID(),
		Currency:    order.TotalPrice().Currency(),
		TotalAmount: order.TotalPrice().Amount(),
		Items:       itemsDTO,
	}

	h.writeJSON(w, orderDTO, http.StatusOK, caller)
}

func (h *OrderHandlers) addItemToOrder(w http.ResponseWriter, r *http.Request) {
	caller := "addItemToOrder"

	orderID := chi.URLParam(r, "orderID")
	if err := uuid.Validate(orderID); err != nil {
		h.logger.Printf("%s: invalid order_id: %v\n", caller, err)
		h.writeJSON(w, errorResponse{Error: "bad request"}, http.StatusBadRequest, caller)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Printf("%s: read request body: %v\n", caller, err)
		h.writeJSON(w, errorResponse{Error: "internal server error"}, http.StatusInternalServerError, caller)
		return
	}

	var req addItemToOrderRequest
	if err := json.Unmarshal(data, &req); err != nil {
		h.logger.Printf("json unmarshalling: %v\n", err)
		h.writeJSON(w, errorResponse{Error: "bad request"}, http.StatusBadRequest, caller)
		return
	}

	// --------------------------------------CHECK--------------------------------------
	if err := req.Validate(); err != nil {
		h.logger.Printf("%s: validation failed: %v\n", caller, err)
		h.writeJSON(w, errorResponse{Error: "bad request"}, http.StatusBadRequest, caller)
		return
	}
	// --------------------------------------CHECK--------------------------------------

	err = h.service.AddItemToOrder(r.Context(), orderID, req.ProductID,
		req.Currency, req.Amount, req.Quantity)
	if err != nil {
		h.logger.Printf("%s: service.AddItemToOrder: %v\n", caller, err)
		h.handleCoreError(w, caller, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *OrderHandlers) deleteOrder(w http.ResponseWriter, r *http.Request) {
	caller := "deleteOrder"
	orderID := chi.URLParam(r, "orderID")
	if err := uuid.Validate(orderID); err != nil {
		h.logger.Printf("%s: invalid order_id: %v\n", caller, err)
		h.writeJSON(w, errorResponse{Error: "bad request"}, http.StatusBadRequest, caller)
		return
	}

	err := h.service.DeleteOrder(r.Context(), orderID)
	if err != nil {
		h.logger.Printf("%s: service.DeleteOrder: %v\n", caller, err)
		h.handleCoreError(w, caller, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeJSON writes the given data to the writer with the given status code.
// It assumes that the writer is already set up to write JSON data.
// If there is an error encoding the data, it logs the error.
// It does not return an error since the error is already logged.
// It also does not modify the writer's status code if there is an error encoding the data.
func (h *OrderHandlers) writeJSON(w http.ResponseWriter, data any, code int, caller string) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		h.logger.Printf("%s: json marshalling: %v\n", caller, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	if _, err := w.Write(dataJSON); err != nil {
		h.logger.Printf("%s: write respone: %v\n", caller, err)
	}
}

func (h *OrderHandlers) handleCoreError(w http.ResponseWriter, caller string, err error) {
	var (
		valErr      domain.ValidateError
		notFoundErr application.NotFoundError
	)
	switch {
	case errors.As(err, &valErr):
		h.writeJSON(w, errorResponse{Error: valErr.Error()}, http.StatusBadRequest, caller)
	case errors.As(err, &notFoundErr):
		h.writeJSON(w, errorResponse{Error: notFoundErr.Error()}, http.StatusNotFound, caller)
	default:
		h.writeJSON(w, errorResponse{Error: "internal server error"}, http.StatusInternalServerError, caller)
	}
}
