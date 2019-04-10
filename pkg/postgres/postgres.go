// Package postgres provides methods for storing accounts and payments
package postgres

import (
	"database/sql"
	"io/ioutil"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	"github.com/sbutakov/wallet/pkg/account"
	"github.com/sbutakov/wallet/pkg/payment"
)

const (
	defaultMaxIdleConnections = 10
	defaultMaxOpenConnections = 10
	defaultConnectionLifeTime = time.Minute

	paymentOutgoingDirection = "outgoing"
	paymentIncomingDirection = "incoming"

	errorCodeConnectionFailure = "08006"
)

// ErrorBadConnection connection failure
type ErrorBadConnection struct {
	parent error
	msg    string
}

func (e *ErrorBadConnection) Error() string {
	return e.msg
}

// Config configuration params for connect to database server
type Config struct {
	DSN                string
	FilePath           string
	MaxIdleConnections int
	MaxOpenConnections int
	ConnectionLifeTime time.Duration
}

// Postgres execute query to database
type Postgres struct {
	connection *sql.DB
}

// New is constructor
func New(config Config) (*Postgres, error) {
	connection, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "error on open connection")
	}

	if config.MaxIdleConnections == 0 {
		config.MaxIdleConnections = defaultMaxIdleConnections
	}

	if config.MaxOpenConnections == 0 {
		config.MaxOpenConnections = defaultMaxOpenConnections
	}

	if config.ConnectionLifeTime == 0 {
		config.ConnectionLifeTime = defaultConnectionLifeTime
	}

	connection.SetMaxIdleConns(config.MaxIdleConnections)
	connection.SetMaxOpenConns(config.MaxOpenConnections)
	connection.SetConnMaxLifetime(config.ConnectionLifeTime)
	postgres := &Postgres{
		connection: connection,
	}

	if config.FilePath != "" {
		initQuery, err := ioutil.ReadFile(config.FilePath)
		if err != nil {
			return nil, errors.Wrap(err, "error on read initialize sql script")
		}

		err = postgres.beginTransaction(func(tx *sql.Tx) error {
			_, err = connection.Exec(string(initQuery))
			return err
		})
		if err != nil {
			return nil, errors.Wrap(err, "error on create tables")
		}
	}

	return postgres, nil
}

func (p *Postgres) beginTransaction(doQuery func(tx *sql.Tx) error) error {
	tx, err := p.connection.Begin()
	if err != nil {
		return errors.Wrap(err, "error on begin transaction")
	}
	defer tx.Rollback() // nolint: errcheck

	if err = doQuery(tx); err != nil {
		if e, ok := err.(*pq.Error); ok && e.Code == errorCodeConnectionFailure {
			return &ErrorBadConnection{parent: err, msg: e.Message}
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "error on commit transaction")
	}
	return nil
}

// CreateAccount create account
func (p *Postgres) CreateAccount(name, currency string, balance float64) (*account.Account, error) {
	acc := new(account.Account)
	return acc, p.beginTransaction(func(tx *sql.Tx) error {
		id := uuid.NewV4().String()
		q := "INSERT INTO accounts(id,name,currency,balance) VALUES($1, $2, $3, $4)"
		_, err := tx.Exec(q, id, name, currency, balance)
		if err != nil {
			return err
		}

		q = "SELECT id,name,currency,balance,created_at FROM accounts WHERE id=$1"
		row := tx.QueryRow(q, id)
		if row == nil {
			return account.ErrorNotFound
		}
		return row.Scan(
			&acc.ID,
			&acc.Name,
			&acc.Currency,
			&acc.Balance,
			&acc.CreatedAt,
		)
	})
}

// ListAccount return stored accounts
func (p *Postgres) ListAccount() ([]*account.Account, error) {
	var accounts []*account.Account
	return accounts, p.beginTransaction(func(tx *sql.Tx) error {
		q := "SELECT id,name,currency,balance,created_at FROM accounts ORDER BY created_at"
		rows, err := tx.Query(q)
		if err != nil {
			return err
		}

		defer rows.Close()
		for rows.Next() {
			account := new(account.Account)
			err = rows.Scan(
				&account.ID,
				&account.Name,
				&account.Currency,
				&account.Balance,
				&account.CreatedAt,
			)
			if err != nil {
				return err
			}
			accounts = append(accounts, account)
		}
		return nil
	})
}

// AssertAccount assert account stored in database
func (p *Postgres) AssertAccount(id string) (*account.Account, error) {
	acc := new(account.Account)
	return acc, p.beginTransaction(func(tx *sql.Tx) error {
		q := "SELECT id,name,currency,balance,created_at FROM accounts WHERE id=$1"
		rows := tx.QueryRow(q, id)
		if rows == nil {
			return account.ErrorNotFound
		}

		return rows.Scan(
			&acc.ID,
			&acc.Name,
			&acc.Currency,
			&acc.Balance,
			&acc.CreatedAt,
		)
	})
}

// TransferMoney transfer money between accounts
func (p *Postgres) TransferMoney(accountFrom, accountTo string, amount float64) (*payment.Payment, error) {
	paymentResult := new(payment.Payment)
	return paymentResult, p.beginTransaction(func(tx *sql.Tx) error {
		q := "UPDATE accounts SET balance = balance - $1 WHERE balance >= $1 AND id=$2"
		res, err := tx.Exec(q, amount, accountFrom)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return payment.ErrorNotEnoughMoney
		}

		q = "UPDATE accounts SET balance = balance + $1 WHERE id=$2"
		if _, err = tx.Exec(q, amount, accountTo); err != nil {
			return err
		}

		outgoingTransactUUID := uuid.NewV4().String()
		q = "INSERT INTO payments(id, account, account_to,amount,direction) " +
			"VALUES($1, $2, $3, $4, $5)"
		_, err = tx.Exec(q,
			outgoingTransactUUID,
			accountFrom,
			accountTo,
			amount,
			paymentOutgoingDirection)
		if err != nil {
			return err
		}

		incomingTransactUUID := uuid.NewV4().String()
		_, err = tx.Exec(q,
			incomingTransactUUID,
			accountTo,
			accountFrom,
			amount,
			paymentIncomingDirection)
		if err != nil {
			return err
		}

		q = "SELECT id,account,account_to,amount,direction,created_at FROM payments WHERE id=$1"
		row := tx.QueryRow(q, outgoingTransactUUID)
		if row == nil {
			return payment.ErrorMoneyTransfer
		}

		return row.Scan(
			&paymentResult.ID,
			&paymentResult.AccountFrom,
			&paymentResult.AccountTo,
			&paymentResult.Amount,
			&paymentResult.Direction,
			&paymentResult.CreatedAt,
		)
	})
}

// PaymentList returned payments stored in database
func (p *Postgres) PaymentList() ([]*payment.Payment, error) {
	var payments []*payment.Payment
	return payments, p.beginTransaction(func(tx *sql.Tx) error {
		q := "SELECT id, account, account_to, amount, direction, created_at FROM payments"
		rows, err := tx.Query(q)
		if err != nil {
			return err
		}

		defer rows.Close()
		for rows.Next() {
			res := new(payment.Payment)
			err = rows.Scan(
				&res.ID,
				&res.AccountFrom,
				&res.AccountTo,
				&res.Amount,
				&res.Direction,
				&res.CreatedAt,
			)
			if err != nil {
				return errors.Wrap(err, "error scan row")
			}

			payments = append(payments, res)
		}

		return nil
	})
}
