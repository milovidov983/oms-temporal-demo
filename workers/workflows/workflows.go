package workflows

import "fmt"

func OrderProcessingWorkflowID(orderID string) string {
	return fmt.Sprintf("OrderProcessing:%s", orderID)
}
