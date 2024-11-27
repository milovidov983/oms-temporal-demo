package handler

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/milovidov983/oms-temporal-demo/shared/events"
	"github.com/milovidov983/oms-temporal-demo/workers/signals"
	"github.com/milovidov983/oms-temporal-demo/workers/signals/channels"
	"github.com/milovidov983/oms-temporal-demo/workers/signals/routes"
	"github.com/milovidov983/oms-temporal-demo/workers/workflows"

	"go.temporal.io/sdk/client"
)

func NewAssemblyHandler(cfg HandlerConfig) (*Handler, error) {
	client, err := newTemporalClient(client.Options{
		HostPort:  cfg.TemporalHost,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return nil, err
	}
	return &Handler{
		logger:   log.New(os.Stdout, "[order-handler]", log.LstdFlags),
		temporal: client,
	}, nil
}

func (h *Handler) HandleAssemblyApplicationEvent(event events.AssemblyApplicationEvent) {
	isFirstTime := true
	var err error
	// fake code
	for isFirstTime || err != nil {
		isFirstTime = false
		switch event.EventType {
		case events.AssemblyCreated:
			err = h.handleAssemblyCreated(event)
		case events.AssemblyCompleted:
			err = h.handleAssemblyCompleted(event)
		case events.OrderCancelled:
			err = h.handleAssemblyCancelled(event)
		default:
			h.logger.Printf("[error] Unknown event type: %s", event.EventType)
		}

		if err != nil {
			h.logger.Printf("[error] Error handling event: %v", err)

			h.logger.Printf("[debug] Sleep 5 seconds...")
			// sleep 5 seconds
			time.Sleep(5 * time.Second)

		}
	}
}

func (h *Handler) handleAssemblyCreated(event events.AssemblyApplicationEvent) error {
	h.logger.Printf("[debug] Handling assembly created event: %v", event)
	// not implemented
	return nil
}
func (h *Handler) handleAssemblyCancelled(event events.AssemblyApplicationEvent) error {
	h.logger.Printf("[debug] Handling assembly cancelled event: %v", event)
	// not implemented
	return nil
}
func (h *Handler) handleAssemblyCompleted(event events.AssemblyApplicationEvent) error {
	h.logger.Printf("[debug] Handling assembly completed event: %v", event)

	workflowID := workflows.OrderProcessingWorkflowID(event.EventData.OrderID)

	update := signals.SignalPayloadCompleteAssembly{
		Route:     routes.RouteTypeCompleteAssembly,
		Collected: event.EventData.Collected,
	}
	signalName := channels.SignalNameCompleteAssemblyChannel

	err := h.temporal.SignalWorkflow(context.Background(), workflowID, "", signalName, update)

	if err != nil {
		h.logger.Printf("[error] Error signaling workflow: %v", err)
		return err
	}

	return nil
}
