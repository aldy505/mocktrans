package main

import "time"

type Expiry time.Duration

const (
	ExpiryShopee Expiry = Expiry(time.Hour * 1)
)

type ErrorStatusCode int

const (
	ErrorValidation          ErrorStatusCode = 400
	ErrorAccessDenied        ErrorStatusCode = 401
	ErrorNotFound            ErrorStatusCode = 404
	ErrorDuplicateOrderId    ErrorStatusCode = 406
	ErrorExpiredTransaction  ErrorStatusCode = 407
	ErrorWrongDataType       ErrorStatusCode = 408
	ErrorTooManyTransactions ErrorStatusCode = 409
	ErrorCannotModify        ErrorStatusCode = 412
	ErrorSyntaxInBody        ErrorStatusCode = 413
	ErrorRefundRejected      ErrorStatusCode = 414
)

type TransactionDetail struct {
	OrderId     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type BankTransfer struct {
	Bank     string         `json:"bank"`
	VaNumber string         `json:"va_number"`
	FreeText map[string]any `json:"free_text"`
	BCA      struct {
		SubCompanyCode string `json:"sub_company_code"`
	} `json:"bca"`
	Permata struct {
		RecipientName string `json:"recipient_name"`
	} `json:"permata"`
}

type CustomerDetail struct {
	// For BCA VA, limit the customer names (first_name and last_name), to only 30 characters.
	FirstName       string          `json:"first_name"`
	LastName        string          `json:"last_name"`
	Email           string          `json:"email"`
	Phone           string          `json:"phone"`
	BillingAddress  CustomerAddress `json:"billing_address"`
	ShippingAddress CustomerAddress `json:"shipping_address"`
}

type CustomerAddress struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	City        string `json:"city"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

// Please avoid using vertical line (`|`) for Alfamart payment type.
// item_details is required for Akulaku and Kredivo payment type.
// Subtotal (item price multiplied by quantity) of all the item details
// needs to be exactly same as the gross_amount inside the transaction_details object.
type ItemDetail struct {
	ID string `json:"id"`
	// You can pass a negative value for price to indicate discount.
	Price        int64  `json:"price"`
	Quantity     int64  `json:"quantity"`
	Name         string `json:"name"`
	Brand        string `json:"brand"`
	Category     string `json:"category"`
	MerchantName string `json:"merchant_name"`
	Tenor        string `json:"tenor"`
	CodePlan     string `json:"code_plan"`
	MID          string `json:"mid"`
	Url          string `json:"url"`
}

type SellerDetail struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Email   string          `json:"email"`
	Url     string          `json:"url"`
	Address CustomerAddress `json:"address"`
}

type PaymentAmount struct {
	PaidAt time.Time `json:"paid_at"`
	Amount string    `json:"amount"`
}

type Action struct {
	Name   string   `json:"name"`
	Method string   `json:"method"`
	Url    string   `json:"url"`
	Fields []string `json:"fields"`
}

type Gopay struct {
	EnableCallback     bool   `json:"enable_callback"`
	CallbackUrl        string `json:"callback_url"`
	AccountId          string `json:"account_id"`
	PaymentOptionToken string `json:"payment_option_token"`
	Recurring          bool   `json:"recurring"`
}

type Qris struct {
	Acquirer string `json:"acquirer"`
}

type Shopeepay struct {
	CallbackUrl string `json:"callback_url"`
}

type CreditCard struct {
	TokenId         string   `json:"token_id"`
	Bank            string   `json:"bank"`
	InstallmentTerm int32    `json:"installment_term"`
	Bins            []string `json:"bins"`
	Type            string   `json:"type"`
	SaveTokenId     bool     `json:"save_token_id"`
}

type CustomExpiry struct {
	OrderTime      time.Time `json:"order_time"`
	ExpiryDuration int64     `json:"expiry_duration"`
	// Possible values are second, minute, hour or day.
	// Default value is minute.
	Unit string `json:"unit"`
}

type NotificationRequest struct {
	CreditCardNotification
	VirtualAccountNotification
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionId     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderId           string `json:"order_id"`
	MerchantId        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
	Currency          string `json:"currency"`
}

type CreditCardNotification struct {
	MaskedCard             string `json:"masked_card,omitempty"`
	Eci                    string `json:"eci,omitempty"`
	ChannelResponseMessage string `json:"channel_response_message,omitempty"`
	ChannelResponseCode    string `json:"channel_response_code,omitempty"`
	CardType               string `json:"card_type,omitempty"`
	Bank                   string `json:"bank,omitempty"`
	ApprovalCode           string `json:"approval_code,omitempty"`
}

type VirtualAccountNumbers struct {
	VaNumber string `json:"va_number"`
	Bank     string `json:"bank"`
}

type VirtualAccountNotification struct {
	VaNumbers      []VirtualAccountNumbers `json:"va_numbers,omitempty"`
	SettlementTime string                  `json:"settlement_time,omitempty"`
	PaymentAmounts []string                `json:"payment_amounts,omitempty"`
}
