package scale

import (
	"time"

	"github.com/shopspring/decimal"
)

type BalanceOutput struct {
	AccountID int64           `json:"account_id"`
	BankID    string          `json:"bank_id"`
	Currency  string          `json:"currency"`
	IBAN      string          `json:"iban"`
	BIC       string          `json:"bic"`
	Balance   decimal.Decimal `json:"balance"`
}

type TransactionOutput struct {
	ID                 int64           `json:"id"`
	Date               time.Time       `json:"date"`
	Amount             decimal.Decimal `json:"amount"`
	Currency           string          `json:"currency"`
	Account            string          `json:"account"`
	AccountName        string          `json:"account_name"`
	BankName           string          `json:"bank_name"`
	BankCode           string          `json:"bank_code"`
	ConstantSymbol     string          `json:"constant_symbol"`
	VariableSymbol     string          `json:"variable_symbol"`
	SpecificSymbol     string          `json:"specific_symbol"`
	UserIdentification string          `json:"user_identification"`
	RecipientMessage   string          `json:"recipient_message"`
	Type               string          `json:"type"`
	Specification      string          `json:"specification"`
	Comment            string          `json:"comment"`
	BIC                string          `json:"bic"`
	OrderID            string          `json:"order_id"`
	PayerReference     string          `json:"payer_reference"`
}
