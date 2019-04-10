// Package payment provides methods for transfer and viewing payments
package payment

import (
	"errors"
	"time"

	"github.com/sbutakov/wallet/pkg/account"
)

var (
	// ErrorDifferentCurrencies different currencies
	ErrorDifferentCurrencies = errors.New("different currencies")
	// ErrorIncorrectAmount amount must be greater than zero
	ErrorIncorrectAmount = errors.New("amount must be greater than zero")
	// ErrorNotEnoughMoney not enough money
	ErrorNotEnoughMoney = errors.New("not enough money")
	// ErrorTransferYourself sending to yourself
	ErrorTransferYourself = errors.New("sending to yourself")
	// ErrorMoneyTransfer error money transfer
	ErrorMoneyTransfer = errors.New("error money transfer")
)

// Payment base type of package
type Payment struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	AccountTo   string    `json:"account_to"`
	AccountFrom string    `json:"account_from"`
	Direction   string    `json:"direction"`
	CreatedAt   time.Time `json:"created_at"`
}

// Storage interface transfer, assert account and view payments
type Storage interface {
	PaymentList() ([]*Payment, error)
	AssertAccount(id string) (*account.Account, error)
	TransferMoney(accountFrom, accountTo string, amount float64) (*Payment, error)
}

// Service handles with payments
type Service struct {
	storage Storage
}

// New is constructor
func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// TransferMoney transfer money between accounts and register transactions in database
func (s *Service) TransferMoney(accountFromID, accountToID string, amount float64) (*Payment, error) {
	if accountFromID == accountToID {
		return nil, ErrorTransferYourself
	}

	if amount <= 0 {
		return nil, ErrorIncorrectAmount
	}

	accountFrom, err := s.storage.AssertAccount(accountFromID)
	if err != nil {
		return nil, account.ErrorNotFound
	}

	accountTo, err := s.storage.AssertAccount(accountToID)
	if err != nil {
		return nil, account.ErrorNotFound
	}

	if accountFrom.Currency != accountTo.Currency {
		return nil, ErrorDifferentCurrencies
	}

	return s.storage.TransferMoney(accountFrom.ID, accountTo.ID, amount)
}

// PaymentList view payments stored in database
func (s *Service) PaymentList() ([]*Payment, error) {
	return s.storage.PaymentList()
}
