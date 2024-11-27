package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/milovidov983/oms-temporal-demo/shared/models"
)

type AssemblyApplicationRepository interface {
	Create(ctx context.Context, orderID string) (*models.AssemblyApplication, error)
	Complete(ctx context.Context, assemblyApplicationID string) (*models.AssemblyApplication, error)
	Cancel(ctx context.Context, assemblyApplicationID string) error
}

type assRepository struct {
	db *sql.DB
}

func NewAssemblyApplicationRepository(db *sql.DB) (AssemblyApplicationRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: database connection is required", ErrInvalidInput)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &assRepository{db: db}, nil
}

func (r *assRepository) Create(ctx context.Context, orderID string) (*models.AssemblyApplication, error) {
	if orderID == "" {
		return nil, fmt.Errorf("%w: order ID is required", ErrInvalidInput)
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction: %v", ErrDatabaseOperation, err)
	}
	defer r.rollbackOnError(tx, &err)

	exists, err := r.checkOrderExists(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%w: order ID %s", ErrOrderNotFound, orderID)
	}

	application, err := r.createAssemblyApplication(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}

	if err = r.updateOrder(ctx, tx, orderID, application.ID); err != nil {
		return nil, err
	}

	items, err := r.fetchOrderItems(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	application.Items = items

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction: %v", ErrDatabaseOperation, err)
	}

	return application, nil
}

func (r *assRepository) Complete(ctx context.Context, assemblyApplicationID string) (*models.AssemblyApplication, error) {
	if assemblyApplicationID == "" {
		return nil, fmt.Errorf("%w: assembly application ID is required", ErrInvalidInput)
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction: %v", ErrDatabaseOperation, err)
	}
	defer r.rollbackOnError(tx, &err)

	exists, err := r.checkAssemblyApplicationExists(ctx, tx, assemblyApplicationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%w: ID %s", ErrAssemblyApplicationNotFound, assemblyApplicationID)
	}

	if err = r.updateAssemblyStatus(ctx, tx, assemblyApplicationID, models.AssemblyStatusComplete); err != nil {
		return nil, err
	}

	if err = r.updateRelatedOrderStatus(ctx, tx, assemblyApplicationID); err != nil {
		return nil, err
	}

	application, err := r.fetchAssemblyApplication(ctx, tx, assemblyApplicationID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction: %v", ErrDatabaseOperation, err)
	}

	return application, nil
}

func (r *assRepository) Cancel(ctx context.Context, assemblyApplicationID string) error {
	if assemblyApplicationID == "" {
		return fmt.Errorf("%w: assembly application ID is required", ErrInvalidInput)
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("%w: failed to begin transaction: %v", ErrDatabaseOperation, err)
	}
	defer r.rollbackOnError(tx, &err)

	exists, err := r.checkAssemblyApplicationExists(ctx, tx, assemblyApplicationID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: ID %s", ErrAssemblyApplicationNotFound, assemblyApplicationID)
	}

	if err = r.updateAssemblyStatus(ctx, tx, assemblyApplicationID, models.AssemblyStatusCanceled); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%w: failed to commit transaction: %v", ErrDatabaseOperation, err)
	}

	return nil
}

// Вспомогательные методы

func (r *assRepository) rollbackOnError(tx *sql.Tx, err *error) {
	if *err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			*err = fmt.Errorf("rollback failed: %v, original error: %w", rbErr, *err)
		}
	}
}

