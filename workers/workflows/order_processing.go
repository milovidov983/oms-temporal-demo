// Yoy need to think how to inject dependencies to activities
// maybe it need to ask

package workflows

import (
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	OrderProcessingStatusQuery = "order-processing-status"
)

type OrderProcessingWorkflowInput struct {
	OrderID       string
	CorrelationID string
}

type OrderProcessingState struct {
	OrderID       string
	CorrelationID string
	State         string
}

type orderProcessingWorkflow struct {
	OrderProcessingState
	processingID string
	logger       log.Logger
}

// newOrderProcessingWorkflow initializes a orderProcessingWorkflow struct
func newOrderProcessingWorkflow(ctx workflow.Context, state *OrderProcessingState) *orderProcessingWorkflow {
	return &orderProcessingWorkflow{
		OrderProcessingState: *state,
		processingID:         workflow.GetInfo(ctx).WorkflowExecution.RunID,
		logger:               workflow.GetLogger(ctx),
	}
}

// pushStatus updates the OrderProcessingStatus search attribute for a order processing workflow execution.
func (w *orderProcessingWorkflow) pushStatus(ctx workflow.Context, status string) error {
	var keywordKey = temporal.NewSearchAttributeKeyKeyword("OrderProcessingStatus")
	return workflow.UpsertTypedSearchAttributes(
		ctx,
		keywordKey.ValueSet(status),
	)
}

// ProcessOrder is a Workflow Definition that calls for the execution of a variable set of Activities and Child Workflows.
// This is the main entry point of the application.
// It accepts an Order ID as the input.
func ProcessOrder(ctx workflow.Context, input *OrderProcessingWorkflowInput) error {
	w := newOrderProcessingWorkflow(
		ctx,
		&OrderProcessingState{
			OrderID:       input.OrderID,
			CorrelationID: input.CorrelationID,
			State:         "processing_started",
		},
	)

	w.logger.Info("Processing order", "order_id", w.OrderID, "correlation_id", w.CorrelationID)

	w.pushStatus(ctx, w.OrderProcessingState.State)

	err := workflow.SetQueryHandler(ctx, OrderProcessingStatusQuery, func() (OrderProcessingState, error) {
		return w.OrderProcessingState, nil
	})
	if err != nil {
		return err
	}

	///
	///

	w.OrderProcessingState.State = "processing_completed"
	w.pushStatus(ctx, w.OrderProcessingState.State)

	return nil
}
