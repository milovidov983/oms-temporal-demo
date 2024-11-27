// Yoy need to think how to inject dependencies to activities
// maybe it need to ask

package workflows

import (
	"github.com/milovidov983/oms-temporal-demo/shared/models"
	"github.com/milovidov983/oms-temporal-demo/workers/activities"
	"github.com/milovidov983/oms-temporal-demo/workers/signals/channels"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	OrderProcessingStatusQuery = "order-processing-status"
)

type OrderProcessingWorkflowInput struct {
	OrderID string
}

type OrderProcessingState struct {
	OrderID      string
	CurrentState OrderProcessingStatus
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
func (w *orderProcessingWorkflow) pushStatus(ctx workflow.Context, status OrderProcessingStatus) error {
	var keywordKey = temporal.NewSearchAttributeKeyKeyword("OrderProcessingStatus")
	return workflow.UpsertTypedSearchAttributes(
		ctx,
		keywordKey.ValueSet(status.String()),
	)
}

// ProcessOrder is a Workflow Definition that calls for the execution of a variable set of Activities and Child Workflows.
// This is the main entry point of the application.
// It accepts an Order ID as the input.
func ProcessOrder(ctx workflow.Context, input *OrderProcessingWorkflowInput) error {
	w := newOrderProcessingWorkflow(
		ctx,
		&OrderProcessingState{
			OrderID:      input.OrderID,
			CurrentState: OrderStatusCreated,
		},
	)

	w.logger.Info("Processing order", "order_id", w.OrderID)

	w.pushStatus(ctx, w.OrderProcessingState.CurrentState)

	err := workflow.SetQueryHandler(ctx, OrderProcessingStatusQuery, func() (OrderProcessingState, error) {
		return w.OrderProcessingState, nil
	})
	if err != nil {
		return err
	}

	// Channels
	startOrderProcessingChannel := workflow.GetSignalChannel(ctx, channels.SignalNameStartOrderProcessingChannel)
	// startAssemblyChannel := workflow.GetSignalChannel(ctx, channels.SignalNameStartAssemblyChannel)
	completeAssemblyChannel := workflow.GetSignalChannel(ctx, channels.SignalNameCompleteAssemblyChannel)
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

	for {
		s := workflow.NewSelector(ctx)
		// Signals handlers
		//
		//
		// The signal handlers check the correctness of the status for the signal
		// and the transfer of the order to a new status according to the business process
		//
		//

		// Signal handler for the assembly complete process
		s.AddReceive(completeAssemblyChannel, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)

			w.logger.Debug("Handling complete assembly channel")

			w.OrderProcessingState.CurrentState = OrderStatusAssembled
			w.pushStatus(ctx, w.OrderProcessingState.CurrentState)
		})
		s.AddReceive(startOrderProcessingChannel, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)

			w.logger.Debug("Handling start order processing channel")

			w.OrderProcessingState.CurrentState = OrderStatusCreated
			w.pushStatus(ctx, w.OrderProcessingState.CurrentState)
		})

		s.Select(ctx)

		w.logger.Debug("Handling order processing workflow", "current_state", w.OrderProcessingState.CurrentState)

		switch w.OrderProcessingState.CurrentState {
		case OrderStatusCreated: // Сборка
			err = w.handleNewOrder(ctx)
		case OrderStatusAssembled:
			err = w.handleAssembledOrder(ctx)

			// debug code
			w.OrderProcessingState.CurrentState = OrderStatusProcessingCompleted
		}

		if err != nil {
			w.logger.Error("Error to handle order", "error", err, "order_id", w.OrderID)
			break
		}
		if w.OrderProcessingState.CurrentState.IsFinalStatus() {
			break
		}
	}

	w.OrderProcessingState.CurrentState = OrderStatusProcessingCompleted
	w.pushStatus(ctx, w.OrderProcessingState.CurrentState)

	return nil
}

func (w *orderProcessingWorkflow) handleNewOrder(ctx workflow.Context) error {
	w.logger.Debug("Handle new order", "order_id", w.OrderID)

	input := &activities.Input{
		OrderID: w.OrderID,
	}
	var output []models.OrderType
	err := workflow.ExecuteActivity(ctx, a.GetOrderTypes, input).Get(ctx, output)
	if err != nil {
		w.logger.Error("Error to get order type", "error", err, "order_id", w.OrderID)
	}

	if len(output) == 0 {
		return nil
	}

	isNeedToAssembly := false
	for _, orderType := range output {
		if orderType == models.OrderTypeAssembly {
			isNeedToAssembly = true
			break
		}
	}

	if isNeedToAssembly {
		err = workflow.ExecuteActivity(ctx, a.CreateAssemblyApplication, input).Get(ctx, nil)
		if err != nil {
			w.logger.Error("Error to start assembly", "error", err, "order_id", w.OrderID)
		}
		w.OrderProcessingState.CurrentState = OrderStatusTransferredToAssembly
		w.pushStatus(ctx, w.OrderProcessingState.CurrentState)
		return nil
	}

	return nil
}
func (w *orderProcessingWorkflow) handleAssembledOrder(ctx workflow.Context) error {
	w.logger.Debug("Handle assembled order", "order_id", w.OrderID)
	// заказ собран если надо передаем на доставку отправляем нотификации и делаем остальные
	// действия согласно бизнес процессу

	return nil
}
