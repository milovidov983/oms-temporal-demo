package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/milovidov983/oms-temporal-demo/shared/models"

	_ "github.com/lib/pq"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(connStr string) (*OrderRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &OrderRepository{db: db}, nil
}

func (r *OrderRepository) SaveOrder(ctx context.Context, order *models.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rollbackErr, err)
			}
		}
	}()
	orderQuery := `
		INSERT INTO orders (
			id, 
			customer_id, 
			total_amount, 
			status, 
			created_at,
			assembly_application_id
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.ExecContext(ctx, orderQuery,
		order.ID,
		order.CustomerID,
		order.TotalAmount,
		order.Status,
		order.CreatedAt,
		order.AssemblyApplicationID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	if len(order.Items) > 0 {
		itemQueryBase := `
			INSERT INTO order_items (
				order_id, 
				product_id, 
				quantity, 
				price
			) VALUES `

		valueStrings := make([]string, 0, len(order.Items))
		valueArgs := make([]interface{}, 0, len(order.Items)*4)

		for i, item := range order.Items {
			n := i * 4
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4))

			valueArgs = append(valueArgs, order.ID)
			valueArgs = append(valueArgs, item.ProductID)
			valueArgs = append(valueArgs, item.Quantity)
			valueArgs = append(valueArgs, item.Price)
		}

		itemQuery := itemQueryBase + strings.Join(valueStrings, ",")

		_, err = tx.ExecContext(ctx, itemQuery, valueArgs...)
		if err != nil {
			return fmt.Errorf("failed to bulk insert order items: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return err
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) error {
	query := `
        UPDATE orders
        SET status = $1
        WHERE id = $2
    `
	_, err := r.db.ExecContext(ctx, query, status, orderID)
	return err
}

func (r *OrderRepository) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	query := `
        SELECT id, customer_id, total_amount, status, created_at
        FROM orders
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, orderID)

	var order models.Order
	err := row.Scan(&order.ID, &order.CustomerID, &order.TotalAmount, &order.Status, &order.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) Close() error {
	return r.db.Close()
}

func (r *OrderRepository) GetOrderStatus(ctx context.Context, orderID string) (string, error) {
	query := `
        SELECT status
        FROM orders
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, orderID)

	var status string
	err := row.Scan(&status)
	if err != nil {
		return "", err
	}

	return status, nil
}

func (r *OrderRepository) SaveAssemblyApplicationID(ctx context.Context, orderID, assemblyApplicationID string) error {
	query := `
        UPDATE orders SET assembly_application_id = $1 WHERE id = $2
    `
	_, err := r.db.ExecContext(ctx, query, assemblyApplicationID, orderID)

	return err
}
