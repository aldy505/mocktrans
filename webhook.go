package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var backoffSchedule = []time.Duration{time.Minute * 2, time.Minute * 10, time.Minute * 30, time.Minute * 90, time.Hour*3 + time.Minute*30}

func (d *Dependencies) SendWebhook(transactionId string, content NotificationRequest) error {
	var amountLeftToRetry = 0

	jsonPayload, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*4)
	defer cancel()

	for _, backoff := range backoffSchedule {
		httpCtx, httpCancel := context.WithTimeout(ctx, time.Minute*3)
		defer httpCancel()

		statusCode, err := d.sendHttpRequest(httpCtx, d.CallbackUrl, bytes.NewReader(jsonPayload))
		if err != nil {
			return fmt.Errorf("failed to send webhook: %w", err)
		}

		err = d.writeWebhookHistoryLog(ctx, statusCode == http.StatusOK, content)
		if err != nil {
			return fmt.Errorf("failed to write webhook history log: %w", err)
		}

		if statusCode <= 299 {
			break
		}

		if amountLeftToRetry == 0 {
			break
		}

		if amountLeftToRetry == 0 && statusCode == 500 {
			amountLeftToRetry = 0
			time.Sleep(backoff)
			continue
		}

		if amountLeftToRetry == 0 && statusCode == 503 {
			amountLeftToRetry = 3
			time.Sleep(backoff)
			continue
		}

		if amountLeftToRetry == 0 && (statusCode == 400 || statusCode == 404) {
			amountLeftToRetry = 1
			time.Sleep(backoff)
			continue
		}

		if amountLeftToRetry == 00 && statusCode != 301 && statusCode != 302 && statusCode != 303 && statusCode != 307 && statusCode != 308 {
			amountLeftToRetry = 4
			time.Sleep(backoff)
			continue
		}

		if amountLeftToRetry > 0 {
			amountLeftToRetry = amountLeftToRetry - 1
			time.Sleep(backoff)
			continue
		}
	}

	return nil
}

func (d *Dependencies) sendHttpRequest(ctx context.Context, url string, content io.Reader) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, content)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: time.Second * 20}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}

	return resp.StatusCode, nil
}

func (d *Dependencies) writeWebhookHistoryLog(ctx context.Context, success bool, content NotificationRequest) error {
	jsonPayload, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %w", err)
	}

	formattedQuery, err := d.formatPlaceholder(`INSERT INTO
		webhook_history
		(
			transaction_id,
			event_type,
			status,
			data,
			success,
			created_at
		)
	VALUES
		($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return fmt.Errorf("failed to format query: %w", err)
	}

	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		formattedQuery,
		content.TransactionId,
		content.TransactionStatus,
		jsonPayload,
		success,
		time.Now(),
	)
	if err != nil {
		if e := tx.Rollback(); e != nil && !errors.Is(err, sql.ErrTxDone) {
			return fmt.Errorf("failed to rollback transaction: %w", e)
		}

		return fmt.Errorf("failed to write webhook history log: %w", err)
	}

	err = tx.Commit()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
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
