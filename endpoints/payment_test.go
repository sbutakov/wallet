package endpoints

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sbutakov/wallet/pkg/payment"

	"github.com/go-kit/kit/log"
)

func (d *dummyStorage) TransferMoney(
	accountFromID, accountToID string, amount float64) (*payment.Payment, error) {

	return &payment.Payment{}, nil
}

func (d *dummyStorage) PaymentList() ([]*payment.Payment, error) {
	return nil, nil
}

func TestMakePaymentEndpoints(t *testing.T) {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "wallet", log.DefaultTimestampUTC)

	storage := &dummyStorage{}
	service := payment.New(storage)
	server := httptest.NewServer(MakePaymentEndpoints(service, logger))

	body, err := json.Marshal(transferMoneyRequest{
		AccountFrom: "dummy_from",
		AccountTo:   "dummy_to",
		Amount:      100.01,
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
