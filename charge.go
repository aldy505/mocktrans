package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type chargeRequest struct {
	PaymentType       string                 `json:"payment_type"`
	TransactionDetail TransactionDetail      `json:"transaction_detail"`
	ItemDetails       []ItemDetail           `json:"item_details"`
	CustomerDetails   CustomerDetail         `json:"customer_details"`
	BankTransfer      BankTransfer           `json:"bank_transfer"`
	CustomExpiry      CustomExpiry           `json:"custom_expiry"`
	Metadata          map[string]interface{} `json:"metadata"`
	CustomField1      string                 `json:"custom_field_1"`
	CustomField2      string                 `json:"custom_field_2"`
	CustomField3      string                 `json:"custom_field_3"`
}

type chargeResponse struct {
	StatusCode        string   `json:"status_code"`
	StatusMessage     string   `json:"status_message"`
	TransactionId     string   `json:"transaction_id"`
	OrderId           string   `json:"order_id"`
	MerchantId        string   `json:"merchant_id,omitempty"`
	GrossAmount       string   `json:"gross_amount"`
	Currency          string   `json:"currency,omitempty"`
	PaymentType       string   `json:"payment_type"`
	TransactionTime   string   `json:"transaction_time"`
	TransactionStatus string   `json:"transaction_status"`
	FraudStatus       string   `json:"fraud_status"`
	ApprovalCode      string   `json:"approval_code,omitempty"`
	MaskedCard        string   `json:"masked_card,omitempty"`
	Bank              string   `json:"bank,omitempty"`
	Acquirer          string   `json:"acquirer,omitempty"`
	Actions           []Action `json:"actions,omitempty"`
}

func (d *Dependencies) Charge(w http.ResponseWriter, r *http.Request) {
	// Validate content type headers
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	// Parse request body
	var req chargeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(int(ErrorSyntaxInBody))
		w.Write([]byte(`{"status": "error", "message": ` + strconv.Quote(err.Error()) + `}`))
		return
	}

	// Validate request body
	errorStatus, reason := req.Validate()
	if errorStatus != 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(int(errorStatus))
		w.Write([]byte(`{"status": "error", "message": ` + strconv.Quote(reason) + `}`))
		return
	}
}

func (c chargeRequest) Validate() (ErrorStatusCode, string) {
	// Payment type must be one of:
	// - credit_card
	// - bank_transfer
	// - bca_klikpay
	// - bca_klikbca
	// - bri_epay
	// - cimb_clicks
	// - danamon_online
	// - uob_ezpay
	// - qris
	// - gopay
	// - shopeepay
	// - cstore
	// - akulaku
	// - kredivo
	var paymentTypeOk = true
	for _, validPaymentTypes := range []string{"credit_card", "bank_transfer", "bca_klikpay", "bca_klikbca", "bri_epay", "cimb_clicks", "danamon_online", "uob_ezpay", "qris", "gopay", "shopeepay", "cstore", "akulaku", "kredivo"} {
		if c.PaymentType == validPaymentTypes {
			paymentTypeOk = true
			break
		}
	}

	if !paymentTypeOk {
		return ErrorValidation, fmt.Sprintf("unknown payment type of %s", c.PaymentType)
	}

	// Order ID must exists and must not contain any other symbols other than dash(-), underscore(_), tilde (~), and dot (.)
	if c.TransactionDetail.OrderId == "" {
		return ErrorValidation, "order_id is required"
	}

	if strings.ContainsAny(c.TransactionDetail.OrderId, "!@#$%^&*()+=[]{}|;':,/<>?\\") {
		return ErrorValidation, "order_id must not contain any other symbols other than dash(-), underscore(_), tilde (~), and dot (.)"
	}

	// Gross Amount must be equal to Item Details total amount
	var totalAmount int64 = 0
	for _, item := range c.ItemDetails {
		totalAmount += item.Quantity * item.Price
	}

	if totalAmount != c.TransactionDetail.GrossAmount {
		return ErrorValidation, "gross_amount must be equal to Item Details total amount"
	}

	// Validate customer
	if len(c.CustomerDetails.Email) > 255 {
		return ErrorValidation, "customer_details.email must not exceed 255 characters"
	}

	if len(c.CustomerDetails.FirstName) > 255 {
		return ErrorValidation, "customer_details.first_name must not exceed 255 characters"
	}

	if len(c.CustomerDetails.LastName) > 255 {
		return ErrorValidation, "customer_details.last_name must not exceed 255 characters"
	}

	if len(c.CustomerDetails.Phone) > 255 {
		return ErrorValidation, "customer_details.phone must not exceed 255 characters"
	}

	// Validate customer's address
	if len(c.CustomerDetails.BillingAddress.Address) > 255 {
		return ErrorValidation, "customer_details.address.address must not exceed 255 characters"
	}

	if len(c.CustomerDetails.BillingAddress.City) > 255 {
		return ErrorValidation, "customer_details.address.city must not exceed 255 characters"
	}

	if len(c.CustomerDetails.BillingAddress.CountryCode) > 3 {
		return ErrorValidation, "customer_details.address.country_code must not exceed 3 characters"
	}

	// Validate customer's shipping address
	if len(c.CustomerDetails.ShippingAddress.Address) > 255 {
		return ErrorValidation, "customer_details.shipping_address.address must not exceed 255 characters"
	}

	if len(c.CustomerDetails.ShippingAddress.City) > 255 {
		return ErrorValidation, "customer_details.shipping_address.city must not exceed 255 characters"
	}

	if len(c.CustomerDetails.ShippingAddress.CountryCode) > 3 {
		return ErrorValidation, "customer_details.shipping_address.country_code must not exceed 3 characters"
	}

	switch c.PaymentType {
	case "bank_transfer":
		break
	case "bca_klikpay":
		for _, itemDetail := range c.ItemDetails {
			if itemDetail.Tenor == "" {
				return ErrorValidation, "tenor is required"
			}

			if len(itemDetail.Tenor) != 2 {
				return ErrorValidation, "tenor must be 2 digits"
			}

			_, err := strconv.Atoi(itemDetail.Tenor)
			if err != nil {
				return ErrorValidation, "tenor must be numeric"
			}

			if itemDetail.CodePlan == "" {
				return ErrorValidation, "code_plan is required"
			}

			if itemDetail.MID == "" {
				return ErrorValidation, "mid is required"
			}
		}
	}

	return 0, ""
}
