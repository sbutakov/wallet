package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-kit/kit/log"

	"github.com/sbutakov/wallet/config"
	"github.com/sbutakov/wallet/endpoints"
	"github.com/sbutakov/wallet/pkg/account"
	"github.com/sbutakov/wallet/pkg/payment"
	"github.com/sbutakov/wallet/pkg/postgres"
)

func main() {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "wallet", log.DefaultTimestampUTC)

	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		logger.Log(err) // nolint: gosec
		return
	}

	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		logger.Log(err) // nolint: gosec
		return
	}

	paymentService := payment.New(db)
	accountsService, err := account.New(cfg.Account, db)
	if err != nil {
		logger.Log(err) // nolint: gosec
		return
	}

	router := chi.NewRouter()
	router.Mount("/accounts", endpoints.MakeAccountEndpoints(accountsService, logger))
	router.Mount("/payments", endpoints.MakePaymentEndpoints(paymentService, logger))
	if err := http.ListenAndServe(cfg.Service.ListenAddress, router); err != nil {
		logger.Log(err) // nolint: gosec
	}
}
