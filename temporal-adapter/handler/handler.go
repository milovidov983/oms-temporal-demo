package handler

import (
	"log"
	"os"

	"github.com/milovidov983/oms-temporal-demo/shared/events"
	"go.temporal.io/sdk/client"
)

type OrderHandler struct {
	logger   *log.Logger
	temporal client.Client
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		logger: log.New(os.Stdout, "adapter: ", log.LstdFlags),
	}
}

func (h *OrderHandler) HandleOrderCreated(event events.OrderEvent) error {
	h.logger.Printf("Processing OrderCreated event: %+v", event.EventData)

	return nil
}

func (h *OrderHandler) HandleOrderCancelled(event events.OrderEvent) error {
	h.logger.Printf("Processing OrderCancelled event: %+v", event.EventData)

	return nil
}
