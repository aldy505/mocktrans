package main

import (
	"context"
	"net/http"
)

func (d *Dependencies) Confirm(w http.ResponseWriter, r *http.Request) {
	transactionId := r.URL.Query().Get("transaction_id")
	if transactionId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Confirm the transaction
}

func (d *Dependencies) updateTransactionToPaid(ctx context.Context, transactionId string) error {

	return nil
}
