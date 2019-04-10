package account

import "testing"

type dummyStorage struct {
}

func (d *dummyStorage) CreateAccount(name, currency string, balance float64) (*Account, error) {
	return &Account{}, nil
}

func (d *dummyStorage) ListAccount() ([]*Account, error) {
	return nil, nil
}

func TestNew(t *testing.T) {
	_, err := New(Config{AllowedCurrency: []string{}}, nil)
	if err == nil {
		t.Error("expected error on create instance")
	}
}

func TestService_Create(t *testing.T) {
	storage := &dummyStorage{}
	instance, err := New(Config{AllowedCurrency: []string{"usd", "eur"}}, storage)
	if err != nil {
		t.Error("unexpected error on create instance")
	}

	if instance == nil {
		t.Fatal("unexpected nil pointer instance")
	}

	_, err = instance.Create("dummy", "usd", 1)
	if err != nil {
		t.Error("unexpected error on create account")
	}

	_, err = instance.Create("dummy", "rub", 1)
	if err != ErrorUnsupportedCurrency {
		t.Error("error on check currency")
	}

	_, err = instance.Create("dummy", "eur", 0)
	if err != ErrorBalanceValue {
		t.Error("error on check balance")
	}
}
