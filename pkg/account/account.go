// Package account provides methods for creating and viewing accounts
package account

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrorNotFound account not found
	ErrorNotFound = errors.New("account not found")
	// ErrorBalanceValue balance must be greater than zero
	ErrorBalanceValue = errors.New("balance must be greater than zero")
	// ErrorUnsupportedCurrency unsupported currency type
	ErrorUnsupportedCurrency = errors.New("unsupported currency type")
)

// Storage interface for creating and viewing account in database
type Storage interface {
	CreateAccount(name, currency string, balance float64) (*Account, error)
	ListAccount() ([]*Account, error)
}

// Account base type of package
type Account struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

// Config configuration params of account service
type Config struct {
	AllowedCurrency []string
}

// Service handles with accounts
type Service struct {
	storage  Storage
	currency []string
}

// New is constructor
func New(config Config, storage Storage) (*Service, error) {
	if len(config.AllowedCurrency) == 0 {
		return nil, errors.New("allowed currency cannot be empty")
	}

	currencies := make([]string, len(config.AllowedCurrency))
	for i, currency := range config.AllowedCurrency {
		currencies[i] = strings.ToLower(currency)
	}

	return &Service{
		storage:  storage,
		currency: currencies,
	}, nil
}

// Create create account and store in database
func (s *Service) Create(name, currency string, balance float64) (*Account, error) {
	if !contains(currency, s.currency) {
		return nil, ErrorUnsupportedCurrency
	}
	if balance <= 0 {
		return nil, ErrorBalanceValue
	}

	account, err := s.storage.CreateAccount(name, currency, balance)
	if err != nil {
		return nil, errors.Wrap(err, "error on create account")
	}
	return account, nil
}

// List view accounts stored in database
func (s *Service) List() ([]*Account, error) {
	return s.storage.ListAccount()
}

func contains(str string, arr []string) bool {
	for _, item := range arr {
		if str == item {
			return true
		}
	}
	return false
}
