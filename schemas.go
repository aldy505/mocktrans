package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

func (d *Dependencies) MigrateSchema(ctx context.Context) error {
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			log.Printf("failed to close database connection: %V", err)
		}
	}()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS transactions (
			id VARCHAR(36) PRIMARY KEY,
			order_id VARCHAR(50) NOT NULL,
			payment_type VARCHAR(50) NOT NULL,
			gross_amount BIGINT NOT NULL,
			merchant_id VARCHAR(255),
			metadata TEXT,
			custom_field_1 VARCHAR(255),
			custom_field_2 VARCHAR(255),
			custom_field_3 VARCHAR(255),
			created_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS transactions_order_id_idx ON transactions (order_id)`,
		`CREATE TABLE IF NOT EXISTS transaction_virtual_account (
			id VARCHAR(36) PRIMARY KEY,
			transaction_id VARCHAR(36) NOT NULL,
			va_number VARCHAR(50) NOT NULL,
			bank VARCHAR(50) NOT NULL,
			created_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS transaction_virtual_account_transaction_id_idx ON transaction_virtual_account (transaction_id)`,
		`CREATE TABLE IF NOT EXISTS webhook_history (
			transaction_id VARCHAR(36) NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			data TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS webhook_history_transaction_id_idx ON webhook_history (transaction_id)`,
	}

	for _, query := range queries {
		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			if e := tx.Rollback(); e != nil && !errors.Is(err, sql.ErrTxDone) {
				return fmt.Errorf("failed to rollback transaction: %w", e)
			}
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil && !errors.Is(err, sql.ErrTxDone) {
			return fmt.Errorf("failed to rollback transaction: %w", e)
		}
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	err = conn.Close()
	if err != nil && !errors.Is(err, sql.ErrConnDone) {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}
