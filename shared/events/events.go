package events

import "github.com/milovidov983/oms-temporal-demo/shared/models"

type EventType string

const (
	OrderCreated   EventType = "OrderCreated"
	OrderCancelled EventType = "OrderCancelled"
)

type OrderEvent struct {
	EventType EventType    `json:"eventType"`
	EventData models.Order `json:"eventData"`
}

const (
	AssemblyCreated   EventType = "AssemblyCreated"
	AssemblyCompleted EventType = "AssemblyCompleted"
	AssemblyCancelled EventType = "AssemblyCancelled"
)

type AssemblyApplicationEvent struct {
	EventType EventType         `json:"eventType"`
	EventData AssemblyEventData `json:"eventData"`
}

type AssemblyEventData struct {
	ID        string `json:"id"`
	OrderID   string `json:"orderId"`
	Collected []models.OrderItem
}
