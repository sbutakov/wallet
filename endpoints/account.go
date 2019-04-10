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

	"github.com/sbutakov/wallet/pkg/account"
)

type accountCreateRequest struct {
	Name     string  `json:"name"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

// AccountService interface for creating and viewing accounts
type AccountService interface {
	List() ([]*account.Account, error)
	Create(name, currency string, balance float64) (*account.Account, error)
}

// MakeAccountEndpoints init router for handling create and view accounts
func MakeAccountEndpoints(service AccountService, logger kitlog.Logger) http.Handler {
	router := chi.NewRouter()
	router.Method(http.MethodPost, "/", kithttp.NewServer(
		createAccount(service), decodeAccountCreateRequest, encodeAccountCreateResponse,
		[]kithttp.ServerOption{
			kithttp.ServerErrorLogger(logger),
			kithttp.ServerErrorEncoder(encodeAccountError),
		}...))

	router.Method(http.MethodGet, "/", kithttp.NewServer(
		listAccount(service), decodeListAccountsRequest, encodeListAccountResponse,
		[]kithttp.ServerOption{
			kithttp.ServerErrorLogger(logger),
			kithttp.ServerErrorEncoder(encodeAccountError),
		}...))

	return router
}

func createAccount(service AccountService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(accountCreateRequest)
		return service.Create(req.Name, req.Currency, req.Balance)
	}
}

func decodeAccountCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := accountCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(err, "error on decode request")
	}
	return req, nil
}

func encodeAccountCreateResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if err, ok := response.(error); ok {
		encodeAccountError(ctx, err, w)
		return nil
	}
	resp := response.(*account.Account)
	return json.NewEncoder(w).Encode(schemaResponse{
		Result: resp,
	})
}

func listAccount(service AccountService) endpoint.Endpoint {
	return func(_ context.Context, _ interface{}) (response interface{}, err error) {
		return service.List()
	}
}

func decodeListAccountsRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeListAccountResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if err, ok := response.(error); ok {
		encodeAccountError(ctx, err, w)
		return nil
	}
	resp := response.([]*account.Account)
	return json.NewEncoder(w).Encode(schemaResponse{
		Result: resp,
	})
}

func encodeAccountError(_ context.Context, err error, w http.ResponseWriter) {
	switch err {
	case account.ErrorUnsupportedCurrency:
	case account.ErrorBalanceValue:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(schemaResponse{ // nolint: errcheck
		Error: err.Error(),
	})
}
