package main

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	klog "github.com/go-kit/kit/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/sbutakov/wallet/config"
	"github.com/sbutakov/wallet/endpoints"
	"github.com/sbutakov/wallet/pkg/account"
	"github.com/sbutakov/wallet/pkg/payment"
	"github.com/sbutakov/wallet/pkg/postgres"
)

var (
	version = "unknown"
	built   = "unknown"
)

func main() {
	log.Logger = zerolog.New(os.Stderr).With().
		Str("@version", version).
		Str("@built", built).
		Timestamp().
		Logger()

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorFieldName = "error"
	zerolog.MessageFieldName = "message"
	zerolog.LevelFieldName = "level"
	zerolog.TimestampFieldName = "@timestamp"
	level, _ := zerolog.ParseLevel(zerolog.ErrorLevel.String())
	zerolog.SetGlobalLevel(level)

	kitlog := klog.NewLogfmtLogger(log.Logger)
	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		log.Panic().
			Err(err).
			Msg("error on load config from env")
	}

	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		log.Panic().
			Err(err).
			Msg("error on connect to database server")
	}

	accountsService, err := account.New(cfg.Account, db)
	if err != nil {
		log.Panic().
			Err(err).
			Msg("error on init account service")
	}
	paymentService := payment.New(db)
	router := chi.NewRouter()
	router.Mount("/accounts", endpoints.MakeAccountEndpoints(accountsService, kitlog))
	router.Mount("/payments", endpoints.MakePaymentEndpoints(paymentService, kitlog))
	if err := http.ListenAndServe(cfg.Service.ListenAddress, router); err != nil {
		log.Panic().
			Err(err).
			Msg("error on listen and serve")
	}
}
