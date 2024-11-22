package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/milovidov983/oms-temporal-demo/oms-core/service"
)

type AssemblyApplicationHandler struct {
	service *service.AssemblyApplicationService
}

func NewAssemblyApplicationHandler(service *service.AssemblyApplicationService) *AssemblyApplicationHandler {
	return &AssemblyApplicationHandler{service: service}
}

func (h *AssemblyApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[warn] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		OrderID string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("[error] Failed to decode request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	application, err := h.service.CreateAssemblyApplication(r.Context(), request.OrderID)
	if err != nil {
		log.Printf("[error] Failed to create assembly application: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"application_id": application.ID})
	log.Printf("[info] Assembly application created: %s", application.ID)
}

func (h *AssemblyApplicationHandler) CompleteApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[warn] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ApplicationID string `json:"application_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("[error] Failed to decode request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CompleteAssembly(r.Context(), request.ApplicationID); err != nil {
		log.Printf("[error] Failed to complete assembly application: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "completed"})
	log.Printf("[info] Assembly application completed: %s", request.ApplicationID)
}

func (h *AssemblyApplicationHandler) CancelApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[warn] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ApplicationID string `json:"application_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("[error] Failed to decode request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CancelAssembly(r.Context(), request.ApplicationID); err != nil {
		log.Printf("[error] Failed to cancel assembly application: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
	log.Printf("[info] Assembly application cancelled: %s", request.ApplicationID)
}
