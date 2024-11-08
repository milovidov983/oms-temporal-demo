package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/milovidov983/oms-temporal-demo/oms-core/service"
	"github.com/milovidov983/oms-temporal-demo/shared/models"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[warn] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		log.Printf("[error] Failed to decode request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateOrder(r.Context(), &order); err != nil {
		log.Printf("[error] Failed to create order: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"order_id": order.ID})
	log.Printf("[info] Order created: %s", order.ID)
}

func (h *OrderHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("[warn] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		log.Printf("[warn] Order ID not provided")
		http.Error(w, "Order ID not provided", http.StatusBadRequest)
		return
	}

	status, err := h.service.GetOrderStatus(r.Context(), orderID)
	if err != nil {
		log.Printf("[error] Failed to get order status: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": string(status)})
	log.Printf("[info] Order status retrieved: %s", status)
}
