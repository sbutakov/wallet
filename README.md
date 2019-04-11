# Wallet coins-challenge
The Wallet is the challenge by Coins.ph

## Description
The service that provides methods for create, view and sends payments between accounts

## Commands
- Build:
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags "-s -X 'main.version=$(git rev-parse --short HEAD)' -X 'main.built=$(date -u '+%Y-%m-%dT%H:%M:%SZ')'" -o wallet
```
- Test: `go test ./...`
- Lint:
```bash
gometalinter --vendor --line-length=100 \
        --exclude='error return value not checked.*(Close|Log|Print).*\(errcheck\)$' \
        --exclude='.*_test\.go' \
        --disable=dupl \
        --concurrency=2 \
        --deadline=300s \
        ./...
```

## Usage
 - Start service: `docker-compose -f deployments/docker-compose.yml up`
 - Create account:
 ```bash
curl -X POST http://localhost:8080/accounts -d '{
    "name":     "Alice",
    "currency": "usd",
    "balance":  1500
}'
```

- List accounts: `curl http://localhost:8080/accounts`

- Send payments:
```bash
curl -X POST http://localhost:8080/payments -d '{
    "account_from": id account_from,
    "account_to":   id account_to,
    "amount":       1500
}'
```

- List payments: `curl http://localhost:8080/payments`
