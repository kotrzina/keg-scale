package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jbub/fio"
	"github.com/shopspring/decimal"
)

func ProvideFioTransactions(token string) (string, error) {
	client := fio.NewClient(token, nil)

	opts := fio.ByPeriodOptions{
		DateFrom: time.Now().Add(-14 * 24 * time.Hour),
		DateTo:   time.Now(),
	}

	resp, err := client.Transactions.ByPeriod(context.Background(), opts)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve transactions: %w", err)
	}

	type TransactionOutput struct {
		ID                 int64           `json:"ID"`
		Date               time.Time       `json:"Date"`
		Amount             decimal.Decimal `json:"Amount"`
		Currency           string          `json:"Currency"`
		Account            string          `json:"Account"`
		AccountName        string          `json:"AccountName"`
		BankName           string          `json:"BankName"`
		BankCode           string          `json:"BankCode"`
		ConstantSymbol     string          `json:"ConstantSymbol"`
		VariableSymbol     string          `json:"VariableSymbol"`
		SpecificSymbol     string          `json:"SpecificSymbol"`
		UserIdentification string          `json:"UserIdentification"`
		RecipientMessage   string          `json:"RecipientMessage"`
		Type               string          `json:"Type"`
		Specification      string          `json:"Specification"`
		Comment            string          `json:"Comment"`
		BIC                string          `json:"BIC"`
		OrderID            string          `json:"OrderID"`
		PayerReference     string          `json:"PayerReference"`
	}

	output := make([]TransactionOutput, len(resp.Transactions))
	for i, t := range resp.Transactions {
		output[i] = TransactionOutput{
			ID:                 t.ID,
			Date:               t.Date,
			Amount:             t.Amount,
			Currency:           t.Currency,
			Account:            t.Account,
			AccountName:        t.AccountName,
			BankName:           t.BankName,
			BankCode:           t.BankCode,
			ConstantSymbol:     t.ConstantSymbol,
			VariableSymbol:     t.VariableSymbol,
			SpecificSymbol:     t.SpecificSymbol,
			UserIdentification: t.UserIdentification,
			RecipientMessage:   t.RecipientMessage,
			Type:               t.Type,
			Specification:      t.Specification,
			Comment:            t.Comment,
			BIC:                t.BIC,
			OrderID:            t.OrderID,
			PayerReference:     t.PayerReference,
		}
	}

	resultJson, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("failed to marshal transactions: %w", err)
	}

	return fmt.Sprintf("Transactions in JSON format:\n\n```json\n%s\n```", string(resultJson)), nil
}
