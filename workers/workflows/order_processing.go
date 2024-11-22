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

type OrderProperties struct {
	IsNeedAssembly    bool
	IsNeedDelivery    bool
	IsNeedMajePayment bool
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

	// Channels
	// startAssemblyChannel := workflow.GetSignalChannel(ctx, channels.SignalNameStartAssemblyChannel)
	// completeAssemblyChannel := workflow.GetSignalChannel(ctx, channels.SignalNameCompleteAssemblyChannel)
	// changeAssemblyCommentChannel := workflow.GetSignalChannel(ctx, channels.SignalNameChangeAssemblyCommentChannel)
	// completeDeliveryChannel := workflow.GetSignalChannel(ctx, channels.SignalNameCompleteDeliveryChannel)
	// changeDeliveryCommentChannel := workflow.GetSignalChannel(ctx, channels.SignalNameChangeDeliveryCommentChannel)
	// cancelOrderChannel := workflow.GetSignalChannel(ctx, channels.SignalNameCancelOrderChannel)

	// Идем в OMS Core и понимаем какой тип заказа перед нами, какие у него свойства и состав
	// и прочие значимые для принятия решения характеристики
	// На основании этих данных запускаем child workflows
	// Например:
	// Если заказ надо собирать, запускаем процедуру сборки и ожидаем завершения сборки
	// Если заказ с доставкой(не самовывоз) то передаем заказ на доставку и ожиадаем ее завершения
	// Если заказ требует оплаты то делаем необходимые действия и завершаем заказ
	// и т.д.

	IsNeedAssembly := true

	if IsNeedAssembly {
		// start assembly
		// wait for assembly completion
		// complete assembly
		// change assembly comment
	}

	w.OrderProcessingState.State = "processing_completed"
	w.pushStatus(ctx, w.OrderProcessingState.State)

	return nil
}
