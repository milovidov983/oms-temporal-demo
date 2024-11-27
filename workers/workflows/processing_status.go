package workflows

type OrderProcessingStatus int

const (
	OrderStatusUndefined = iota
	OrderStatusCreated
	OrderStatusTransferredToAssembly
	OrderStatusAssemblyInProgress
	OrderStatusAssembled
	OrderStatusTransferredToDelivery
	OrderStatusDeliveryInProgress
	OrderStatusDelivered
	OrderStatusCanceled
	OrderStatusProcessingCompleted
)

var statusName = map[OrderProcessingStatus]string{
	OrderStatusUndefined:             "undefined",
	OrderStatusCreated:               "created",
	OrderStatusTransferredToAssembly: "transferred_to_assembly",
	OrderStatusAssemblyInProgress:    "assembly_in_progress",
	OrderStatusAssembled:             "assembled",
	OrderStatusTransferredToDelivery: "transferred_to_delivery",
	OrderStatusDeliveryInProgress:    "delivery_in_progress",
	OrderStatusDelivered:             "delivered",
	OrderStatusCanceled:              "canceled",
	OrderStatusProcessingCompleted:   "order_processing_completed",
}

func (os OrderProcessingStatus) String() string {
	return statusName[os]
}

var finalOrderStatuses = map[OrderProcessingStatus]bool{
	OrderStatusCanceled:            true,
	OrderStatusDelivered:           true,
	OrderStatusProcessingCompleted: true,
}

func (os OrderProcessingStatus) IsFinalStatus() bool {
	value := finalOrderStatuses[os]
	return value
}
