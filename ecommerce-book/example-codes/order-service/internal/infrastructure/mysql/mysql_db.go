package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"order-service/internal/infrastructure/logger"
	"order-service/internal/model"
)

type MySQLDB struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewMySQLDBFromEnv(logger *logger.Logger) (*MySQLDB, error) {
	dsn := os.Getenv("ORDER_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/order_service?parseTime=true&charset=utf8mb4&loc=Local"
	}

	// The mysql driver must be imported by the program at runtime.
	rawDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql failed: %w", err)
	}
	rawDB.SetMaxOpenConns(20)
	rawDB.SetMaxIdleConns(10)
	rawDB.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rawDB.PingContext(ctx); err != nil {
		_ = rawDB.Close()
		return nil, fmt.Errorf("ping mysql failed: %w", err)
	}

	mysqlDB := &MySQLDB{db: rawDB, logger: logger}
	if err := mysqlDB.initSchema(ctx); err != nil {
		_ = rawDB.Close()
		return nil, err
	}

	logger.Info("infra.mysql", "connected mysql and schema is ready")
	return mysqlDB, nil
}

func (db *MySQLDB) Close() error {
	return db.db.Close()
}

func (db *MySQLDB) NextOrderID(ctx context.Context) (string, error) {
	result, err := db.db.ExecContext(ctx, "INSERT INTO order_id_seq(created_at) VALUES(?)", time.Now())
	if err != nil {
		return "", fmt.Errorf("insert order sequence failed: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("get order sequence id failed: %w", err)
	}
	return fmt.Sprintf("ORD-%d", id), nil
}

func (db *MySQLDB) SaveOrder(ctx context.Context, order model.Order) error {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("marshal order items failed: %w", err)
	}

	now := time.Now()
	_, err = db.db.ExecContext(
		ctx,
		`INSERT INTO orders (
			order_id, customer_id, items_json, total_cents, status, created_at, paid_at, cancelled_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			customer_id = VALUES(customer_id),
			items_json = VALUES(items_json),
			total_cents = VALUES(total_cents),
			status = VALUES(status),
			paid_at = VALUES(paid_at),
			cancelled_at = VALUES(cancelled_at),
			updated_at = VALUES(updated_at)`,
		order.ID, order.CustomerID, string(itemsJSON), order.TotalCents, string(order.Status),
		order.CreatedAt, order.PaidAt, order.CancelledAt, now,
	)
	if err != nil {
		return fmt.Errorf("save order failed: %w", err)
	}
	return nil
}

func (db *MySQLDB) GetOrder(ctx context.Context, id string) (model.Order, bool, error) {
	row := db.db.QueryRowContext(
		ctx,
		`SELECT order_id, customer_id, items_json, total_cents, status, created_at, paid_at, cancelled_at
		 FROM orders WHERE order_id = ?`,
		id,
	)

	order, err := scanOrder(row)
	if err == sql.ErrNoRows {
		return model.Order{}, false, nil
	}
	if err != nil {
		return model.Order{}, false, fmt.Errorf("get order failed: %w", err)
	}
	return order, true, nil
}

func (db *MySQLDB) ListPendingPaymentBefore(ctx context.Context, before time.Time, limit int) ([]model.Order, error) {
	rows, err := db.db.QueryContext(
		ctx,
		`SELECT order_id, customer_id, items_json, total_cents, status, created_at, paid_at, cancelled_at
		 FROM orders
		 WHERE status = ? AND created_at < ?
		 ORDER BY created_at ASC
		 LIMIT ?`,
		string(model.OrderStatusPendingPayment), before, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list timeout orders failed: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order failed: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders failed: %w", err)
	}
	return orders, nil
}

func (db *MySQLDB) initSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS orders (
			order_id VARCHAR(64) PRIMARY KEY,
			customer_id VARCHAR(64) NOT NULL,
			items_json JSON NOT NULL,
			total_cents BIGINT NOT NULL,
			status VARCHAR(32) NOT NULL,
			created_at DATETIME(6) NOT NULL,
			paid_at DATETIME(6) NULL,
			cancelled_at DATETIME(6) NULL,
			updated_at DATETIME(6) NOT NULL,
			INDEX idx_orders_status_created_at (status, created_at)
		)`,
		`CREATE TABLE IF NOT EXISTS order_id_seq (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			created_at DATETIME(6) NOT NULL
		)`,
	}

	for _, sqlStmt := range statements {
		if _, err := db.db.ExecContext(ctx, sqlStmt); err != nil {
			return fmt.Errorf("init schema failed: %w", err)
		}
	}
	return nil
}

type orderScanner interface {
	Scan(dest ...any) error
}

func scanOrder(s orderScanner) (model.Order, error) {
	var (
		order       model.Order
		itemsJSON   string
		status      string
		paidAt      sql.NullTime
		cancelledAt sql.NullTime
	)

	if err := s.Scan(
		&order.ID,
		&order.CustomerID,
		&itemsJSON,
		&order.TotalCents,
		&status,
		&order.CreatedAt,
		&paidAt,
		&cancelledAt,
	); err != nil {
		return model.Order{}, err
	}

	if err := json.Unmarshal([]byte(itemsJSON), &order.Items); err != nil {
		return model.Order{}, fmt.Errorf("unmarshal order items failed: %w", err)
	}
	order.Status = model.OrderStatus(status)
	if paidAt.Valid {
		order.PaidAt = &paidAt.Time
	}
	if cancelledAt.Valid {
		order.CancelledAt = &cancelledAt.Time
	}
	return order, nil
}
