package handler

import (
	"log"
	"os"

	"github.com/milovidov983/oms-temporal-demo/shared/events"
	"go.temporal.io/sdk/client"
)

type OrderHandlerConfig struct {
	TemporalHost string
	Namespace    string
}

func (cfg *OrderHandlerConfig) Check() {
	if cfg.TemporalHost == "" {
		log.Fatal("[fatal] Temporal host is not set")
	}
	if cfg.Namespace == "" {
		log.Fatal("[fatal] Temporal Namespace is not set")
	}
}

type OrderHandler struct {
	logger   *log.Logger
	temporal client.Client
}

func newTemporalClient(options client.Options) (client.Client, error) {
	if options.HostPort == "" {
		options.HostPort = os.Getenv("TEMPORAL_GRPC_ENDPOINT")
	}

	return client.NewLazyClient(options)
}

func NewOrderHandler(cfg OrderHandlerConfig) (*OrderHandler, error) {
	client, err := newTemporalClient(client.Options{
		HostPort:  cfg.TemporalHost,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return nil, err
	}
	return &OrderHandler{
		logger:   log.New(os.Stdout, "[order-handler]", log.LstdFlags),
		temporal: client,
	}, nil
}

func (h *OrderHandler) HandleOrderEvent(event events.OrderEvent) {
	var err error
	switch event.EventType {
	case events.OrderCreated:
		err = h.handleOrderCreated(event)
	case events.OrderCancelled:
		err = h.handleOrderCancelled(event)
	default:
		h.logger.Printf("[error] Unknown event type: %s", event.EventType)
	}

	if err != nil {
		h.logger.Printf("[error] Error handling event: %v", err)
	}
}

func (h *OrderHandler) handleOrderCreated(event events.OrderEvent) error {
	h.logger.Printf("[debug] Processing OrderCreated event: %+v", event.EventData)

	return nil
}

func (h *OrderHandler) handleOrderCancelled(event events.OrderEvent) error {
	h.logger.Printf("[debug] Processing OrderCancelled event: %+v", event.EventData)

	return nil
}

func getOrderWorkflowID(orderID string) string {
	return "order-" + orderID
}
