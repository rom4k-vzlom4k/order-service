package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rom4k-vzlom4k/oder-service/internal/kafka"
	"github.com/rom4k-vzlom4k/oder-service/internal/service"
	"github.com/rom4k-vzlom4k/oder-service/pkg/models"
)

type orderHandler struct {
	service  service.OrderService
	producer kafka.Producer
}

type OrderHandler interface {
	GetByOrderId(w http.ResponseWriter, r *http.Request)
	SaveOrder(w http.ResponseWriter, r *http.Request)
}

func NewOrderHandler(service service.OrderService, producer kafka.Producer) OrderHandler {
	return &orderHandler{
		service:  service,
		producer: producer,
	}
}

func (h *orderHandler) GetByOrderId(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/") // /orders/b563feb7b2b84b6test  ["", "orders", "b563feb7b2b84b6test"]
	if len(parts) < 3 {
		http.Error(w, "missing id", http.StatusBadRequest)
	}
	id := parts[2]

	order, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *orderHandler) SaveOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var order models.Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid json: %v", err), http.StatusBadRequest)
		return
	}

	// сообщение в кафку
	orderBytes, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "failed to marshal order", http.StatusInternalServerError)
		return
	}

	err = h.producer.WriteMessage(context.Background(), orderBytes)
	if err != nil {
		http.Error(w, "failed to publish message to Kafka", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("order saved successfully"))
}
