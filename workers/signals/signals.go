package signals

import "github.com/milovidov983/oms-temporal-demo/shared/models"

type SignalPayloadCompleteAssembly struct {
	Route     string
	Collected []models.OrderItem
}

type SignalPayloadChangeAssemblyComment struct {
	Route   string
	Comment string
}

type SignalPayloadCompleteDelivery struct {
	Route     string
	Delivered []models.OrderItem
}

type SignalPayloadChangeDeliveryComment struct {
	Route   string
	Comment string
}

type SignalPayloadCancelOrder struct {
	Route  string
	Reason string
}
