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
