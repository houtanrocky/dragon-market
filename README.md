# Dragon Market

A Go implementation of the Aethoria marketplace challenge. Common and rare items are traded through fixed-price limit orders; unique legendary items are traded through timed auctions.

## Architecture

Domain services live in `internal/guild`, `internal/item`, `internal/order`, and `internal/auction`. They depend on small repository and service interfaces. PostgreSQL implementations are isolated in `internal/postgres`, while HTTP transport is in `internal/api`. Decisions and tradeoffs are documented in `docs/adr`.

PostgreSQL was selected because purchases, bids, wallet reservations, and ownership transfers require strong transactional consistency. Row locks serialize competing purchases and bids, and a partial unique index guarantees one active auction per legendary item.

## Run locally

Requirements: Go 1.25+, Docker, and Docker Compose.

```sh
docker compose up -d postgres oracle
```

Then run the API. PowerShell:

```powershell
$env:DATABASE_URL = "postgres://market:market@localhost:5432/market?sslmode=disable"
$env:PRICE_ORACLE_URL = "http://localhost:8090"
go run ./cmd/server
```

The API listens on `:8080`. `GET /healthz` provides a process health check. If `PRICE_ORACLE_URL` is omitted, the API runs without automatic price refresh.

## Price oracle

The external contract is:

```http
GET /prices/{itemID}
```

```json
{"base_price": 1250}
```

Prices are refreshed immediately at startup and every 30 seconds. Zero, negative, malformed, timed-out, and non-200 responses are rejected. The last valid database price is retained. Each item is refreshed independently, so one bad response does not stop the batch. Docker Compose includes a deterministic mock oracle for local use.

## Core API

- `POST /items`
- `GET /items` and `GET /items/{id}`
- `POST /orders`
- `POST /orders/{id}/buy` — requires `Idempotency-Key`
- `DELETE /orders/{id}`
- `POST /auctions`
- `GET /auctions/{id}`
- `POST /auctions/{id}/bids` — requires `Idempotency-Key`
- `GET /bids/{id}`
- `DELETE /auctions/{auctionID}/bids/{bidID}?bidder_id={guildID}`
- `GET /guilds/{id}/wallet`

Example purchase:

```sh
curl -X POST http://localhost:8080/orders/order-1/buy \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: purchase-001' \
  -d '{"buyer_id":"guild-2"}'
```

Guild IDs are treated as trusted caller input because authentication is outside this challenge's scope.

## Reliability and business rules

- Purchases and bid changes run in PostgreSQL transactions with row locks.
- Available wallet balance is total gold minus reserved gold.
- Daily spending resets on the next UTC calendar day.
- Wallet mutations append an audit record in the same transaction.
- Bids reserve funds; replacing a top bid releases the prior reservation.
- Bids must exceed the current bid by at least 5%.
- Sellers cannot bid on their own items.
- Auctions extend five minutes when a bid arrives in their final five minutes.
- A background worker settles expired auctions; repeated workers are safe due to locks and status checks.
- Completed HTTP mutations can be safely replayed with the same idempotency key and request.

## Verification

```sh
go build ./...
go test ./...
go test -race ./...
```

PostgreSQL repository tests use Testcontainers and require Docker.

## Assumptions

- Gold is stored as an integer smallest unit; floating-point arithmetic is not used.
- Auction deadlines and daily-limit dates use UTC.
- A highest bidder may cancel only while its bid is still the active top bid. Cancellation releases its reservation and leaves the auction without a winning bid until another bid is placed.
- Limit-order prices are fixed when listed and are not modified by later oracle updates.
