package endpoints

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-kit/kit/log"

	"github.com/sbutakov/wallet/pkg/account"
)

type dummyStorage struct {
}

func (d *dummyStorage) CreateAccount(name, currency string, balance float64) (*account.Account, error) {
	return &account.Account{}, nil
}

func (d *dummyStorage) AssertAccount(id string) (*account.Account, error) {
	return &account.Account{}, nil
}

func (d *dummyStorage) ListAccount() ([]*account.Account, error) {
	return nil, nil
}

func TestMakeAccountEndpoints(t *testing.T) {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "wallet", log.DefaultTimestampUTC)

	storage := &dummyStorage{}
	service, err := account.New(account.Config{AllowedCurrency: []string{"usd"}}, storage)
	server := httptest.NewServer(MakeAccountEndpoints(service, logger))

	body, err := json.Marshal(accountCreateRequest{
		Name:     "dummy",
		Balance:  100.21,
		Currency: "usd",
	})
	if err != nil {
		t.Fatal("unexpected marshal error")
	}

	response, err := http.Post(server.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal("unexpected error on request")
	}

	resp := schemaResponse{}
	if err = json.NewDecoder(response.Body).Decode(&resp); err != nil {
		t.Fatal("error on decode response")
	}

	if resp.Error != nil {
		t.Error("unexpected error on response")
	}

	if resp.Result == nil {
		t.Error("unexpected nil no result")
	}
}
