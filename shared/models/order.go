package models

import "time"

type OrderStatus string

const (
	OrderStatusNew              OrderStatus = "NEW"
	OrderStatusCreated          OrderStatus = "CREATED"
	OrderStatusPassedToAssembly OrderStatus = "PASSED_TO_ASSEMBLY"
	OrderStatusAssembled        OrderStatus = "ASSEMBLED"
	OrderStatusCanceled         OrderStatus = "CANCELED"
)

type Order struct {
	ID                    string      `json:"id"`
	CustomerID            string      `json:"customer_id"`
	Items                 []OrderItem `json:"items"`
	TotalAmount           float64     `json:"total_amount"`
	Status                OrderStatus `json:"status"`
	CreatedAt             time.Time   `json:"created_at"`
	AssemblyApplicationID string      `json:"assembly_application_id"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
