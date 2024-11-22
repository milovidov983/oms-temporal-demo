package models

import "time"

type AssemblyStatus string

const (
	AssemblyStatusNew      AssemblyStatus = "NEW"
	AssemblyStatusCreated  AssemblyStatus = "CREATED"
	AssemblyStatusSent     AssemblyStatus = "SENT"
	AssemblyStatusComplete AssemblyStatus = "COMPLETE"
	AssemblyStatusCanceled AssemblyStatus = "CANCELED"
)

type AssemblyApplication struct {
	ID        string         `json:"id"`
	OrderID   string         `json:"order_id"`
	Items     []AssemblyItem `json:"items"`
	Status    AssemblyStatus `json:"status"`
	Comment   string         `json:"comment"`
	CreatedAt time.Time      `json:"created_at"`
}

type AssemblyItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}
