package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/pkg/errors"

	"github.com/sbutakov/wallet/pkg/payment"
)

type transferMoneyRequest struct {
	AccountFrom string  `json:"account_from"`
	AccountTo   string  `json:"account_to"`
	Amount      float64 `json:"amount"`
}

// PaymentService interface for transfer and viewing payments
type PaymentService interface {
	PaymentList() ([]*payment.Payment, error)
	TransferMoney(accountFromID, accountToID string, amount float64) (*payment.Payment, error)
}

// MakePaymentEndpoints init router for handling create and view payments
func MakePaymentEndpoints(service PaymentService, logger kitlog.Logger) http.Handler {
	router := chi.NewRouter()
	router.Method(http.MethodPost, "/", kithttp.NewServer(
		transferMoney(service), decodeTransferMoneyRequest, encodeTransferMoneyResponse,
		[]kithttp.ServerOption{
			kithttp.ServerErrorLogger(logger),
			kithttp.ServerErrorEncoder(encodePaymentError),
		}...))

	router.Method(http.MethodGet, "/", kithttp.NewServer(
		listPayments(service), decodeListPaymentsRequest, encodeListPaymentsResponse,
		[]kithttp.ServerOption{
			kithttp.ServerErrorLogger(logger),
			kithttp.ServerErrorEncoder(encodePaymentError),
		}...))

	return router
}

func transferMoney(service PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(transferMoneyRequest)
		return service.TransferMoney(req.AccountFrom, req.AccountTo, req.Amount)
	}
}

func decodeTransferMoneyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := transferMoneyRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(err, "error on decode request")
	}
	return req, nil
}

func encodeTransferMoneyResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if err, ok := response.(error); ok {
		encodePaymentError(ctx, err, w)
		return nil
	}
	resp := response.(*payment.Payment)
	return json.NewEncoder(w).Encode(schemaResponse{
		Result: resp,
	})
}

func listPayments(service PaymentService) endpoint.Endpoint {
	return func(_ context.Context, _ interface{}) (response interface{}, err error) {
		return service.PaymentList()
	}
}

func decodeListPaymentsRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeListPaymentsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if err, ok := response.(error); ok {
		encodeAccountError(ctx, err, w)
		return nil
	}
	resp := response.([]*payment.Payment)
	return json.NewEncoder(w).Encode(schemaResponse{
		Result: resp,
	})
}

func encodePaymentError(_ context.Context, err error, w http.ResponseWriter) {
	switch err {
	case payment.ErrorMoneyTransfer:
	case payment.ErrorTransferYourself:
	case payment.ErrorIncorrectAmount:
		w.WriteHeader(http.StatusBadRequest)

	case payment.ErrorDifferentCurrencies:
	case payment.ErrorNotEnoughMoney:
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(schemaResponse{ // nolint: errcheck
		Error: err.Error(),
	})
}
