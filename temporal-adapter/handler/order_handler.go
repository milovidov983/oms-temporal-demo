package handler

import (
	"context"
	"time"

	"github.com/milovidov983/oms-temporal-demo/shared/events"
	"github.com/milovidov983/oms-temporal-demo/workers/queue"
	"github.com/milovidov983/oms-temporal-demo/workers/signals/channels"
	"github.com/milovidov983/oms-temporal-demo/workers/workflows"

	"go.temporal.io/sdk/client"
)

func (h *Handler) HandleOrderEvent(event events.OrderEvent) {
	isFirstTime := true
	var err error
	// fake code
	for isFirstTime || err != nil {
		isFirstTime = false
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

			h.logger.Printf("[debug] Sleep 5 seconds...")
			// sleep 5 seconds
			time.Sleep(5 * time.Second)

		}
	}
}

func (h *Handler) handleOrderCreated(event events.OrderEvent) error {
	h.logger.Printf("[debug] Processing OrderCreated event: %+v", event.EventData)

	taskQueue := queue.TaskQueueNameOrder

	options := client.StartWorkflowOptions{
		ID:        workflows.OrderProcessingWorkflowID(event.EventData.ID),
		TaskQueue: taskQueue,
	}
	h.logger.Printf("[debug] Starting workflow: %+v", options)
	input := &workflows.OrderProcessingWorkflowInput{
		OrderID: event.EventData.ID,
	}
	//we, err := h.temporal.ExecuteWorkflow(context.Background(), options, workflows.ProcessOrder, input)
	we, err := h.temporal.SignalWithStartWorkflow(context.Background(), options.ID, channels.SignalNameStartOrderProcessingChannel, nil, options, workflows.ProcessOrder, input)

	// Check if the workflow is already running
	if err != nil && we != nil {
		h.logger.Printf("[debug] Workflow already running: %+v", we)
		return nil
	}
	if err != nil {
		h.logger.Printf("[error] %s error starting workflow: %+v", options.ID, err)
		return err
	}

	return nil
}

func (h *Handler) handleOrderCancelled(event events.OrderEvent) error {
	h.logger.Printf("[debug] Processing OrderCancelled event: %+v", event.EventData)
	// not implemented
	return nil
}

func getOrderWorkflowID(orderID string) string {
	return "order-" + orderID
}
