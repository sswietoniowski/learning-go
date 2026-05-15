package payments

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// StubClient is a test double for the payments service.
// It tracks preauthorized payments and captured payments in memory.
type StubClient struct {
	mu       sync.Mutex
	preauths map[string]preauth // nonce -> preauth
	accounts map[string]*stubAccount
}

type preauth struct {
	CardNumber string
	Amount     decimal.Decimal
	Currency   string
}

type stubAccount struct {
	Number     string
	MerchantID string
	Balance    decimal.Decimal
	Card       stubCard
	History    []StubTransfer
}

type stubCard struct {
	CardNumber string
}

// StubTransfer records a balance change on an account.
type StubTransfer struct {
	Amount                decimal.Decimal
	Currency              string
	ExternalAccountNumber string
	Reference             string
	ReceiverDetails       string
}

func NewStub() *StubClient {
	return &StubClient{
		preauths: make(map[string]preauth),
		accounts: make(map[string]*stubAccount),
	}
}

// CapturePayment implements app.PaymentsService.
// It looks up the preauthorized payment by nonce and transfers the amount to the merchant's account.
func (s *StubClient) CapturePayment(_ context.Context, nonce string, amount decimal.Decimal, merchantID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pa, ok := s.preauths[nonce]
	if !ok {
		return fmt.Errorf("payment nonce %q not found", nonce)
	}

	// Find the cardholder and merchant accounts first
	var cardholderAcc, merchantAcc *stubAccount
	for _, acc := range s.accounts {
		if acc.Card.CardNumber == pa.CardNumber {
			cardholderAcc = acc
		}
		if acc.MerchantID == merchantID {
			merchantAcc = acc
		}
	}

	// Debit the card holder's account
	if cardholderAcc != nil {
		merchantAccountNumber := ""
		if merchantAcc != nil {
			merchantAccountNumber = merchantAcc.Number
		}
		cardholderAcc.Balance = cardholderAcc.Balance.Sub(amount)
		cardholderAcc.History = append(cardholderAcc.History, StubTransfer{
			Amount:                amount.Neg(),
			Currency:              pa.Currency,
			ExternalAccountNumber: merchantAccountNumber,
		})
	}

	// Credit the merchant's account
	if merchantAcc != nil {
		cardholderAccountNumber := ""
		if cardholderAcc != nil {
			cardholderAccountNumber = cardholderAcc.Number
		}
		merchantAcc.Balance = merchantAcc.Balance.Add(amount)
		merchantAcc.History = append(merchantAcc.History, StubTransfer{
			Amount:                amount,
			Currency:              pa.Currency,
			ExternalAccountNumber: cardholderAccountNumber,
		})
	}

	delete(s.preauths, nonce)
	return nil
}

// CreateAccount creates an in-memory bank account. Returns (accountNumber, cardNumber).
func (s *StubClient) CreateAccount(merchantID string, initialBalance decimal.Decimal) (string, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountNumber := "DE00STUB" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
	cardNumber := "4111" + uuid.NewString()[:12]

	s.accounts[accountNumber] = &stubAccount{
		Number:     accountNumber,
		MerchantID: merchantID,
		Balance:    initialBalance,
		Card:       stubCard{CardNumber: cardNumber},
	}

	return accountNumber, cardNumber
}

// PreauthorizePayment creates a preauth and returns a nonce.
func (s *StubClient) PreauthorizePayment(cardNumber string, amount decimal.Decimal, currency string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	nonce := uuid.NewString()
	s.preauths[nonce] = preauth{
		CardNumber: cardNumber,
		Amount:     amount,
		Currency:   currency,
	}

	return nonce
}

// GetAccountBalance returns the current balance of an account.
func (s *StubClient) GetAccountBalance(accountNumber string) (decimal.Decimal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc, ok := s.accounts[accountNumber]
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("account %q not found", accountNumber)
	}

	return acc.Balance, nil
}

// GetAccountHistory returns the transfer history for an account.
func (s *StubClient) GetAccountHistory(accountNumber string) ([]StubTransfer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc, ok := s.accounts[accountNumber]
	if !ok {
		return nil, fmt.Errorf("account %q not found", accountNumber)
	}

	result := make([]StubTransfer, len(acc.History))
	copy(result, acc.History)
	return result, nil
}

// RecordBankTransfer records a bank transfer: credits the target account and debits the source account.
// Called by the bank transfer stub to keep account balances in sync.
func (s *StubClient) RecordBankTransfer(targetAccount string, amount decimal.Decimal, currency string, sourceAccount string, reference string, receiver string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Credit target account
	if acc, ok := s.accounts[targetAccount]; ok {
		acc.Balance = acc.Balance.Add(amount)
		acc.History = append(acc.History, StubTransfer{
			Amount:                amount,
			Currency:              currency,
			ExternalAccountNumber: sourceAccount,
			Reference:             reference,
			ReceiverDetails:       receiver,
		})
	}

	// Debit source account
	if acc, ok := s.accounts[sourceAccount]; ok {
		acc.Balance = acc.Balance.Sub(amount)
		acc.History = append(acc.History, StubTransfer{
			Amount:                amount.Neg(),
			Currency:              currency,
			ExternalAccountNumber: targetAccount,
			Reference:             reference,
			ReceiverDetails:       receiver,
		})
	}
}
