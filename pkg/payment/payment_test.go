package payment

import (
	"testing"

	"github.com/sbutakov/wallet/pkg/account"
)

type dummyStorage struct {
	accounts map[string]account.Account
}

func (d *dummyStorage) PaymentList() ([]*Payment, error) {
	return nil, nil
}

func (d *dummyStorage) AssertAccount(id string) (*account.Account, error) {
	if res, ok := d.accounts[id]; ok {
		return &res, nil
	}
	return nil, account.ErrorNotFound
}

func (d *dummyStorage) TransferMoney(string, string, float64) (*Payment, error) {
	return &Payment{}, nil
}

func TestService_TransferMoney(t *testing.T) {
	storage := &dummyStorage{
		accounts: map[string]account.Account{
			"dummy_from": {Currency: "usd"},
			"dummy_to":   {Currency: "usd"},
			"dummy_eur":  {Currency: "eur"},
		},
	}

	instance := New(storage)
	_, err := instance.TransferMoney("dummy_from", "dummy_to", 1)
	if err != nil {
		t.Error("unexpected error on transfer money")
	}

	_, err = instance.TransferMoney("dummy", "dummy", 1)
	if err != ErrorTransferYourself {
		t.Error("error on check accounts for transfer money")
	}

	_, err = instance.TransferMoney("dummy_from", "dummy_to", 0)
	if err != ErrorIncorrectAmount {
		t.Error("error on check correct amount")
	}

	_, err = instance.TransferMoney("dummy", "dummy_to", 1)
	if err != account.ErrorNotFound {
		t.Error("error on assert account_from")
	}

	_, err = instance.TransferMoney("dummy_from", "dummy", 1)
	if err != account.ErrorNotFound {
		t.Error("error on assert account_to")
	}

	_, err = instance.TransferMoney("dummy_from", "dummy_eur", 1)
	if err != ErrorDifferentCurrencies {
		t.Error("error on check equal currency")
	}
}