func (r *assRepository) checkOrderExists(ctx context.Context, tx *sql.Tx, orderID string) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)`, orderID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check order existence: %v", ErrDatabaseOperation, err)
	}
	return exists, nil
}

func (r *assRepository) createAssemblyApplication(ctx context.Context, tx *sql.Tx, orderID string) (*models.AssemblyApplication, error) {
	application := &models.AssemblyApplication{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		Status:    models.AssemblyStatus(models.AssemblyStatusCreated),
		CreatedAt: time.Now(),
	}

	_, err := tx.ExecContext(ctx, `
        INSERT INTO assembly_applications (id, order_id, status, created_at)
        VALUES ($1, $2, $3, $4)
    `, application.ID, application.OrderID, application.Status, application.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("%w: failed to create assembly application: %v", ErrDatabaseOperation, err)
	}

	return application, nil
}

func (r *assRepository) updateOrder(ctx context.Context, tx *sql.Tx, orderID, assemblyApplicationID string) error {
	result, err := tx.ExecContext(ctx, `
        UPDATE orders
        SET assembly_application_id = $1, status = $2
        WHERE id = $3
    `, assemblyApplicationID, models.OrderStatusPassedToAssembly, orderID)

	if err != nil {
		return fmt.Errorf("%w: failed to update order: %v", ErrDatabaseOperation, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: failed to get rows affected: %v", ErrDatabaseOperation, err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: order ID %s", ErrOrderNotFound, orderID)
	}

	return nil
}

func (r *assRepository) fetchOrderItems(ctx context.Context, tx *sql.Tx, orderID string) ([]models.AssemblyItem, error) {
	rows, err := tx.QueryContext(ctx, `
        SELECT product_id, quantity
        FROM order_items
        WHERE order_id = $1
    `, orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch order items: %v", ErrDatabaseOperation, err)
	}
	defer rows.Close()

	var items []models.AssemblyItem
	for rows.Next() {
		var item models.AssemblyItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("%w: failed to scan order item: %v", ErrDatabaseOperation, err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: failed to iterate over rows: %v", ErrDatabaseOperation, err)
	}

	return items, nil
}

func (r *assRepository) checkAssemblyApplicationExists(ctx context.Context, tx *sql.Tx, assemblyID string) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM assembly_applications WHERE id = $1)`, assemblyID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%w: failed to check assembly application existence: %v", ErrDatabaseOperation, err)
	}
	return exists, nil
}

func (r *assRepository) updateAssemblyStatus(ctx context.Context, tx *sql.Tx, assemblyApplicationID string, status models.AssemblyStatus) error {
	result, err := tx.ExecContext(ctx, `
        UPDATE assembly_applications
        SET status = $1, completed_at = $2
        WHERE id = $3
    `, status, time.Now(), assemblyApplicationID)

	if err != nil {
		return fmt.Errorf("%w: failed to update assembly status: %v", ErrDatabaseOperation, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: failed to get rows affected: %v", ErrDatabaseOperation, err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: ID %s", ErrAssemblyApplicationNotFound, assemblyApplicationID)
	}

	return nil
}

func (r *assRepository) updateRelatedOrderStatus(ctx context.Context, tx *sql.Tx, assemblyApplicationID string) error {
	result, err := tx.ExecContext(ctx, `
        UPDATE orders
        SET status = $1
        WHERE assembly_application_id = $2
    `, models.OrderStatusAssembled, assemblyApplicationID)

	if err != nil {
		return fmt.Errorf("%w: failed to update order status: %v", ErrDatabaseOperation, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: failed to get rows affected: %v", ErrDatabaseOperation, err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: no order found for assembly application ID %s", ErrOrderNotFound, assemblyApplicationID)
	}

	return nil
}

func (r *assRepository) fetchAssemblyApplication(ctx context.Context, tx *sql.Tx, assemblyApplicationID string) (*models.AssemblyApplication, error) {
	application := &models.AssemblyApplication{}

	err := tx.QueryRowContext(ctx, `
        SELECT id, order_id, status, created_at, comment
        FROM assembly_applications
        WHERE id = $1
    `, assemblyApplicationID).Scan(
		&application.ID,
		&application.OrderID,
		&application.Status,
		&application.CreatedAt,
		&application.Comment,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: ID %s", ErrAssemblyApplicationNotFound, assemblyApplicationID)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch assembly application: %v", ErrDatabaseOperation, err)
	}

	items, err := r.fetchOrderItems(ctx, tx, application.OrderID)
	if err != nil {
		return nil, err
	}
	application.Items = items

	return application, nil
}
